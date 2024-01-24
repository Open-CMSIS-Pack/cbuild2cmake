/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands_test

import (
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/cmd/cbuild2cmake/commands"
	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/inittest"
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
