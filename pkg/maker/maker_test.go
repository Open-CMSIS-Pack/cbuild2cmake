/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
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
	var m maker.Maker

	t.Run("test maker", func(t *testing.T) {
		m.Params.InputFile = testRoot + "/run/generic/solutionName1.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
	})

	t.Run("test maker with invalid input param", func(t *testing.T) {
		m.Params.InputFile = testRoot + "invalid.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Error(err)
	})
}
