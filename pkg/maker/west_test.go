/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"os"
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
				require.Contains(t, string(content), testCase.pathVar+`="${ZEPHYR_TOOLCHAIN_PATH}"`)
				require.Contains(t, string(content), `ZEPHYR_TOOLCHAIN_VARIANT="`+testCase.variant+`"`)
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
				pathVar := strings.ToUpper(strings.TrimPrefix(variant, "host/")) + `_TOOLCHAIN_PATH="${ZEPHYR_TOOLCHAIN_PATH}"`
				assert.Contains(string(content), pathVar, "compiler: %s", toolchain.Name+suffix)
				assert.Contains(string(content), `ZEPHYR_TOOLCHAIN_VARIANT="`+variant+`"`, "compiler: %s", toolchain.Name+suffix)
				content, err = os.ReadFile(filepath.Join(contextDir, "toolchain.cmake"))
				assert.Nil(err)
				assert.Contains(string(content), `set(REGISTERED_TOOLCHAIN_ROOT "`+toolchain.Path+`")`)
			}
		}
	})
}
