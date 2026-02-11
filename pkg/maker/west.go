/*
 * Copyright (c) 2025-2026 Arm Limited. All rights reserved.
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

var WestToolchainMap = map[string]string{
	"AC6":   "armclang",
	"GCC":   "gnuarmemb",
	"IAR":   "iar",
	"CLANG": "llvm",
}

func (m *Maker) CreateWestCMakeLists(index int) error {
	cbuild := &m.Cbuilds[index]
	cbuild.ContextRoot, _ = filepath.Rel(m.SolutionRoot, cbuild.BaseDir)
	cbuild.ContextRoot = filepath.ToSlash(cbuild.ContextRoot)
	cbuild.Toolchain = m.RegisteredToolchains[m.SelectedToolchainVersion[index]].Name
	outDir := cbuild.AddRootPrefix(cbuild.ContextRoot, cbuild.BuildDescType.OutputDirs.Outdir)
	contextDir := path.Join(m.SolutionTmpDir, cbuild.BuildDescType.Context)
	westApp := cbuild.AddRootPrefix(cbuild.ContextRoot, cbuild.BuildDescType.West.AppPath)
	westToolchain := WestToolchainMap[cbuild.BuildDescType.Compiler]

	var westOptions, westDefs string
	var westOptionsRef, westDefsRef string
	for _, opt := range cbuild.BuildDescType.West.WestOpt {
		westOptions += "\n  " + opt
	}
	if len(westOptions) > 0 {
		westOptions = "\nset(WEST_OPTIONS" + westOptions + "\n)"
		westOptionsRef = " ${WEST_OPTIONS}"
	}

	for _, define := range cbuild.BuildDescType.West.WestDefs {
		key, value := utils.GetDefine(define)
		def := key
		if len(value) > 0 {
			def += "=" + value
		}
		westDefs += "\n  -D" + def
	}
	if len(westDefs) > 0 {
		westDefs = "\nset(WEST_DEFS" + westDefs + "\n)"
		westDefsRef = " -- ${WEST_DEFS}"
	}

	// Create toolchain.cmake
	err := m.CMakeCreateToolchain(index, contextDir, false)
	if err != nil {
		return err
	}

	// Create CMakeLists content
	content := `cmake_minimum_required(VERSION 3.27)

# Roots
include("../roots.cmake")

set(CONTEXT ` + strings.ReplaceAll(cbuild.BuildDescType.Context, " ", "_") + `)
set(TARGET ${CONTEXT})
set(OUT_DIR "` + outDir + `")
set(WEST_BOARD "` + cbuild.BuildDescType.West.Board + `")
set(WEST_APP "` + westApp + `")

# Toolchain config map
include("toolchain.cmake")

# Setup project
project(${CONTEXT} LANGUAGES NONE)
` + westOptions + westDefs + `

# Enable color diagnostics
set(CMAKE_COLOR_DIAGNOSTICS ON)

# Environment variables
set(ZEPHYR_TOOLCHAIN_PATH "${REGISTERED_TOOLCHAIN_ROOT}/..")
cmake_path(ABSOLUTE_PATH ZEPHYR_TOOLCHAIN_PATH NORMALIZE OUTPUT_VARIABLE ZEPHYR_TOOLCHAIN_PATH)
set(ENV_VARS
  ` + strings.ToUpper(westToolchain) + `_TOOLCHAIN_PATH="${ZEPHYR_TOOLCHAIN_PATH}"
  ZEPHYR_TOOLCHAIN_VARIANT="` + westToolchain + `"
)

# West setup and compilation database
add_custom_target(database DEPENDS "${OUT_DIR}/CMakeCache.txt")
add_custom_command(
  OUTPUT "${OUT_DIR}/CMakeCache.txt"
  BYPRODUCTS "${OUT_DIR}/compile_commands.json"
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Setup west build directory"
  COMMAND cmake -E env ${ENV_VARS} west` + westOptionsRef + ` build -b ${WEST_BOARD} -d "${OUT_DIR}" -p auto --cmake-only "${WEST_APP}"` + westDefsRef + `
  COMMENT "" 
  DEPENDS "${CMAKE_CURRENT_LIST_FILE}"
  USES_TERMINAL
)

# West build
add_custom_target(west
  COMMAND ${CMAKE_COMMAND} -E cmake_echo_color --blue " Run west build"
  COMMAND west build -d "${OUT_DIR}"
  COMMENT ""
  DEPENDS database
  USES_TERMINAL
)
`
	// Update CMakeLists.txt
	contextCMakeLists := path.Join(contextDir, "CMakeLists.txt")
	err = utils.UpdateFile(contextCMakeLists, content)
	if err != nil {
		return err
	}
	return err
}
