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
)

func (m *Maker) CreateContextCMakeLists(index int, cbuild Cbuild) error {

	outputByProducts, outputFile, outputType, customCommands := OutputFiles(cbuild.BuildDescType.Output)
	outputExt := path.Ext(outputFile)
	outputName := strings.TrimSuffix(outputFile, outputExt)
	cbuild.ContextRoot, _ = filepath.Rel(m.CbuildIndex.BaseDir, cbuild.BaseDir)
	outDir := AddRootPrefix(cbuild.ContextRoot, cbuild.BuildDescType.OutputDirs.Outdir)
	contextDir := path.Join(m.SolutionIntDir, cbuild.BuildDescType.Context)

	var cmakeTargetType, outputDirType string
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

	// Create CMakeLists content
	content := `cmake_minimum_required(VERSION 3.22)

set(CONTEXT ` + cbuild.BuildDescType.Context + `)
set(TARGET ${CONTEXT})
set(OUT_DIR "` + outDir + `")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)` + outputByProducts + `

# Processor Options` + cbuild.ProcessorOptions() + `

# Toolchain config map
set(REGISTERED_TOOLCHAIN_ROOT "` + m.RegisteredToolchains[m.SelectedToolchainVersion[index]].Path + `")
set(REGISTERED_TOOLCHAIN_VERSION "` + m.SelectedToolchainVersion[index].String() + `")
include("` + m.SelectedToolchainConfig[index] + `")

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

# Add groups and components
include("groups.cmake")
include("components.cmake")
target_link_libraries(${CONTEXT}
  ${CONTEXT}_GLOBAL` + ListGroupsAndComponents(cbuild) + `
)

# Linker options` + cbuild.LinkerOptions() + customCommands + `
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
	abstractions := CompilerAbstractions{c.BuildDescType.Debug, c.BuildDescType.Optimize, c.BuildDescType.Warnings}
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
		// default private scope
		scope := "PRIVATE"
		if hasChildren {
			// make scope public to its children
			scope = "PUBLIC"
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
		abstractions := InheritCompilerAbstractions(parentAbstractions, CompilerAbstractions{group.Debug, group.Optimize, group.Warnings})
		content += c.CompilerAbstractions(abstractions)
		// target_compile_options
		if !IsCompileMiscEmpty(group.Misc) {
			content += c.CMakeTargetCompileOptions(name, scope, group.Misc, CompilerAbstractions{})
		}
		// target_link_libraries
		content += "\ntarget_link_libraries(" + name + " PRIVATE ${CONTEXT}_GLOBAL"
		if len(parent) > 5 {
			content += " " + parent
		}
		content += ")\n"
		// file properties
		for _, file := range group.Files {
			content += c.CMakeSetFileProperties(file, abstractions)
		}
		// create children groups recursively
		if hasChildren {
			content += c.CMakeCreateGroupRecursively(name, group.Groups, abstractions)
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
		abstractions := InheritCompilerAbstractions(
			CompilerAbstractions{c.BuildDescType.Debug, c.BuildDescType.Optimize, c.BuildDescType.Warnings},
			CompilerAbstractions{component.Debug, component.Optimize, component.Warnings})

		// target_compile_options
		if !IsCompileMiscEmpty(component.Misc) || !IsAbstractionEmpty(abstractions) {
			content += c.CMakeTargetCompileOptions(name, scope, component.Misc, abstractions)
		}
		// target_link_libraries
		content += "\ntarget_link_libraries(" + name + " " + scope + " ${CONTEXT}_GLOBAL)\n"
	}

	filename := path.Join(contextDir, "components.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}
