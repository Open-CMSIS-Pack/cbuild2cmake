/*
 * Copyright (c) 2024-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker

import (
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	sortedmap "github.com/gobs/sortedmap"
)

type BuildFiles struct {
	Interface       bool
	Include         ScopeMap
	Source          LanguageMap
	Custom          LanguageMap
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
type DependenciesMap map[string][]string

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
	pattern := regexp.MustCompile(`::|:|&|@>=|@|\.|/|\(|\)| `)
	return pattern.ReplaceAllString(identifier, "_")
}

func MergeLanguageCommonIncludes(languages LanguageMap) LanguageMap {
	intersection := utils.Intersection(languages["C"], languages["CXX"])
	if len(intersection) > 0 {
		languages["C,CXX"] = utils.AppendUniquely(languages["C,CXX"], intersection...)
		languages["C"] = utils.RemoveIncludes(languages["C"], intersection...)
		languages["CXX"] = utils.RemoveIncludes(languages["CXX"], intersection...)
	}
	intersection = utils.Intersection(languages["ASM"], languages["C,CXX"])
	if len(intersection) > 0 {
		languages["ALL"] = utils.AppendUniquely(languages["ALL"], intersection...)
		languages["ASM"] = utils.RemoveIncludes(languages["ASM"], intersection...)
		languages["C,CXX"] = utils.RemoveIncludes(languages["C,CXX"], intersection...)
	}
	languages["C"] = utils.RemoveIncludes(languages["C"], utils.Intersection(languages["ALL"], languages["C"])...)
	languages["CXX"] = utils.RemoveIncludes(languages["CXX"], utils.Intersection(languages["ALL"], languages["CXX"])...)
	languages["C,CXX"] = utils.RemoveIncludes(languages["C,CXX"], utils.Intersection(languages["ALL"], languages["C,CXX"])...)
	languages["ASM"] = utils.RemoveIncludes(languages["ASM"], utils.Intersection(languages["ALL"], languages["ASM"])...)
	return languages
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

func (c *Cbuild) CMakeAddLibraryCustomFile(name string, file Files) string {
	return "\nadd_library(" + name + " OBJECT\n  \"" + c.AddRootPrefix(c.ContextRoot, file.File) + "\"\n)"
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
		case "map":
			outputByProducts += "\nset(LD_MAP_FILE \"" + output.File + "\")"
		case "cmse-lib":
			outputByProducts += "\nset(CMSE_LIB \"" + output.File + "\")"
			customCommands += "\n\n# CMSE Library\n add_custom_command(TARGET ${CONTEXT} PRE_LINK COMMAND \"\" BYPRODUCTS ${OUT_DIR}/${CMSE_LIB})"
		case "elf", "lib":
			outputFile = output.File
			outputType = output.Type
		}
	}
	return outputByProducts, outputFile, outputType, customCommands
}

func HasMapFile(outputList []Output) bool {
	for _, output := range outputList {
		if output.Type == "map" {
			return true
		}
	}
	return false
}

func (m *Maker) AddStepSuffix(name string) string {
	if slices.Contains(m.Contexts, name) {
		name += "-build"
	}
	return name
}

func (m *Maker) CMakeTargetAddDependencies(name string, dependencies []string) string {
	var content string
	dependencies = utils.AppendUniquely(dependencies, m.GetIndependentRunAlways(name)...)
	if len(dependencies) > 0 {
		content += "\nadd_dependencies(" + m.AddStepSuffix(name)
		for _, dependency := range dependencies {
			content += "\n  " + m.AddStepSuffix(dependency)
		}
		content += "\n)"
	}
	return content
}

func (m *Maker) BuildDependencies() string {
	var content string
	for _, cbuild := range m.CbuildIndex.BuildIdx.Cbuilds {
		if m.Options.UseContextSet && !slices.Contains(m.Contexts, cbuild.Project+cbuild.Configuration) {
			continue
		}
		content += m.CMakeTargetAddDependencies(cbuild.Project+cbuild.Configuration, cbuild.DependsOn)
	}
	var postBuildDependencies = make(DependenciesMap)
	for _, item := range m.CbuildIndex.BuildIdx.Executes {
		content += m.CMakeTargetAddDependencies(item.Execute, item.DependsOn)
		postBuildDependencies = m.GetContextDependencies(item.Execute, item.DependsOn, postBuildDependencies)
	}
	// add executes statement to ${CONTEXT}-executes target of context
	for context, dependencies := range postBuildDependencies {
		content += m.CMakeTargetAddDependencies(context+"-executes", dependencies)
	}
	if len(content) > 0 {
		content = "\n\n# Build dependencies" + content
	}
	return content
}

func (m *Maker) GetContextDependencies(execute string, dependsOn []string, deps DependenciesMap) DependenciesMap {
	if m.GetExecute(execute).Always == nil {
		for _, item := range dependsOn {
			if slices.Contains(m.Contexts, item) {
				// collect dependency on context (post build step)
				deps[item] = utils.AppendUniquely(deps[item], execute)
			} else {
				// check recursively further dependencies
				deps = m.GetContextDependencies(execute, m.GetExecute(item).DependsOn, deps)
			}
		}
	}
	return deps
}

func (m *Maker) GetExecute(execute string) Executes {
	for _, item := range m.CbuildIndex.BuildIdx.Executes {
		if item.Execute == execute {
			return item
		}
	}
	return Executes{}
}

func (m *Maker) GetIndependentRunAlways(execute string) (elements []string) {
	if m.GetExecute(execute).Always == nil {
		for _, item := range m.CbuildIndex.BuildIdx.Executes {
			if item.Always != nil && len(item.DependsOn) == 0 {
				elements = utils.AppendUniquely(elements, item.Execute)
			}
		}
	}
	return elements
}

func CMakeTargetIncludeDirectories(name string, includes ScopeMap) string {
	if len(includes) == 0 {
		return ""
	}
	content := "\ntarget_include_directories(" + name
	scopeIndentation := " "
	fileIndentation := "\n  "
	if len(includes) > 1 {
		scopeIndentation = "\n  "
		fileIndentation = "\n    "
	}
	for _, scope := range sortedmap.AsSortedMap(includes) {
		content += scopeIndentation + scope.Key
		var allLanguagesContent, specificLanguageContent string
		for _, language := range sortedmap.AsSortedMap(MergeLanguageCommonIncludes(scope.Value)) {
			if language.Key == "ALL" {
				for _, file := range language.Value {
					if strings.Contains(file, "$<TARGET_PROPERTY:") {
						content += fileIndentation + file
					} else {
						allLanguagesContent += fileIndentation + "\"" + file + "\""
					}
				}
			} else {
				if len(language.Value) > 0 {
					specificLanguageContent += fileIndentation + "$<$<COMPILE_LANGUAGE:" + language.Key + ">:"
					for _, file := range language.Value {
						specificLanguageContent += fileIndentation + "  \"" + file + "\""
					}
					specificLanguageContent += fileIndentation + ">"
				}
			}
		}
		content += specificLanguageContent + allLanguagesContent
	}
	content += "\n)"
	return content
}

func CMakeTargetCompileDefinitions(name string, parent string, scope string, define []interface{}, undefine []string) string {
	content := "\ntarget_compile_definitions(" + name + " " + scope
	if len(define) > 0 {
		content += "\n  $<$<COMPILE_LANGUAGE:C,CXX>:\n    "
		content += ListCompileDefinitions(define, "\n    ")
		content += "\n  >"
	}
	if len(parent) > 0 {
		if len(undefine) > 0 {
			content += "\n  $<LIST:FILTER,$<TARGET_PROPERTY:" + parent + ",INTERFACE_COMPILE_DEFINITIONS>,EXCLUDE,^" + strings.Join(undefine, ".*,^") + ".*>"
		} else {
			content += "\n  $<TARGET_PROPERTY:" + parent + ",INTERFACE_COMPILE_DEFINITIONS>"
		}
	}
	content += "\n)"
	return content
}

func ListIncludeDirectories(includes []string, delimiter string) string {
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
		optionsMap[language] = append(optionsMap[language], "${"+prefix+"_CPU}")
		optionsMap[language] = append(optionsMap[language], "${"+prefix+"_FLAGS}")
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
	c.GetCompileOptionsLanguageMap((c.BuildDescType.Lto), c.BuildDescType.Misc, &optionsMap)

	// pre-includes global
	for _, preInclude := range c.PreIncludeGlobal {
		optionsMap["C,CXX"] = append(optionsMap["C,CXX"], "${_PI}\""+preInclude+"\"")
	}

	// target compile options
	content := "\ntarget_compile_options(" + name + " " + scope
	for _, language := range sortedmap.AsSortedMap(optionsMap) {
		content += c.LanguageSpecificCompileOptions(language.Key, language.Value...)
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

func (c *Cbuild) CMakeTargetCompileOptions(name string, scope string, lto bool, misc Misc, preIncludes []string, parent string) string {
	content := "\ntarget_compile_options(" + name + " " + scope
	content += "\n  $<TARGET_PROPERTY:" + parent + ",INTERFACE_COMPILE_OPTIONS>"
	optionsMap := make(map[string][]string)
	c.GetCompileOptionsLanguageMap(lto, misc, &optionsMap)
	for _, preInclude := range preIncludes {
		optionsMap["C,CXX"] = append(optionsMap["C,CXX"], "${_PI}\""+preInclude+"\"")
	}
	for _, language := range sortedmap.AsSortedMap(optionsMap) {
		content += c.LanguageSpecificCompileOptions(language.Key, language.Value...)
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
		if !AreAbstractionsEmpty(abstractions, []string{language}) {
			content += "\ncbuild_set_options_flags(" + prefix
			content += c.SetOptionsFlags(abstractions, language)
			content += " " + prefix + "_OPTIONS_FLAGS_" + name + ")"
			options += c.LanguageSpecificCompileOptions(language, "${"+prefix+"_OPTIONS_FLAGS_"+name+"}")
		}
	}
	if len(options) > 0 {
		content += "\ntarget_compile_options(" + name + "_ABSTRACTIONS INTERFACE" + options + "\n)"
	}
	return content
}

func (c *Cbuild) GetCompileOptionsLanguageMap(lto bool, misc Misc, optionsMap *map[string][]string) {
	for _, language := range c.Languages {
		switch language {
		case "ASM":
			if len(misc.ASM) > 0 {
				(*optionsMap)[language] = append((*optionsMap)[language], misc.ASM...)
			}
		case "C", "CXX":
			if language == "C" {
				if lto {
					c.LinkerLto = true
					(*optionsMap)[language] = append((*optionsMap)[language], "${CC_LTO}")
				}
				if len(misc.C) > 0 {
					(*optionsMap)[language] = append((*optionsMap)[language], misc.C...)
				}
			}
			if language == "CXX" {
				if lto {
					c.LinkerLto = true
					(*optionsMap)[language] = append((*optionsMap)[language], "${CXX_LTO}")
				}
				if len(misc.CPP) > 0 {
					(*optionsMap)[language] = append((*optionsMap)[language], misc.CPP...)
				}
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
	if len(abstractions.Debug) > 0 || len(abstractions.Optimize) > 0 || len(abstractions.Warnings) > 0 {
		return false
	}
	for _, language := range languages {
		if (language == "C" && len(abstractions.LanguageC) > 0) ||
			(language == "CXX" && len(abstractions.LanguageCpp) > 0) {
			return false
		}
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
	if file.Lto {
		options = append(options, "${"+prefix+"_LTO}")
	}
	if hasAbstractions {
		options = append(options, "${"+prefix+"_OPTIONS_FLAGS}")
	}
	return strings.Join(options, delimiter)
}

func (c *Cbuild) LanguageSpecificCompileOptions(language string, options ...string) string {
	content := "\n  " + "$<$<COMPILE_LANGUAGE:" + language + ">:"
	for _, option := range options {
		content += "\n    " + AddShellPrefix(c.AdjustRelativePath(option))
	}
	content += "\n  >"
	return content
}

func AddShellPrefix(input string) string {
	return "\"SHELL:" + strings.ReplaceAll(input, "\"", "\\\"") + "\""
}

func (c *Cbuild) AddRootPrefix(base string, input string) string {
	return AddRootPrefix(base, input, c.SolutionRoot)
}

func AddRootPrefix(base string, input string, solutionRoot string) string {
	if !strings.HasPrefix(input, "${") && !filepath.IsAbs(input) {
		if strings.Contains(input, "..") {
			relPath, _ := filepath.Rel(solutionRoot, path.Join(solutionRoot, base, input))
			return "${SOLUTION_ROOT}/" + filepath.ToSlash(relPath)
		} else {
			return "${SOLUTION_ROOT}/" + path.Join(base, input)
		}
	}
	return input
}

func (c *Cbuild) AddRootPrefixes(base string, input []string) []string {
	var list []string
	for _, element := range input {
		list = append(list, c.AddRootPrefix(base, element))
	}
	return list
}

func (c *Cbuild) ClassifyFiles(files []Files) BuildFiles {
	var buildFiles BuildFiles
	buildFiles.Include = make(ScopeMap)
	buildFiles.Source = make(LanguageMap)
	buildFiles.Custom = make(LanguageMap)
	buildFiles.Interface = true
	for _, file := range files {
		if strings.Contains(file.Category, "source") && file.Attr != "template" && !HasFileCustomOptions(file) {
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
			if file.Attr == "config" {
				buildFiles.Include[scope][language] = utils.PrependUniquely(buildFiles.Include[scope][language], c.AddRootPrefix(c.ContextRoot, includePath))
			} else {
				buildFiles.Include[scope][language] = utils.AppendUniquely(buildFiles.Include[scope][language], c.AddRootPrefix(c.ContextRoot, includePath))
			}
		case "source", "sourceAsm", "sourceC", "sourceCpp":
			language := GetLanguage(file)
			c.AddContextLanguage(language)
			if HasFileCustomOptions(file) {
				buildFiles.Custom[language] = append(buildFiles.Custom[language], c.AddRootPrefix(c.ContextRoot, file.File))
			} else {
				buildFiles.Source[language] = append(buildFiles.Source[language], c.AddRootPrefix(c.ContextRoot, file.File))
			}
		case "library":
			buildFiles.Library = append(buildFiles.Library, c.AddRootPrefix(c.ContextRoot, file.File))
		case "object":
			buildFiles.Object = append(buildFiles.Object, c.AddRootPrefix(c.ContextRoot, file.File))
		case "preIncludeLocal":
			buildFiles.PreIncludeLocal = append(buildFiles.PreIncludeLocal, c.AddRootPrefix(c.ContextRoot, file.File))
		case "preIncludeGlobal":
			c.PreIncludeGlobal = append(c.PreIncludeGlobal, c.AddRootPrefix(c.ContextRoot, file.File))
		}
	}

	return buildFiles
}

func (c *Cbuild) MergeIncludes(includes ScopeMap, scope string, parent string, addPaths []string, addPathsAsm []string, delPaths []string) ScopeMap {
	if _, ok := includes[scope]; !ok {
		includes[scope] = make(LanguageMap)
	}
	if len(addPaths) > 0 {
		includes[scope]["C,CXX"] = utils.AppendUniquely(c.AddRootPrefixes(c.ContextRoot, addPaths), includes[scope]["C,CXX"]...)
	}
	if len(addPathsAsm) > 0 {
		includes[scope]["ASM"] = utils.AppendUniquely(c.AddRootPrefixes(c.ContextRoot, addPathsAsm), includes[scope]["ASM"]...)
	}
	if len(parent) > 0 {
		if len(delPaths) > 0 {
			includes[scope]["ALL"] = utils.PrependUniquely(includes[scope]["ALL"], "$<LIST:REMOVE_ITEM,$<TARGET_PROPERTY:"+
				parent+",INTERFACE_INCLUDE_DIRECTORIES>,"+ListIncludeDirectories(c.AddRootPrefixes(c.ContextRoot, delPaths), ",")+">")
		} else {
			includes[scope]["ALL"] = utils.PrependUniquely(includes[scope]["ALL"], "$<TARGET_PROPERTY:"+parent+",INTERFACE_INCLUDE_DIRECTORIES>")
		}
	}
	return includes
}

func AppendGlobalIncludes(includes LanguageMap, elements ScopeMap) LanguageMap {
	for scope, languages := range elements {
		if scope != "PRIVATE" {
			if includes == nil {
				includes = make(LanguageMap)
			}
			for language, paths := range languages {
				includes[language] = utils.AppendUniquely(includes[language], paths...)
			}
		}
	}
	return includes
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
		"secure":      "Secure",
		"secure-only": "Secure-only",
		"non-secure":  "Non-secure",
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
			hasFileAbstractions = !AreAbstractionsEmpty(fileAbstractions, []string{GetLanguage(file)})
			if hasFileAbstractions {
				break
			}
		}
	}
	return hasFileAbstractions
}

func HasFileCustomOptions(file Files) bool {
	if len(file.AddPath) > 0 || len(file.AddPathAsm) > 0 || len(file.DelPath) > 0 ||
		(GetLanguage(file) != "ASM" && (len(file.Define) > 0 || len(file.Undefine) > 0)) {
		return true
	}
	return false
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
	// file build options
	language := GetLanguage(file)
	hasMisc := !IsCompileMiscEmpty(file.Misc)
	c.LinkerLto = c.LinkerLto || file.Lto
	// file compiler abstractions
	hasAbstractions := !AreAbstractionsEmpty(abstractions, []string{language})
	if hasAbstractions {
		content += c.CompilerAbstractions(abstractions, language)
	}
	// file is generated by executes
	filename := c.AddRootPrefix(c.ContextRoot, file.File)
	isGenerated := slices.Contains(c.GeneratedFiles, filename)
	// set file properties
	if hasMisc || file.Lto || hasAbstractions || isGenerated {
		content += "\nset_source_files_properties(\"" + filename + "\" PROPERTIES"
		if hasMisc || file.Lto || hasAbstractions {
			content += "\n  COMPILE_OPTIONS \"" + GetFileOptions(file, hasAbstractions, ";") + "\""
		}
		if isGenerated {
			content += "\n  GENERATED TRUE"
		}
		content += "\n)"
	}
	return content
}

func (c *Cbuild) SetFileAsmDefines(file Files, parentMiscAsm []string) string {
	var content string
	if len(file.DefineAsm) > 0 {
		flags := utils.AppendUniquely(parentMiscAsm, file.Misc.ASM...)
		if (c.Toolchain == "AC6" || c.Toolchain == "GCC") && path.Ext(file.File) != ".S" && !strings.Contains(utils.FindLast(flags, "-x"), "assembler-with-cpp") {
			syntax := "AS_GNU"
			masm := utils.FindLast(flags, "-masm")
			if c.Toolchain == "AC6" && (strings.Contains(masm, "armasm") || strings.Contains(masm, "auto")) {
				syntax = "AS_ARM"
			}
			content += "\nset(COMPILE_DEFINITIONS\n  " + ListCompileDefinitions(file.DefineAsm, "\n  ") + "\n)"
			content += "\ncbuild_set_defines(" + syntax + " COMPILE_DEFINITIONS)"
			content += "\nset_source_files_properties(\"" + c.AddRootPrefix(c.ContextRoot, file.File) +
				"\" PROPERTIES\n  COMPILE_FLAGS \"${COMPILE_DEFINITIONS}\"\n)"
		} else {
			content += "\nset_source_files_properties(\"" + c.AddRootPrefix(c.ContextRoot, file.File) +
				"\" PROPERTIES\n  COMPILE_DEFINITIONS \"" + ListCompileDefinitions(file.DefineAsm, ";") + "\"\n)"
		}
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
	linkerVars += "\nset(LD_SCRIPT \"" + c.AddRootPrefix(c.ContextRoot, c.BuildDescType.Linker.Script) + "\")"
	if len(c.BuildDescType.Linker.Regions) > 0 {
		linkerVars += "\nset(LD_REGIONS \"" + c.AddRootPrefix(c.ContextRoot, c.BuildDescType.Linker.Regions) + "\")"
	}
	if len(c.BuildDescType.Linker.Define) > 0 {
		linkerVars += "\nset(LD_SCRIPT_PP_DEFINES\n  "
		linkerVars += ListCompileDefinitions(c.BuildDescType.Linker.Define, "\n  ")
		linkerVars += "\n)"
	}
	linkerOptions += "\n# Linker options\ntarget_link_options(${CONTEXT} PUBLIC\n  " +
		AddShellPrefix("${LD_CPU}") + "\n  " +
		AddShellPrefix("${_LS}\"${LD_SCRIPT_PP}\"")
	if HasMapFile(c.BuildDescType.Output) {
		linkerOptions += "\n  " + AddShellPrefix("${LD_MAP}")
	}
	if len(c.BuildDescType.Processor.Trustzone) > 0 {
		linkerOptions += "\n  " + AddShellPrefix("${LD_SECURE}")
	}
	if c.LinkerLto || c.BuildDescType.Lto {
		linkerOptions += "\n  " + AddShellPrefix("${LD_LTO}")
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
		linkerOptions += "\n  " + AddShellPrefix(c.AdjustRelativePath(option))
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
	if !strings.Contains(option, "${SOLUTION_ROOT}") {
		pattern := regexp.MustCompile(`\./.*|\.\./.*`)
		if pattern.MatchString(option) {
			relativePath := pattern.FindString(option)
			option = strings.Replace(option, relativePath, c.AddRootPrefix(c.ContextRoot, relativePath), 1)
		}
	}
	return option
}

func QuoteArguments(cmd string) string {
	pattern := regexp.MustCompile(`(\${INPUT(_\d)?}|\${OUTPUT(_\d)?})`)
	return pattern.ReplaceAllString(cmd, "\"${1}\"")
}

func (m *Maker) ListExecutesIOs(io string, list []string, run string) string {
	content := "\nset(" + io
	var listItems string
	for index, input := range list {
		content += "\n  " + AddRootPrefix("", input, m.SolutionRoot)
		if strings.Contains(run, "${"+io+"_"+strconv.Itoa(index)+"}") {
			listItems += "\nlist(GET " + io + " " + strconv.Itoa(index) + " " + io + "_" + strconv.Itoa(index) + ")"
		}
	}
	content += "\n)"
	content += listItems
	return content
}

func (m *Maker) GetGeneratedFiles(list []string) {
	for _, input := range list {
		file := AddRootPrefix("", input, m.SolutionRoot)
		m.GeneratedFiles = utils.AppendUniquely(m.GeneratedFiles, file)
	}
}

func (m *Maker) ExecutesCommands(executes []Executes) string {
	var content string
	for _, item := range executes {
		content += "\n\n# Execute: " + item.Execute
		customTarget := "\nadd_custom_target(" + item.Execute + " ALL"
		runAlways := item.Always != nil
		if runAlways {
			customTarget += "\n  COMMAND ${CMAKE_COMMAND} -E echo \"Executing: " + item.Execute + "\""
			customTarget += "\n  COMMAND " + QuoteArguments(item.Run)
			if len(item.Output) > 0 {
				customTarget += "\n  BYPRODUCTS ${OUTPUT}"
			}
			customTarget += "\n  USES_TERMINAL\n)"
		} else {
			customTarget += " DEPENDS ${OUTPUT})"
		}
		customCommand := "\nadd_custom_command(OUTPUT ${OUTPUT}"
		if len(item.Input) > 0 {
			content += m.ListExecutesIOs("INPUT", item.Input, item.Run)
			customCommand += " DEPENDS ${INPUT}"
		}
		executeCommandNameAdded := false
		if !runAlways && len(item.Output) == 0 {
			item.Output = append(item.Output, "${CMAKE_CURRENT_BINARY_DIR}/"+item.Execute+".stamp")
			customCommand += "\n  COMMAND ${CMAKE_COMMAND} -E echo \"Executing: " + item.Execute + "\""
			customCommand += "\n  COMMAND ${CMAKE_COMMAND} -E touch \"" + item.Execute + ".stamp\""
			executeCommandNameAdded = true
		}
		if len(item.Output) > 0 {
			content += m.ListExecutesIOs("OUTPUT", item.Output, item.Run)
			m.GetGeneratedFiles(item.Output)
		}
		content += customTarget
		if !runAlways {
			if !executeCommandNameAdded {
				customCommand += "\n  COMMAND ${CMAKE_COMMAND} -E echo \"Executing: " + item.Execute + "\""
			}
			customCommand += "\n  COMMAND " + QuoteArguments(item.Run) + "\n  USES_TERMINAL\n)"
			content += customCommand
		}
	}
	return content
}

func (c *Cbuild) GetAPIFiles(id string) []Files {
	for _, api := range c.BuildDescType.Apis {
		if api.API == id {
			return api.Files
		}
	}
	return nil
}
