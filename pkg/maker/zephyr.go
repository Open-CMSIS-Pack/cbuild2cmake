/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	utils "github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	sortedmap "github.com/gobs/sortedmap"
)

type ZephyrLayer struct {
	Clayer     Clayer
	Components []Components
	Packs      []Packs
}

type ZephyrMaker struct {
	Layers           []ZephyrLayer
	PackPaths        map[string]string
	ComponentFiles   map[string]BuildFiles
	CompileOptions   map[string][]string
	Cbuild           *Cbuild
	RteComponents    bool
	PreIncludeGlobal bool
	DupComponents    []string
}

func (m *Maker) GenerateZephyrModules() error {
	// Parse clayer files
	err := m.ParseClayerFiles()
	if err != nil {
		return err
	}

	m.ZephyrMaker.Cbuild = &m.Cbuilds[0]
	m.ZephyrMaker.Cbuild.ContextRoot, _ = filepath.Rel(m.SolutionRoot, m.ZephyrMaker.Cbuild.BaseDir)
	m.ZephyrMaker.Cbuild.ContextRoot = filepath.ToSlash(m.ZephyrMaker.Cbuild.ContextRoot)
	m.ZephyrMaker.Layers = nil
	m.ZephyrMaker.DupComponents = nil
	m.ZephyrMaker.PackPaths = make(map[string]string)
	m.ZephyrMaker.ComponentFiles = make(map[string]BuildFiles)
	m.ZephyrMaker.CompileOptions = make(map[string][]string)
	allComponents := []string{}

	// Iterate over clayers
	for clayerIndex := range m.Clayers {
		zephyrLayer := ZephyrLayer{
			Clayer: m.Clayers[clayerIndex],
		}
		// Packs
		for _, layerPack := range zephyrLayer.Clayer.Layer.Packs {
			for _, pack := range m.ZephyrMaker.Cbuild.BuildDescType.Packs {
				vendor, name, _ := utils.ExtractPackIdParts(pack.Pack)
				layerVendor, layerName, _ := utils.ExtractPackIdParts(layerPack.Pack)
				if vendor == layerVendor && name == layerName {
					zephyrLayer.Packs = append(zephyrLayer.Packs, pack)
					if filepath.IsAbs(pack.Path) || strings.HasPrefix(pack.Path, "${") {
						m.ZephyrMaker.PackPaths[pack.Pack] = pack.Path
					} else {
						packPath, _ := filepath.Rel(path.Join(m.SolutionRoot, m.SolutionName), (path.Join(m.ZephyrMaker.Cbuild.BaseDir, pack.Path)))
						m.ZephyrMaker.PackPaths[pack.Pack] = "${CMAKE_CURRENT_LIST_DIR}/" + filepath.ToSlash(packPath)
					}
				}
			}
		}
		// Components
		for _, clayerComponent := range zephyrLayer.Clayer.Layer.Components {
			for _, cbuildComponent := range m.ZephyrMaker.Cbuild.BuildDescType.Components {
				if cbuildComponent.SelectedBy == clayerComponent.Component {
					// track duplicate components across layers
					if slices.Contains(allComponents, cbuildComponent.Component) {
						m.ZephyrMaker.DupComponents = append(m.ZephyrMaker.DupComponents, cbuildComponent.Component)
					} else {
						allComponents = append(allComponents, cbuildComponent.Component)
					}
					// Add component to layer and classify files
					zephyrLayer.Components = append(zephyrLayer.Components, cbuildComponent)
					m.ZephyrMaker.ComponentFiles[cbuildComponent.Component] = m.ZephyrMaker.Cbuild.ClassifyFiles(cbuildComponent.Files)
					break
				}
			}
		}
		m.ZephyrMaker.Layers = append(m.ZephyrMaker.Layers, zephyrLayer)
	}

	// Compile options
	m.ZephyrMaker.Cbuild.GetCompileOptionsLanguageMap(false, m.ZephyrMaker.Cbuild.BuildDescType.Misc, &m.ZephyrMaker.CompileOptions)

	// Copy CMSIS constructed files
	m.ZephyrMaker.RteComponents = false
	m.ZephyrMaker.PreIncludeGlobal = false
	for _, file := range m.ZephyrMaker.Cbuild.BuildDescType.ConstructedFiles {
		if strings.HasSuffix(file.File, "RTE_Components.h") {
			m.ZephyrMaker.RteComponents = true
		} else if strings.HasSuffix(file.File, "Pre_Include_Global.h") {
			m.ZephyrMaker.PreIncludeGlobal = true
		} else {
			continue
		}
		err = utils.CopyFile(path.Join(m.ZephyrMaker.Cbuild.BaseDir, file.File), path.Join(m.SolutionRoot, m.SolutionName, path.Base(file.File)))
		if err != nil {
			return err
		}
	}

	// Generate Zephyr module files
	err = m.GenerateModuleYml()
	if err != nil {
		return err
	}
	// Generate Kconfig
	err = m.GenerateModuleKconfig()
	if err != nil {
		return err
	}
	// Generate CMakeLists.txt
	err = m.GenerateModuleCMakeLists()
	if err != nil {
		return err
	}
	// Generate sources.cmake
	err = m.GenerateModuleCMakeSources()
	if err != nil {
		return err
	}
	return nil
}

