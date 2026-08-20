/*
 * Copyright (c) 2024-2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const CMAKE_MIN_REQUIRED = "3.27"

func (m *Maker) CreateSuperCMakeLists() error {
	// Iterate over cbuilds
	var contexts, dirs, westContextFlags, nativeCMakeContextFlags, contextOutputs, compilers string
	west := false
	nativeCMake := false
	for i, cbuild := range m.Cbuilds {
		contexts = contexts + "  \"" + strings.ReplaceAll(cbuild.BuildDescType.Context, " ", "_") + "\"\n"
		dirs = dirs + "  \"${CMAKE_CURRENT_SOURCE_DIR}/" + cbuild.BuildDescType.Context + "\"\n"
		westContext := cbuild.WestContext
		nativeCMakeContext := cbuild.NativeCMakeContext
		west = west || westContext
		nativeCMake = nativeCMake || nativeCMakeContext
		westContextFlags += "  \"" + strconv.FormatBool(westContext) + "\"\n"
		nativeCMakeContextFlags += "  \"" + strconv.FormatBool(nativeCMakeContext) + "\"\n"

		compilers += "  \"" + m.RegisteredToolchains[m.SelectedToolchainVersion[i]].Name +
			" V" + m.SelectedToolchainVersion[i].String() + "\"\n"

		var contextOutputsName = "OUTPUTS_" + strconv.Itoa(i+1)
		contextOutputs += "\nset(" + contextOutputsName + "\n"

		cbuildRelativePath, err := filepath.Rel(m.SolutionRoot, cbuild.BaseDir)
		if err != nil {
			log.Error("derive cbuild path for ", cbuild.BaseDir, " relative to solution root ", m.SolutionRoot, ": ", err)
			return err
		}
		cbuildRelativePath = filepath.ToSlash(cbuildRelativePath)
		if nativeCMakeContext {
			for _, image := range cbuild.BuildDescType.CMake.Images {
				output := cbuild.AddRootPrefix(cbuildRelativePath, path.Join(cbuild.BuildDescType.OutputDirs.Outdir, image.Image))
				contextOutputs += "  \"" + output + "\"\n"
			}
		} else {
			for _, buildOutput := range cbuild.BuildDescType.Output {
				output := cbuild.AddRootPrefix(cbuildRelativePath, path.Join(cbuild.BuildDescType.OutputDirs.Outdir, buildOutput.File))
				contextOutputs += "  \"" + output + "\"\n"
			}
		}

		contextOutputs += ")"
	}

	var westContexts, westContextCheck, westTarget string
	var nativeCMakeContexts, nativeCMakeContextCheck, nativeCMakeTarget, excludeFromMain string
	if west {
		westContexts = "\nset(WEST_CONTEXTS\n" + westContextFlags + ")\n"
		westContextCheck = "\n  list(GET WEST_CONTEXTS ${INDEX} WEST_CONTEXT)\n  if(WEST_CONTEXT)\n    set(WEST_TARGET \"--target west\")\n  endif()"
		westTarget = " ${WEST_TARGET}"
		excludeFromMain = "\n    EXCLUDE_FROM_MAIN TRUE"
	}
	if nativeCMake {
		nativeCMakeContexts = "\nset(NATIVE_CMAKE_CONTEXTS\n" + nativeCMakeContextFlags + ")\n"
		nativeCMakeContextCheck = "\n  list(GET NATIVE_CMAKE_CONTEXTS ${INDEX} NATIVE_CMAKE_CONTEXT)\n  if(NATIVE_CMAKE_CONTEXT)\n    set(NATIVE_CMAKE_TARGET \"--target cmake\")\n  endif()"
		nativeCMakeTarget = " ${NATIVE_CMAKE_TARGET}"
		excludeFromMain = "\n    EXCLUDE_FROM_MAIN TRUE"
	}

	var verbosity, logConfigure, stepLog string
	if m.Options.Debug || m.Options.Verbose {
		verbosity = " --verbose"
	} else {
		logConfigure = "\n    LOG_CONFIGURE         ON"
		logConfigure += "\n    LOG_OUTPUT_ON_FAILURE ON"
		if !west && !nativeCMake {
			stepLog = "\n    LOG               TRUE"
		}
	}

	// Write content
	content :=
		`cmake_minimum_required(VERSION ` + CMAKE_MIN_REQUIRED + `)
include(ExternalProject)
	
project("` + m.SolutionName + `" NONE)

# Enable color diagnostics
set(CMAKE_COLOR_DIAGNOSTICS ON)

# Roots
include("roots.cmake")

# Context specific lists
set(CONTEXTS
` + contexts + `)
list(LENGTH CONTEXTS CONTEXTS_LENGTH)
math(EXPR CONTEXTS_LENGTH "${CONTEXTS_LENGTH}-1")

set(COMPILERS
` + compilers + `)

set(DIRS
` + dirs + `)
` + westContexts + nativeCMakeContexts + contextOutputs + `

set(ARGS
  "-DSOLUTION_ROOT=${SOLUTION_ROOT}"
  "-DCMSIS_PACK_ROOT=${CMSIS_PACK_ROOT}"
  "-DCMSIS_COMPILER_ROOT=${CMSIS_COMPILER_ROOT}"
)

# Compilation database
add_custom_target(database)

# Iterate over contexts
foreach(INDEX RANGE ${CONTEXTS_LENGTH})

  math(EXPR N "${INDEX}+1")
  list(GET CONTEXTS ${INDEX} CONTEXT)
  list(GET COMPILERS ${INDEX} COMPILER)
  list(GET DIRS ${INDEX} DIR)` + westContextCheck + nativeCMakeContextCheck + `

  # Create external project, set configure and build steps
  ExternalProject_Add(${CONTEXT}
    PREFIX                ${DIR}
    SOURCE_DIR            ${DIR}
    BINARY_DIR            ${N}
    INSTALL_COMMAND       ""
    TEST_COMMAND          ""
    CONFIGURE_COMMAND     ${CMAKE_COMMAND} -G Ninja -S <SOURCE_DIR> -B <BINARY_DIR> ${ARGS} 
    BUILD_COMMAND         ${CMAKE_COMMAND} -E cmake_echo_color --blue --bold "Building CMake target '${CONTEXT}'"
    COMMAND               ${CMAKE_COMMAND} -E echo "Using compiler: ${COMPILER}"
    COMMAND               ${CMAKE_COMMAND} --build <BINARY_DIR>` + westTarget + nativeCMakeTarget + verbosity + `
    BUILD_ALWAYS          TRUE
    BUILD_BYPRODUCTS      ${OUTPUTS_${N}}` + logConfigure + `
    USES_TERMINAL_BUILD   ON
  )

  # Executes command step
  ExternalProject_Add_Step(${CONTEXT} executes
    DEPENDEES         build
  )

  ExternalProject_Add_StepTargets(${CONTEXT} build configure executes)

  # Debug
  message(VERBOSE "Configure Context: ${CMAKE_COMMAND} -G Ninja -S ${DIR} -B ${N}")

  # Database generation step
  ExternalProject_Add_Step(${CONTEXT} database
    COMMAND           ${CMAKE_COMMAND} --build <BINARY_DIR> --target database` + verbosity + excludeFromMain + `
    ALWAYS            TRUE` + stepLog + `
    USES_TERMINAL     ON
    DEPENDEES         configure
  )
  ExternalProject_Add_StepTargets(${CONTEXT} database)
  add_dependencies(database ${CONTEXT}-database)

endforeach()` + m.ExecutesCommands(m.CbuildIndex.BuildIdx.Executes) + m.BuildDependencies() + `
`
	superCMakeLists := path.Join(m.SolutionTmpDir, "CMakeLists.txt")
	err := utils.UpdateFile(superCMakeLists, content)
	if err != nil {
		return err
	}

	log.Info("CMakeLists were successfully generated in the " + m.SolutionTmpDir + " directory")
	return nil
}

func (m *Maker) CMakeCreateRoots(solutionRoot string) error {
	content :=
		`# roots.cmake
set(CMSIS_PACK_ROOT "` + m.EnvVars.PackRoot + `" CACHE PATH "CMSIS pack root")
cmake_path(ABSOLUTE_PATH CMSIS_PACK_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_PACK_ROOT)
set(CMSIS_COMPILER_ROOT "` + m.EnvVars.CompilerRoot + `" CACHE PATH "CMSIS compiler root")
cmake_path(ABSOLUTE_PATH CMSIS_COMPILER_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_COMPILER_ROOT)
set(SOLUTION_ROOT "` + solutionRoot + `" CACHE PATH "CMSIS solution root")
cmake_path(ABSOLUTE_PATH SOLUTION_ROOT NORMALIZE OUTPUT_VARIABLE SOLUTION_ROOT)
`

	filename := path.Join(m.SolutionTmpDir, "roots.cmake")
	err := utils.UpdateFile(filename, content)
	if err != nil {
		return err
	}

	return err
}

func (m *Maker) CreateCMakeListsImageOnly() error {
	// Write content
	content :=
		`cmake_minimum_required(VERSION ` + CMAKE_MIN_REQUIRED + `)

project("` + m.SolutionName + `" NONE)

# Roots
include("roots.cmake")` + m.ExecutesCommands(m.CbuildIndex.BuildIdx.Executes) + m.BuildDependencies() + `
`
	pathCMakeLists := path.Join(m.SolutionTmpDir, "CMakeLists.txt")
	err := utils.UpdateFile(pathCMakeLists, content)
	if err != nil {
		return err
	}

	log.Info("CMakeLists was successfully generated in the " + m.SolutionTmpDir + " directory")
	return nil
}
