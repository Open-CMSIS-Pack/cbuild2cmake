/*
 * Copyright (c) 2023 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"encoding/json"
	"os"
	"path"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type CbuildIndex struct {
	BuildIdx struct {
		GeneratedBy string `yaml:"generated-by"`
		Cdefault    string `yaml:"cdefault"`
		Csolution   string `yaml:"csolution"`
		Cprojects   []struct {
			Cproject string `yaml:"cproject"`
		} `yaml:"cprojects"`
		Licenses interface{} `yaml:"licenses"`
		Cbuilds  []struct {
			Cbuild        string `yaml:"cbuild"`
			Project       string `json:"project"`
			Configuration string `json:"configuration"`
		} `yaml:"cbuilds"`
	} `yaml:"build-idx"`
	BaseDir string
}

type Cbuild struct {
	BuildDescType struct {
		GeneratedBy      string        `yaml:"generated-by"`
		CurrentGenerator []struct{}    `yaml:"current-generator"`
		Solution         string        `yaml:"solution"`
		Project          string        `yaml:"project"`
		Context          string        `yaml:"context"`
		Compiler         string        `yaml:"compiler"`
		Board            string        `yaml:"board"`
		BoardPack        string        `yaml:"board-pack"`
		Device           string        `yaml:"device"`
		DevicePack       string        `yaml:"device-pack"`
		Processor        struct{}      `yaml:"processor"`
		Packs            []struct{}    `yaml:"packs"`
		Optimize         string        `yaml:"optimize"`
		Debug            string        `yaml:"debug"`
		Warnings         string        `yaml:"warnings"`
		Misc             struct{}      `yaml:"misc"`
		Define           []interface{} `yaml:"define"`
		AddPath          []string      `yaml:"add-path"`
		OutputDirs       struct {
			Intdir  string `yaml:"intdir"`
			Outdir  string `yaml:"outdir"`
			Cprjdir string `yaml:"cprjdir"`
		} `yaml:"output-dirs"`
		Output []struct {
			File string `yaml:"file"`
			Type string `yaml:"type"`
		} `yaml:"output"`
		Components       []struct{} `yaml:"components"`
		Linker           struct{}   `yaml:"linker"`
		Groups           []struct{} `yaml:"groups"`
		Generators       []struct{} `yaml:"generators"`
		ConstructedFiles []struct{} `yaml:"constructed-files"`
		Licenses         []struct{} `yaml:"licenses"`
	} `yaml:"build"`
	BaseDir string
}

func (m *Maker) ParseCbuildIndexFile(cbuildIndexFile string) (data CbuildIndex, err error) {
	yfile, err := os.ReadFile(cbuildIndexFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yfile, &data)
	return
}

func (m *Maker) ParseCbuildFile(cbuildFile string) (data Cbuild, err error) {
	yfile, err := os.ReadFile(cbuildFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yfile, &data)
	return
}

func (m *Maker) ParseCbuildFiles() error {
	// Parse cbuild-idx file
	cbuildIndex, err := m.ParseCbuildIndexFile(m.Params.InputFile)
	if err != nil {
		return err
	}
	cbuildIndex.BaseDir = path.Dir(m.Params.InputFile)
	m.CbuildIndex = cbuildIndex

	// Debug
	if m.Params.Options.Debug {
		s, _ := json.MarshalIndent(cbuildIndex, "", "\t")
		log.Debug(string(s))
	}

	// Parse cbuild files
	for _, cbuildRef := range m.CbuildIndex.BuildIdx.Cbuilds {
		cbuildFile := path.Join(m.CbuildIndex.BaseDir, cbuildRef.Cbuild)
		if _, err := os.Stat(cbuildFile); os.IsNotExist(err) {
			log.Warn("file " + cbuildFile + " was not found")
			continue
		}
		cbuild, err := m.ParseCbuildFile(cbuildFile)
		if err != nil {
			return err
		}
		cbuild.BaseDir = path.Dir(cbuildFile)
		m.Cbuilds = append(m.Cbuilds, cbuild)

		// Debug
		if m.Params.Options.Debug {
			s, _ := json.MarshalIndent(cbuild, "", "\t")
			log.Debug(string(s))
		}
	}
	return err
}
