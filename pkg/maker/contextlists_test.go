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
		assert.Contains(options, "set(CPP_OPTIONS_C ${CC_FLAGS} \"-c-option\" \"-common-option\")")
		assert.Contains(options, "set(CPP_OPTIONS_CXX ${CXX_FLAGS} \"-cxx-option\" \"-common-option\")")
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
		assert.Contains(options, `set(CPP_OPTIONS_C ${CC_FLAGS} "-DTEST=1" "-include" "\"path with spaces/header.h\"" "-Wall")`)
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
		assert.Contains(options, "set(CPP_OPTIONS_C ${CC_FLAGS} \"--c-option\")")
		assert.Contains(options, "set(CPP_OPTIONS_CXX ${CXX_FLAGS} \"--cxx-option\")")
		assert.Contains(commands, "COMMAND ${CPP} ${CPP_OPTIONS_C}")
		assert.Contains(commands, "COMMAND ${CPP} --c++ ${CPP_OPTIONS_CXX}")
		assert.NotContains(options, "-xc")
	})
}

func TestPreprocessorCPUArguments(t *testing.T) {
	cmake, err := exec.LookPath("cmake")
	if err != nil {
		t.Skip("CMake is required to check preprocessor argument boundaries")
	}
	for _, test := range []struct {
		toolchain string
		cpu       string
		expected  string
	}{
		{"IAR", "--cpu Cortex-M0 --endian little", "--cpu;Cortex-M0;--endian;little"},
		{"AC6", "--target=arm-arm-none-eabi -mcpu=cortex-m0", "--target=arm-arm-none-eabi;-mcpu=cortex-m0"},
		{"GCC", "-mcpu=cortex-m0 -mthumb", "-mcpu=cortex-m0;-mthumb"},
		{"CLANG", "--target=arm-none-eabi -mcpu=cortex-m0", "--target=arm-none-eabi;-mcpu=cortex-m0"},
		{"empty", "", ""},
	} {
		t.Run(test.toolchain, func(t *testing.T) {
			cbuild := maker.Cbuild{Toolchain: test.toolchain, Languages: []string{"C", "CXX"}}
			cbuild.BuildDescType.Misc.CCPP = []string{"-e --dlib_config DLib_Config_Full.h"}
			options, macros, _, commands := cbuild.PreprocessorOptions()
			content := `cmake_minimum_required(VERSION 3.27)
set(CC_CPU "` + test.cpu + `")
set(CXX_CPU "` + test.cpu + `")
set(CC_FLAGS -DFLAG=1 "-Ipath with spaces")
set(CXX_FLAGS ${CC_FLAGS})
` + options + cbuild.CompilerDiscoveryOptions() + `
set(expected -DFLAG=1 "-Ipath with spaces" -e --dlib_config DLib_Config_Full.h)
string(JOIN " " expected_discovery "${CC_CPU}" ${expected})
foreach(language C CXX)
	if(NOT "${CPP_OPTIONS_${language}}" STREQUAL "${expected}")
		message(FATAL_ERROR "Incorrect ${language} arguments: ${CPP_OPTIONS_${language}}; expected: ${expected}")
	endif()
	if(NOT "${CMAKE_${language}_FLAGS}" STREQUAL "${expected_discovery}")
		message(FATAL_ERROR "Incorrect ${language} discovery flags: ${CMAKE_${language}_FLAGS}")
	endif()
endforeach()
function(add_custom_command)
	set(cpu_arguments "` + test.expected + `")
	if(cpu_arguments)
		list(GET cpu_arguments 0 cpu_option)
		set(count 0)
		foreach(argument IN LISTS ARGN)
			if(argument STREQUAL cpu_option)
				math(EXPR count "${count} + 1")
			endif()
		endforeach()
		if(NOT count EQUAL 2)
			message(FATAL_ERROR "Expected one CPU option per language, found ${count}: ${ARGN}")
		endif()
	endif()
endfunction()
set(CPP compiler)
set(CPP_DUMP_MACROS "` + test.expected + `")
list(APPEND CPP_DUMP_MACROS --dump)
` + macros + "\n" + commands + `
`
			script := filepath.Join(t.TempDir(), "arguments.cmake")
			require.NoError(t, os.WriteFile(script, []byte(content), 0600))
			output, err := exec.Command(cmake, "-P", script).CombinedOutput()
			require.NoError(t, err, string(output))
		})
	}
}

func TestCompilerDiscoveryOptions(t *testing.T) {
	cbuild := maker.Cbuild{Languages: []string{"ASM", "C", "CXX"}}
	cbuild.BuildDescType.Misc = maker.Misc{
		C:    []string{`--sysroot="path with spaces"`},
		CPP:  []string{"--config=llvmlibc.cfg"},
		CCPP: []string{"-ffunction-sections"},
	}
	setup := cbuild.CompilerDiscoveryOptions()
	assert.Equal(t, "\nstring(JOIN \" \" CMAKE_C_FLAGS \"${CC_CPU}\" ${CPP_OPTIONS_C})\nstring(JOIN \" \" CMAKE_CXX_FLAGS \"${CXX_CPU}\" ${CPP_OPTIONS_CXX})", setup)
	cbuild.BuildDescType.Misc = maker.Misc{}
	emptySetup := cbuild.CompilerDiscoveryOptions()
	assert.Equal(t, setup, emptySetup)
}

