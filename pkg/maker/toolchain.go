/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
)

type Toolchain struct {
	Name string
	Path string
}

func (m *Maker) ProcessToolchain() error {

	toolchainFiles, err := os.ReadDir(m.EnvVars.CompilerRoot)
	if err != nil {
		err := errors.New("reading directory failed: " + m.EnvVars.CompilerRoot)
		return err
	}

	// Toolchain configs
	m.ToolchainConfigs = make(map[*semver.Version]Toolchain)
	pattern := regexp.MustCompile(`(\w+)\.(\d+\.\d+\.\d+).cmake`)
	for _, toolchainFile := range toolchainFiles {
		matched := pattern.FindAllStringSubmatch(toolchainFile.Name(), -1)
		if matched == nil {
			continue
		}
		var toolchain Toolchain
		toolchain.Name = matched[0][1]
		toolchain.Path = path.Join(m.EnvVars.CompilerRoot, toolchainFile.Name())
		version, _ := semver.NewVersion(matched[0][2])
		m.ToolchainConfigs[version] = toolchain

		// Debug
		if m.Params.Options.Debug {
			log.Debug("Found config file: " + toolchain.Name + " " + version.String() + " " + toolchain.Path)
		}
	}

	if len(m.ToolchainConfigs) == 0 {
		err := errors.New("no toolchain configuration file was found in " + m.EnvVars.CompilerRoot)
		return err
	}

	// Registered toolchains
	m.RegisteredToolchains = make(map[*semver.Version]Toolchain)
	systemEnvVars := os.Environ()
	pattern = regexp.MustCompile(`(\w+)_TOOLCHAIN_(\d+)_(\d+)_(\d+)=(.*)`)
	for _, systemEnvVar := range systemEnvVars {
		matched := pattern.FindAllStringSubmatch(systemEnvVar, -1)
		if matched == nil {
			continue
		}
		var toolchain Toolchain
		toolchain.Name = matched[0][1]
		toolchain.Path = filepath.ToSlash(matched[0][5])
		version, _ := semver.NewVersion(matched[0][2] + "." + matched[0][3] + "." + matched[0][4])
		m.RegisteredToolchains[version] = toolchain

		// Debug
		if m.Params.Options.Debug {
			log.Debug("Found registered toolchain: " + toolchain.Name + " " + version.String() + " " + toolchain.Path)
		}
	}

	// Get each context's toolchain
	m.SelectedToolchainVersion = make([]*semver.Version, len(m.Cbuilds))
	m.SelectedToolchainConfig = make([]string, len(m.Cbuilds))
	for index, cbuild := range m.Cbuilds {
		contextConstraints := make(map[*semver.Constraints]bool)
		var contextToolchain string
		if strings.Contains(cbuild.BuildDescType.Compiler, "@") {
			contextToolchain = cbuild.BuildDescType.Compiler[:strings.Index(cbuild.BuildDescType.Compiler, "@")]
			constraint, _ := semver.NewConstraint(cbuild.BuildDescType.Compiler[strings.Index(cbuild.BuildDescType.Compiler, "@")+1:])
			contextConstraints[constraint] = true
		} else {
			contextToolchain = cbuild.BuildDescType.Compiler
		}

		// Debug
		if m.Params.Options.Debug {
			var constraints string
			for constraint := range contextConstraints {
				constraints = constraints + " " + constraint.String()
			}
			log.Debug("Context toolchain: " + contextToolchain + " - Constraints:" + constraints)
		}

		// Sort config versions and registered versions
		var configVersions []*semver.Version
		for version, toolchainConfig := range m.ToolchainConfigs {
			if toolchainConfig.Name == contextToolchain {
				configVersions = append(configVersions, version)
			}
		}
		if len(configVersions) == 0 {
			err := errors.New("no toolchain configuration file was found for " + contextToolchain)
			return err
		}
		sort.Sort(sort.Reverse(semver.Collection(configVersions)))
		var registeredVersions []*semver.Version
		for version, registeredToolchain := range m.RegisteredToolchains {
			if registeredToolchain.Name == contextToolchain {
				registeredVersions = append(registeredVersions, version)
			}
		}
		sort.Sort(sort.Reverse(semver.Collection(registeredVersions)))

		// Get latest compatible registered version
		compatible := false
		for _, registeredVersion := range registeredVersions {
			for _, configVersion := range configVersions {
				if !registeredVersion.LessThan(configVersion) {
					m.SelectedToolchainVersion[index] = registeredVersion
					m.SelectedToolchainConfig[index] = m.ToolchainConfigs[configVersion].Path
					compatible = true
					break
				}
			}
			if compatible {
				for constraint := range contextConstraints {
					if !constraint.Check(registeredVersion) {
						compatible = false
						break
					}
				}
			}
			if compatible {
				break
			}
		}
		if !compatible {
			err := errors.New("no compatible registered toolchain was found for " + contextToolchain)
			return err
		}

		// Debug
		if m.Params.Options.Debug {
			log.Debug("Latest compatible registered toolchain: " + m.RegisteredToolchains[m.SelectedToolchainVersion[index]].Name + " " + m.SelectedToolchainVersion[index].String())
			log.Debug("Compatible config file: " + m.SelectedToolchainConfig[index])
		}
	}

	return nil
}
