/*
 * Copyright (c) 2024-2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/inittest"
	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
)

const testRoot = "../../test"

func init() {
	inittest.TestInitialization(testRoot)
}

func TestMaker(t *testing.T) {
	assert := assert.New(t)

	t.Run("test maker", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/generic/solutionName1.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
	})

	t.Run("test maker with cbuild in subfolder", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/generic/solutionName4.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
		assert.Equal("project/subfolder", m.Cbuilds[0].ContextRoot)
	})

	t.Run("test maker with invalid input param", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "invalid.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Error(err)
	})

	t.Run("test maker with image only solution", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/generic/imageOnly.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
	})

	t.Run("test maker with west solution", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/solutions/west/solution.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
	})
}
