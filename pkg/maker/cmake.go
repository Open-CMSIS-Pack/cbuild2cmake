/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func (m *Maker) CreateNativeCMakeLists(index int) error {
	cbuild := &m.Cbuilds[index]
	contextRoot, err := filepath.Rel(m.SolutionRoot, cbuild.BaseDir)
	if err != nil {
		log.Error("derive context root for ", cbuild.BaseDir, " relative to solution root ", m.SolutionRoot, ": ", err)
		return err
	}
	cbuild.ContextRoot = contextRoot
	cbuild.ContextRoot = filepath.ToSlash(cbuild.ContextRoot)

	nativeCMake := cbuild.BuildDescType.CMake
	generator := nativeCMake.Generator
	if generator == "" {
		generator = "Ninja"
	}
	contextDir := path.Join(m.SolutionTmpDir, cbuild.BuildDescType.Context)
	sourceDir := cbuild.AddRootPrefix(cbuild.ContextRoot, nativeCMake.Source)
	outDir := cbuild.AddRootPrefix(cbuild.ContextRoot, cbuild.BuildDescType.OutputDirs.Outdir)

	configureOptions := "\n  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON"
	for _, option := range nativeCMake.Configure {
		configureOptions += "\n  " + option
	}
	configureOptions = "\nset(CMAKE_CONFIGURE_OPTIONS" + configureOptions + "\n)"
	configureCommand := `  COMMAND ${CMAKE_COMMAND} -S "${CMAKE_SOURCE_DIR_NATIVE}" -B "${OUT_DIR}" -G "${CMAKE_GENERATOR_NATIVE}" ${CMAKE_CONFIGURE_OPTIONS}`

	buildTarget := ""
	if nativeCMake.Target != "" {
		buildTarget = " --target \"" + nativeCMake.Target + "\""
	}

	content := `cmake_minimum_required(VERSION 3.27)

# Roots
include("../roots.cmake")

set(CONTEXT ` + strings.ReplaceAll(cbuild.BuildDescType.Context, " ", "_") + `)
set(TARGET ${CONTEXT})
set(OUT_DIR "` + outDir + `")
set(CMAKE_SOURCE_DIR_NATIVE "` + sourceDir + `")
set(CMAKE_GENERATOR_NATIVE "` + generator + `")

project(${CONTEXT} LANGUAGES NONE)
` + configureOptions + `

# Native CMake setup and compilation database
add_custom_target(database DEPENDS "${OUT_DIR}/CMakeCache.txt")
add_custom_command(
  OUTPUT "${OUT_DIR}/CMakeCache.txt"
  BYPRODUCTS "${OUT_DIR}/compile_commands.json"
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Setup native CMake build directory"
` + configureCommand + `
  COMMENT ""
  DEPENDS "${CMAKE_CURRENT_LIST_FILE}"
  USES_TERMINAL
)

# Native CMake build
add_custom_target(cmake
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Run native CMake build"
  COMMAND ${CMAKE_COMMAND} --build "${OUT_DIR}"` + buildTarget + `
  COMMENT ""
  DEPENDS database
  USES_TERMINAL
)
`

	return utils.UpdateFile(path.Join(contextDir, "CMakeLists.txt"), content)
}
