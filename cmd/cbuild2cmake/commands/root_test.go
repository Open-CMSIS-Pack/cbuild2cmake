/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package commands_test

import (
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/cmd/cbuild2cmake/commands"

	"github.com/stretchr/testify/assert"
)

func init() {

}

func TestCommands(t *testing.T) {
	assert := assert.New(t)

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
}
