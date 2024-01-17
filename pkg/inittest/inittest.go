/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// This package is used as a common test setup
// avoiding duplicate setup for all the packages
// under test

package inittest

import (
	"os"
	"path/filepath"
	"time"

	cp "github.com/otiai10/copy"
)

func TestInitialization(testRoot string) {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	time.Sleep(2 * time.Second)
	_ = cp.Copy(testRoot+"/data", testRoot+"/run")

	// Set compiler root
	compilerRoot, _ := filepath.Abs(testRoot + "/run/etc")
	os.Setenv("CMSIS_COMPILER_ROOT", compilerRoot)

	// Set toolchain root
	os.Setenv("AC6_TOOLCHAIN_6_19_0", testRoot+"/run/path/to/ac6/toolchain")
}
