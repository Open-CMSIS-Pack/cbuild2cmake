/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"regexp"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	sortedmap "github.com/gobs/sortedmap"
)

type BuildFiles struct {
	LibraryType string
	Interface   bool
	Include     ScopeMap
	Source      LanguageMap
	Library     []string
	Object      []string
}

type ScopeMap map[string]map[string][]string
type LanguageMap map[string][]string

var CategoryLanguageMap = map[string]string{
	"headerAsm":  "ASM",
	"headerC":    "C",
	"headerCpp":  "CXX",
	"includeAsm": "ASM",
	"includeC":   "C",
	"includeCpp": "CXX",
	"sourceAsm":  "ASM",
	"sourceC":    "C",
	"sourceCpp":  "CXX",
}

var LanguageReMap = map[string]string{
	"asm":   "ASM",
	"c":     "C",
	"cpp":   "CXX",
	"c-cpp": "C,CXX",
}

func GetLanguage(file Files) string {
	language := CategoryLanguageMap[file.Category]
	if len(language) == 0 {
		language = LanguageReMap[file.Language]
	}
	if len(language) == 0 {
		language = "ALL"
	}
	return language
}

func GetScope(file Files) string {
	if len(file.Scope) > 0 && (file.Scope == "private" || file.Scope == "hidden") {
		return "PRIVATE"
	}
	return "PUBLIC"
}

func ReplaceDelimiters(identifier string) string {
	pattern := regexp.MustCompile(`::|:|&|@>=|@|\.`)
	return pattern.ReplaceAllString(identifier, "_")
}

func ClassifyFiles(files []Files) BuildFiles {
	var buildFiles BuildFiles
	buildFiles.Include = make(ScopeMap)
	buildFiles.Source = make(LanguageMap)
	buildFiles.Interface = true
	for _, file := range files {
		if strings.Contains(file.Category, "source") {
			buildFiles.Interface = false
			break
		}
	}

	for _, file := range files {
		switch file.Category {
		case "header", "headerAsm", "headerC", "headerCpp", "include", "includeAsm", "includeC", "includeCpp":
			var scope string
			if buildFiles.Interface {
				scope = "INTERFACE"
			} else {
				scope = GetScope(file)
			}
			language := GetLanguage(file)
			includePath := path.Clean(file.File)
			if strings.Contains(file.Category, "header") {
				includePath = path.Dir(includePath)
			}
			if _, ok := buildFiles.Include[scope]; !ok {
				buildFiles.Include[scope] = make(LanguageMap)
			}
			buildFiles.Include[scope][language] = append(buildFiles.Include[scope][language], includePath)
		case "source", "sourceAsm", "sourceC", "sourceCpp":
			language := GetLanguage(file)
			buildFiles.Source[language] = append(buildFiles.Source[language], path.Clean(file.File))
		case "library":
			buildFiles.Library = append(buildFiles.Library, path.Clean(file.File))
		case "object":
			buildFiles.Object = append(buildFiles.Object, path.Clean(file.File))
		}
	}

	return buildFiles
}

func CMakeAddLibrary(name string, buildFiles BuildFiles) string {
	content := "\nadd_library(" + name
	if buildFiles.Interface {
		content += " INTERFACE)"
	} else {
		content += " OBJECT"
		for _, language := range sortedmap.AsSortedMap(buildFiles.Source) {
			for _, file := range language.Value {
				content += "\n  \"" + file + "\""
			}
		}
		content += "\n)"
	}
	return content
}

func CMakeTargetIncludeDirectoriesFromFiles(name string, buildFiles BuildFiles) string {
	content := "\ntarget_include_directories(" + name
	for _, scope := range sortedmap.AsSortedMap(buildFiles.Include) {
		content += "\n  " + scope.Key
		for _, language := range sortedmap.AsSortedMap(scope.Value) {
			if language.Key == "ALL" {
				for _, file := range language.Value {
					content += "\n    \"" + file + "\""
				}
			} else {
				content += "\n    " + "$<$<COMPILE_LANGUAGE:" + language.Key + ">:"
				for _, file := range language.Value {
					content += "\n      \"" + file + "\""
				}
				content += "\n    >"
			}
		}
	}
	content += "\n)"
	return content
}

func CMakeTargetIncludeDirectories(name string, scope string, includes []string) string {
	content := "\ntarget_include_directories(" + name + " " + scope
	for _, include := range includes {
		content += "\n  \"" + include + "\""
	}
	content += "\n)"
	return content
}

func CMakeTargetCompileDefinitions(name string, scope string, defines []interface{}) string {
	content := "\ntarget_compile_definitions(" + name + " " + scope
	for _, define := range defines {
		key, value := utils.GetDefine(define)
		content += "\n  " + key
		if len(value) > 0 {
			content += "=" + value
		}
	}
	content += "\n)"
	return content
}

func GroupsAndComponentsList(cbuild Cbuild) string {
	var content string
	for _, groups := range cbuild.BuildDescType.Groups {
		content += "\n  Group_" + ReplaceDelimiters(groups.Group)
	}
	for _, component := range cbuild.BuildDescType.Components {
		content += "\n  " + ReplaceDelimiters(component.Component)
	}
	return content
}

func ProcessorOptions(cbuild Cbuild) string {
	content := "\nset(CPU " + cbuild.BuildDescType.Processor.Core + ")"

	var FpuMap = map[string]string{
		"dp":  "DP_FPU",
		"sp":  "SP_FPU",
		"off": "NO_FPU",
	}
	fpu := FpuMap[cbuild.BuildDescType.Processor.Fpu]
	if len(fpu) > 0 {
		content += "\nset(FPU " + fpu + ")"
	}

	var DspMap = map[string]string{
		"on":  "DSP",
		"off": "NO_DSP",
	}
	dsp := DspMap[cbuild.BuildDescType.Processor.Dsp]
	if len(dsp) > 0 {
		content += "\nset(DSP " + dsp + ")"
	}

	var SecureMap = map[string]string{
		"secure":     "Secure",
		"non-secure": "Non-secure",
	}
	secure := SecureMap[cbuild.BuildDescType.Processor.Trustzone]
	if len(secure) > 0 {
		content += "\nset(SECURE " + secure + ")"
	}

	var MveMap = map[string]string{
		"fp":  "FP_FVE",
		"int": "MVE",
		"off": "NO_MVE",
	}
	mve := MveMap[cbuild.BuildDescType.Processor.Mve]
	if len(mve) > 0 {
		content += "\nset(MVE " + mve + ")"
	}

	var BranchProtectionMap = map[string]string{
		"bti":         "BTI",
		"bti-signret": "BTI_SIGNRET",
		"off":         "NO_BRANCHPROT",
	}
	branchProtection := BranchProtectionMap[cbuild.BuildDescType.Processor.BranchProtection]
	if len(branchProtection) > 0 {
		content += "\nset(BRANCHPROT " + branchProtection + ")"
	}

	var EndianMap = map[string]string{
		"big":    "Big-endian",
		"little": "Little-endian",
	}
	endian := EndianMap[cbuild.BuildDescType.Processor.Endian]
	if len(endian) > 0 {
		content += "\nset(BYTE_ORDER " + endian + ")"
	}

	return content
}

func LinkerOptions(cbuild Cbuild) string {
	content := `
set(LD_SCRIPT "` + cbuild.BuildDescType.Linker.Script + `")
target_link_options(${CONTEXT} PUBLIC ${GLOBAL_OPTIONS_LD})
set_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS ${LD_SCRIPT})`
	return content
}
