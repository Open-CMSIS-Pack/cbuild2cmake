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
	"strings"
	"time"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
)

// Update original references
const updateReferences = false

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

func CompareFiles(reference string, actual string) (error, bool) {
	referenceEntries, err := os.ReadDir(reference)
	if err != nil {
		return err, false
	}
	mismatch := false
	for _, referenceEntry := range referenceEntries {
		if path.Ext(referenceEntry.Name()) == ".bak" {
			continue
		}
		referencePath := path.Join(reference, referenceEntry.Name())
		actualPath := path.Join(actual, referenceEntry.Name())
		if referenceEntry.IsDir() {
			err, childMismatch := CompareFiles(referencePath, actualPath)
			if err != nil {
				return err, false
			}
			if childMismatch {
				mismatch = true
			}
		} else {
			referenceContent, _ := utils.ReadFileContent(referencePath)
			actualContent, err := utils.ReadFileContent(actualPath)
			if err != nil {
				return err, false
			}
			actualContent = strings.ReplaceAll(actualContent, "\r\n", "\n")
			referenceContent = strings.ReplaceAll(referenceContent, "\r\n", "\n")
			if actualContent != referenceContent {
				if updateReferences {
					originalReference := strings.ReplaceAll(referencePath, "/run/", "/data/")
					_ = cp.Copy(actualPath, originalReference)
				}
				log.Error("files " + referencePath + " and " + actualPath + " do not match")
				mismatch = true
			}
		}
	}
	return nil, mismatch
}
