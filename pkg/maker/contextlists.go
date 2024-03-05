/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	"golang.org/x/exp/maps"
)

func (m *Maker) CreateContextCMakeLists(index int, cbuild Cbuild) error {

	outputByProducts, outputFile, outputType, customCommands := OutputFiles(cbuild.BuildDescType.Output)
	outputExt := path.Ext(outputFile)
	outputName := strings.TrimSuffix(outputFile, outputExt)
	cbuild.ContextRoot, _ = filepath.Rel(m.CbuildIndex.BaseDir, cbuild.BaseDir)
	outDir := AddRootPrefix(cbuild.ContextRoot, cbuild.BuildDescType.OutputDirs.Outdir)
	contextDir := path.Join(m.SolutionIntDir, cbuild.BuildDescType.Context)

	var cmakeTargetType, outputDirType, linkerVars, linkerOptions string
	if outputType == "elf" {
		cmakeTargetType = "add_executable"
		outputDirType = "RUNTIME_OUTPUT_DIRECTORY"
	} else if outputType == "lib" {
		cmakeTargetType = "add_library"
		outputDirType = "ARCHIVE_OUTPUT_DIRECTORY"
	}

	// Create components.cmake
	err := cbuild.CMakeCreateComponents(contextDir)
	if err != nil {
		return err
	}

	// Create groups.cmake
	err = cbuild.CMakeCreateGroups(contextDir)
	if err != nil {
		return err
	}

	// Libraries
	libraries := []string{"${CONTEXT}_GLOBAL"}
	libraries = append(libraries, cbuild.ListGroupsAndComponents()...)

	// Linker options
	if outputType == "elf" {
		linkerVars, linkerOptions = cbuild.LinkerOptions()
	}

	// Toolchain config
	toolchainConfig, _ := filepath.EvalSymlinks(m.SelectedToolchainConfig[index])
	toolchainConfig = filepath.ToSlash(toolchainConfig)

	// Global pre-includes
	for _, file := range cbuild.BuildDescType.ConstructedFiles {
		if file.Category == "preIncludeGlobal" {
			cbuild.PreIncludeGlobal = append(cbuild.PreIncludeGlobal, AddRootPrefix(cbuild.ContextRoot, file.File))
		}
	}

	// Global compile options abstractions
	abstractions := CompilerAbstractions{cbuild.BuildDescType.Debug, cbuild.BuildDescType.Optimize, cbuild.BuildDescType.Warnings, cbuild.BuildDescType.LanguageC, cbuild.BuildDescType.LanguageCpp}
	var globalCompilerAbstractions string
	if !AreAbstractionsEmpty(abstractions, cbuild.Languages) {
		globalCompilerAbstractions = "\n# Compile Options Abstractions" + cbuild.CMakeTargetCompileOptionsAbstractions("${CONTEXT}", abstractions, cbuild.Languages)
	}

	// Create CMakeLists content
	content := `cmake_minimum_required(VERSION 3.22)

set(CONTEXT ` + cbuild.BuildDescType.Context + `)
set(TARGET ${CONTEXT})
set(OUT_DIR "` + outDir + `")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)` + outputByProducts + linkerVars + `

# Processor Options` + cbuild.ProcessorOptions() + `

# Toolchain config map
set(REGISTERED_TOOLCHAIN_ROOT "` + m.RegisteredToolchains[m.SelectedToolchainVersion[index]].Path + `")
set(REGISTERED_TOOLCHAIN_VERSION "` + m.SelectedToolchainVersion[index].String() + `")
include("` + toolchainConfig + `")

# Setup project
project(${CONTEXT} LANGUAGES ` + strings.Join(cbuild.Languages, " ") + `)

# Compilation database
add_custom_target(database COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_BINARY_DIR}/compile_commands.json" "${OUT_DIR}")

# Setup context
` + cmakeTargetType + `(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX "` + outputExt + `" OUTPUT_NAME "` + outputName + `")
set_target_properties(${CONTEXT} PROPERTIES ` + outputDirType + ` ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes
add_library(${CONTEXT}_INCLUDES INTERFACE)` + CMakeTargetIncludeDirectories("${CONTEXT}_INCLUDES", "INTERFACE", AddRootPrefixes(cbuild.ContextRoot, cbuild.BuildDescType.AddPath)) + `

# Defines
add_library(${CONTEXT}_DEFINES INTERFACE)` + CMakeTargetCompileDefinitions("${CONTEXT}_DEFINES", "INTERFACE", cbuild.BuildDescType.Define) + `

# Compile options` + cbuild.CMakeTargetCompileOptionsGlobal("${CONTEXT}_GLOBAL", "INTERFACE") + `
` + globalCompilerAbstractions + `

# Add groups and components
include("groups.cmake")
include("components.cmake")
` + cbuild.CMakeTargetLinkLibraries("${CONTEXT}", "PUBLIC", libraries...) + `
` + linkerOptions + customCommands + `
`
	// Update CMakeLists.txt
	contextCMakeLists := path.Join(contextDir, "CMakeLists.txt")
	err = utils.UpdateFile(contextCMakeLists, content)
	if err != nil {
		return err
	}

	return err
}

