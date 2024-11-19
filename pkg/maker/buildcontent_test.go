/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"path/filepath"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
)

func TestReplaceDelimiters(t *testing.T) {
	assert := assert.New(t)

	t.Run("test replace delimiters", func(t *testing.T) {
		assert.Equal("Cvendor_Cbundle_Cclass_Cgroup_Cvariant_Cversion", maker.ReplaceDelimiters("Cvendor&Cbundle::Cclass:Cgroup&Cvariant@Cversion"))
		assert.Equal("ARM_CMSIS_CORE_A", maker.ReplaceDelimiters("ARM::CMSIS.CORE A"))
		assert.Equal("AC6_6_16_0", maker.ReplaceDelimiters("AC6@>=6.16.0"))
		assert.Equal("path_with_spaces", maker.ReplaceDelimiters("path/with spaces"))
		assert.Equal("Handlers__GCC_", maker.ReplaceDelimiters("Handlers (GCC)"))
	})
}

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
		preIncludeLocal := maker.Files{
			File:     "./RTE/class/pre-include.h",
			Category: "preIncludeLocal",
		}
		files = append(files, preIncludeLocal)
		preIncludeGlobal := maker.Files{
			File:     "${CMSIS_PACK_ROOT}/vendor/pack/inc/pre-include.h",
			Category: "preIncludeGlobal",
		}
		files = append(files, preIncludeGlobal)
		template := maker.Files{
			File:     "./template.c",
			Category: "source",
			Attr:     "template",
		}
		files = append(files, template)
		config := maker.Files{
			File:     "./configs/config.h",
			Category: "header",
			Attr:     "config",
		}
		files = append(files, config)
		buildFiles = cbuild.ClassifyFiles(files)
		assert.False(buildFiles.Interface)
		assert.Equal("${SOLUTION_ROOT}/project/configs", buildFiles.Include["PUBLIC"]["ALL"][0])
		assert.Equal("${SOLUTION_ROOT}/project/headers", buildFiles.Include["PUBLIC"]["ALL"][1])
		assert.Equal("${SOLUTION_ROOT}/project/includes", buildFiles.Include["PUBLIC"]["ALL"][2])
		assert.Equal("${SOLUTION_ROOT}/project/source.c", buildFiles.Source["C"][0])
		assert.Equal("${SOLUTION_ROOT}/project/lib.a", buildFiles.Library[0])
		assert.Equal("${SOLUTION_ROOT}/project/obj.o", buildFiles.Object[0])
		assert.Equal("${SOLUTION_ROOT}/project/RTE/class/pre-include.h", buildFiles.PreIncludeLocal[0])
		assert.Equal("${CMSIS_PACK_ROOT}/vendor/pack/inc/pre-include.h", cbuild.PreIncludeGlobal[0])
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
		content := maker.CMakeTargetIncludeDirectories("TARGET", buildFiles.Include)
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
		var cbuild maker.Cbuild
		cbuild.Languages = []string{"ASM", "C", "CXX"}
		preIncludes := []string{"${SOLUTION_ROOT}/project/RTE/class/pre-include.h"}
		content := cbuild.CMakeTargetCompileOptions("TARGET", "PUBLIC", misc, preIncludes, "${CONTEXT}")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:ASM>:\n    \"SHELL:-asm-flag\"")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:C>:\n    \"SHELL:-c-flag\"\n    \"SHELL:-c-cpp-flag\"")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:CXX>:\n    \"SHELL:-cpp-flag\"\n    \"SHELL:-c-cpp-flag\"")
		assert.Contains(content, "\"SHELL:${_PI}\\\"${SOLUTION_ROOT}/project/RTE/class/pre-include.h\\\"\"")
	})

	t.Run("test language specific compile options", func(t *testing.T) {
		var misc = maker.Misc{
			ASM: []string{"-asm-flag"},
		}
		var cbuild maker.Cbuild
		content := cbuild.LanguageSpecificCompileOptions("ASM", misc.ASM...)
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:ASM>:\n    \"SHELL:-asm-flag\"")
	})

	t.Run("test add context language", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.AddContextLanguage("ALL")
		assert.Empty(cbuild.Languages)
		cbuild.AddContextLanguage("C")
		assert.Equal("C", cbuild.Languages[0])
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
		content := maker.GetFileOptions(files[0], false, ";")
		assert.Contains(content, "-asm-flag")
		content = maker.GetFileOptions(files[1], false, ";")
		assert.Contains(content, "-c-flag")
		content = maker.GetFileOptions(files[2], false, ";")
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

	t.Run("test inherit compile abstractions", func(t *testing.T) {
		var parent = maker.CompilerAbstractions{
			Debug:       "on",
			Optimize:    "speed",
			Warnings:    "all",
			LanguageC:   "c90",
			LanguageCpp: "c++98",
		}
		var child = maker.CompilerAbstractions{
			Debug:       "off",
			Optimize:    "size",
			Warnings:    "off",
			LanguageC:   "c11",
			LanguageCpp: "c++11",
		}
		inherited := maker.InheritCompilerAbstractions(parent, child)
		assert.Equal(inherited, child)

		var emptyChild = maker.CompilerAbstractions{}
		inherited = maker.InheritCompilerAbstractions(parent, emptyChild)
		assert.Equal(inherited, parent)
	})

	t.Run("test abstractions empty", func(t *testing.T) {
		areAbstractionsEmpty := maker.AreAbstractionsEmpty(maker.CompilerAbstractions{}, []string{"ASM", "C", "CXX"})
		assert.Equal(true, areAbstractionsEmpty)
		areAbstractionsEmpty = maker.AreAbstractionsEmpty(maker.CompilerAbstractions{Debug: "on"}, []string{"ASM"})
		assert.Equal(false, areAbstractionsEmpty)
		areAbstractionsEmpty = maker.AreAbstractionsEmpty(maker.CompilerAbstractions{LanguageC: "c11"}, []string{"C"})
		assert.Equal(false, areAbstractionsEmpty)
		areAbstractionsEmpty = maker.AreAbstractionsEmpty(maker.CompilerAbstractions{LanguageCpp: "c++11"}, []string{"CXX"})
		assert.Equal(false, areAbstractionsEmpty)
		areAbstractionsEmpty = maker.AreAbstractionsEmpty(maker.CompilerAbstractions{LanguageC: "c11"}, []string{"CXX"})
		assert.Equal(true, areAbstractionsEmpty)
		areAbstractionsEmpty = maker.AreAbstractionsEmpty(maker.CompilerAbstractions{LanguageCpp: "c++11"}, []string{"C"})
		assert.Equal(true, areAbstractionsEmpty)
	})

	t.Run("test compile abstractions", func(t *testing.T) {
		var abstractions = maker.CompilerAbstractions{
			Debug:       "on",
			Optimize:    "speed",
			Warnings:    "all",
			LanguageC:   "c90",
			LanguageCpp: "c++98",
		}
		var cbuild maker.Cbuild
		cbuild.Languages = []string{"ASM", "C", "CXX"}
		content := cbuild.CompilerAbstractions(abstractions, "ASM")
		assert.Contains(content, "cbuild_set_options_flags(ASM \"speed\" \"on\" \"all\" \"\" ASM_OPTIONS_FLAGS)")
		content = cbuild.CompilerAbstractions(abstractions, "C")
		assert.Contains(content, "cbuild_set_options_flags(CC \"speed\" \"on\" \"all\" \"c90\" CC_OPTIONS_FLAGS")
		content = cbuild.CompilerAbstractions(abstractions, "CXX")
		assert.Contains(content, "cbuild_set_options_flags(CXX \"speed\" \"on\" \"all\" \"c++98\" CXX_OPTIONS_FLAGS)")
	})

	t.Run("test cmake target compile options abstractions", func(t *testing.T) {
		var abstractions = maker.CompilerAbstractions{
			Debug:       "on",
			Optimize:    "speed",
			Warnings:    "all",
			LanguageC:   "c90",
			LanguageCpp: "c++98",
		}
		var cbuild maker.Cbuild
		cbuild.Languages = []string{"ASM", "C", "CXX"}
		content := cbuild.CMakeTargetCompileOptionsAbstractions("NAME", abstractions, cbuild.Languages)
		assert.Contains(content, "add_library(NAME_ABSTRACTIONS INTERFACE)")
		assert.Contains(content, "cbuild_set_options_flags(ASM \"speed\" \"on\" \"all\" \"\" ASM_OPTIONS_FLAGS_NAME)")
		assert.Contains(content, "cbuild_set_options_flags(CC \"speed\" \"on\" \"all\" \"c90\" CC_OPTIONS_FLAGS_NAME")
		assert.Contains(content, "cbuild_set_options_flags(CXX \"speed\" \"on\" \"all\" \"c++98\" CXX_OPTIONS_FLAGS_NAME)")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:ASM>:\n    \"SHELL:${ASM_OPTIONS_FLAGS_NAME}\"\n  >")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:C>:\n    \"SHELL:${CC_OPTIONS_FLAGS_NAME}\"\n  >")
		assert.Contains(content, "$<$<COMPILE_LANGUAGE:CXX>:\n    \"SHELL:${CXX_OPTIONS_FLAGS_NAME}\"\n  >")
	})

	t.Run("test get file options", func(t *testing.T) {
		var files = []maker.Files{
			{
				File: "source.asm",
			},
			{
				File: "source.c",
			},
			{
				File: "source.cxx",
			},
		}
		content := maker.GetFileOptions(files[0], true, ";")
		assert.Contains(content, "${ASM_OPTIONS_FLAGS}")
		content = maker.GetFileOptions(files[1], true, ";")
		assert.Contains(content, "${CC_OPTIONS_FLAGS}")
		content = maker.GetFileOptions(files[2], true, ";")
		assert.Contains(content, "${CXX_OPTIONS_FLAGS}")
	})

	t.Run("test build dependencies", func(t *testing.T) {
		var m maker.Maker
		m.CbuildIndex.BuildIdx.Cbuilds = []maker.Cbuilds{
			{
				Project:       "project",
				Configuration: ".debug+target",
				DependsOn:     []string{"dependentContext"},
			},
		}
		m.Vars.Contexts = []string{
			"project.debug+target",
		}
		m.CbuildIndex.BuildIdx.Executes = []maker.Executes{
			{
				Execute: "RunAlways1",
				Always:  make(map[string]interface{}),
			},
			{
				Execute: "RunAlways2",
				Always:  make(map[string]interface{}),
			},
			{
				// Test safety net to ensure cyclic dependencies are avoided
				// RunAlways 'executes' usually do not have input dependencies
				Execute:   "RunAlways_Dependent",
				Always:    make(map[string]interface{}),
				DependsOn: []string{"SecondLevel_PostBuild"},
			},
			{
				Execute:   "PostBuild",
				DependsOn: []string{"project.debug+target"},
			},
			{
				Execute:   "SecondLevel_PostBuild",
				DependsOn: []string{"PostBuild"},
			},
		}
		content := m.BuildDependencies()
		assert.Contains(content, `
# Build dependencies
add_dependencies(project.debug+target-build
  dependentContext
  RunAlways1
  RunAlways2
)
add_dependencies(RunAlways_Dependent
  SecondLevel_PostBuild
)
add_dependencies(PostBuild
  project.debug+target-build
  RunAlways1
  RunAlways2
)
add_dependencies(SecondLevel_PostBuild
  PostBuild
  RunAlways1
  RunAlways2
)
add_dependencies(project.debug+target-executes
  PostBuild
  SecondLevel_PostBuild
  RunAlways1
  RunAlways2
)`)
	})

	t.Run("test linker options", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.Languages = []string{"C", "CXX"}
		cbuild.BuildDescType.Linker = maker.Linker{
			Script: "./path/to/script.ld",
		}
		cbuild.BuildDescType.Processor.Trustzone = "secure"
		cbuild.BuildDescType.Misc.Link = []string{"--link-flag"}
		cbuild.BuildDescType.Misc.LinkC = []string{"--linkC-flag"}
		cbuild.BuildDescType.Misc.LinkCPP = []string{"--linkCPP-flag"}
		linkerVars, linkerOptions := cbuild.LinkerOptions()
		assert.Contains(linkerVars, "set(LD_SCRIPT \"${SOLUTION_ROOT}/path/to/script.ld\")")
		assert.Contains(linkerOptions, "${LD_SECURE}")
		assert.Contains(linkerOptions, "--link-flag")
		assert.Contains(linkerOptions, "--linkC-flag")
		assert.Contains(linkerOptions, "--linkCPP-flag")
	})

	t.Run("test linker options with pre-processing", func(t *testing.T) {
		var cbuild maker.Cbuild
		define := make([]interface{}, 1)
		define[0] = "DEF_LD_PP"
		cbuild.BuildDescType.Linker = maker.Linker{
			Script:  "./path/to/script.ld.src",
			Regions: "./path/to/regions.h",
			Define:  define,
		}
		linkerVars, _ := cbuild.LinkerOptions()
		assert.Contains(linkerVars, "set(LD_SCRIPT \"${SOLUTION_ROOT}/path/to/script.ld.src\")")
		assert.Contains(linkerVars, "set(LD_REGIONS \"${SOLUTION_ROOT}/path/to/regions.h\")")
		assert.Contains(linkerVars, "DEF_LD_PP")
	})

	t.Run("test adjust relative path", func(t *testing.T) {
		var cbuild maker.Cbuild
		cbuild.ContextRoot = "./context/folder"
		adjustedOption := cbuild.AdjustRelativePath("-map=./out/file.map")
		assert.Equal("-map=${SOLUTION_ROOT}/context/folder/out/file.map", adjustedOption)
		adjustedOption = cbuild.AdjustRelativePath("-map=../../out/file.map")
		assert.Equal("-map=${SOLUTION_ROOT}/out/file.map", adjustedOption)
		adjustedOption = cbuild.AdjustRelativePath("-map=${SOLUTION_ROOT}/out/file.map")
		assert.Equal("-map=${SOLUTION_ROOT}/out/file.map", adjustedOption)
	})

	t.Run("merge common language includes", func(t *testing.T) {
		languages := maker.LanguageMap{
			"C":   []string{"inc-C", "inc-C-CXX", "inc-ASM-C-CXX"},
			"CXX": []string{"inc-CXX", "inc-C-CXX", "inc-ASM-C-CXX"},
			"ASM": []string{"inc-ASM", "inc-ASM-C-CXX"},
		}
		commonlanguages := maker.LanguageMap{
			"C":     []string{"inc-C"},
			"CXX":   []string{"inc-CXX"},
			"ASM":   []string{"inc-ASM"},
			"ALL":   []string{"inc-ASM-C-CXX"},
			"C,CXX": []string{"inc-C-CXX"},
		}
		assert.Equal(commonlanguages, maker.MergeLanguageCommonIncludes(languages))
	})

	t.Run("add root prefixes", func(t *testing.T) {
		var cbuild maker.Cbuild
		absTestRoot, _ := filepath.Abs(testRoot)
		absTestRoot = filepath.ToSlash(absTestRoot)
		cbuild.SolutionRoot = absTestRoot + "/solution"
		testData := []string{
			"relative/path",
			"../solution/path",
			"${CMSIS_PACK_ROOT}/Pack/Name/0.0.0",
			absTestRoot + "/absolute/path",
		}
		expectedData := []string{
			"${SOLUTION_ROOT}/relative/path",
			"${SOLUTION_ROOT}/path",
			"${CMSIS_PACK_ROOT}/Pack/Name/0.0.0",
			absTestRoot + "/absolute/path",
		}
		assert.Equal(expectedData, cbuild.AddRootPrefixes("", testData))
	})

}
