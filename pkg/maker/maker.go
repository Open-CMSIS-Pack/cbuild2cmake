/*
 * Copyright (c) 2024-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"

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
	Quiet         bool
	Debug         bool
	Verbose       bool
	UseContextSet bool
}

type Vars struct {
	CbuildIndex              CbuildIndex
	CbuildSet                CbuildSet
	Cbuilds                  []Cbuild
	Contexts                 []string
	EnvVars                  utils.EnvVars
	GeneratedFiles           []string
	ToolchainConfigs         map[*semver.Version]Toolchain
	RegisteredToolchains     map[*semver.Version]Toolchain
	SelectedToolchainVersion []*semver.Version
	SelectedToolchainConfig  []string
	SolutionTmpDir           string
	SolutionRoot             string
	SolutionName             string
}

type Maker struct {
	Params
	Vars
}

func (m *Maker) GenerateCMakeLists() error {
	// Update environment variables
	m.EnvVars = utils.UpdateEnvVars(m.InstallConfigs.BinPath, m.InstallConfigs.EtcPath)
	m.EnvVars.PackRoot, _ = filepath.EvalSymlinks(m.EnvVars.PackRoot)
	m.EnvVars.PackRoot = filepath.ToSlash(m.EnvVars.PackRoot)
	m.EnvVars.CompilerRoot, _ = filepath.EvalSymlinks(m.EnvVars.CompilerRoot)
	m.EnvVars.CompilerRoot = filepath.ToSlash(m.EnvVars.CompilerRoot)

	// Parse cbuild files
	err := m.ParseCbuildFiles()
	if err != nil {
		return err
	}

	// Get tmp directory
	if len(m.CbuildIndex.BuildIdx.TmpDir) == 0 {
		m.CbuildIndex.BuildIdx.TmpDir = "tmp"
	}
	m.SolutionTmpDir = path.Join(m.CbuildIndex.BaseDir, m.CbuildIndex.BuildIdx.TmpDir)

	// Create roots.cmake
	err = m.CMakeCreateRoots(m.SolutionRoot)
	if err != nil {
		return err
	}

	// Create CMakeLists.txt for image only solution
	if m.CbuildIndex.BuildIdx.ImageOnly {
		return m.CreateCMakeListsImageOnly()
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
		err = m.CreateContextCMakeLists(index)
		if err != nil {
			return err
		}
	}

	return err
}
