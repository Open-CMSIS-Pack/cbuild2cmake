/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
)

func TestBuildContent(t *testing.T) {
	assert := assert.New(t)

	t.Run("test get file language", func(t *testing.T) {
		var file maker.Files
		assert.Empty(file.Language)
		assert.Empty(file.Category)
		assert.Equal("ALL", maker.GetLanguage(file))
		file.Language = "asm"
		assert.Equal("ASM", maker.GetLanguage(file))
		file.Language = "c"
		assert.Equal("C", maker.GetLanguage(file))
		file.Category = "headerAsm"
		assert.Equal("ASM", maker.GetLanguage(file))
		file.Category = "headerC"
		assert.Equal("C", maker.GetLanguage(file))
	})

	t.Run("test get scope", func(t *testing.T) {
		var file maker.Files
		assert.Empty(file.Scope)
		assert.Equal("PUBLIC", maker.GetScope(file))
		file.Scope = "public"
		assert.Equal("PUBLIC", maker.GetScope(file))
		file.Scope = "private"
		assert.Equal("PRIVATE", maker.GetScope(file))
		file.Scope = "hidden"
		assert.Equal("PRIVATE", maker.GetScope(file))
	})

	t.Run("test classify files", func(t *testing.T) {
		var files []maker.Files
		var cbuild maker.Cbuild
		cbuild.ContextRoot = "project"
		header := maker.Files{
			File:     "./headers/header.h",
			Category: "header",
		}
		files = append(files, header)
		buildFiles := cbuild.ClassifyFiles(files)
		assert.True(buildFiles.Interface)
		assert.Equal("${SOLUTION_ROOT}/project/headers", buildFiles.Include["INTERFACE"]["ALL"][0])

		include := maker.Files{
			File:     "./includes",
			Category: "include",
		}
		files = append(files, include)
		buildFiles = cbuild.ClassifyFiles(files)
		assert.True(buildFiles.Interface)
		assert.Equal("${SOLUTION_ROOT}/project/includes", buildFiles.Include["INTERFACE"]["ALL"][1])

		source := maker.Files{
			File:     "./source.c",
			Category: "source",
		}
		files = append(files, source)
		library := maker.Files{
			File:     "./lib.a",
			Category: "library",
		}
		files = append(files, library)
		object := maker.Files{
			File:     "./obj.o",
			Category: "object",
		}
		files = append(files, object)
		buildFiles = cbuild.ClassifyFiles(files)
		assert.False(buildFiles.Interface)
		assert.Equal("${SOLUTION_ROOT}/project/headers", buildFiles.Include["PUBLIC"]["ALL"][0])
		assert.Equal("${SOLUTION_ROOT}/project/includes", buildFiles.Include["PUBLIC"]["ALL"][1])
		assert.Equal("${SOLUTION_ROOT}/project/source.c", buildFiles.Source["ALL"][0])
		assert.Equal("${SOLUTION_ROOT}/project/lib.a", buildFiles.Library[0])
		assert.Equal("${SOLUTION_ROOT}/project/obj.o", buildFiles.Object[0])
	})

	t.Run("test cmake target include directories from files", func(t *testing.T) {
		var files = []maker.Files{
			{
				File:     "./includes",
				Category: "include",
			},
			{
				File:     "./includes-c",
				Category: "includeC",
			},
		}
		var cbuild maker.Cbuild
		buildFiles := cbuild.ClassifyFiles(files)
		content := maker.CMakeTargetIncludeDirectoriesFromFiles("TARGET", buildFiles)
		assert.Contains(content, "includes")
		assert.Contains(content, "includes-c")
	})

	t.Run("test cmake target compile options", func(t *testing.T) {
		var misc = maker.Misc{
			ASM:  []string{"-asm-flag"},
			C:    []string{"-c-flag"},
			CPP:  []string{"-cpp-flag"},
			CCPP: []string{"-c-cpp-flag"},
		}
		content := maker.CMakeTargetCompileOptions("TARGET", "PUBLIC", misc)
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:ASM>:\n    -asm-flag")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:C>:\n    -c-flag")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:CXX>:\n    -cpp-flag")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:C,CXX>:\n    -c-cpp-flag")
	})

	t.Run("test language specific compile options", func(t *testing.T) {
		var misc = maker.Misc{
			ASM: []string{"-asm-flag"},
		}
		content := maker.LangugeSpecificCompileOptions("ASM", misc.ASM)
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:ASM>:\n    -asm-flag")
	})

	t.Run("test get file misc", func(t *testing.T) {
		var files = []maker.Files{
			{
				Category: "sourceAsm",
				Misc: maker.Misc{
					ASM: []string{"-asm-flag"},
				},
			},
			{
				Category: "sourceC",
				Misc: maker.Misc{
					C: []string{"-c-flag"},
				},
			},
			{
				Category: "sourceCpp",
				Misc: maker.Misc{
					CPP: []string{"-cpp-flag"},
				},
			},
		}
		content := maker.GetFileMisc(files[0], ";")
		assert.Contains(content, "-asm-flag")
		content = maker.GetFileMisc(files[1], ";")
		assert.Contains(content, "-c-flag")
		content = maker.GetFileMisc(files[2], ";")
		assert.Contains(content, "-cpp-flag")
	})

	t.Run("test get output files info for secure executable", func(t *testing.T) {
		var output = []maker.Output{
			{
				File: "./arfifact.elf",
				Type: "elf",
			},
			{
				File: "./binary.bin",
				Type: "bin",
			},
			{
				File: "./hexadecimal.hex",
				Type: "hex",
			},
			{
				File: "./secure.lib",
				Type: "cmse-lib",
			},
		}
		outputByProducts, outputFile, outputType, customCommands := maker.OutputFiles(output)
		assert.Equal(outputFile, "./arfifact.elf")
		assert.Equal(outputType, "elf")
		assert.Contains(outputByProducts, "binary.bin")
		assert.Contains(outputByProducts, "hexadecimal.hex")
		assert.Contains(outputByProducts, "secure.lib")
		assert.Contains(customCommands, "${ELF2BIN}")
		assert.Contains(customCommands, "${ELF2HEX}")
	})

	t.Run("test get output files info for library", func(t *testing.T) {
		var output = []maker.Output{
			{
				File: "./library.a",
				Type: "lib",
			},
		}
		_, outputFile, outputType, _ := maker.OutputFiles(output)
		assert.Equal(outputFile, "./library.a")
		assert.Equal(outputType, "lib")
	})
}
