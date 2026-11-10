/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextLists(t *testing.T) {
	assert := assert.New(t)

	t.Run("test context cmakelists creation", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/solutions/build-cpp/solution.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
		for index := range m.Cbuilds {
			err = m.CreateContextCMakeLists(index)
			assert.Nil(err)
		}
	})

	t.Run("test preprocessor options", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.Languages = []string{"ASM", "C", "CXX"}
		cbuild.BuildDescType.Misc = maker.Misc{
			C:    []string{"-c-option"},
			CPP:  []string{"-cxx-option"},
			CCPP: []string{"-common-option"},
		}

		options, macros, dependencies, commands := cbuild.PreprocessorOptions()
		assert.Contains(options, "set(CPP_OPTIONS_C \"-c-option\" \"-common-option\")")
		assert.Contains(options, "set(CPP_OPTIONS_CXX \"-cxx-option\" \"-common-option\")")
		assert.Equal("\nset(COMPILE_MACROS_C ${OUT_DIR}/compile_macros_c.h)\nset(COMPILE_MACROS_CXX ${OUT_DIR}/compile_macros_cxx.h)", macros)
		assert.Equal(" ${COMPILE_MACROS_C} ${COMPILE_MACROS_CXX}", dependencies)
		assert.Equal("add_custom_command(OUTPUT ${COMPILE_MACROS_C} ${COMPILE_MACROS_CXX}\n  COMMAND ${CPP} -xc ${CPP_OPTIONS_C} ${CPP_DUMP_MACROS} \"${COMPILE_MACROS_C}\"\n  COMMAND ${CPP} -xc++ ${CPP_OPTIONS_CXX} ${CPP_DUMP_MACROS} \"${COMPILE_MACROS_CXX}\"\n)", commands)
		assert.NotContains(options+macros+dependencies+commands, "ASM")
	})

	t.Run("test preprocessor options with spaces", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.Languages = []string{"C"}
		cbuild.BuildDescType.Misc = maker.Misc{
			C: []string{`-DTEST=1 -include "path with spaces/header.h" -Wall`},
		}

		options, _, _, _ := cbuild.PreprocessorOptions()
		assert.Contains(options, `set(CPP_OPTIONS_C "-DTEST=1" "-include" "\"path with spaces/header.h\"" "-Wall")`)
	})

	t.Run("test IAR preprocessor options", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.Toolchain = "IAR"
		cbuild.Languages = []string{"C", "CXX"}
		cbuild.BuildDescType.Misc = maker.Misc{
			C:   []string{"--c-option"},
			CPP: []string{"--cxx-option"},
		}

		options, _, _, commands := cbuild.PreprocessorOptions()
		assert.Contains(options, "set(CPP_OPTIONS_C \"--c-option\")")
		assert.Contains(options, "set(CPP_OPTIONS_CXX \"--cxx-option\")")
		assert.Contains(commands, "COMMAND ${CPP} ${CPP_OPTIONS_C}")
		assert.Contains(commands, "COMMAND ${CPP} --c++ ${CPP_OPTIONS_CXX}")
		assert.NotContains(options, "-xc")
	})
}

func TestCompilerDiscoveryOptions(t *testing.T) {
	cbuild := maker.Cbuild{Languages: []string{"ASM", "C", "CXX"}}
	cbuild.BuildDescType.Misc = maker.Misc{
		C:    []string{`--sysroot="path with spaces"`},
		CPP:  []string{"--config=llvmlibc.cfg"},
		CCPP: []string{"-ffunction-sections"},
	}
	setup, restore := cbuild.CompilerDiscoveryOptions()
	assert.Equal(t, "\nstring(JOIN \" \" CMAKE_C_FLAGS ${CPP_OPTIONS_C})\nstring(JOIN \" \" CMAKE_CXX_FLAGS ${CPP_OPTIONS_CXX})", setup)
	assert.Equal(t, "\nunset(CMAKE_C_FLAGS)\nunset(CMAKE_CXX_FLAGS)", restore)
	assert.NotContains(t, setup+restore, "CACHE")
	assert.NotContains(t, setup+restore, "ASM")
	assert.NotContains(t, setup, "-xc")
	cbuild.BuildDescType.Misc = maker.Misc{}
	setup, restore = cbuild.CompilerDiscoveryOptions()
	assert.Empty(t, setup)
	assert.Empty(t, restore)
}

func TestCompilerDiscoveryConfigure(t *testing.T) {
	for _, tool := range []string{"cmake", "ninja", "gcc", "g++"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is required for the compiler discovery configure test")
		}
	}
	directory := t.TempDir()
	include := filepath.Join(directory, "discovery include")
	require.NoError(t, os.Mkdir(include, 0700))
	cbuild := maker.Cbuild{Languages: []string{"C", "CXX"}}
	cbuild.BuildDescType.Misc.CCPP = []string{`-isystem "` + filepath.ToSlash(include) + `" -DDISCOVERY_MISC=1`}
	setup, restore := cbuild.CompilerDiscoveryOptions()
	options, _, _, _ := cbuild.PreprocessorOptions()
	content := `cmake_minimum_required(VERSION 3.27)
set(CMAKE_SYSTEM_NAME Generic)
set(CMAKE_TRY_COMPILE_TARGET_TYPE STATIC_LIBRARY)
set(CMAKE_C_COMPILER gcc)
set(CMAKE_CXX_COMPILER g++)
` + options + setup + `
project(discovery LANGUAGES C CXX)
` + restore + `
foreach(language C CXX)
  if(CMAKE_${language}_FLAGS MATCHES "DISCOVERY_MISC")
    message(FATAL_ERROR "Discovery flags leaked")
  endif()
  if(NOT CMAKE_${language}_IMPLICIT_INCLUDE_DIRECTORIES MATCHES "/discovery include(;|$)")
    message(FATAL_ERROR "Discovery include directory missing for ${language}: ${CMAKE_${language}_IMPLICIT_INCLUDE_DIRECTORIES}")
  endif()
endforeach()
`
	require.NoError(t, os.WriteFile(filepath.Join(directory, "CMakeLists.txt"), []byte(content), 0600))
	output, err := exec.Command("cmake", "-S", directory, "-B", filepath.Join(directory, "build"), "-G", "Ninja").CombinedOutput()
	require.NoError(t, err, string(output))
	cache, err := os.ReadFile(filepath.Join(directory, "build", "CMakeCache.txt"))
	require.NoError(t, err)
	assert.NotContains(t, string(cache), "DISCOVERY_MISC")
}
