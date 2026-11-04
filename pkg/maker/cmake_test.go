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
	"runtime"
	"strings"
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCMakeVersionWarning(t *testing.T) {
	for _, imageOnly := range []bool{false, true} {
		name := "super"
		if imageOnly {
			name = "image-only"
		}
		t.Run(name, func(t *testing.T) {
			m := maker.Maker{SolutionTmpDir: t.TempDir(), SolutionName: "version-warning"}
			if imageOnly {
				require.NoError(t, m.CreateCMakeListsImageOnly())
			} else {
				require.NoError(t, m.CreateSuperCMakeLists())
			}
			content, err := os.ReadFile(filepath.Join(m.SolutionTmpDir, "CMakeLists.txt"))
			require.NoError(t, err)
			header, _, found := strings.Cut(string(content), "endif()")
			require.True(t, found)
			require.Contains(t, header, "cmake_minimum_required(VERSION 3.27)")
			require.Contains(t, header, `if(CMAKE_VERSION VERSION_LESS "4.2.1")`)

			cmake, err := exec.LookPath("cmake")
			if err != nil {
				t.Skip("CMake is required to execute the generated version check")
			}
			for _, version := range []string{"3.27.0", "4.2.0", "4.2.1", "4.3.0"} {
				t.Run(version, func(t *testing.T) {
					scriptDir := t.TempDir()
					root, err := os.OpenRoot(scriptDir)
					require.NoError(t, err)
					err = root.WriteFile("version.cmake", []byte("set(CMAKE_VERSION "+version+")\n"+header+"endif()\n"), 0600)
					require.NoError(t, root.Close())
					require.NoError(t, err)
					command := exec.Command(cmake, "-P", "version.cmake")
					command.Dir = scriptDir
					output, err := command.CombinedOutput()
					require.NoError(t, err, string(output))
					if version == "3.27.0" || version == "4.2.0" {
						assert.Contains(t, string(output), "CMake Warning")
						assert.Contains(t, string(output), "CMake 4.2.1 or newer is recommended.")
						assert.Contains(t, string(output), "You are using CMake "+version)
					} else {
						assert.Empty(t, string(output))
					}
				})
			}
		})
	}
}

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
	assert.Contains(t, string(super), "if(NATIVE_CMAKE_CONTEXT)\n    set(NATIVE_CMAKE_TARGET \"--target cmake\")\n  endif()")
	assert.NotContains(t, string(super), "set(NATIVE_CMAKE_TARGET \"\")")
	assert.Contains(t, string(super), "set(OUTPUTS_1\n  \"${SOLUTION_ROOT}/out/core0/CM0/Debug/build/core0.elf\"\n)")
}

func TestSuperCMakeListsRejectsCbuildOnDifferentVolume(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("filepath.Rel only rejects different volumes on Windows")
	}

	solutionRoot := t.TempDir()
	solutionVolume := filepath.VolumeName(solutionRoot)
	otherVolume := "C:"
	if solutionVolume == otherVolume {
		otherVolume = "D:"
	}

	version := semver.MustParse("1.0.0")
	m := maker.Maker{}
	m.SolutionRoot = solutionRoot
	m.SolutionTmpDir = t.TempDir()
	m.Cbuilds = []maker.Cbuild{{BaseDir: filepath.Join(otherVolume+`\`, "context")}}
	m.SelectedToolchainVersion = []*semver.Version{version}
	m.RegisteredToolchains = map[*semver.Version]maker.Toolchain{
		version: {Name: "GCC"},
	}

	err := m.CreateSuperCMakeLists()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Rel:")
}
