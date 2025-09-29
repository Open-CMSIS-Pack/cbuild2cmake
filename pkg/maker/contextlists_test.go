/*
 * Copyright (c) 2025 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"testing"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
)

func TestContextLists(t *testing.T) {
	assert := assert.New(t)

	t.Run("test context cmakelists creation", func(t *testing.T) {
		var m maker.Maker
		m.Params.InputFile = testRoot + "/run/solutions/build-cpp/solution.cbuild-idx.yml"
		err := m.GenerateCMakeLists()
		assert.Nil(err)
		for index := range m.Cbuilds {
			err = m.CreateContextCMakeLists(index)
			assert.Nil(err)
		}
	})
}