func (m *Maker) GetFilePath(file string, packId string) string {
	packPath := strings.ReplaceAll(m.ZephyrMaker.PackPaths[packId], "${CMAKE_CURRENT_LIST_DIR}", path.Join(m.SolutionRoot, m.SolutionName))
	file = strings.ReplaceAll(file, "${SOLUTION_ROOT}", m.SolutionRoot)
	file, _ = filepath.Rel(packPath, file)
	return "${" + strings.ToUpper(ReplaceSpecialChars(strings.SplitN(packId, "@", 2)[0])) + "}/" + filepath.ToSlash(file)
}

func (m *Maker) GenerateModuleYml() error {
	content := `name: ` + m.SolutionName + `
build:
  cmake: .
  kconfig: Kconfig
`
	// Write module.yml
	moduleYml := path.Join(m.SolutionRoot, m.SolutionName, "zephyr", "module.yml")
	err := utils.UpdateFile(moduleYml, content)
	if err != nil {
		return err
	}
	return nil
}

func (m *Maker) GenerateModuleKconfig() error {
	content := ""
	// Iterate over zephyr layers
	for _, zephyrLayer := range m.ZephyrMaker.Layers {
		menuConfigName := "CMSIS_" + strings.ToUpper(ReplaceSpecialChars(zephyrLayer.Clayer.Name))
		source, _ := filepath.Rel(m.SolutionRoot, zephyrLayer.Clayer.File)
		source = filepath.ToSlash(source)
		cpp := ""
		if slices.Contains(m.ZephyrMaker.Cbuild.Languages, "CXX") {
			cpp = "\n    depends on CPP && STD_CPP17"
		}
		content += `menuconfig ` + menuConfigName + `
    bool "` + zephyrLayer.Clayer.Name + ` (CMSIS-Pack)"` + cpp + `
    default y
    help
        ` + zephyrLayer.Clayer.Layer.Description + `
        Source: ` + source + "\n\n"

		// Iterate over components
		content += "if " + menuConfigName + "\n\n"
		for _, component := range zephyrLayer.Components {
			componentName := strings.SplitN(component.Component, "@", 2)[0]
			configName := menuConfigName + "_" + strings.ToUpper(ReplaceSpecialChars(componentName))
			content += `config ` + configName + `
    bool "` + componentName + `"
    default y
    help
      Component: ` + component.Component + "\n\n"
		}
		content += "endif # " + menuConfigName + "\n\n"
	}
	content = strings.TrimSuffix(content, "\n")

	// Write Kconfig
	kconfig := path.Join(m.SolutionRoot, m.SolutionName, "Kconfig")
	err := utils.UpdateFile(kconfig, content)
	if err != nil {
		return err
	}
	return nil
}

func (m *Maker) GenerateModuleCMakeLists() error {
	// Iterate over zephyr layers
	var configList []string
	for _, zephyrLayer := range m.ZephyrMaker.Layers {
		configList = append(configList, "CONFIG_CMSIS_"+strings.ToUpper(ReplaceSpecialChars(zephyrLayer.Clayer.Name)))
	}
	content := "include(${CMAKE_CURRENT_LIST_DIR}/sources.cmake)\n"
	if len(configList) > 0 {
		content = "if(" + strings.Join(configList, " OR ") + ")\n  " + content + "endif()\n"
	}

	// Write CMakeLists.txt
	cmakeLists := path.Join(m.SolutionRoot, m.SolutionName, "CMakeLists.txt")
	err := utils.UpdateFile(cmakeLists, content)
	if err != nil {
		return err
	}
	return nil
}

