/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/inittest"
	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
)

func TestToolchain(t *testing.T) {
	assert := assert.New(t)
	var m maker.Maker

	// Update environment variables
	m.EnvVars = utils.UpdateEnvVars(m.InstallConfigs.BinPath, m.InstallConfigs.EtcPath)
	inittest.ClearToolchainRegistration()
	absTestRoot, _ := filepath.Abs(testRoot)
	absTestRoot = filepath.ToSlash(absTestRoot)
	os.Setenv("AC6_TOOLCHAIN_6_19_0", path.Join(absTestRoot, "run/path/to/ac619/bin"))
	os.Setenv("AC6_TOOLCHAIN_6_21_0", path.Join(absTestRoot, "run/path/to/ac621/bin"))

	t.Run("test toolchain minimum version", func(t *testing.T) {
		m.Cbuilds = make([]maker.Cbuild, 1)
		m.Cbuilds[0].BuildDescType.Compiler = "AC6@>=6.18.0"
		err := m.ProcessToolchain()
		assert.Nil(err)
		expectedConfig := path.Join(m.EnvVars.CompilerRoot, "AC6.6.18.0.cmake")
		expectedPath := path.Join(absTestRoot, "run/path/to/ac621/bin")
		expectedVersion, _ := semver.NewVersion("6.21.0")
		assert.Equal(expectedConfig, m.SelectedToolchainConfig[0])
		assert.Equal(expectedPath, m.RegisteredToolchains[m.SelectedToolchainVersion[0]].Path)
		assert.Equal(expectedVersion, m.SelectedToolchainVersion[0])
	})

	t.Run("test toolchain fixed version", func(t *testing.T) {
		m.Cbuilds = make([]maker.Cbuild, 1)
		m.Cbuilds[0].BuildDescType.Compiler = "AC6@6.19.0"
		err := m.ProcessToolchain()
		assert.Nil(err)
		expectedConfig := path.Join(m.EnvVars.CompilerRoot, "AC6.6.18.0.cmake")
		expectedPath := path.Join(absTestRoot, "run/path/to/ac619/bin")
		expectedVersion, _ := semver.NewVersion("6.19.0")
		assert.Equal(expectedConfig, m.SelectedToolchainConfig[0])
		assert.Equal(expectedPath, m.RegisteredToolchains[m.SelectedToolchainVersion[0]].Path)
		assert.Equal(expectedVersion, m.SelectedToolchainVersion[0])
	})

	t.Run("test toolchain with debug flag", func(t *testing.T) {
		m.Cbuilds = make([]maker.Cbuild, 1)
		m.Cbuilds[0].BuildDescType.Compiler = "AC6@>=6.18.0"
		m.Params.Options.Debug = true
		err := m.ProcessToolchain()
		assert.Nil(err)
	})

	t.Run("test toolchain not registered", func(t *testing.T) {
		m.Cbuilds = make([]maker.Cbuild, 1)
		m.Cbuilds[0].BuildDescType.Compiler = "AC6@>=6.22.0"
		err := m.ProcessToolchain()
		assert.Error(err)
		assert.ErrorContains(err, "no compatible registered toolchain was found")
	})

	t.Run("test toolchain without config files", func(t *testing.T) {
		m.EnvVars.CompilerRoot = path.Join(absTestRoot, "empty")
		_ = os.MkdirAll(m.EnvVars.CompilerRoot, 0755)
		err := m.ProcessToolchain()
		assert.Error(err)
		assert.ErrorContains(err, "no toolchain configuration file was found")
	})

	t.Run("test toolchain with invalid compiler root", func(t *testing.T) {
		m.EnvVars.CompilerRoot = path.Join(absTestRoot, "invalid")
		err := m.ProcessToolchain()
		assert.Error(err)
		assert.ErrorContains(err, "reading directory failed")
	})
}
