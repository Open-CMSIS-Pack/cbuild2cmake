/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"os"
	"path"
	"strconv"
)

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

func UpdateFile(filename string, content string) error {
	// Check whether file content is the same
	fileContent, err := os.ReadFile(filename)
	if err == nil {
		if string(fileContent) == content {
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
