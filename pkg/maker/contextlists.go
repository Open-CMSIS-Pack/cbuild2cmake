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
		globalCompilerAbstractions = "\n# Compile Options Abstractions" + cbuild.CMakeTargetCompileOptionsAbstractions("${CONTEXT}_GLOBAL", abstractions, cbuild.Languages)
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

# Includes` + CMakeTargetIncludeDirectories("${CONTEXT}_GLOBAL", "INTERFACE", AddRootPrefixes(cbuild.ContextRoot, cbuild.BuildDescType.AddPath)) + `

# Defines` + CMakeTargetCompileDefinitions("${CONTEXT}_GLOBAL", "INTERFACE", cbuild.BuildDescType.Define) + `

# Compile options` + cbuild.CMakeTargetCompileOptionsGlobal("${CONTEXT}_GLOBAL", "INTERFACE") + `
` + globalCompilerAbstractions + `

# Add groups and components
include("groups.cmake")
include("components.cmake")
target_link_libraries(${CONTEXT}
  ${CONTEXT}_GLOBAL` + cbuild.ListGroupsAndComponents() + `
)
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
	content += c.CMakeCreateGroupRecursively("Group", c.BuildDescType.Groups, abstractions)

	filename := path.Join(contextDir, "groups.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}

func (c *Cbuild) CMakeCreateGroupRecursively(parent string, groups []Groups, parentAbstractions CompilerAbstractions) string {
	var content string
	for _, group := range groups {
		buildFiles := c.ClassifyFiles(group.Files)
		name := parent + "_" + ReplaceDelimiters(group.Group)
		hasChildren := len(group.Groups) > 0
		if !hasChildren && len(buildFiles.Source) == 0 {
			continue
		}
		// default private scope
		scope := "PRIVATE"
		if hasChildren {
			if len(buildFiles.Source) == 0 {
				scope = "INTERFACE"
			} else {
				scope = "PUBLIC"
			}
		}
		// add_library
		content += "\n# group " + group.Group
		content += CMakeAddLibrary(name, buildFiles)
		// target_include_directories
		if len(buildFiles.Include) > 0 {
			content += CMakeTargetIncludeDirectoriesFromFiles(name, buildFiles)
		}
		if len(group.AddPath) > 0 {
			content += CMakeTargetIncludeDirectories(name, scope, AddRootPrefixes(c.ContextRoot, group.AddPath))
		}
		// target_compile_definitions
		if len(group.Define) > 0 {
			content += CMakeTargetCompileDefinitions(name, scope, group.Define)
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
		content += "\ntarget_link_libraries(" + name + " " + scope + "\n  ${CONTEXT}_GLOBAL"
		if len(parent) > 5 {
			content += "\n  " + parent
		}
		if !hasFileAbstractions {
			if !AreAbstractionsEmpty(groupAbstractions, languages) {
				content += "\n  " + name + "_ABSTRACTIONS"
			} else if !AreAbstractionsEmpty(parentAbstractions, languages) {
				if len(parent) > 5 {
					content += "\n  " + parent + "_ABSTRACTIONS"
				} else {
					content += "\n  ${CONTEXT}_GLOBAL_ABSTRACTIONS"
				}
			}
		}
		content += "\n)\n"

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
		// create children groups recursively
		if hasChildren {
			content += c.CMakeCreateGroupRecursively(name, group.Groups, abstractions)
		} else {
			c.BuildGroups = append(c.BuildGroups, name)
		}
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
		// target_include_directories
		if len(buildFiles.Include) > 0 {
			content += CMakeTargetIncludeDirectoriesFromFiles(name, buildFiles)
		}
		if len(component.AddPath) > 0 {
			content += CMakeTargetIncludeDirectories(name, scope, AddRootPrefixes(c.ContextRoot, component.AddPath))
		}
		// target_compile_definitions
		if len(component.Define) > 0 {
			content += CMakeTargetCompileDefinitions(name, scope, component.Define)
		}
		// compiler abstractions
		componentAbstractions := CompilerAbstractions{component.Debug, component.Optimize, component.Warnings, component.LanguageC, component.LanguageCpp}
		globalAbstractions := CompilerAbstractions{c.BuildDescType.Debug, c.BuildDescType.Optimize, c.BuildDescType.Warnings, c.BuildDescType.LanguageC, c.BuildDescType.LanguageCpp}
		languages := maps.Keys(buildFiles.Source)
		if !AreAbstractionsEmpty(componentAbstractions, languages) {
			abstractions := InheritCompilerAbstractions(globalAbstractions, componentAbstractions)
			content += c.CMakeTargetCompileOptionsAbstractions(name, abstractions, languages)
		}
		// target_compile_options
		if !IsCompileMiscEmpty(component.Misc) || len(buildFiles.PreIncludeLocal) > 0 {
			content += c.CMakeTargetCompileOptions(name, scope, component.Misc, buildFiles.PreIncludeLocal)
		}
		// target_link_libraries
		content += "\ntarget_link_libraries(" + name + " " + scope + "\n  ${CONTEXT}_GLOBAL"
		if !AreAbstractionsEmpty(componentAbstractions, languages) {
			content += "\n  " + name + "_ABSTRACTIONS"
		} else if !AreAbstractionsEmpty(globalAbstractions, languages) {
			content += "\n  ${CONTEXT}_GLOBAL_ABSTRACTIONS"
		}
		content += "\n)\n"
	}

	filename := path.Join(contextDir, "components.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}
