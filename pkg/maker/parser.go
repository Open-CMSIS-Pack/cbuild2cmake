/*
 * Copyright (c) 2024-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

type CbuildIndex struct {
	BuildIdx struct {
		GeneratedBy string      `yaml:"generated-by"`
		Cdefault    string      `yaml:"cdefault"`
		Csolution   string      `yaml:"csolution"`
		ImageOnly   bool        `yaml:"image-only"`
		TmpDir      string      `yaml:"tmpdir"`
		Cprojects   []Cprojects `yaml:"cprojects"`
		Cbuilds     []Cbuilds   `yaml:"cbuilds"`
		Executes    []Executes  `yaml:"executes"`
	} `yaml:"build-idx"`
	BaseDir string
}

type CbuildSet struct {
	BuildSet struct {
		GeneratedBy string     `yaml:"generated-by"`
		Contexts    []Contexts `yaml:"contexts"`
		Compiler    string     `yaml:"compiler"`
	} `yaml:"cbuild-set"`
}

type Cbuild struct {
	BuildDescType struct {
		GeneratedBy      string        `yaml:"generated-by"`
		CurrentGenerator struct{}      `yaml:"current-generator"`
		Solution         string        `yaml:"solution"`
		Project          string        `yaml:"project"`
		Context          string        `yaml:"context"`
		Compiler         string        `yaml:"compiler"`
		Board            string        `yaml:"board"`
		BoardPack        string        `yaml:"board-pack"`
		Device           string        `yaml:"device"`
		DevicePack       string        `yaml:"device-pack"`
		Processor        Processor     `yaml:"processor"`
		Packs            []Packs       `yaml:"packs"`
		Optimize         string        `yaml:"optimize"`
		Debug            string        `yaml:"debug"`
		Warnings         string        `yaml:"warnings"`
		LanguageC        string        `yaml:"language-C"`
		LanguageCpp      string        `yaml:"language-CPP"`
		Lto              bool          `yaml:"link-time-optimize"`
		Misc             Misc          `yaml:"misc"`
		Define           []interface{} `yaml:"define"`
		DefineAsm        []interface{} `yaml:"define-asm"`
		AddPath          []string      `yaml:"add-path"`
		AddPathAsm       []string      `yaml:"add-path-asm"`
		OutputDirs       OutputDirs    `yaml:"output-dirs"`
		Output           []Output      `yaml:"output"`
		Components       []Components  `yaml:"components"`
		Apis             []Apis        `yaml:"apis"`
		Linker           Linker        `yaml:"linker"`
		Groups           []Groups      `yaml:"groups"`
		Generators       []struct{}    `yaml:"generators"`
		ConstructedFiles []Files       `yaml:"constructed-files"`
		Licenses         []struct{}    `yaml:"licenses"`
		West             West          `yaml:"west"`
	} `yaml:"build"`
	BaseDir          string
	ContextRoot      string
	SolutionRoot     string
	Languages        []string
	PreIncludeGlobal []string
	LibraryGlobal    []string
	IncludeGlobal    LanguageMap
	UserIncGlobal    LanguageMap
	BuildGroups      []string
	Toolchain        string
	GeneratedFiles   []string
	LinkerLto        bool
}

type Cbuilds struct {
	Cbuild        string   `yaml:"cbuild"`
	Project       string   `yaml:"project"`
	Configuration string   `yaml:"configuration"`
	DependsOn     []string `yaml:"depends-on"`
	West          bool     `yaml:"west"`
}

type Clayers struct {
	Clayer string `yaml:"clayer"`
}

type Contexts struct {
	Context string `yaml:"context"`
}

type Cprojects struct {
	Cproject string    `yaml:"cproject"`
	Clayers  []Clayers `yaml:"clayers"`
}

type Apis struct {
	API      string  `yaml:"api"`
	Files    []Files `yaml:"files"`
	FromPack string  `yaml:"from-pack"`
}

type Components struct {
	Component   string        `yaml:"component"`
	Condition   string        `yaml:"condition"`
	SelectedBy  string        `yaml:"selected-by"`
	Implements  string        `yaml:"implements"`
	Rtedir      string        `yaml:"rtedir"`
	Optimize    string        `yaml:"optimize"`
	Debug       string        `yaml:"debug"`
	Warnings    string        `yaml:"warnings"`
	LanguageC   string        `yaml:"language-C"`
	LanguageCpp string        `yaml:"language-CPP"`
	Lto         bool          `yaml:"link-time-optimize"`
	Define      []interface{} `yaml:"define"`
	DefineAsm   []interface{} `yaml:"define-asm"`
	Undefine    []string      `yaml:"undefine"`
	AddPath     []string      `yaml:"add-path"`
	AddPathAsm  []string      `yaml:"add-path-asm"`
	DelPath     []string      `yaml:"del-path"`
	Misc        Misc          `yaml:"misc"`
	Files       []Files       `yaml:"files"`
	Generator   Generator     `yaml:"generator"`
	FromPack    string        `yaml:"from-pack"`
}

type Executes struct {
	Execute   string                 `yaml:"execute"`
	Run       string                 `yaml:"run"`
	Always    map[string]interface{} `yaml:"always,inline"`
	Input     []string               `yaml:"input"`
	Output    []string               `yaml:"output"`
	DependsOn []string               `yaml:"depends-on"`
}

type Files struct {
	File        string        `yaml:"file"`
	Category    string        `yaml:"category"`
	Scope       string        `yaml:"scope"`
	Language    string        `yaml:"language"`
	Attr        string        `yaml:"attr"`
	Version     string        `yaml:"version"`
	Optimize    string        `yaml:"optimize"`
	Debug       string        `yaml:"debug"`
	Warnings    string        `yaml:"warnings"`
	LanguageC   string        `yaml:"language-C"`
	LanguageCpp string        `yaml:"language-CPP"`
	Lto         bool          `yaml:"link-time-optimize"`
	Define      []interface{} `yaml:"define"`
	DefineAsm   []interface{} `yaml:"define-asm"`
	Undefine    []string      `yaml:"undefine"`
	AddPath     []string      `yaml:"add-path"`
	AddPathAsm  []string      `yaml:"add-path-asm"`
	DelPath     []string      `yaml:"del-path"`
	Misc        Misc          `yaml:"misc"`
}

type Generator struct {
	ID       string  `yaml:"id"`
	Path     string  `yaml:"path"`
	FromPack string  `yaml:"from-pack"`
	Files    []Files `yaml:"files"`
}

type Groups struct {
	Group       string        `yaml:"group"`
	Groups      []Groups      `yaml:"groups"`
	Files       []Files       `yaml:"files"`
	Optimize    string        `yaml:"optimize"`
	Debug       string        `yaml:"debug"`
	Warnings    string        `yaml:"warnings"`
	LanguageC   string        `yaml:"language-C"`
	LanguageCpp string        `yaml:"language-CPP"`
	Lto         bool          `yaml:"link-time-optimize"`
	Define      []interface{} `yaml:"define"`
	DefineAsm   []interface{} `yaml:"define-asm"`
	Undefine    []string      `yaml:"undefine"`
	AddPath     []string      `yaml:"add-path"`
	AddPathAsm  []string      `yaml:"add-path-asm"`
	DelPath     []string      `yaml:"del-path"`
	Misc        Misc          `yaml:"misc"`
}

type Linker struct {
	Regions string        `yaml:"regions"`
	Script  string        `yaml:"script"`
	Define  []interface{} `yaml:"define"`
}

type Misc struct {
	C       []string `yaml:"C"`
	CPP     []string `yaml:"CPP"`
	CCPP    []string `yaml:"C-CPP"`
	ASM     []string `yaml:"ASM"`
	Link    []string `yaml:"Link"`
	LinkC   []string `yaml:"Link-C"`
	LinkCPP []string `yaml:"Link-CPP"`
	Library []string `yaml:"Library"`
	Lib     []string `yaml:"Lib"`
}

type OutputDirs struct {
	Intdir  string `yaml:"intdir"`
	Outdir  string `yaml:"outdir"`
	Cprjdir string `yaml:"cprjdir"`
	Rtedir  string `yaml:"rtedir"`
}

type Output struct {
	File string `yaml:"file"`
	Type string `yaml:"type"`
}

type Processor struct {
	Fpu              string `yaml:"fpu"`
	Dsp              string `yaml:"dsp"`
	Mve              string `yaml:"mve"`
	Endian           string `yaml:"endian"`
	Trustzone        string `yaml:"trustzone"`
	BranchProtection string `yaml:"branch-protection"`
	Core             string `yaml:"core"`
}

type Packs struct {
	Pack string `yaml:"pack"`
	Path string `yaml:"path"`
}

type West struct {
	ProjectId string        `yaml:"project-id"`
	AppPath   string        `yaml:"app-path"`
	Board     string        `yaml:"board"`
	Device    string        `yaml:"device"`
	WestDefs  []interface{} `yaml:"west-defs"`
	WestOpt   []string      `yaml:"west-opt"`
}

func (m *Maker) ParseCbuildIndexFile(cbuildIndexFile string) (data CbuildIndex, err error) {
	yfile, err := os.ReadFile(cbuildIndexFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yfile, &data)
	return
}

func (m *Maker) ParseCbuildSetFile(cbuildSetFile string) (data CbuildSet, err error) {
	yfile, err := os.ReadFile(cbuildSetFile)
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
	cbuildIndex.BaseDir, _ = filepath.Abs(path.Dir(m.Params.InputFile))
	cbuildIndex.BaseDir = filepath.ToSlash(cbuildIndex.BaseDir)
	m.CbuildIndex = cbuildIndex
	m.SolutionRoot = filepath.Dir(filepath.Join(cbuildIndex.BaseDir, cbuildIndex.BuildIdx.Csolution))
	m.SolutionRoot, _ = filepath.EvalSymlinks(m.SolutionRoot)
	m.SolutionRoot = filepath.ToSlash(m.SolutionRoot)
	m.SolutionName = filepath.Base(m.CbuildIndex.BuildIdx.Csolution)
	reg := regexp.MustCompile(`(.*)\.csolution.ya?ml`)
	m.SolutionName = reg.ReplaceAllString(m.SolutionName, "$1")

	// Parse cbuild-set file
	if m.Options.UseContextSet {
		cbuildSetFile, _ := filepath.Abs(m.Params.InputFile[:len(m.Params.InputFile)-len(".cbuild-idx.yml")] + ".cbuild-set.yml")
		cbuildSetFile = filepath.ToSlash(cbuildSetFile)
		cbuildSet, err := m.ParseCbuildSetFile(cbuildSetFile)
		if err != nil {
			return err
		}
		for _, item := range cbuildSet.BuildSet.Contexts {
			m.Contexts = append(m.Contexts, item.Context)
		}
		m.CbuildSet = cbuildSet
	}

	// Parse cbuild files
	for _, cbuildRef := range m.CbuildIndex.BuildIdx.Cbuilds {
		if m.Options.UseContextSet && !slices.Contains(m.Contexts, cbuildRef.Project+cbuildRef.Configuration) {
			continue
		}
		cbuildFile := path.Join(m.CbuildIndex.BaseDir, cbuildRef.Cbuild)
		if _, err := os.Stat(cbuildFile); os.IsNotExist(err) {
			log.Warn("file " + cbuildFile + " was not found")
			continue
		}
		cbuild, err := m.ParseCbuildFile(cbuildFile)
		if err != nil {
			return err
		}
		if !m.Options.UseContextSet {
			m.Contexts = append(m.Contexts, cbuild.BuildDescType.Context)
		}
		cbuild.BaseDir, _ = filepath.Abs(path.Dir(cbuildFile))
		cbuild.BaseDir = filepath.ToSlash(cbuild.BaseDir)
		cbuild.SolutionRoot = m.SolutionRoot
		m.Cbuilds = append(m.Cbuilds, cbuild)
	}
	return err
}
