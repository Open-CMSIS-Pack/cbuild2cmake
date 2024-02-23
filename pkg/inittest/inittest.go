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
	"path"
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
	absTestRoot = filepath.ToSlash(absTestRoot)
	_ = cp.Copy(path.Join(absTestRoot, "data"), path.Join(absTestRoot, "run"))

	// Set compiler root
	compilerRoot := path.Join(absTestRoot, "run/etc")
	os.Setenv("CMSIS_COMPILER_ROOT", compilerRoot)

	// Set toolchain root
	os.Setenv("AC6_TOOLCHAIN_6_19_0", path.Join(absTestRoot, "run/path/to/ac619/bin"))
	os.Setenv("GCC_TOOLCHAIN_12_3_0", path.Join(absTestRoot, "run/path/to/gcc1230/bin"))
	os.Setenv("IAR_TOOLCHAIN_9_32_5", path.Join(absTestRoot, "run/path/to/iar9325/bin"))
	os.Setenv("CLANG_TOOLCHAIN_18_0_0", path.Join(absTestRoot, "run/path/to/clang1800/bin"))
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
