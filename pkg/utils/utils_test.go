/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package utils_test

import (
	"os"
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"

func TestUtils(t *testing.T) {
	assert := assert.New(t)

	t.Run("test AppendUniquely", func(t *testing.T) {
		assert.Equal([]string{"one", "two", "three"}, utils.AppendUniquely([]string{"one", "two"}, "three"))
		assert.Equal([]string{"one", "two"}, utils.AppendUniquely([]string{"one", "two"}, "one"))
	})

	t.Run("test PrependUniquely", func(t *testing.T) {
		assert.Equal([]string{"three", "one", "two"}, utils.PrependUniquely([]string{"one", "two"}, "three"))
		assert.Equal([]string{"one", "two"}, utils.PrependUniquely([]string{"one", "two"}, "one"))
	})

	t.Run("test FindLast", func(t *testing.T) {
		assert.Equal("option two", utils.FindLast([]string{"option one", "option two", "other"}, "option"))
		assert.Equal("", utils.FindLast([]string{"option one", "option two"}, "other"))
	})

	t.Run("test Intersection", func(t *testing.T) {
		assert.Equal([]string{"two"}, utils.Intersection([]string{"one", "two"}, []string{"two", "three"}))
		assert.Nil(utils.Intersection([]string{"one", "two"}, []string{"three", "four"}))
	})

	t.Run("test RemoveIncludes", func(t *testing.T) {
		assert.Equal([]string{"one", "three"}, utils.RemoveIncludes([]string{"one", "two", "three"}, "two"))
		assert.Equal([]string{"one", "two", "three"}, utils.RemoveIncludes([]string{"one", "two", "three"}, "four"))
	})

	t.Run("test AppendDefines", func(t *testing.T) {
		assert.Equal([]interface{}{"one", "two", "three", "four"}, utils.AppendDefines([]interface{}{"one", "two"}, []interface{}{"three", "four"}))
	})

	t.Run("test RemoveDefines", func(t *testing.T) {
		assert.Equal([]interface{}{"one", "three"}, utils.RemoveDefines([]interface{}{"one", "two", "three"}, "two"))
		assert.Equal([]interface{}{"one", "two", "three"}, utils.RemoveDefines([]interface{}{"one", "two", "three"}, "four"))
	})

	t.Run("test GetDefine", func(t *testing.T) {
		key, value := utils.GetDefine(map[string]interface{}{"key": "value"})
		assert.Equal("key", key)
		assert.Equal("value", value)

		key, value = utils.GetDefine(map[string]interface{}{"key": true})
		assert.Equal("key", key)
		assert.Equal("true", value)

		key, value = utils.GetDefine(map[string]interface{}{"key": 1})
		assert.Equal("key", key)
		assert.Equal("1", value)

		key, _ = utils.GetDefine(interface{}("key"))
		assert.Equal("key", key)
	})

	t.Run("test file operations", func(t *testing.T) {
		// write file and read content back
		filename := testRoot + "/test.txt"
		assert.Nil(utils.UpdateFile(filename, "line1\nline2\nline3\n"))
		content, err := utils.ReadFileContent(filename)
		assert.Equal("line1\nline2\nline3\n", content)
		assert.Nil(err)

		// update file with same content (no modification)
		info, _ := os.Stat(filename)
		timestamp := os.FileInfo.ModTime(info)
		assert.Nil(utils.UpdateFile(filename, "line1\nline2\nline3\n"))
		info, _ = os.Stat(filename)
		assert.Equal(timestamp, os.FileInfo.ModTime(info))

		os.Remove(filename)
	})
}