func (c *Cbuild) CMakeCreateGroups(contextDir string) error {
	content := "# groups.cmake\n"
	abstractions := CompilerAbstractions{c.BuildDescType.Debug, c.BuildDescType.Optimize, c.BuildDescType.Warnings, c.BuildDescType.LanguageC, c.BuildDescType.LanguageCpp}
	content += c.CMakeCreateGroupRecursively("", c.BuildDescType.Groups, AddRootPrefixes(c.ContextRoot, c.BuildDescType.AddPath), "${CONTEXT}_INCLUDES", c.BuildDescType.Define, "${CONTEXT}_DEFINES", abstractions)

	filename := path.Join(contextDir, "groups.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}

func (c *Cbuild) CMakeCreateGroupRecursively(parent string, groups []Groups,
	parentIncludes []string, parentIncludesInterface string,
	parentDefines []interface{}, parentDefinesInterface string,
	parentAbstractions CompilerAbstractions) string {
	var content string
	for _, group := range groups {
		buildFiles := c.ClassifyFiles(group.Files)
		hasChildren := len(group.Groups) > 0
		if !hasChildren && len(buildFiles.Source) == 0 && len(buildFiles.Library) == 0 && len(buildFiles.Object) == 0 {
			continue
		}
		firstLevelGroup := len(parent) == 0
		name := parent + "_" + ReplaceDelimiters(group.Group)
		parentName := parent
		if firstLevelGroup {
			name = "Group" + name
			parentName = "${CONTEXT}"
		}
		// default private scope
		scope := "PRIVATE"
		if len(buildFiles.Source) == 0 {
			scope = "INTERFACE"
		} else if hasChildren {
			scope = "PUBLIC"
		}
		// add_library
		content += "\n# group " + group.Group
		content += CMakeAddLibrary(name, buildFiles)
		// target_include_directories
		if len(buildFiles.Include) > 0 {
			content += CMakeTargetIncludeDirectoriesFromFiles(name, buildFiles)
		}
		includes := parentIncludes
		includesInterface := parentIncludesInterface
		if len(group.DelPath) == 0 {
			if len(group.AddPath) > 0 {
				includes = append(parentIncludes, AddRootPrefixes(c.ContextRoot, group.AddPath)...)
				includesInterface = name + "_INCLUDES"
				content += "\nadd_library(" + includesInterface + " INTERFACE)"
				content += CMakeTargetIncludeDirectories(includesInterface, "INTERFACE", AddRootPrefixes(c.ContextRoot, group.AddPath))
				content += c.CMakeTargetLinkLibraries(includesInterface, "INTERFACE", parentIncludesInterface)
			}
		} else {
			includes = append(parentIncludes, AddRootPrefixes(c.ContextRoot, group.AddPath)...)
			includes = utils.RemoveIncludes(includes, AddRootPrefixes(c.ContextRoot, group.DelPath)...)
			includesInterface = name + "_INCLUDES"
			content += "\nadd_library(" + includesInterface + " INTERFACE)"
			content += CMakeTargetIncludeDirectories(includesInterface, "INTERFACE", includes)
		}
		// target_compile_definitions
		defines := parentDefines
		definesInterface := parentDefinesInterface
		if len(group.Undefine) == 0 {
			if len(group.Define) > 0 {
				defines = append(parentDefines, group.Define...)
				definesInterface = name + "_DEFINES"
				content += "\nadd_library(" + definesInterface + " INTERFACE)"
				content += CMakeTargetCompileDefinitions(definesInterface, "INTERFACE", group.Define)
				content += c.CMakeTargetLinkLibraries(definesInterface, "INTERFACE", parentDefinesInterface)
			}
		} else {
			defines = append(parentDefines, group.Define...)
			defines = utils.RemoveDefines(defines, group.Undefine...)
			definesInterface = name + "_DEFINES"
			content += "\nadd_library(" + definesInterface + " INTERFACE)"
			content += CMakeTargetCompileDefinitions(definesInterface, "INTERFACE", defines)
		}
		// compiler abstractions
		hasFileAbstractions := HasFileAbstractions(group.Files)
		groupAbstractions := CompilerAbstractions{group.Debug, group.Optimize, group.Warnings, group.LanguageC, group.LanguageCpp}
		languages := maps.Keys(buildFiles.Source)
		var abstractions CompilerAbstractions
		if !AreAbstractionsEmpty(groupAbstractions, c.Languages) {
			abstractions = InheritCompilerAbstractions(parentAbstractions, groupAbstractions)
			if !hasFileAbstractions {
				content += c.CMakeTargetCompileOptionsAbstractions(name, abstractions, languages)
			}
		}
		// target_compile_options
		if !IsCompileMiscEmpty(group.Misc) || len(buildFiles.PreIncludeLocal) > 0 {
			content += c.CMakeTargetCompileOptions(name, scope, group.Misc, buildFiles.PreIncludeLocal)
		}
		// target_link_libraries
		libraries := []string{"${CONTEXT}_GLOBAL", includesInterface, definesInterface}
		if !hasFileAbstractions {
			if !AreAbstractionsEmpty(groupAbstractions, languages) {
				libraries = append(libraries, name+"_ABSTRACTIONS")
			} else if !AreAbstractionsEmpty(parentAbstractions, languages) {
				libraries = append(libraries, parentName+"_ABSTRACTIONS")
			}
		}
		libraries = append(libraries, buildFiles.Library...)
		libraries = append(libraries, buildFiles.Object...)
		content += c.CMakeTargetLinkLibraries(name, scope, libraries...)
		// file properties
		for _, file := range group.Files {
			if strings.Contains(file.Category, "source") {
				fileAbstractions := CompilerAbstractions{file.Debug, file.Optimize, file.Warnings, file.LanguageC, file.LanguageCpp}
				if hasFileAbstractions {
					fileAbstractions = InheritCompilerAbstractions(abstractions, fileAbstractions)
				}
				content += c.CMakeSetFileProperties(file, fileAbstractions)
			}
		}
		content += "\n"

		// create children groups recursively
		if hasChildren {
			content += c.CMakeCreateGroupRecursively(name, group.Groups, includes, includesInterface, defines, definesInterface, abstractions)
		}
		c.BuildGroups = append(c.BuildGroups, name)
	}
	return content
}

func (c *Cbuild) CMakeCreateComponents(contextDir string) error {
	content := "# components.cmake\n"
	for _, component := range c.BuildDescType.Components {
		buildFiles := c.ClassifyFiles(component.Files)
		name := ReplaceDelimiters(component.Component)
		var scope string
		if buildFiles.Interface {
			scope = "INTERFACE"
		} else {
			scope = "PRIVATE"
		}
		// add_library
		content += "\n# component " + component.Component
		content += CMakeAddLibrary(name, buildFiles)
		libraries := []string{"${CONTEXT}_GLOBAL"}
		// target_include_directories
		if len(buildFiles.Include) > 0 {
			content += CMakeTargetIncludeDirectoriesFromFiles(name, buildFiles)
		}
		if len(component.DelPath) == 0 {
			libraries = append(libraries, "${CONTEXT}_INCLUDES")
			if len(component.AddPath) > 0 {
				content += "\nadd_library(" + name + "_INCLUDES INTERFACE)"
				content += CMakeTargetIncludeDirectories(name+"_INCLUDES", "INTERFACE", AddRootPrefixes(c.ContextRoot, component.AddPath))
				libraries = append(libraries, name+"_INCLUDES")
			}
		} else {
			includes := append(AddRootPrefixes(c.ContextRoot, c.BuildDescType.AddPath), AddRootPrefixes(c.ContextRoot, component.AddPath)...)
			includes = utils.RemoveIncludes(includes, AddRootPrefixes(c.ContextRoot, component.DelPath)...)
			content += "\nadd_library(" + name + "_INCLUDES INTERFACE)"
			content += CMakeTargetIncludeDirectories(name+"_INCLUDES", "INTERFACE", includes)
			libraries = append(libraries, name+"_INCLUDES")
		}
		// target_compile_definitions
		if len(component.Undefine) == 0 {
			libraries = append(libraries, "${CONTEXT}_DEFINES")
			if len(component.Define) > 0 {
				content += "\nadd_library(" + name + "_DEFINES INTERFACE)"
				content += CMakeTargetCompileDefinitions(name+"_DEFINES", "INTERFACE", component.Define)
				libraries = append(libraries, name+"_DEFINES")
			}
		} else {
			defines := append(c.BuildDescType.Define, component.Define...)
			defines = utils.RemoveDefines(defines, component.Undefine...)
			content += "\nadd_library(" + name + "_DEFINES INTERFACE)"
			content += CMakeTargetCompileDefinitions(name+"_DEFINES", "INTERFACE", defines)
			libraries = append(libraries, name+"_DEFINES")
		}
		// compiler abstractions
		componentAbstractions := CompilerAbstractions{component.Debug, component.Optimize, component.Warnings, component.LanguageC, component.LanguageCpp}
		globalAbstractions := CompilerAbstractions{c.BuildDescType.Debug, c.BuildDescType.Optimize, c.BuildDescType.Warnings, c.BuildDescType.LanguageC, c.BuildDescType.LanguageCpp}
		languages := maps.Keys(buildFiles.Source)
		if !AreAbstractionsEmpty(componentAbstractions, languages) {
			abstractions := InheritCompilerAbstractions(globalAbstractions, componentAbstractions)
			content += c.CMakeTargetCompileOptionsAbstractions(name, abstractions, languages)
			libraries = append(libraries, name+"_ABSTRACTIONS")
		} else if !AreAbstractionsEmpty(globalAbstractions, languages) {
			libraries = append(libraries, "${CONTEXT}_ABSTRACTIONS")
		}
		// target_compile_options
		if !IsCompileMiscEmpty(component.Misc) || len(buildFiles.PreIncludeLocal) > 0 {
			content += c.CMakeTargetCompileOptions(name, scope, component.Misc, buildFiles.PreIncludeLocal)
		}
		// target_link_libraries
		libraries = append(libraries, buildFiles.Library...)
		libraries = append(libraries, buildFiles.Object...)
		content += c.CMakeTargetLinkLibraries(name, scope, libraries...)

		content += "\n"
	}

	filename := path.Join(contextDir, "components.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}
