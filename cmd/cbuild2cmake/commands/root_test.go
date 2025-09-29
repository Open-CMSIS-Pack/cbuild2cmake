/*
 * Copyright (c) 2024-2025 Arm Limited. All rights reserved.
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

	t.Run("missing cbuild-set.yml", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{cbuildIdxFile, "--context-set"})
		err := cmd.Execute()
		assert.Error(err)
	})

	t.Run("test minimal cbuild-idx", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		cmd.SetArgs([]string{cbuildIdxFile})
		err := cmd.Execute()
		assert.Nil(err)
		assert.FileExists(testRoot + "/run/minimal/custom/tmp/path/CMakeLists.txt")
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
set(OUTPUTS_1
  "${SOLUTION_ROOT}/out/project/ARMCM0/AC6/project.axf"
)
set(OUTPUTS_2
  "${SOLUTION_ROOT}/out/project/ARMCM0/CLANG/project.elf"
)
set(OUTPUTS_3
  "${SOLUTION_ROOT}/out/project/ARMCM0/GCC/project.elf"
)
set(OUTPUTS_4
  "${SOLUTION_ROOT}/out/project/ARMCM0/IAR/project.out"
)`)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test build-set", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/build-set"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug", "--context-set"})
		err := cmd.Execute()
		assert.Nil(err)

		// check super CMakeLists contents
		content, err := utils.ReadFileContent(testCaseRoot + "/tmp/CMakeLists.txt")
		assert.Nil(err)

		content = strings.ReplaceAll(content, "\r\n", "\n")
		assert.Contains(content, `
set(CONTEXTS
  "project.Release+ARMCM0"
)`)
		assert.Contains(content, `
set(DIRS
  "${CMAKE_CURRENT_SOURCE_DIR}/project.Release+ARMCM0"
)`)
		assert.Contains(content, `
set(OUTPUTS_1
  "${SOLUTION_ROOT}/out/project/ARMCM0/Release/project.axf"
)`)
	})

	t.Run("test build asm", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/build-asm"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test build cpp", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/build-cpp"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test blanks", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/blanks"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test linker preprocessing", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/linker-pre-processing"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test global and local pre-includes", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/pre-include"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test pre-includes in out-of-tree build", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/pre-include-oot"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test add-path, del-path, define, undefine", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/include-define"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test language and scope", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/language-scope"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test library rtos", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/library-rtos"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test executes", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/executes"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test abstractions", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/abstractions"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test image-only solution", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/image-only"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})

	t.Run("test west solution", func(t *testing.T) {
		cmd := commands.NewRootCmd()
		testCaseRoot := testRoot + "/run/solutions/west"
		cbuildIdxFile := testCaseRoot + "/solution.cbuild-idx.yml"
		cmd.SetArgs([]string{cbuildIdxFile, "--debug"})
		err := cmd.Execute()
		assert.Nil(err)

		// check golden references
		err, mismatch := inittest.CompareFiles(testCaseRoot+"/ref", testCaseRoot+"/tmp")
		assert.Nil(err)
		assert.False(mismatch)
	})
}
