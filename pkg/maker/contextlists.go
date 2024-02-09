/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"strconv"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
)

func (m *Maker) CreateContextCMakeLists(index int, cbuild Cbuild) error {

	outDir := path.Join(cbuild.BaseDir, cbuild.BuildDescType.OutputDirs.Outdir)
	intDir := path.Join(cbuild.BaseDir, cbuild.BuildDescType.OutputDirs.Intdir)
	objDir := path.Join(m.SolutionIntDir, strconv.Itoa(index))

	var outputFile, outputType string
	for _, output := range cbuild.BuildDescType.Output {
		if output.Type == "elf" || output.Type == "lib" {
			outputFile = output.File
			outputType = output.Type
			break
		}
	}

	outputExt := path.Ext(outputFile)
	outputName := strings.TrimSuffix(outputFile, outputExt)

	var outputDirType string
	if outputType == "elf" {
		outputDirType = "RUNTIME_OUTPUT_DIRECTORY"
	} else if outputType == "lib" {
		outputDirType = "ARCHIVE_OUTPUT_DIRECTORY"
	}

	// Write content
	content := `cmake_minimum_required(VERSION 3.22)

set(CONTEXT ` + cbuild.BuildDescType.Project + `)
set(OUT_DIR "` + outDir + `")
set(OBJ_DIR "` + objDir + `")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Processor Options` + ProcessorOptions(cbuild) + `

# Toolchain config map
set(REGISTERED_TOOLCHAIN_ROOT "` + m.RegisteredToolchains[m.SelectedToolchainVersion[index]].Path + `")
set(REGISTERED_TOOLCHAIN_VERSION "` + m.SelectedToolchainVersion[index].String() + `")
include("` + m.SelectedToolchainConfig[index] + `")

# Setup project
project(${CONTEXT} LANGUAGES C)

# Compilation database
add_custom_target(database COMMAND ${CMAKE_COMMAND} -E copy_if_different "${INT_DIR}/compile_commands.json" "${OUT_DIR}")

# Setup context
add_library(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX "` + outputExt + `" OUTPUT_NAME "` + outputName + `")
set_target_properties(${CONTEXT} PROPERTIES ` + outputDirType + ` ${OUT_DIR})
add_library(${CONTEXT}_GLOBAL INTERFACE)

# Includes` + CMakeTargetIncludeDirectories("${CONTEXT}_GLOBAL", "INTERFACE", cbuild.BuildDescType.AddPath) + `

# Defines` + CMakeTargetCompileDefinitions("${CONTEXT}_GLOBAL", "INTERFACE", cbuild.BuildDescType.Define) + `

# Compile options` + CMakeTargetCompileOptionsGlobal("${CONTEXT}_GLOBAL", "INTERFACE", cbuild) + `

# Add groups and components
include("groups.cmake")
include("components.cmake")
target_link_libraries(${CONTEXT}
  ${CONTEXT}_GLOBAL` + ListGroupsAndComponents(cbuild) + `
)

# Linker options` + LinkerOptions(cbuild) + `
`
	// Update CMakeLists.txt
	contextDir := path.Join(intDir, cbuild.BuildDescType.Context)
	contextCMakeLists := path.Join(contextDir, "CMakeLists.txt")
	err := utils.UpdateFile(contextCMakeLists, content)
	if err != nil {
		return err
	}

	// Create components.cmake
	err = m.CMakeCreateComponents(cbuild.BuildDescType.Components, contextDir)
	if err != nil {
		return err
	}

	// Create groups.cmake
	err = m.CMakeCreateGroups(cbuild.BuildDescType.Groups, contextDir)
	if err != nil {
		return err
	}

	return err
}

func (m *Maker) CMakeCreateGroups(groups []Groups, contextDir string) error {
	content := "# groups.cmake\n"
	content += CMakeCreateGroupRecursively("Group", groups)

	filename := path.Join(contextDir, "groups.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}

func CMakeCreateGroupRecursively(parent string, groups []Groups) string {
	var content string
	for _, group := range groups {
		buildFiles := ClassifyFiles(group.Files)
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
			content += CMakeTargetIncludeDirectories(name, scope, group.AddPath)
		}
		// target_compile_definitions
		if len(group.Define) > 0 {
			content += CMakeTargetCompileDefinitions(name, scope, group.Define)
		}
		// target_compile_options
		if !IsCompileMiscEmpty(group.Misc) {
			content += CMakeTargetCompileOptions(name, scope, group.Misc)
		}
		// target_link_libraries
		content += "\ntarget_link_libraries(" + name + " PRIVATE ${CONTEXT}_GLOBAL"
		if len(parent) > 5 {
			content += " " + parent
		}
		content += ")\n"
		// file properties
		for _, file := range group.Files {
			content += CMakeSetFileProperties(file)
		}
		// create children groups recursively
		if hasChildren {
			content += CMakeCreateGroupRecursively(name, group.Groups)
		}
	}
	return content
}

func (m *Maker) CMakeCreateComponents(components []Components, contextDir string) error {
	content := "# components.cmake\n"
	for _, component := range components {
		buildFiles := ClassifyFiles(component.Files)
		name := ReplaceDelimiters(component.Component)
		// add_library
		content += "\n# component " + component.Component
		content += CMakeAddLibrary(name, buildFiles)
		// target_include_directories
		if len(buildFiles.Include) > 0 {
			content += CMakeTargetIncludeDirectoriesFromFiles(name, buildFiles)
		}
		if len(component.AddPath) > 0 {
			content += CMakeTargetIncludeDirectories(name, "PRIVATE", component.AddPath)
		}
		// target_compile_definitions
		if len(component.Define) > 0 {
			content += CMakeTargetCompileDefinitions(name, "PRIVATE", component.Define)
		}
		// target_compile_options
		if !IsCompileMiscEmpty(component.Misc) {
			content += CMakeTargetCompileOptions(name, "PRIVATE", component.Misc)
		}
		// target_link_libraries
		content += "\ntarget_link_libraries(" + name + " PRIVATE ${CONTEXT}_GLOBAL)\n"
	}

	filename := path.Join(contextDir, "components.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}
