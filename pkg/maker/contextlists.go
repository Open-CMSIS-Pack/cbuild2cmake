/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
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

func (m *Maker) CreateContextCMakeLists(index int, cbuild *Cbuild) error {

	outDir := path.Join(cbuild.BaseDir, cbuild.BuildDescType.OutputDirs.Outdir)
	intDir := path.Join(m.CbuildIndex.BaseDir, "tmp")
	objDir := path.Join(intDir, strconv.Itoa(index))

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
	content :=
		`cmake_minimum_required(VERSION 3.22)

set(CONTEXT ` + cbuild.BuildDescType.Project + `)
set(OUT_DIR "` + outDir + `")
set(OBJ_DIR "` + objDir + `")
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Setup project
project(${CONTEXT} LANGUAGES C)

# Compilation database
add_custom_target(database COMMAND ${CMAKE_COMMAND} -E copy_if_different "${INT_DIR}/compile_commands.json" "${OUT_DIR}")

# Setup context
add_library(${CONTEXT})
set_target_properties(${CONTEXT} PROPERTIES PREFIX "" SUFFIX "` + outputExt + `" OUTPUT_NAME "` + outputName + `")
set_target_properties(${CONTEXT} PROPERTIES ` + outputDirType + ` ${OUT_DIR})

# Add groups and components
include("groups.cmake")
include("components.cmake")
`
	contextCMakeLists := path.Join(path.Join(intDir, cbuild.BuildDescType.Context), "CMakeLists.txt")
	err := utils.UpdateFile(contextCMakeLists, content)
	if err != nil {
		return err
	}

	return err
}
