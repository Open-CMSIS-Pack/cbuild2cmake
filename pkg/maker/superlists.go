/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"
	"regexp"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
)

func (m *Maker) CreateSuperCMakeLists() error {
	csolution := filepath.Base(m.CbuildIndex.BuildIdx.Csolution)
	reg := regexp.MustCompile(`(.*)\.csolution.ya?ml`)
	csolution = reg.ReplaceAllString(csolution, "$1")

	var contexts, dirs, outputs string
	for _, cbuild := range m.Cbuilds {
		contexts = contexts + "  \"" + cbuild.BuildDescType.Context + "\"\n"
		dirs = dirs + "  \"${CMAKE_CURRENT_SOURCE_DIR}/" + cbuild.BuildDescType.Context + "\"\n"

		var outputFile string
		for _, output := range cbuild.BuildDescType.Output {
			if output.Type == "elf" || output.Type == "lib" {
				outputFile = output.File
				break
			}
		}
		output := path.Join(path.Join(cbuild.BaseDir, cbuild.BuildDescType.OutputDirs.Outdir), outputFile)
		outputs = outputs + "  \"${SOLUTION_ROOT}/" + output + "\"\n"
	}

	// Write content
	content :=
		`cmake_minimum_required(VERSION 3.22)
include(ExternalProject)
	
project("` + csolution + `" NONE)

# Context specific lists
set(CONTEXTS
` + contexts + `)
list(LENGTH CONTEXTS CONTEXTS_LENGTH)
math(EXPR CONTEXTS_LENGTH "${CONTEXTS_LENGTH}-1")

set(DIRS
` + dirs + `)

set(OUTPUTS
` + outputs + `)

# Iterate over contexts
foreach(INDEX RANGE ${CONTEXTS_LENGTH})

  math(EXPR N "${INDEX}+1")
  list(GET CONTEXTS ${INDEX} CONTEXT)
  list(GET DIRS ${INDEX} DIR)
  list(GET OUTPUTS ${INDEX} OUTPUT)

  # Create external project, set configure and build steps
  ExternalProject_Add(${CONTEXT}
    PREFIX            ${DIR}
    SOURCE_DIR        ${DIR}
    BINARY_DIR        ${N}
    INSTALL_COMMAND   ""
    TEST_COMMAND      ""
    CONFIGURE_COMMAND ${CMAKE_COMMAND} -G Ninja -S <SOURCE_DIR> -B <BINARY_DIR>
    BUILD_COMMAND     ${CMAKE_COMMAND} --build <BINARY_DIR>
    BUILD_ALWAYS      TRUE
    BUILD_BYPRODUCTS  ${OUTPUT}
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

endforeach()
`
	intDir := path.Join(m.CbuildIndex.BaseDir, "tmp")
	superCMakeLists := path.Join(intDir, "CMakeLists.txt")
	err := utils.UpdateFile(superCMakeLists, content)
	if err != nil {
		return err
	}
	return err
}
