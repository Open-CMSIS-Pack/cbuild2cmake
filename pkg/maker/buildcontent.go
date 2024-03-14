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
	log "github.com/sirupsen/logrus"
)

type BuildFiles struct {
	Interface       bool
	Include         ScopeMap
	Source          LanguageMap
	Library         []string
	Object          []string
	PreIncludeLocal []string
}

type CompilerAbstractions struct {
	Debug       string
	Optimize    string
	Warnings    string
	LanguageC   string
	LanguageCpp string
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
	if len(language) > 0 {
		return language
	}
	language = LanguageReMap[file.Language]
	if len(language) > 0 {
		return language
	}
	switch path.Ext(file.File) {
	case ".c", ".C":
		return "C"
	case ".cpp", ".c++", ".C++", ".cxx", ".cc", ".CC":
		return "CXX"
	case ".asm", ".s", ".S":
		return "ASM"
	}
	return "ALL"
}

func GetScope(file Files) string {
	if len(file.Scope) > 0 && (file.Scope == "private" || file.Scope == "hidden") {
		return "PRIVATE"
	}
	return "PUBLIC"
}

func ReplaceDelimiters(identifier string) string {
	pattern := regexp.MustCompile(`::|:|&|@>=|@|\.|/| `)
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

func BuildDependencies(cbuilds []Cbuilds) string {
	var content string
	for _, cbuild := range cbuilds {
		if len(cbuild.DependsOn) > 0 {
			content += "\n\nExternalProject_Add_StepDependencies(" + cbuild.Project + cbuild.Configuration + " build"
			for _, dependency := range cbuild.DependsOn {
				content += "\n  " + dependency + "-build"
			}
			content += "\n)"
		}
	}
	if len(content) > 0 {
		content = "\n# Build dependencies" + content
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

func (c *Cbuild) ListGroupsAndComponents() []string {
	// get last child group names
	groupsAndComponents := c.BuildGroups
	// get component names
	for _, component := range c.BuildDescType.Components {
		groupsAndComponents = append(groupsAndComponents, ReplaceDelimiters(component.Component))
	}
	return groupsAndComponents
}

func (c *Cbuild) CMakeTargetCompileOptionsGlobal(name string, scope string) string {
	// options from context settings
	optionsMap := make(map[string][]string)
	for _, language := range c.Languages {
		prefix := language
		if language == "C" {
			prefix = "CC"
		}
		optionsMap[language] = append(optionsMap[language], "SHELL:${"+prefix+"_CPU}")
		optionsMap[language] = append(optionsMap[language], "SHELL:${"+prefix+"_FLAGS}")
		if len(c.BuildDescType.Processor.Trustzone) > 0 {
			optionsMap[language] = append(optionsMap[language], "${"+prefix+"_SECURE}")
		}
		if len(c.BuildDescType.Processor.BranchProtection) > 0 {
			optionsMap[language] = append(optionsMap[language], "${"+prefix+"_BRANCHPROT}")
		}
		if len(c.BuildDescType.Processor.Endian) > 0 {
			optionsMap[language] = append(optionsMap[language], "${"+prefix+"_BYTE_ORDER}")
		}
	}
	// add global misc options
	c.GetCompileOptionsLanguageMap(c.BuildDescType.Misc, &optionsMap)

	// target compile options
	content := "\ntarget_compile_options(" + name + " " + scope
	for language, options := range optionsMap {
		content += c.LanguageSpecificCompileOptions(language, options...)
	}
	// pre-includes global
	for _, preInclude := range c.PreIncludeGlobal {
		content += "\n  SHELL:${_PI}\"" + preInclude + "\""
	}
	content += "\n)"
	return content
}

func (c *Cbuild) CMakeTargetLinkLibraries(name string, scope string, libraries ...string) string {
	content := "\ntarget_link_libraries(" + name + " " + scope
	for _, library := range libraries {
		content += "\n  " + library
	}
	content += "\n)"
	return content
}

func (c *Cbuild) CMakeTargetCompileOptions(name string, scope string, misc Misc, preIncludes []string) string {
	content := "\ntarget_compile_options(" + name + " " + scope
	optionsMap := make(map[string][]string)
	c.GetCompileOptionsLanguageMap(misc, &optionsMap)
	for language, options := range optionsMap {
		content += c.LanguageSpecificCompileOptions(language, options...)
	}
	for _, preInclude := range preIncludes {
		content += "\n  SHELL:${_PI}\"" + preInclude + "\""
	}
	content += "\n)"
	return content
}

func (c *Cbuild) CMakeTargetCompileOptionsAbstractions(name string, abstractions CompilerAbstractions, languages []string) string {
	content := "\nadd_library(" + name + "_ABSTRACTIONS INTERFACE)"
	var options string
	for _, language := range languages {
		prefix := language
		if language == "C" {
			prefix = "CC"
		}
		if !IsAbstractionEmpty(abstractions, language) {
			content += "\ncbuild_set_options_flags(" + prefix
			content += c.SetOptionsFlags(abstractions, language)
			content += " " + prefix + "_OPTIONS_FLAGS_" + name + ")"
			options += c.LanguageSpecificCompileOptions(language, "SHELL:${"+prefix+"_OPTIONS_FLAGS_"+name+"}")
		}
	}
	if len(content) > 0 {
		content += "\ntarget_compile_options(" + name + "_ABSTRACTIONS INTERFACE" + options + "\n)"
	}
	return content
}

func (c *Cbuild) GetCompileOptionsLanguageMap(misc Misc, optionsMap *map[string][]string) {
	for _, language := range c.Languages {
		switch language {
		case "ASM":
			if len(misc.ASM) > 0 {
				(*optionsMap)[language] = append((*optionsMap)[language], misc.ASM...)
			}
		case "C", "CXX":
			if language == "C" && len(misc.C) > 0 {
				(*optionsMap)[language] = append((*optionsMap)[language], misc.C...)
			}
			if language == "CXX" && len(misc.CPP) > 0 {
				(*optionsMap)[language] = append((*optionsMap)[language], misc.CPP...)
			}
			if len(misc.CCPP) > 0 {
				(*optionsMap)[language] = append((*optionsMap)[language], misc.CCPP...)
			}
		}
	}
}

func IsCompileMiscEmpty(misc Misc) bool {
	if len(misc.ASM) > 0 || len(misc.C) > 0 || len(misc.CPP) > 0 || len(misc.CCPP) > 0 {
		return false
	}
	return true
}

func AreAbstractionsEmpty(abstractions CompilerAbstractions, languages []string) bool {
	for _, language := range languages {
		if !IsAbstractionEmpty(abstractions, language) {
			return false
		}
	}
	return true
}

func IsAbstractionEmpty(abstractions CompilerAbstractions, language string) bool {
	if len(abstractions.Debug) > 0 || len(abstractions.Optimize) > 0 || len(abstractions.Warnings) > 0 ||
		(language == "C" && len(abstractions.LanguageC) > 0) ||
		(language == "CXX" && len(abstractions.LanguageCpp) > 0) {
		return false
	}
	return true
}

func GetFileOptions(file Files, hasAbstractions bool, delimiter string) string {
	var options []string
	language := GetLanguage(file)
	prefix := language
	switch language {
	case "ASM":
		options = file.Misc.ASM
	case "C":
		options = append(file.Misc.C, file.Misc.CCPP...)
		prefix = "CC"
	case "CXX":
		options = append(file.Misc.CPP, file.Misc.CCPP...)
	}
	if hasAbstractions {
		options = append(options, "${"+prefix+"_OPTIONS_FLAGS}")
	}
	return strings.Join(options, delimiter)
}

func (c *Cbuild) LanguageSpecificCompileOptions(language string, options ...string) string {
	content := "\n  " + "$<$<COMPILE_LANGUAGE:" + language + ">:"
	for _, option := range options {
		content += "\n    " + c.AdjustRelativePath(option)
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
		if strings.Contains(file.Category, "source") && file.Attr != "template" {
			buildFiles.Interface = false
			break
		}
	}

	for _, file := range files {
		if file.Attr == "template" {
			continue
		}
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
			buildFiles.Include[scope][language] = utils.AppendUniquely(buildFiles.Include[scope][language], AddRootPrefix(c.ContextRoot, includePath))
		case "source", "sourceAsm", "sourceC", "sourceCpp":
			language := GetLanguage(file)
			c.AddContextLanguage(language)
			buildFiles.Source[language] = append(buildFiles.Source[language], AddRootPrefix(c.ContextRoot, file.File))
		case "library":
			buildFiles.Library = append(buildFiles.Library, AddRootPrefix(c.ContextRoot, file.File))
		case "object":
			buildFiles.Object = append(buildFiles.Object, AddRootPrefix(c.ContextRoot, file.File))
		case "preIncludeLocal":
			buildFiles.PreIncludeLocal = append(buildFiles.PreIncludeLocal, AddRootPrefix(c.ContextRoot, file.File))
		case "preIncludeGlobal":
			c.PreIncludeGlobal = append(c.PreIncludeGlobal, AddRootPrefix(c.ContextRoot, file.File))
		}
	}

	return buildFiles
}

func (c *Cbuild) ProcessorOptions() string {
	content := "\nset(CPU " + c.BuildDescType.Processor.Core + ")"

	var FpuMap = map[string]string{
		"dp":  "DP_FPU",
		"sp":  "SP_FPU",
		"off": "NO_FPU",
	}
	fpu := FpuMap[c.BuildDescType.Processor.Fpu]
	if len(fpu) > 0 {
		content += "\nset(FPU " + fpu + ")"
	}

	var DspMap = map[string]string{
		"on":  "DSP",
		"off": "NO_DSP",
	}
	dsp := DspMap[c.BuildDescType.Processor.Dsp]
	if len(dsp) > 0 {
		content += "\nset(DSP " + dsp + ")"
	}

	var SecureMap = map[string]string{
		"secure":     "Secure",
		"non-secure": "Non-secure",
	}
	secure := SecureMap[c.BuildDescType.Processor.Trustzone]
	if len(secure) > 0 {
		content += "\nset(SECURE " + secure + ")"
	}

	var MveMap = map[string]string{
		"fp":  "FP_FVE",
		"int": "MVE",
		"off": "NO_MVE",
	}
	mve := MveMap[c.BuildDescType.Processor.Mve]
	if len(mve) > 0 {
		content += "\nset(MVE " + mve + ")"
	}

	var BranchProtectionMap = map[string]string{
		"bti":         "BTI",
		"bti-signret": "BTI_SIGNRET",
		"off":         "NO_BRANCHPROT",
	}
	branchProtection := BranchProtectionMap[c.BuildDescType.Processor.BranchProtection]
	if len(branchProtection) > 0 {
		content += "\nset(BRANCHPROT " + branchProtection + ")"
	}

	var EndianMap = map[string]string{
		"big":    "Big-endian",
		"little": "Little-endian",
	}
	endian := EndianMap[c.BuildDescType.Processor.Endian]
	if len(endian) > 0 {
		content += "\nset(BYTE_ORDER " + endian + ")"
	}

	return content
}

func InheritCompilerAbstractions(parent CompilerAbstractions, child CompilerAbstractions) CompilerAbstractions {
	if len(child.Debug) == 0 {
		child.Debug = parent.Debug
	}
	if len(child.Optimize) == 0 {
		child.Optimize = parent.Optimize
	}
	if len(child.Warnings) == 0 {
		child.Warnings = parent.Warnings
	}
	if len(child.LanguageC) == 0 {
		child.LanguageC = parent.LanguageC
	}
	if len(child.LanguageCpp) == 0 {
		child.LanguageCpp = parent.LanguageCpp
	}
	return child
}

func (c *Cbuild) SetOptionsFlags(abstractions CompilerAbstractions, language string) string {
	languageStandard := map[string]string{
		"C":   abstractions.LanguageC,
		"CXX": abstractions.LanguageCpp,
	}
	flags := []string{
		abstractions.Optimize,
		abstractions.Debug,
		abstractions.Warnings,
		languageStandard[language],
	}
	var content string
	for _, flag := range flags {
		content += " \""
		if len(flag) > 0 {
			content += flag
		}
		content += "\""
	}
	return content
}

func HasFileAbstractions(files []Files) bool {
	hasFileAbstractions := false
	for _, file := range files {
		if strings.Contains(file.Category, "source") {
			fileAbstractions := CompilerAbstractions{file.Debug, file.Optimize, file.Warnings, file.LanguageC, file.LanguageCpp}
			hasFileAbstractions = !IsAbstractionEmpty(fileAbstractions, GetLanguage(file))
			if hasFileAbstractions {
				break
			}
		}
	}
	return hasFileAbstractions
}

func (c *Cbuild) CompilerAbstractions(abstractions CompilerAbstractions, language string) string {
	languageStandard := map[string]string{
		"C":   abstractions.LanguageC,
		"CXX": abstractions.LanguageCpp,
	}
	flags := []string{
		abstractions.Optimize,
		abstractions.Debug,
		abstractions.Warnings,
		languageStandard[language],
	}
	prefix := language
	if language == "C" {
		prefix = "CC"
	}
	content := "\nset(" + prefix + "_OPTIONS_FLAGS)"
	content += "\ncbuild_set_options_flags(" + prefix
	for _, flag := range flags {
		content += " \""
		if len(flag) > 0 {
			content += flag
		}
		content += "\""
	}
	content += " " + prefix + "_OPTIONS_FLAGS)"
	content += "\nseparate_arguments(" + prefix + "_OPTIONS_FLAGS)"
	return content
}

func (c *Cbuild) CMakeSetFileProperties(file Files, abstractions CompilerAbstractions) string {
	var content string
	// del-path and undefine are currently not supported at file level
	if len(file.DelPath) > 0 {
		log.Warn("del-path is not supported for file " + AddRootPrefix(c.ContextRoot, file.File))
	}
	if len(file.Undefine) > 0 {
		log.Warn("undefine is not supported for file " + AddRootPrefix(c.ContextRoot, file.File))
	}
	// file build options
	hasIncludes := len(file.AddPath) > 0
	hasDefines := len(file.Define) > 0
	hasMisc := !IsCompileMiscEmpty(file.Misc)
	// file compiler abstractions
	language := GetLanguage(file)
	hasAbstractions := !IsAbstractionEmpty(abstractions, language)
	if hasAbstractions {
		content += c.CompilerAbstractions(abstractions, language)
	}
	// set file properties
	if hasIncludes || hasDefines || hasMisc || hasAbstractions {
		content += "\nset_source_files_properties(\"" + AddRootPrefix(c.ContextRoot, file.File) + "\" PROPERTIES"
		if hasIncludes {
			content += "\n  INCLUDE_DIRECTORIES \"" + ListIncludeDirectories(AddRootPrefixes(c.ContextRoot, file.AddPath), ";", false) + "\""
		}
		if hasDefines {
			content += "\n  COMPILE_DEFINITIONS \"" + ListCompileDefinitions(file.Define, ";") + "\""
		}
		if hasMisc || hasAbstractions {
			content += "\n  COMPILE_OPTIONS \"" + GetFileOptions(file, hasAbstractions, ";") + "\""
		}
		content += "\n)\n"
	}
	return content
}

func (c *Cbuild) AddContextLanguage(language string) {
	if language == "ALL" {
		return
	}
	for _, stored := range c.Languages {
		if stored == language {
			return
		}
	}
	c.Languages = append(c.Languages, language)
}

func (c *Cbuild) LinkerOptions() (linkerVars string, linkerOptions string) {
	linkerVars += "\nset(LD_SCRIPT \"" + AddRootPrefix(c.ContextRoot, c.BuildDescType.Linker.Script) + "\")"
	if len(c.BuildDescType.Linker.Regions) > 0 {
		linkerVars += "\nset(LD_REGIONS \"" + AddRootPrefix(c.ContextRoot, c.BuildDescType.Linker.Regions) + "\")"
	}
	if len(c.BuildDescType.Linker.Define) > 0 {
		linkerVars += "\nset(LD_SCRIPT_PP_DEFINES\n  "
		linkerVars += ListCompileDefinitions(c.BuildDescType.Linker.Define, "\n  ")
		linkerVars += "\n)"
	}
	linkerOptions += "\n# Linker options\ntarget_link_options(${CONTEXT} PUBLIC\n  SHELL:${LD_CPU}\n  SHELL:${_LS}\"${LD_SCRIPT_PP}\""
	if len(c.BuildDescType.Processor.Trustzone) > 0 {
		linkerOptions += "\n  SHELL:${LD_SECURE}"
	}
	options := c.BuildDescType.Misc.Link
	for _, language := range c.Languages {
		if language == "C" {
			options = append(options, c.BuildDescType.Misc.LinkC...)
		}
		if language == "CXX" {
			options = append(options, c.BuildDescType.Misc.LinkCPP...)
		}
	}
	for _, option := range options {
		linkerOptions += "\n  " + c.AdjustRelativePath(option)
	}
	linkerOptions += "\n)"
	linkerOptions += "\nset_target_properties(${CONTEXT} PROPERTIES LINK_DEPENDS ${LD_SCRIPT})"
	if path.Ext(c.BuildDescType.Linker.Script) == ".src" || len(c.BuildDescType.Linker.Regions) > 0 || len(c.BuildDescType.Linker.Define) > 0 {
		linkerScriptPP := strings.TrimSuffix(path.Base(c.BuildDescType.Linker.Script), ".src")
		linkerVars += "\nset(LD_SCRIPT_PP \"${CMAKE_CURRENT_BINARY_DIR}/" + linkerScriptPP + "\")"
		linkerOptions += "\n\n# Linker script pre-processing\nadd_custom_command(TARGET ${CONTEXT} PRE_LINK COMMAND ${CPP} ARGS ${CPP_ARGS_LD_SCRIPT} BYPRODUCTS ${LD_SCRIPT_PP})"
	} else {
		linkerVars += "\nset(LD_SCRIPT_PP ${LD_SCRIPT})"
	}
	return linkerVars, linkerOptions
}

func (c *Cbuild) AdjustRelativePath(option string) string {
	pattern := regexp.MustCompile(`\./.*|\.\./.*`)
	if pattern.MatchString(option) {
		relativePath := pattern.FindString(option)
		option = strings.Replace(option, relativePath, AddRootPrefix(c.ContextRoot, relativePath), 1)
	}
	return option
}
