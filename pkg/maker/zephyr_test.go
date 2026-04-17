/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZephyr(t *testing.T) {
	t.Run("test GetFilePath", func(t *testing.T) {
		assert := assert.New(t)
		var m maker.Maker
		m.SolutionRoot = filepath.ToSlash(t.TempDir())
		m.SolutionName = "solution"
		m.ZephyrMaker.PackPaths = map[string]string{
			"Vendor::Pack@1.0.0": "${CMAKE_CURRENT_LIST_DIR}/packs/vendor/pack",
		}

		filePath := m.GetFilePath("${SOLUTION_ROOT}/solution/packs/vendor/pack/src/main.c", "Vendor::Pack@1.0.0")
		assert.Equal("${VENDOR_PACK}/src/main.c", filePath)
	})

	t.Run("test module.yml generation", func(t *testing.T) {
		assert := assert.New(t)
		var m maker.Maker
		m.SolutionRoot = filepath.ToSlash(t.TempDir())
		m.SolutionName = "zephyr-solution"

		err := m.GenerateModuleYml()
		assert.NoError(err)

		content, err := os.ReadFile(path.Join(m.SolutionRoot, m.SolutionName, "zephyr", "module.yml"))
		assert.NoError(err)
		assert.Equal("name: zephyr-solution\nbuild:\n  cmake: .\n  kconfig: Kconfig\n", string(content))
	})

	t.Run("test Kconfig generation", func(t *testing.T) {
		assert := assert.New(t)
		var m maker.Maker
		m.SolutionRoot = filepath.ToSlash(t.TempDir())
		m.SolutionName = "solution"
		cbuild := maker.Cbuild{Languages: []string{"C", "CXX"}}
		m.ZephyrMaker.Cbuild = &cbuild

		layer := maker.ZephyrLayer{
			Components: []maker.Components{
				{Component: "Vendor::Pack:Comp@1.0.0"},
			},
		}
		layer.Clayer.Name = "layer-main"
		layer.Clayer.File = path.Join(m.SolutionRoot, "layers", "layer-main.clayer.yml")
		layer.Clayer.Layer.Description = "Main layer description"
		m.ZephyrMaker.Layers = []maker.ZephyrLayer{layer}

		err := m.GenerateModuleKconfig()
		assert.NoError(err)

		content, err := os.ReadFile(path.Join(m.SolutionRoot, m.SolutionName, "Kconfig"))
		assert.NoError(err)
		text := string(content)
		assert.Contains(text, "menuconfig CMSIS_LAYER_MAIN")
		assert.Contains(text, "depends on CPP && STD_CPP17")
		assert.Contains(text, "Source: layers/layer-main.clayer.yml")
		assert.Contains(text, "config CMSIS_LAYER_MAIN_VENDOR_PACK_COMP")
		assert.Contains(text, "Component: Vendor::Pack:Comp@1.0.0")
	})

	t.Run("test CMakeLists generation", func(t *testing.T) {
		assert := assert.New(t)
		var m maker.Maker
		m.SolutionRoot = filepath.ToSlash(t.TempDir())
		m.SolutionName = "solution"

		layerA := maker.ZephyrLayer{}
		layerA.Clayer.Name = "layer-a"
		layerB := maker.ZephyrLayer{}
		layerB.Clayer.Name = "layer b"
		m.ZephyrMaker.Layers = []maker.ZephyrLayer{layerA, layerB}

		err := m.GenerateModuleCMakeLists()
		assert.NoError(err)

		content, err := os.ReadFile(path.Join(m.SolutionRoot, m.SolutionName, "CMakeLists.txt"))
		assert.NoError(err)
		assert.Equal("if(CONFIG_CMSIS_LAYER_A OR CONFIG_CMSIS_LAYER_B)\n  include(${CMAKE_CURRENT_LIST_DIR}/sources.cmake)\nendif()\n", string(content))
	})

	t.Run("test sources.cmake generation", func(t *testing.T) {
		assert := assert.New(t)
		var m maker.Maker
		m.SolutionRoot = filepath.ToSlash(t.TempDir())
		m.SolutionName = "solution"

		componentID := "Vendor::Pack:Comp@1.0.0"
		component := maker.Components{
			Component: componentID,
			FromPack:  "Vendor::Pack@1.0.0",
		}

		cbuild := maker.Cbuild{}
		cbuild.BuildDescType.Define = []interface{}{"DEF_SCALAR", map[string]interface{}{"DEF_KEY": "VALUE"}}
		m.ZephyrMaker.Cbuild = &cbuild
		m.ZephyrMaker.PackPaths = map[string]string{
			"Vendor::Pack@1.0.0": "${CMAKE_CURRENT_LIST_DIR}/packs/vendor/pack",
		}
		m.ZephyrMaker.CompileOptions = map[string][]string{
			"C":   {"-O2"},
			"CXX": {"-fno-exceptions"},
		}
		m.ZephyrMaker.RteComponents = true
		m.ZephyrMaker.PreIncludeGlobal = true
		m.ZephyrMaker.DupComponents = []string{componentID}

		layer := maker.ZephyrLayer{Components: []maker.Components{component}}
		layer.Clayer.Name = "layer-1"
		m.ZephyrMaker.Layers = []maker.ZephyrLayer{layer}
		m.ZephyrMaker.ComponentFiles = map[string]maker.BuildFiles{
			componentID: {
				Source: maker.LanguageMap{
					"C":   {"${SOLUTION_ROOT}/solution/packs/vendor/pack/src/main.c"},
					"CXX": {"${SOLUTION_ROOT}/solution/packs/vendor/pack/src/main.cpp"},
				},
				Library:      []string{"${SOLUTION_ROOT}/solution/packs/vendor/pack/lib/liba.a"},
				WholeArchive: []string{"${SOLUTION_ROOT}/solution/packs/vendor/pack/lib/libwa.a"},
				Object:       []string{"${SOLUTION_ROOT}/solution/packs/vendor/pack/obj/startup.o"},
				Include: maker.ScopeMap{
					"PUBLIC": {
						"C": {"${SOLUTION_ROOT}/solution/packs/vendor/pack/include"},
					},
				},
			},
		}

		err := m.GenerateModuleCMakeSources()
		assert.NoError(err)

		content, err := os.ReadFile(path.Join(m.SolutionRoot, m.SolutionName, "sources.cmake"))
		assert.NoError(err)
		text := string(content)
		assert.Contains(text, "set(CMSIS_PACK_ROOT $ENV{CMSIS_PACK_ROOT})")
		assert.Contains(text, "cmake_path(SET VENDOR_PACK NORMALIZE \"${CMAKE_CURRENT_LIST_DIR}/packs/vendor/pack\")")
		assert.Contains(text, "if(NOT EXISTS \"${VENDOR_PACK}/Vendor.Pack.pdsc\")")
		assert.Contains(text, "cmake_path(SET VENDOR_PACK NORMALIZE \"${CMSIS_PACK_ROOT}/Vendor/Pack/1.0.0\")")
		assert.Contains(text, "message(FATAL_ERROR \"Pack(s) not found. Set pack path or CMSIS_PACK_ROOT.\")")
		assert.Contains(text, "zephyr_compile_definitions(")
		assert.Contains(text, "DEF_SCALAR")
		assert.Contains(text, "DEF_KEY=VALUE")
		assert.Contains(text, "# RTE_Components.h (CMSIS-Pack component selection header)")
		assert.Contains(text, "# Pre_Include_Global.h (CMSIS-Pack global pre-include header)")
		assert.Contains(text, "# Misc compile options")
		assert.Contains(text, "if(CONFIG_CMSIS_LAYER_1_VENDOR_PACK_COMP AND NOT VENDOR_PACK_COMP_INCLUDED)")
		assert.Contains(text, "zephyr_library_named(layer_1_vendor_pack_comp)")
		assert.Contains(text, "\"${VENDOR_PACK}/src/main.c\"")
		assert.Contains(text, "\"${VENDOR_PACK}/src/main.cpp\"")
		assert.Contains(text, "\"${VENDOR_PACK}/lib/liba.a\"")
		assert.Contains(text, "\"${VENDOR_PACK}/lib/libwa.a\"")
		assert.Contains(text, "\"${VENDOR_PACK}/obj/startup.o\"")
		assert.Contains(text, "\"${VENDOR_PACK}/include\"")
		assert.Contains(text, "zephyr_library_compile_options(-std=c++17 -fno-rtti -fno-exceptions)")
	})

	t.Run("test Zephyr modules generation", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		root := filepath.ToSlash(t.TempDir())
		baseDir := path.Join(root, "build")
		require.NoError(os.MkdirAll(baseDir, 0o755))

		clayerFile := path.Join(baseDir, "layer-a.clayer.yml")
		clayerContent := `layer:
  description: Layer A description
  packs:
    - pack: Vendor::Pack@1.0.0
  components:
    - component: CORE
`
		require.NoError(os.WriteFile(clayerFile, []byte(clayerContent), 0o600))

		require.NoError(os.MkdirAll(path.Join(baseDir, "generated"), 0o755))
		require.NoError(os.WriteFile(path.Join(baseDir, "generated", "RTE_Components.h"), []byte("/* generated */"), 0o600))
		require.NoError(os.WriteFile(path.Join(baseDir, "generated", "Pre_Include_Global.h"), []byte("/* generated */"), 0o600))

		componentID := "Vendor::Pack:Comp@1.0.0"
		cbuild := maker.Cbuild{
			BaseDir:   baseDir,
			Languages: []string{"C", "CXX"},
		}
		cbuild.BuildDescType.Packs = []maker.Packs{
			{Pack: "Vendor::Pack@1.0.0", Path: "packs/vendor/pack"},
		}
		cbuild.BuildDescType.Components = []maker.Components{
			{
				Component:  componentID,
				SelectedBy: "CORE",
				FromPack:   "Vendor::Pack@1.0.0",
				Files: []maker.Files{
					{File: "packs/vendor/pack/src/main.c", Category: "sourceC"},
					{File: "packs/vendor/pack/include/main.h", Category: "headerC"},
				},
			},
		}
		cbuild.BuildDescType.Define = []interface{}{"DEF_GEN"}
		cbuild.BuildDescType.Misc = maker.Misc{C: []string{"-Wall"}}
		cbuild.BuildDescType.ConstructedFiles = []maker.Files{
			{File: "generated/RTE_Components.h"},
			{File: "generated/Pre_Include_Global.h"},
		}

		var m maker.Maker
		m.SolutionRoot = root
		m.SolutionName = "solution"
		m.CbuildIndex.BaseDir = baseDir
		m.CbuildIndex.BuildIdx.Cbuilds = []maker.Cbuilds{
			{
				Clayers: []maker.Clayers{
					{Clayer: "layer-a.clayer.yml"},
				},
			},
		}
		m.Cbuilds = []maker.Cbuild{cbuild}

		err := m.GenerateZephyrModules()
		assert.NoError(err)
		assert.True(m.ZephyrMaker.RteComponents)
		assert.True(m.ZephyrMaker.PreIncludeGlobal)
		assert.Len(m.ZephyrMaker.Layers, 1)
		assert.NotEmpty(m.ZephyrMaker.PackPaths["Vendor::Pack@1.0.0"])
		assert.Contains(m.ZephyrMaker.ComponentFiles, componentID)

		expectedOutputs := []string{
			path.Join(root, "solution", "zephyr", "module.yml"),
			path.Join(root, "solution", "Kconfig"),
			path.Join(root, "solution", "CMakeLists.txt"),
			path.Join(root, "solution", "sources.cmake"),
			path.Join(root, "solution", "RTE_Components.h"),
			path.Join(root, "solution", "Pre_Include_Global.h"),
		}
		for _, output := range expectedOutputs {
			_, statErr := os.Stat(output)
			assert.NoError(statErr)
		}

		sourcesContent, err := os.ReadFile(path.Join(root, "solution", "sources.cmake"))
		assert.NoError(err)
		assert.Contains(string(sourcesContent), "CONFIG_CMSIS_LAYER_A_VENDOR_PACK_COMP")
		assert.True(strings.Contains(string(sourcesContent), "\"${VENDOR_PACK}/src/main.c\""))
		assert.True(strings.Contains(string(sourcesContent), "\"${VENDOR_PACK}/include\""))
	})
}
