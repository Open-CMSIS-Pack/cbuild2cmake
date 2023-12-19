/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	//"cbuild2cmake/pkg/utils"
	//log "github.com/sirupsen/logrus"
	semver "github.com/Masterminds/semver/v3"
	utils "github.com/Open-CMSIS-Pack/cbuild/v2/pkg/utils"
)

type Params struct {
	Runner         utils.RunnerInterface
	Options        Options
	InputFile      string
	InstallConfigs utils.Configurations
}

type Options struct {
	Quiet   bool
	Debug   bool
	Verbose bool
	Clean   bool
}

type Vars struct {
	CbuildIndex              CbuildIndex
	Cbuilds                  []Cbuild
	EnvVars                  utils.EnvVars
	ToolchainConfigs         map[*semver.Version]Toolchain
	RegisteredToolchains     map[*semver.Version]Toolchain
	SelectedToolchainVersion *semver.Version
	SelectedToolchainConfig  string
}

type Maker struct {
	Params
	Vars
}

func (m *Maker) GenerateCMakeLists() error {
	// Update environment variables
	m.EnvVars = utils.UpdateEnvVars(m.InstallConfigs.BinPath, m.InstallConfigs.EtcPath)

	// Parse cbuild files
	err := m.ParseCbuildFiles()
	if err != nil {
		return err
	}

	// Process toolchain
	err = m.ProcessToolchain()
	if err != nil {
		return err
	}

	// Create super project CMakeLists.txt
	err = m.CreateSuperCMakeLists()
	if err != nil {
		return err
	}

	// Create context specific CMake files
	for index := range m.Cbuilds {
		err = m.CreateContextCMakeLists(index, &m.Cbuilds[index])
		if err != nil {
			return err
		}
	}

	return err
}
