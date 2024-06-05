/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func (m *Maker) CreateSuperCMakeLists() error {
	csolution := filepath.Base(m.CbuildIndex.BuildIdx.Csolution)
	reg := regexp.MustCompile(`(.*)\.csolution.ya?ml`)
	csolution = reg.ReplaceAllString(csolution, "$1")

	var contexts, dirs, contextOutputs string
	for i, cbuild := range m.Cbuilds {
		contexts = contexts + "  \"" + cbuild.BuildDescType.Context + "\"\n"
		dirs = dirs + "  \"${CMAKE_CURRENT_SOURCE_DIR}/" + cbuild.BuildDescType.Context + "\"\n"

		var contextOutputsName = "OUTPUTS_" + strconv.Itoa(i+1)
		contextOutputs += "set(" + contextOutputsName + "\n"

		var outputFile string
		for _, output := range cbuild.BuildDescType.Output {
			outputFile = output.File

			cbuildRelativePath, _ := filepath.Rel(m.CbuildIndex.BaseDir, cbuild.BaseDir)
			cbuildRelativePath = filepath.ToSlash(cbuildRelativePath)
			output := AddRootPrefix(cbuildRelativePath, path.Join(cbuild.BuildDescType.OutputDirs.Outdir, outputFile))
			contextOutputs += "  \"" + output + "\"\n"
		}

		contextOutputs += ")\n"
	}

	solutionRoot, _ := filepath.EvalSymlinks(m.CbuildIndex.BaseDir)
	solutionRoot = filepath.ToSlash(solutionRoot)

	var verbosity string
	if m.Options.Debug || m.Options.Verbose {
		verbosity = " --verbose"
	}

	// Write content
	content :=
		`cmake_minimum_required(VERSION 3.22)
include(ExternalProject)
	
project("` + csolution + `" NONE)

# Roots
set(CMSIS_PACK_ROOT "` + m.EnvVars.PackRoot + `")
cmake_path(ABSOLUTE_PATH CMSIS_PACK_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_PACK_ROOT)
set(CMSIS_COMPILER_ROOT "` + m.EnvVars.CompilerRoot + `")
cmake_path(ABSOLUTE_PATH CMSIS_COMPILER_ROOT NORMALIZE OUTPUT_VARIABLE CMSIS_COMPILER_ROOT)
set(SOLUTION_ROOT "` + solutionRoot + `")
cmake_path(ABSOLUTE_PATH SOLUTION_ROOT NORMALIZE OUTPUT_VARIABLE SOLUTION_ROOT)

# Context specific lists
set(CONTEXTS
` + contexts + `)
list(LENGTH CONTEXTS CONTEXTS_LENGTH)
math(EXPR CONTEXTS_LENGTH "${CONTEXTS_LENGTH}-1")

set(DIRS
` + dirs + `)

` + contextOutputs + `

set(ARGS
  "-DSOLUTION_ROOT=${SOLUTION_ROOT}"
  "-DCMSIS_PACK_ROOT=${CMSIS_PACK_ROOT}"
  "-DCMSIS_COMPILER_ROOT=${CMSIS_COMPILER_ROOT}"
)

# Iterate over contexts
foreach(INDEX RANGE ${CONTEXTS_LENGTH})

  math(EXPR N "${INDEX}+1")
  list(GET CONTEXTS ${INDEX} CONTEXT)
  list(GET DIRS ${INDEX} DIR)

  # Create external project, set configure and build steps
  ExternalProject_Add(${CONTEXT}
    PREFIX            ${DIR}
    SOURCE_DIR        ${DIR}
    BINARY_DIR        ${N}
    INSTALL_COMMAND   ""
    TEST_COMMAND      ""
    CONFIGURE_COMMAND ${CMAKE_COMMAND} -G Ninja -S <SOURCE_DIR> -B <BINARY_DIR> ${ARGS} 
    BUILD_COMMAND     ${CMAKE_COMMAND} --build <BINARY_DIR>` + verbosity + `
    BUILD_ALWAYS      TRUE
    BUILD_BYPRODUCTS  ${OUTPUTS_${N}}
  )
  ExternalProject_Add_StepTargets(${CONTEXT} build configure)

  # Debug
  message(VERBOSE "Configure Context: ${CMAKE_COMMAND} -G Ninja -S ${DIR} -B ${N}")

  # Database generation step
  ExternalProject_Add_Step(${CONTEXT} database
    COMMAND           ${CMAKE_COMMAND} --build <BINARY_DIR> --target database
    EXCLUDE_FROM_MAIN TRUE
    ALWAYS            TRUE
    DEPENDEES         configure
  )
  ExternalProject_Add_StepTargets(${CONTEXT} database)

endforeach()` + ExecutesCommands(m.CbuildIndex.BuildIdx.Executes) + m.BuildDependencies() + `
`
	superCMakeLists := path.Join(m.SolutionIntDir, "CMakeLists.txt")
	err := utils.UpdateFile(superCMakeLists, content)
	if err != nil {
		return err
	}

	log.Info("CMakeLists were successfully generated in the " + m.SolutionIntDir + " directory")
	return nil
}
