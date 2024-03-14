/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands_test

import (
	"strings"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/cmd/cbuild2cmake/commands"
	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/inittest"
	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../../test"

func init() {
	inittest.TestInitialization(testRoot)
}

func TestCommands(t *testing.T) {
	assert := assert.New(t)
	cbuildIdxFile := testRoot + "/run/minimal/minimal.cbuild-idx.yml"

	t.Run("test version", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--version"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test help", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("invalid flag", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--invalid"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("invalid argument", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"./invalid.yml"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("multiple arguments", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{cbuildIdxFile, cbuildIdxFile})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test minimal cbuild-idx", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{cbuildIdxFile})
		err := cmd.Execute()
		assert.Nil(err)
	})

	t.Run("test quiet verbosity level", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--quiet", "--version"})
		err := cmd.Execute()
		assert.Nil(err)
		assert.Equal(log.ErrorLevel, log.GetLevel())
	})

	t.Run("test debug level", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{"--debug", "--version"})
		err := cmd.Execute()
		assert.Nil(err)
		assert.Equal(log.DebugLevel, log.GetLevel())
	})
}

func TestSolutions(t *testing.T) {
	assert := assert.New(t)

	t.Run("test build c", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/build-c"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check super CMakeLists contents
		content, err := utils.ReadFileContent(testCaseRoot + "/tmp/CMakeLists.txt")
		assert.Nil(err)

		content = strings.ReplaceAll(content, "\r\n", "\n")
		assert.Contains(content, `
set(CONTEXTS
  "project.AC6+ARMCM0"
  "project.CLANG+ARMCM0"
  "project.GCC+ARMCM0"
  "project.IAR+ARMCM0"
)`)
		assert.Contains(content, `
set(DIRS
  "${CMAKE_CURRENT_SOURCE_DIR}/project.AC6+ARMCM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/project.CLANG+ARMCM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/project.GCC+ARMCM0"
  "${CMAKE_CURRENT_SOURCE_DIR}/project.IAR+ARMCM0"
)`)
		assert.Contains(content, `
set(OUTPUTS
  "${SOLUTION_ROOT}/out/project/ARMCM0/AC6/project.axf"
  "${SOLUTION_ROOT}/out/project/ARMCM0/CLANG/project.elf"
  "${SOLUTION_ROOT}/out/project/ARMCM0/GCC/project.elf"
  "${SOLUTION_ROOT}/out/project/ARMCM0/IAR/project.out"
)`)

		// check golden references
		assert.Nil(inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp"))
	})

	t.Run("test linker preprocessing", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/linker-pre-processing"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		assert.Nil(inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp"))
	})

	t.Run("test global and local pre-includes", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/pre-include"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		assert.Nil(inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp"))
	})

	t.Run("test add-path, del-path, define, undefine", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/include-define"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		assert.Nil(inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp"))
	})
}