func TestCompilerDiscoveryConfigure(t *testing.T) {
	for _, tool := range []string{"cmake", "ninja", "gcc", "g++"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is required for the compiler discovery configure test")
		}
	}
	for _, binding := range []string{"cache", "normal"} {
		t.Run(binding, func(t *testing.T) {
			directory := t.TempDir()
			include := filepath.Join(directory, "discovery include")
			require.NoError(t, os.Mkdir(include, 0700))
			baseInclude := filepath.Join(directory, "base include")
			require.NoError(t, os.Mkdir(baseInclude, 0700))
			baseFlags := `-isystem "` + filepath.ToSlash(baseInclude) + `" -DBASE_FLAGS=1`
			var initialFlags string
			for _, language := range []string{"C", "CXX"} {
				flags := "CMAKE_" + language + "_FLAGS"
				initialFlags += "\nset(" + flags + " [=[" + baseFlags + "]=]"
				if binding == "cache" {
					initialFlags += ` CACHE STRING "Base flags"`
				}
				initialFlags += ")"
			}
			cbuild := maker.Cbuild{Languages: []string{"C", "CXX"}}
			cbuild.BuildDescType.Misc.CCPP = []string{`-isystem "` + filepath.ToSlash(include) + `" -DDISCOVERY_MISC=1`}
			setup := cbuild.CompilerDiscoveryOptions()
			options, _, _, _ := cbuild.PreprocessorOptions()
			content := `cmake_minimum_required(VERSION 3.27)
set(CMAKE_SYSTEM_NAME Generic)
set(CMAKE_TRY_COMPILE_TARGET_TYPE STATIC_LIBRARY)
set(CMAKE_C_COMPILER gcc)
set(CMAKE_CXX_COMPILER g++)
` + initialFlags + `
set(CC_CPU -DPROCESSOR_OPTIONS=1)
set(CXX_CPU -DPROCESSOR_OPTIONS=1)
set(CC_FLAGS -DTOOLCHAIN_OPTIONS=1 -DTOOLCHAIN_EXTRA=1)
set(CXX_FLAGS -DTOOLCHAIN_OPTIONS=1 -DTOOLCHAIN_EXTRA=1)
` + options + setup + `
project(discovery LANGUAGES C CXX)
foreach(language C CXX)
	if(NOT CMAKE_${language}_FLAGS MATCHES "DISCOVERY_MISC")
		message(FATAL_ERROR "Shared flags missing")
  endif()
	if(CMAKE_${language}_FLAGS MATCHES "BASE_FLAGS")
		message(FATAL_ERROR "External flags unexpectedly included")
	endif()
	if(NOT CMAKE_${language}_FLAGS MATCHES "PROCESSOR_OPTIONS")
		message(FATAL_ERROR "Processor options missing")
	endif()
	string(REGEX MATCHALL "DISCOVERY_MISC" occurrences "${CMAKE_${language}_FLAGS}")
	list(LENGTH occurrences count)
	if(NOT count EQUAL 1)
		message(FATAL_ERROR "Shared flags duplicated")
	endif()
	foreach(flag TOOLCHAIN_OPTIONS TOOLCHAIN_EXTRA)
		string(REGEX MATCHALL "${flag}" occurrences "${CMAKE_${language}_FLAGS}")
		list(LENGTH occurrences count)
		if(NOT count EQUAL 1)
			message(FATAL_ERROR "Toolchain flag ${flag} missing or duplicated for ${language}")
		endif()
	endforeach()
	if(CMAKE_${language}_IMPLICIT_INCLUDE_DIRECTORIES MATCHES "/base include(;|$)")
		message(FATAL_ERROR "External include directory unexpectedly included")
	endif()
  if(NOT CMAKE_${language}_IMPLICIT_INCLUDE_DIRECTORIES MATCHES "/discovery include(;|$)")
    message(FATAL_ERROR "Discovery include directory missing for ${language}: ${CMAKE_${language}_IMPLICIT_INCLUDE_DIRECTORIES}")
  endif()
endforeach()
`
			require.NoError(t, os.WriteFile(filepath.Join(directory, "CMakeLists.txt"), []byte(content), 0600))
			for configure := 0; configure < 2; configure++ {
				output, err := exec.Command("cmake", "-S", directory, "-B", filepath.Join(directory, "build"), "-G", "Ninja").CombinedOutput()
				require.NoError(t, err, string(output))
			}
			cache, err := os.ReadFile(filepath.Join(directory, "build", "CMakeCache.txt"))
			require.NoError(t, err)
			assert.NotContains(t, string(cache), "DISCOVERY_MISC")
		})
	}
}
