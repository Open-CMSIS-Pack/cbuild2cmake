/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWest(t *testing.T) {
	assert := assert.New(t)

	t.Run("test west toolchain environment", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/solutions/west/solution.cbuild-idx.yml"
		require.NoError(t, m.GenerateCMakeLists())
		m.SolutionTmpDir = filepath.ToSlash(t.TempDir())
		selectedVersion := m.SelectedToolchainVersion[0]
		registeredToolchain := m.RegisteredToolchains[selectedVersion]
		for _, testCase := range []struct {
			compiler string
			pathVar  string
			variant  string
		}{
			{"AC6", "ARMCLANG_TOOLCHAIN_PATH", "armclang"},
			{"GCC", "GNUARMEMB_TOOLCHAIN_PATH", "gnuarmemb"},
			{"IAR", "IAR_TOOLCHAIN_PATH", "iar"},
			{"CLANG", "LLVM_TOOLCHAIN_PATH", "host/llvm"},
		} {
			t.Run(testCase.compiler, func(t *testing.T) {
				toolchain := registeredToolchain
				toolchain.Name = testCase.compiler
				m.RegisteredToolchains[selectedVersion] = toolchain
				require.NoError(t, m.CreateWestCMakeLists(0))
				content, err := os.ReadFile(filepath.Join(m.SolutionTmpDir, m.Cbuilds[0].BuildDescType.Context, "CMakeLists.txt"))
				require.NoError(t, err)
				require.Contains(t, string(content), `"`+testCase.pathVar+`=${ZEPHYR_TOOLCHAIN_PATH}"`)
				require.Contains(t, string(content), `"ZEPHYR_TOOLCHAIN_VARIANT=`+testCase.variant+`"`)
			})
		}
	})

	t.Run("test west cmakelists creation", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/solutions/west/solution.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
		for index := range m.Cbuilds {
			toolchain := m.RegisteredToolchains[m.SelectedToolchainVersion[index]]
			version := m.SelectedToolchainVersion[index].String()
			for _, suffix := range []string{"", "@" + version, "@>=" + version} {
				m.Cbuilds[index].BuildDescType.Compiler = toolchain.Name + suffix
				err = m.CreateWestCMakeLists(index)
				assert.Nil(err)
				contextDir := filepath.Join(m.SolutionTmpDir, m.Cbuilds[index].BuildDescType.Context)
				content, err := os.ReadFile(filepath.Join(contextDir, "CMakeLists.txt"))
				assert.Nil(err)
				variant := maker.WestToolchainMap[toolchain.Name]
				assert.Contains(string(content), `"`+strings.ToUpper(strings.TrimPrefix(variant, "host/"))+`_TOOLCHAIN_PATH=${ZEPHYR_TOOLCHAIN_PATH}"`, "compiler: %s", toolchain.Name+suffix)
				assert.Contains(string(content), `"ZEPHYR_TOOLCHAIN_VARIANT=`+variant+`"`, "compiler: %s", toolchain.Name+suffix)
				content, err = os.ReadFile(filepath.Join(contextDir, "toolchain.cmake"))
				assert.Nil(err)
				assert.Contains(string(content), `set(REGISTERED_TOOLCHAIN_ROOT "`+toolchain.Path+`")`)
			}
		}
	})

	t.Run("test west definitions quoting", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/solutions/west/solution.cbuild-idx.yml"
		require.NoError(t, m.GenerateCMakeLists())
		m.SolutionTmpDir = filepath.ToSlash(t.TempDir())
		m.Cbuilds[0].BuildDescType.West.WestDefs = []interface{}{
			map[string]interface{}{"MESSAGE": "hello world"},
			map[string]interface{}{"QUOTED": `say "hello world"`},
			map[string]interface{}{"DIRECTORY": `C:\Program Files\SDK`},
			map[string]interface{}{"TAB": "hello\tworld"},
			map[string]interface{}{"ENABLED": true},
			map[string]interface{}{"COUNT": 2},
			"FLAG",
			map[string]interface{}{"CONF_FILE": "base.conf;extra config.conf"},
			map[string]interface{}{"REFERENCE": "${SOLUTION_ROOT}/config path"},
			map[string]interface{}{"UNDEFINED": "${WEST_DEFS_UNDEFINED_VARIABLE}"},
		}
		require.NoError(t, m.CreateWestCMakeLists(0))
		content, err := os.ReadFile(filepath.Join(m.SolutionTmpDir, m.Cbuilds[0].BuildDescType.Context, "CMakeLists.txt"))
		require.NoError(t, err)
		assert.Contains(string(content), `set(WEST_DEFS
  "-DMESSAGE=hello world"
  "-DQUOTED=say \"hello world\""
  "-DDIRECTORY=C:\\Program Files\\SDK"
  "-DTAB=hello`+"\t"+`world"
  "-DENABLED=true"
  "-DCOUNT=2"
  "-DFLAG"
`)
		assert.Contains(string(content), `"-DCONF_FILE=base.conf\;extra config.conf"`)
		assert.Contains(string(content), `"-DREFERENCE=${SOLUTION_ROOT}/config path"`)
		assert.Contains(string(content), `"-DUNDEFINED=${WEST_DEFS_UNDEFINED_VARIABLE}"`)
		assert.Regexp(`USES_TERMINAL\s+VERBATIM\s+\)`, string(content))
		assert.Contains(string(content), `"ARMCLANG_TOOLCHAIN_PATH=${ZEPHYR_TOOLCHAIN_PATH}"`)
		assert.Contains(string(content), `"ZEPHYR_TOOLCHAIN_VARIANT=armclang"`)

		t.Run("CMake command arguments", func(t *testing.T) {
			for _, tool := range []string{"cmake", "ninja"} {
				if _, err := exec.LookPath(tool); err != nil {
					t.Skipf("%s is required: %v", tool, err)
				}
			}
			source := t.TempDir()
			sourceRoot, err := os.OpenRoot(source)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, sourceRoot.Close())
			})
			cmakeContent := strings.ReplaceAll(string(content), `include("../roots.cmake")`, `set(SOLUTION_ROOT "${CMAKE_CURRENT_SOURCE_DIR}")`)
			cmakeContent = strings.ReplaceAll(cmakeContent, `include("toolchain.cmake")`, `set(REGISTERED_TOOLCHAIN_ROOT "${CMAKE_CURRENT_SOURCE_DIR}/SDK path/bin")`)
			lines := strings.Split(cmakeContent, "\n")
			for index, line := range lines {
				if strings.Contains(line, "COMMAND cmake -E env") {
					lines[index] = `  COMMAND "${CMAKE_COMMAND}" -E env ${ENV_VARS} "${CMAKE_COMMAND}" -P "${CMAKE_CURRENT_SOURCE_DIR}/check.cmake" -- ${WEST_DEFS}`
				}
			}
			require.NoError(t, sourceRoot.WriteFile("CMakeLists.txt", []byte(strings.Join(lines, "\n")), 0600))
			check := `cmake_minimum_required(VERSION 3.27)
if(NOT CMAKE_ARGC EQUAL 14
	OR NOT CMAKE_ARGV4 STREQUAL [[-DMESSAGE=hello world]]
	OR NOT CMAKE_ARGV5 STREQUAL [[-DQUOTED=say "hello world"]]
	OR NOT CMAKE_ARGV6 STREQUAL [[-DDIRECTORY=C:\Program Files\SDK]]
	OR NOT CMAKE_ARGV7 STREQUAL "-DTAB=hello\tworld"
	OR NOT CMAKE_ARGV8 STREQUAL [[-DENABLED=true]]
	OR NOT CMAKE_ARGV9 STREQUAL [[-DCOUNT=2]]
	OR NOT CMAKE_ARGV10 STREQUAL [[-DFLAG]]
	OR NOT CMAKE_ARGV11 STREQUAL [[-DCONF_FILE=base.conf;extra config.conf]]
	OR NOT CMAKE_ARGV12 STREQUAL "-DREFERENCE=${CMAKE_CURRENT_LIST_DIR}/config path"
	OR NOT CMAKE_ARGV13 STREQUAL [[-DUNDEFINED=]])
  message(FATAL_ERROR "West definitions changed across the command boundary")
endif()
if(NOT "$ENV{ZEPHYR_TOOLCHAIN_VARIANT}" STREQUAL "armclang"
	OR NOT "$ENV{ARMCLANG_TOOLCHAIN_PATH}" STREQUAL "${CMAKE_CURRENT_LIST_DIR}/SDK path/")
  message(FATAL_ERROR "Toolchain environment contains unexpected quotes or spaces")
endif()
`
			require.NoError(t, sourceRoot.WriteFile("check.cmake", []byte(check), 0600))
			build := filepath.Join(source, "build")
			output, err := exec.Command("cmake", "-G", "Ninja", "-S", source, "-B", build).CombinedOutput()
			require.NoError(t, err, "%s", output)
			output, err = exec.Command("cmake", "--build", build, "--target", "database").CombinedOutput()
			require.NoError(t, err, "%s", output)
		})
	})
}
