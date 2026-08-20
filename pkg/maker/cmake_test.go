/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"os"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNativeCMake(t *testing.T) {
	var m maker.Maker
	m.Params.InputFile = testRoot + "/data/solutions/cmake-support/solution.cbuild-idx.yml"
	require.NoError(t, m.GenerateCMakeLists())

	configured, err := os.ReadFile(testRoot + "/data/solutions/cmake-support/tmp/CM0/default/core0.Debug+CM0/CMakeLists.txt")
	require.NoError(t, err)
	assert.Contains(t, string(configured), `set(OUT_DIR "${SOLUTION_ROOT}/out/core0/CM0/Debug")`)
	assert.Contains(t, string(configured), `set(CMAKE_GENERATOR_NATIVE "Ninja Multi-Config")`)
	assert.Contains(t, string(configured), "set(CMAKE_CONFIGURE_OPTIONS\n  -DCMAKE_EXPORT_COMPILE_COMMANDS=ON\n  -DCMAKE_BUILD_TYPE=Debug\n)")
	assert.Contains(t, string(configured), `COMMAND ${CMAKE_COMMAND} -S "${CMAKE_SOURCE_DIR_NATIVE}" -B "${OUT_DIR}" -G "${CMAKE_GENERATOR_NATIVE}" ${CMAKE_CONFIGURE_OPTIONS}`)
	assert.Contains(t, string(configured), `COMMAND ${CMAKE_COMMAND} --build "${OUT_DIR}" --target "core0-target"`)

	minimal, err := os.ReadFile(testRoot + "/data/solutions/cmake-support/tmp/CM0/default/explicit-core1.Debug+CM0/CMakeLists.txt")
	require.NoError(t, err)
	assert.Contains(t, string(minimal), `set(CMAKE_GENERATOR_NATIVE "Ninja")`)
	assert.Contains(t, string(minimal), `COMMAND ${CMAKE_COMMAND} --build "${OUT_DIR}"`)
	assert.NotContains(t, string(minimal), `--target`)

	super, err := os.ReadFile(testRoot + "/data/solutions/cmake-support/tmp/CM0/default/CMakeLists.txt")
	require.NoError(t, err)
	assert.Contains(t, string(super), "if(NATIVE_CMAKE_CONTEXT)\n    set(NATIVE_CMAKE_TARGET \"--target cmake\")\n  else()\n    set(NATIVE_CMAKE_TARGET \"\")\n  endif()")
	assert.Contains(t, string(super), "set(OUTPUTS_1\n  \"${SOLUTION_ROOT}/out/core0/CM0/Debug/build/core0.elf\"\n)")
}
