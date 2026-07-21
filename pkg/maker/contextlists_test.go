/*
 * Copyright (c) 2026 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
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
		assert.Contains(options, "set(CPP_OPTIONS_C \"-xc\" \"-c-option\" \"-common-option\")")
		assert.Contains(options, "set(CPP_OPTIONS_CXX \"-xc++\" \"-cxx-option\" \"-common-option\")")
		assert.Equal("\nset(COMPILE_MACROS_C ${OUT_DIR}/compile_macros_c.h)\nset(COMPILE_MACROS_CXX ${OUT_DIR}/compile_macros_cxx.h)", macros)
		assert.Equal(" ${COMPILE_MACROS_C} ${COMPILE_MACROS_CXX}", dependencies)
		assert.Equal("add_custom_command(OUTPUT ${COMPILE_MACROS_C} ${COMPILE_MACROS_CXX}\n  COMMAND ${CPP} ${CPP_OPTIONS_C} ${CPP_DUMP_MACROS} \"${COMPILE_MACROS_C}\"\n  COMMAND ${CPP} ${CPP_OPTIONS_CXX} ${CPP_DUMP_MACROS} \"${COMPILE_MACROS_CXX}\"\n)", commands)
		assert.NotContains(options+macros+dependencies+commands, "ASM")
	})

	t.Run("test preprocessor options with spaces", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.Languages = []string{"C"}
		cbuild.BuildDescType.Misc = maker.Misc{
			C: []string{`-DTEST=1 -include "path with spaces/header.h" -Wall`},
		}

		options, _, _, _ := cbuild.PreprocessorOptions()
		assert.Contains(options, `set(CPP_OPTIONS_C "-xc" "-DTEST=1" "-include" "\"path with spaces/header.h\"" "-Wall")`)
	})

	t.Run("test IAR preprocessor options", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.Toolchain = "IAR"
		cbuild.Languages = []string{"C", "CXX"}
		cbuild.BuildDescType.Misc = maker.Misc{
			C:   []string{"--c-option"},
			CPP: []string{"--cxx-option"},
		}

		options, _, _, _ := cbuild.PreprocessorOptions()
		assert.Contains(options, "set(CPP_OPTIONS_C \"--c-option\")")
		assert.Contains(options, "set(CPP_OPTIONS_CXX \"--c++\" \"--cxx-option\")")
		assert.NotContains(options, "-xc")
	})
}
