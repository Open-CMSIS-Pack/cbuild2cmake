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
	"regexp"
	"time"

	cp "github.com/otiai10/copy"
)

func TestInitialization(testRoot string) {
	// Prepare test data
	_ = os.RemoveAll(testRoot + "/run")
	time.Sleep(2 * time.Second)
	absTestRoot, _ := filepath.Abs(testRoot)
	_ = cp.Copy(filepath.Join(absTestRoot, "data"), filepath.Join(absTestRoot, "run"))

	// Set compiler root
	compilerRoot := filepath.Join(absTestRoot, "run/etc")
	os.Setenv("CMSIS_COMPILER_ROOT", compilerRoot)

	// Set toolchain root
	os.Setenv("AC6_TOOLCHAIN_6_19_0", filepath.Join(absTestRoot, "run/path/to/ac619/bin"))
}

func ClearToolchainRegistration() {
	// Unset environment variables
	systemEnvVars := os.Environ()
	pattern := regexp.MustCompile(`(\w+)_TOOLCHAIN_(\d+)_(\d+)_(\d+)=(.*)`)
	for _, systemEnvVar := range systemEnvVars {
		matched := pattern.FindAllStringSubmatch(systemEnvVar, -1)
		if matched == nil {
			continue
		}
		os.Unsetenv(systemEnvVar)
	}
}
