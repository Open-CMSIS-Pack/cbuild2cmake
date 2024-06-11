/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

func AppendUniquely(list []string, elements ...string) []string {
	for _, element := range elements {
		if !slices.Contains(list, element) {
			list = append(list, element)
		}
	}
	return list
}

func FindLast(list []string, substr string) string {
	for i := len(list) - 1; i >= 0; i-- {
		if strings.Contains(list[i], substr) {
			return list[i]
		}
	}
	return ""
}

func Intersection(slice1, slice2 []string) []string {
	var intersection []string
	hash := make(map[string]bool)
	for _, element := range slice1 {
		hash[element] = true
	}
	for _, element := range slice2 {
		if hash[element] {
			intersection = AppendUniquely(intersection, element)
		}
	}
	return intersection
}

func RemoveIncludes(includes []string, delpaths ...string) []string {
	for _, delpath := range delpaths {
		index := slices.Index(includes, delpath)
		if index > -1 {
			includes = append(includes[:index], includes[index+1:]...)
		}
	}
	return includes
}

func AppendDefines(defines []interface{}, elements []interface{}) []interface{} {
	return append(defines, elements...)
}

func RemoveDefines(defines []interface{}, undefines ...string) []interface{} {
	for _, undefine := range undefines {
		for index, define := range defines {
			key, _ := GetDefine(define)
			if key == undefine {
				defines = append(defines[:index], defines[index+1:]...)
				break
			}
		}
	}
	return defines
}

func GetDefine(define interface{}) (key string, value string) {
	switch def := define.(type) {
	case string:
		key = def
	case map[string]interface{}:
		for k, v := range def {
			key = k
			switch val := v.(type) {
			case string:
				value = val
			case bool:
				value = strconv.FormatBool(val)
			case int:
				value = strconv.Itoa(val)
			}
		}
	}
	return key, value
}

func ReadFileContent(filename string) (string, error) {
	// Read file content
	fileContent, err := os.ReadFile(filename)
	if err == nil {
		return string(fileContent), nil
	}
	return "", err
}

func UpdateFile(filename string, content string) error {
	// Check whether file content is the same
	fileContent, err := ReadFileContent(filename)
	if err == nil {
		if fileContent == content {
			return nil
		}
	}

	// Create or truncate file
	_ = os.MkdirAll(path.Dir(filename), 0755)
	file, err := os.Create(filename)
	if err != nil {
		file.Close()
		return err
	}

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	// Close created file
	err = file.Close()
	if err != nil {
		return err
	}
	return err
}
