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
	Interface bool
	Include   ScopeMap
	Source    LanguageMap
	Library   []string
	Object    []string
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

func OutputFiles(outputList []Output) (outputByProducts string, outputFile string, outputType string, customCommands string) {
	for _, output := range outputList {
		switch output.Type {
		case "hex":
			outputByProducts += "\nset(HEX_FILE \"" + output.File + "\")"
			customCommands += "\n\n# Hex Conversion\n add_custom_command(TARGET ${CONTEXT} POST_BUILD COMMAND ${CMAKE_OBJCOPY} ${ELF2HEX})"
		case "bin":
			outputByProducts += "\nset(BIN_FILE \"" + output.File + "\")"
			customCommands += "\n\n# Bin Conversion\n add_custom_command(TARGET ${CONTEXT} POST_BUILD COMMAND ${CMAKE_OBJCOPY} ${ELF2BIN})"
		case "cmse-lib":
			outputByProducts += "\nset(CMSE_LIB \"" + output.File + "\")"
		case "elf", "lib":
			outputFile = output.File
			outputType = output.Type
		}
	}
	return outputByProducts, outputFile, outputType, customCommands
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
	content := "\ntarget_include_directories(" + name + " " + scope + "\n  "
	content += ListIncludeDirectories(includes, "\n  ", true)
	content += "\n)"
	return content
}

func CMakeTargetCompileDefinitions(name string, scope string, defines []interface{}) string {
	content := "\ntarget_compile_definitions(" + name + " " + scope + "\n  "
	content += ListCompileDefinitions(defines, "\n  ")
	content += "\n)"
	return content
}

func ListIncludeDirectories(includes []string, delimiter string, quoted bool) string {
	if quoted {
		var includesList []string
		for _, include := range includes {
			includesList = append(includesList, "\""+include+"\"")
		}
		return strings.Join(includesList, delimiter)
	}
	return strings.Join(includes, delimiter)
}

func ListCompileDefinitions(defines []interface{}, delimiter string) string {
	var definesList []string
	for _, define := range defines {
		key, value := utils.GetDefine(define)
		pair := key
		if len(value) > 0 {
			pair += "=" + value
		}
		definesList = append(definesList, pair)
	}
	return strings.Join(definesList, delimiter)
}

func ListGroupsAndComponents(cbuild Cbuild) string {
	// get last child group names
	content := GetLastChildGroupNamesRecursively("Group", cbuild.BuildDescType.Groups)
	// get component names
	for _, component := range cbuild.BuildDescType.Components {
		content += "\n  " + ReplaceDelimiters(component.Component)
	}
	return content
}

func GetLastChildGroupNamesRecursively(parent string, groups []Groups) string {
	var content string
	for _, group := range groups {
		name := parent + "_" + ReplaceDelimiters(group.Group)
		if len(group.Groups) > 0 {
			// get children group names recursively
			content += GetLastChildGroupNamesRecursively(name, group.Groups)
		} else {
			// last child
			content += "\n  " + name
		}
	}
	return content
}

func CMakeTargetCompileOptionsGlobal(name string, scope string, cbuild Cbuild) string {
	// options from context settings
	var content string
	for _, language := range cbuild.Languages {
		prefix := language
		if language == "C" {
			prefix = "CC"
		}
		content += "\nset(CMAKE_" + language + "_FLAGS \"${" + prefix + "_CPU} ${" + prefix + "_FLAGS}"
		if len(cbuild.BuildDescType.Processor.Trustzone) > 0 {
			content += " ${" + prefix + "_SECURE}"
		}
		if len(cbuild.BuildDescType.Processor.BranchProtection) > 0 {
			content += " {" + prefix + "_BRANCHPROT}"
		}
		if len(cbuild.BuildDescType.Processor.Endian) > 0 {
			content += " ${" + prefix + "_BYTE_ORDER}"
		}
		content += "\")"
	}
	// misc options
	content += CMakeTargetCompileOptions(name, scope, cbuild.BuildDescType.Misc)
	return content
}

func CMakeTargetCompileOptions(name string, scope string, misc Misc) string {
	content := "\ntarget_compile_options(" + name + " " + scope
	content += ListMiscOptions(misc)
	content += "\n)"
	return content
}

