/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"os"
	"path/filepath"
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
				m.Cbuilds[0].BuildDescType.Compiler = testCase.compiler
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
			err = m.CreateWestCMakeLists(index)
			assert.Nil(err)
		}
	})
}
