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
	cbuild.Toolchain = m.RegisteredToolchains[m.SelectedToolchainVersion[index]].Name
	outDir := AddRootPrefix(cbuild.ContextRoot, cbuild.BuildDescType.OutputDirs.Outdir)
	contextDir := path.Join(m.SolutionIntDir, cbuild.BuildDescType.Context)
	cbuild.IncludeGlobal = make(LanguageMap)

	var cmakeTargetType, outputDirType, linkerVars, linkerOptions string
	if outputType == "elf" {
		cmakeTargetType = "add_executable"
		outputDirType = "RUNTIME_OUTPUT_DIRECTORY"
	} else if outputType == "lib" {
		cmakeTargetType = "add_library"
		outputDirType = "ARCHIVE_OUTPUT_DIRECTORY"
	}

	// Create toolchain.cmake
	err := m.CMakeCreateToolchain(index, contextDir)
	if err != nil {
		return err
	}

	// Create groups.cmake
	err = cbuild.CMakeCreateGroups(contextDir)
	if err != nil {
		return err
	}

	// Create components.cmake
	err = cbuild.CMakeCreateComponents(contextDir)
	if err != nil {
		return err
	}

	// Libraries
	var libraries []string
	libraries = append(libraries, cbuild.ListGroupsAndComponents()...)
	libraries = append(libraries, cbuild.BuildDescType.Misc.Library...)

	// Linker options
	if outputType == "elf" {
		linkerVars, linkerOptions = cbuild.LinkerOptions()
	}

	// Make system includes explicit for compilation database completeness
	var systemIncludes string
	for _, language := range cbuild.Languages {
		switch language {
		case "C", "CXX":
			systemIncludes += "\nset(CMAKE_" + language + "_STANDARD_INCLUDE_DIRECTORIES ${CMAKE_" + language + "_IMPLICIT_INCLUDE_DIRECTORIES})"
		}
	}

	// Global classified includes
	includeGlobal := make(ScopeMap)
	includeGlobal["PUBLIC"] = cbuild.IncludeGlobal
	includeGlobal["PUBLIC"]["ALL"] = utils.AppendUniquely(includeGlobal["PUBLIC"]["ALL"], AddRootPrefixes(cbuild.ContextRoot, cbuild.BuildDescType.AddPath)...)

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
		globalCompilerAbstractions = "\n\n# Compile Options Abstractions" + cbuild.CMakeTargetCompileOptionsAbstractions("${CONTEXT}", abstractions, cbuild.Languages)
	}

	// Create CMakeLists content
	content := `cmake_minimum_required(VERSION 3.22)

set(CONTEXT ` + cbuild.BuildDescType.Context + `)
set(TARGET ${CONTEXT})
set(OUT_DIR "` + outDir + `")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)` + outputByProducts + linkerVars + `

# Processor Options` + cbuild.ProcessorOptions() + `

# Toolchain config map
include("toolchain.cmake")

# Setup project
project(${CONTEXT} LANGUAGES ` + strings.Join(cbuild.Languages, " ") + `)

# Compilation database
add_custom_target(database COMMAND ${CMAKE_COMMAND} -E copy_if_different "${CMAKE_CURRENT_BINARY_DIR}/compile_commands.json" "${OUT_DIR}")` + systemIncludes + `

# Setup context
` + cmakeTargetType + `(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX "` + outputExt + `" OUTPUT_NAME "` + outputName + `")
set_target_properties(${CONTEXT} PROPERTIES ` + outputDirType + ` ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes` + CMakeTargetIncludeDirectoriesClassified("${CONTEXT}", includeGlobal) + `

# Defines` + CMakeTargetCompileDefinitions("${CONTEXT}", "", "PUBLIC", cbuild.BuildDescType.Define, []string{}) + `

# Compile options` + cbuild.CMakeTargetCompileOptionsGlobal("${CONTEXT}", "PUBLIC") + globalCompilerAbstractions + `

# Add groups and components
include("groups.cmake")
include("components.cmake")
` + cbuild.CMakeTargetLinkLibraries("${CONTEXT}", "PUBLIC", libraries...) + RescanLinkLibraries("${CONTEXT}", cbuild.Toolchain) + `
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

func (m *Maker) CMakeCreateToolchain(index int, contextDir string) error {
	toolchainConfig, _ := filepath.Rel(m.EnvVars.CompilerRoot, m.SelectedToolchainConfig[index])
	toolchainConfig = "${CMSIS_COMPILER_ROOT}/" + filepath.ToSlash(toolchainConfig)
	content := `# toolchain.cmake

