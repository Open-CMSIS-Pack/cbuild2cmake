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
)

func TestWest(t *testing.T) {
	assert := assert.New(t)

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
				assert.Contains(string(content), strings.ToUpper(variant)+`_TOOLCHAIN_PATH="${ZEPHYR_TOOLCHAIN_PATH}"`, "compiler: %s", toolchain.Name+suffix)
				assert.Contains(string(content), `ZEPHYR_TOOLCHAIN_VARIANT="`+variant+`"`, "compiler: %s", toolchain.Name+suffix)
				content, err = os.ReadFile(filepath.Join(contextDir, "toolchain.cmake"))
				assert.Nil(err)
				assert.Contains(string(content), `set(REGISTERED_TOOLCHAIN_ROOT "`+toolchain.Path+`")`)
			}
		}
	})
}