func IsCompileMiscEmpty(misc Misc) bool {
	if len(misc.ASM) > 0 || len(misc.C) > 0 || len(misc.CPP) > 0 || len(misc.CCPP) > 0 {
		return false
	}
	return true
}

func ListMiscOptions(misc Misc) string {
	var content string
	if len(misc.ASM) > 0 {
		content += LangugeSpecificCompileOptions("ASM", misc.ASM)
	}
	if len(misc.C) > 0 {
		content += LangugeSpecificCompileOptions("C", misc.C)
	}
	if len(misc.CPP) > 0 {
		content += LangugeSpecificCompileOptions("CXX", misc.CPP)
	}
	if len(misc.CCPP) > 0 {
		content += LangugeSpecificCompileOptions("C,CXX", misc.CCPP)
	}
	return content
}

func GetFileMisc(file Files, delimiter string) string {
	var misc []string
	switch file.Category {
	case "sourceAsm":
		misc = file.Misc.ASM
	case "sourceC":
		misc = file.Misc.C
	case "sourceCpp":
		misc = file.Misc.CPP
	}
	return strings.Join(misc, delimiter)
}

func LangugeSpecificCompileOptions(language string, misc []string) string {
	content := "\n  " + "$<$<COMPILE_LANGUAGE:" + language + ">:"
	for _, option := range misc {
		content += "\n    " + option
	}
	content += "\n  >"
	return content
}

func AddRootPrefix(base string, input string) string {
	if !strings.HasPrefix(input, "${") {
		return "${SOLUTION_ROOT}/" + path.Join(base, input)
	}
	return input
}

func AddRootPrefixes(base string, input []string) []string {
	var list []string
	for _, element := range input {
		list = append(list, AddRootPrefix(base, element))
	}
	return list
}

func (c *Cbuild) ClassifyFiles(files []Files) BuildFiles {
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
			buildFiles.Include[scope][language] = append(buildFiles.Include[scope][language], AddRootPrefix(c.ContextRoot, includePath))
		case "source", "sourceAsm", "sourceC", "sourceCpp":
			language := GetLanguage(file)
			c.AddContextLanguage(language)
			buildFiles.Source[language] = append(buildFiles.Source[language], AddRootPrefix(c.ContextRoot, file.File))
		case "library":
			buildFiles.Library = append(buildFiles.Library, AddRootPrefix(c.ContextRoot, file.File))
		case "object":
			buildFiles.Object = append(buildFiles.Object, AddRootPrefix(c.ContextRoot, file.File))
		}
	}

	return buildFiles
}

func (c *Cbuild) CMakeSetFileProperties(file Files) string {
	var content string
	hasIncludes := len(file.AddPath) > 0
	hasDefines := len(file.Define) > 0
	hasMisc := !IsCompileMiscEmpty(file.Misc)
	if hasIncludes || hasDefines || hasMisc {
		content = "\nset_source_files_properties(\"" + file.File + "\" PROPERTIES"
		if hasIncludes {
			content += "\n  INCLUDE_DIRECTORIES \"" + ListIncludeDirectories(AddRootPrefixes(c.ContextRoot, file.AddPath), ";", false) + "\""
		}
		if hasDefines {
			content += "\n  COMPILE_DEFINITIONS \"" + ListCompileDefinitions(file.Define, ";") + "\""
		}
		if hasMisc {
			content += "\n  COMPILE_OPTIONS \"" + GetFileMisc(file, ";") + "\""
		}
		content += "\n)\n"
	}
	return content
}

func (c *Cbuild) AddContextLanguage(language string) {
	for _, stored := range c.Languages {
		if stored == language {
			return
		}
	}
	c.Languages = append(c.Languages, language)
}

func (c *Cbuild) LinkerOptions() string {
	content := `
set(LD_SCRIPT "` + AddRootPrefix(c.ContextRoot, c.BuildDescType.Linker.Script) + `")
target_link_options(${CONTEXT} PUBLIC ${GLOBAL_OPTIONS_LD})
set_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS ${LD_SCRIPT})`
	return content
}