func (m *Maker) GenerateModuleCMakeSources() error {
	content := "set(CMSIS_PACK_ROOT $ENV{CMSIS_PACK_ROOT})\n"
	if len(m.ZephyrMaker.PackPaths) > 0 {
		fallback := ""
		pdscs := []string{}
		// Iterate over used packs and set pack paths
		for packId, packPath := range m.ZephyrMaker.PackPaths {
			packName := strings.ToUpper(ReplaceSpecialChars(strings.SplitN(packId, "@", 2)[0]))
			content += "cmake_path(SET " + packName + " NORMALIZE \"" + packPath + "\")\n"
			vendor, name, version := utils.ExtractPackIdParts(packId)
			pdsc := vendor + "." + name + ".pdsc"
			pdscs = append(pdscs, "\"${"+packName+"}/"+pdsc+"\"")
			if strings.HasPrefix(packPath, "${CMAKE_CURRENT_LIST_DIR}") {
				fallback += `if(NOT EXISTS "${` + packName + `}/` + pdsc + `")
  # Fallback to CMSIS_PACK_ROOT
  cmake_path(SET ` + packName + ` NORMALIZE "${CMSIS_PACK_ROOT}/` + vendor + `/` + name + `/` + version + `")
endif()

`
			}
		}
		content += "\n" + fallback
		content += `if(NOT EXISTS ` + strings.Join(pdscs, "\n  OR NOT EXISTS ") + `)
  message(FATAL_ERROR "Pack(s) not found. Set pack path or CMSIS_PACK_ROOT.")
endif()

`
	}

	// Add compile definitions
	defineList := ListCompileDefinitions(m.ZephyrMaker.Cbuild.BuildDescType.Define, "\n  ")
	if len(defineList) > 0 {
		content += "zephyr_compile_definitions(\n  " + defineList + "\n)\n\n"
	}

	// Add include directories
	if m.ZephyrMaker.RteComponents {
		content += `# RTE_Components.h (CMSIS-Pack component selection header)
zephyr_include_directories(
  "${CMAKE_CURRENT_LIST_DIR}"
)

`
	}

	// Add global pre-include header
	if m.ZephyrMaker.PreIncludeGlobal {
		content += `# Pre_Include_Global.h (CMSIS-Pack global pre-include header)
zephyr_compile_options(
  "SHELL:-include \"${CMAKE_CURRENT_LIST_DIR}/Pre_Include_Global.h\""
)

`
	}

	// Add compile options
	if len(m.ZephyrMaker.CompileOptions) > 0 {
		compileOptions := ""
		for _, language := range sortedmap.AsSortedMap(m.ZephyrMaker.CompileOptions) {
			compileOptions += m.ZephyrMaker.Cbuild.LanguageSpecificCompileOptions(language.Key, language.Value...)
		}
		content += `# Misc compile options
zephyr_compile_options(` + compileOptions + `
)

`
	}

	// Iterate over zephyr layers
	for _, zephyrLayer := range m.ZephyrMaker.Layers {
		for _, component := range zephyrLayer.Components {
			componentName := strings.SplitN(component.Component, "@", 2)[0]
			configName := ReplaceSpecialChars(zephyrLayer.Clayer.Name) + "_" + ReplaceSpecialChars(componentName)
			commentLine := "# ── " + componentName + " "
			content += commentLine + strings.Repeat("─", max(0, 80-len(commentLine)))
			content += "\nif(CONFIG_CMSIS_" + strings.ToUpper(configName)
			if slices.Contains(m.ZephyrMaker.DupComponents, component.Component) {
				includeGuard := strings.ToUpper(ReplaceSpecialChars(componentName)) + "_INCLUDED"
				content += " AND NOT " + includeGuard + ")\n\n  set(" + includeGuard + " TRUE"
			}
			content += ")\n\n"

			// Set zephyr library name
			content += "  zephyr_library_named(" + strings.ToLower(configName) + ")\n\n"

			// Set zephyr sources
			if len(m.ZephyrMaker.ComponentFiles[component.Component].Source) > 0 {
				content += "  zephyr_library_sources("
				for _, language := range sortedmap.AsSortedMap(m.ZephyrMaker.ComponentFiles[component.Component].Source) {
					for _, file := range language.Value {
						content += "\n    \"" + m.GetFilePath(file, component.FromPack) + "\""
					}
				}
				content += "\n  )\n\n"
			}

			// Set zephyr libraries (object files and libraries)
			libraries := m.ZephyrMaker.ComponentFiles[component.Component].Library
			libraries = append(libraries, m.ZephyrMaker.ComponentFiles[component.Component].WholeArchive...)
			libraries = append(libraries, m.ZephyrMaker.ComponentFiles[component.Component].Object...)
			if len(libraries) > 0 {
				content += "  zephyr_library_import("
				for _, file := range libraries {
					content += "\n    \"" + m.GetFilePath(file, component.FromPack) + "\""
				}
				content += "\n  )\n\n"
			}

			// Set zephyr include directories
			if len(m.ZephyrMaker.ComponentFiles[component.Component].Include) > 0 {
				content += "  zephyr_include_directories("
				for _, languages := range sortedmap.AsSortedMap(m.ZephyrMaker.ComponentFiles[component.Component].Include) {
					for _, files := range languages.Value {
						for _, file := range files {
							content += "\n    \"" + m.GetFilePath(file, component.FromPack) + "\""
						}
					}
				}
				content += "\n  )\n\n"
			}

			// Set zephyr compile options for C++ files
			if _, ok := m.ZephyrMaker.ComponentFiles[component.Component].Source["CXX"]; ok {
				content += "  zephyr_library_compile_options(-std=c++17 -fno-rtti -fno-exceptions)\n\n"
			}

			content = strings.TrimSuffix(content, "\n")
			content += "\nendif()\n\n"
		}
	}
	content = strings.TrimSuffix(content, "\n")

	// Write sources.cmake
	cmakeSources := path.Join(m.SolutionRoot, m.SolutionName, "sources.cmake")
	err := utils.UpdateFile(cmakeSources, content)
	if err != nil {
		return err
	}
	return nil
}