set(REGISTERED_TOOLCHAIN_ROOT "` + m.RegisteredToolchains[m.SelectedToolchainVersion[index]].Path + `")
set(REGISTERED_TOOLCHAIN_VERSION "` + m.SelectedToolchainVersion[index].String() + `")
include("` + toolchainConfig + `")
`
	filename := path.Join(contextDir, "toolchain.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}
	return err
}

func (c *Cbuild) CMakeCreateGroups(contextDir string) error {
	content := "# groups.cmake\n"
	abstractions := CompilerAbstractions{c.BuildDescType.Debug, c.BuildDescType.Optimize, c.BuildDescType.Warnings, c.BuildDescType.LanguageC, c.BuildDescType.LanguageCpp}
	content += c.CMakeCreateGroupRecursively("", c.BuildDescType.Groups, abstractions, c.BuildDescType.Misc.ASM)
	filename := path.Join(contextDir, "groups.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}

func (c *Cbuild) CMakeCreateGroupRecursively(parent string, groups []Groups,
	parentAbstractions CompilerAbstractions, parentMiscAsm []string) string {
	var content string
	for _, group := range groups {
		miscAsm := utils.AppendUniquely(parentMiscAsm, group.Misc.ASM...)
		buildFiles := c.ClassifyFiles(group.Files)
		hasChildren := len(group.Groups) > 0
		if !hasChildren && len(buildFiles.Source) == 0 && len(buildFiles.Include) == 0 && len(buildFiles.Library) == 0 && len(buildFiles.Object) == 0 {
			continue
		}
		firstLevelGroup := len(parent) == 0
		name := parent + "_" + ReplaceDelimiters(group.Group)
		parentName := parent
		if firstLevelGroup {
			name = "Group" + name
			parentName = "${CONTEXT}"
		}
		// default scope
		scope := "PUBLIC"
		if buildFiles.Interface {
			scope = "INTERFACE"
		}
		// add_library
		content += "\n# group " + group.Group
		content += CMakeAddLibrary(name, buildFiles)
		// target_include_directories
		if len(buildFiles.Include) > 0 {
			content += CMakeTargetIncludeDirectoriesClassified(name, buildFiles.Include)
			c.IncludeGlobal = AppendGlobalIncludes(c.IncludeGlobal, buildFiles.Include)
		}
		content += CMakeTargetIncludeDirectories(name, parentName, scope, AddRootPrefixes(c.ContextRoot, group.AddPath), AddRootPrefixes(c.ContextRoot, group.DelPath))
		// target_compile_definitions
		content += CMakeTargetCompileDefinitions(name, parentName, scope, group.Define, group.Undefine)
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
		var libraries []string
		if !buildFiles.Interface && !hasFileAbstractions {
			if !AreAbstractionsEmpty(groupAbstractions, languages) {
				libraries = append(libraries, name+"_ABSTRACTIONS")
			} else if !AreAbstractionsEmpty(parentAbstractions, languages) {
				libraries = append(libraries, parentName+"_ABSTRACTIONS")
			}
		}
		// target_compile_options
		if !buildFiles.Interface || !IsCompileMiscEmpty(group.Misc) || len(buildFiles.PreIncludeLocal) > 0 {
			content += c.CMakeTargetCompileOptions(name, scope, group.Misc, buildFiles.PreIncludeLocal, parentName)
		}
		// target_link_libraries
		libraries = append(libraries, buildFiles.Library...)
		libraries = append(libraries, buildFiles.Object...)
		if len(libraries) > 0 {
			content += c.CMakeTargetLinkLibraries(name, scope, libraries...)
		}
		// file properties
		for _, file := range group.Files {
			if strings.Contains(file.Category, "source") {
				fileAbstractions := CompilerAbstractions{file.Debug, file.Optimize, file.Warnings, file.LanguageC, file.LanguageCpp}
				if hasFileAbstractions {
					fileAbstractions = InheritCompilerAbstractions(abstractions, fileAbstractions)
				}
				content += c.CMakeSetFileProperties(file, fileAbstractions, miscAsm)
			}
		}
		content += "\n"

		// create children groups recursively
		if hasChildren {
			content += c.CMakeCreateGroupRecursively(name, group.Groups, abstractions, miscAsm)
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
		// default scope
		scope := "PUBLIC"
		if buildFiles.Interface {
			scope = "INTERFACE"
		}
		// add_library
		content += "\n# component " + component.Component
		content += CMakeAddLibrary(name, buildFiles)
		// target_include_directories
		if len(buildFiles.Include) > 0 {
			content += CMakeTargetIncludeDirectoriesClassified(name, buildFiles.Include)
			c.IncludeGlobal = AppendGlobalIncludes(c.IncludeGlobal, buildFiles.Include)
		}
		content += CMakeTargetIncludeDirectories(name, "${CONTEXT}", scope, AddRootPrefixes(c.ContextRoot, component.AddPath), AddRootPrefixes(c.ContextRoot, component.DelPath))
		// target_compile_definitions
		content += CMakeTargetCompileDefinitions(name, "${CONTEXT}", scope, component.Define, component.Undefine)
		// compiler abstractions
		var libraries []string
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
		if !buildFiles.Interface || !IsCompileMiscEmpty(component.Misc) || len(buildFiles.PreIncludeLocal) > 0 {
			content += c.CMakeTargetCompileOptions(name, scope, component.Misc, buildFiles.PreIncludeLocal, "${CONTEXT}")
		}
		// target_link_libraries
		libraries = append(libraries, buildFiles.Library...)
		libraries = append(libraries, buildFiles.Object...)
		if len(libraries) > 0 {
			content += c.CMakeTargetLinkLibraries(name, scope, libraries...)
		}

		content += "\n"
	}

	filename := path.Join(contextDir, "components.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}

func RescanLinkLibraries(name string, compiler string) string {
	var content string
	if compiler == "GCC" {
		content += "\nget_target_property(LINK_LIBRARIES " + name + " LINK_LIBRARIES)"
		content += "\nset_target_properties(" + name + " PROPERTIES LINK_LIBRARIES \"-Wl,--start-group;${LINK_LIBRARIES};-Wl,--end-group\")"
	}
	return content
}
