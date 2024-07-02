/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package maker_test

import (
	"testing"

	utils "github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/utils"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/pkg/maker"
	"github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	assert := assert.New(t)
	var m maker.Maker

	t.Run("test parsing cbuild-idx.yml", func(t *testing.T) {
		data, err := m.ParseCbuildIndexFile(testRoot + "/run/generic/solutionName0.cbuild-idx.yml")
		assert.Nil(err)
		assert.Equal("csolution version 2.2.1", data.BuildIdx.GeneratedBy)
		assert.Equal("cdefault.yml", data.BuildIdx.Cdefault)
		assert.Equal("solutionName.csolution.yml", data.BuildIdx.Csolution)
		assert.Equal("custom/tmp/path", data.BuildIdx.TmpDir)
		assert.Equal("projectName.cproject.yml", data.BuildIdx.Cprojects[0].Cproject)
		assert.Equal("layerName.clayer.yml", data.BuildIdx.Cprojects[0].Clayers[0].Clayer)
		assert.Equal("projectName.BuildType0+TargetType0.cbuild.yml", data.BuildIdx.Cbuilds[0].Cbuild)
		assert.Equal("projectName", data.BuildIdx.Cbuilds[0].Project)
		assert.Equal(".BuildType0+TargetType0", data.BuildIdx.Cbuilds[0].Configuration)
		assert.Equal("projectName.BuildType1+TargetType1.cbuild.yml", data.BuildIdx.Cbuilds[0].DependsOn[0])
	})

	t.Run("test parsing cbuild.yml", func(t *testing.T) {
		data, err := m.ParseCbuildFile(testRoot + "/run/generic/contextName0.cbuild.yml")
		assert.Nil(err)
		assert.Equal("csolution version 2.2.1", data.BuildDescType.GeneratedBy)
		assert.Equal("solutionName.csolution.yml", data.BuildDescType.Solution)
		assert.Equal("projectName.cproject.yml", data.BuildDescType.Project)
		assert.Equal("projectName.BuildType+TargetType", data.BuildDescType.Context)
		assert.Equal("AC6@>=6.6.6", data.BuildDescType.Compiler)
		assert.Equal("deviceName", data.BuildDescType.Device)
		assert.Equal("vendorName::DFP@8.8.8", data.BuildDescType.DevicePack)
		assert.Equal("dp", data.BuildDescType.Processor.Fpu)
		assert.Equal("on", data.BuildDescType.Processor.Dsp)
		assert.Equal("fp", data.BuildDescType.Processor.Mve)
		assert.Equal("little", data.BuildDescType.Processor.Endian)
		assert.Equal("bti-signret", data.BuildDescType.Processor.BranchProtection)
		assert.Equal("non-secure", data.BuildDescType.Processor.Trustzone)
		assert.Equal("Cortex-M0", data.BuildDescType.Processor.Core)

		assert.Equal("vendorName::DFP@8.8.8", data.BuildDescType.Packs[0].Pack)
		assert.Equal("${CMSIS_PACK_ROOT}/vendorName/DFP/8.8.8", data.BuildDescType.Packs[0].Path)

		key, value := utils.GetDefine(data.BuildDescType.Define[0])
		assert.Equal("DEF_SCALAR", key)
		assert.Empty(value)
		key, value = utils.GetDefine(data.BuildDescType.Define[1])
		assert.Equal("DEF_KEY", key)
		assert.Equal("VALUE", value)

		assert.Equal("RTE/_BuildType_TargetType", data.BuildDescType.AddPath[0])
		assert.Equal("${CMSIS_PACK_ROOT}/vendorName/DFP/8.8.8/Include", data.BuildDescType.AddPath[1])

		assert.Equal("tmp/projectName/TargetType/BuildType", data.BuildDescType.OutputDirs.Intdir)
		assert.Equal("out/projectName/TargetType/BuildType", data.BuildDescType.OutputDirs.Outdir)
		assert.Equal("RTE", data.BuildDescType.OutputDirs.Rtedir)

		assert.Equal("lib", data.BuildDescType.Output[0].Type)
		assert.Equal("projectName.lib", data.BuildDescType.Output[0].File)

		assert.Equal("ac6.sct.src", data.BuildDescType.Linker.Script)
		assert.Equal("regions_deviceName.h", data.BuildDescType.Linker.Regions)

		key, _ = utils.GetDefine(data.BuildDescType.Linker.Define[0])
		assert.Equal("LD_PP_DEF0", key)

		assert.Equal("vendorName::DFP:CORE@7.7.7", data.BuildDescType.Components[0].Component)
		assert.Equal("Cortex-M Condition", data.BuildDescType.Components[0].Condition)
		assert.Equal("vendorName::DFP@8.8.8", data.BuildDescType.Components[0].FromPack)
		assert.Equal("CORE", data.BuildDescType.Components[0].SelectedBy)

		assert.Equal("Source", data.BuildDescType.Groups[0].Group)
		assert.Equal("./TestSource.c", data.BuildDescType.Groups[0].Files[0].File)
		assert.Equal("sourceC", data.BuildDescType.Groups[0].Files[0].Category)

		assert.Equal("Subgroup", data.BuildDescType.Groups[0].Groups[0].Group)
		assert.Equal("./TestSubgroup.c", data.BuildDescType.Groups[0].Groups[0].Files[0].File)
		assert.Equal("sourceC", data.BuildDescType.Groups[0].Groups[0].Files[0].Category)

		assert.Equal("config", data.BuildDescType.Groups[0].Files[0].Attr)
		assert.Equal("9.9.9", data.BuildDescType.Groups[0].Files[0].Version)
		assert.Equal("speed", data.BuildDescType.Groups[0].Files[0].Optimize)
		assert.Equal("on", data.BuildDescType.Groups[0].Files[0].Debug)
		assert.Equal("all", data.BuildDescType.Groups[0].Files[0].Warnings)
		assert.Equal("c90", data.BuildDescType.Groups[0].Files[0].LanguageC)
		assert.Equal("c++20", data.BuildDescType.Groups[0].Files[0].LanguageCpp)

		key, _ = utils.GetDefine(data.BuildDescType.Groups[0].Files[0].Define[0])
		assert.Equal("DEF_FILE", key)

		assert.Equal("UNDEF_FILE", data.BuildDescType.Groups[0].Files[0].Undefine[0])
		assert.Equal("./add/path/file", data.BuildDescType.Groups[0].Files[0].AddPath[0])
		assert.Equal("./del/path/file", data.BuildDescType.Groups[0].Files[0].DelPath[0])
		assert.Equal("-ASM-file", data.BuildDescType.Groups[0].Files[0].Misc.ASM[0])
		assert.Equal("-C-file", data.BuildDescType.Groups[0].Files[0].Misc.C[0])
		assert.Equal("-CPP-file", data.BuildDescType.Groups[0].Files[0].Misc.CPP[0])
		assert.Equal("-C-CPP-file", data.BuildDescType.Groups[0].Files[0].Misc.CCPP[0])
		assert.Equal("-Lib-file", data.BuildDescType.Groups[0].Files[0].Misc.Lib[0])
		assert.Equal("-lgcc", data.BuildDescType.Groups[0].Files[0].Misc.Library[0])
		assert.Equal("-Link-file", data.BuildDescType.Groups[0].Files[0].Misc.Link[0])
		assert.Equal("-Link-C-file", data.BuildDescType.Groups[0].Files[0].Misc.LinkC[0])
		assert.Equal("-Link-CPP-file", data.BuildDescType.Groups[0].Files[0].Misc.LinkCPP[0])

		assert.Equal("RTE/__BuildType_TargetType/RTE_Components.h", data.BuildDescType.ConstructedFiles[0].File)
		assert.Equal("header", data.BuildDescType.ConstructedFiles[0].Category)
	})

	t.Run("test parsing invalid cbuild-idx.yml", func(t *testing.T) {
		_, err := m.ParseCbuildIndexFile(testRoot + "/invalid.cbuild-idx.yml")
		assert.Error(err)
	})

	t.Run("test parsing invalid cbuild.yml", func(t *testing.T) {
		_, err := m.ParseCbuildFile(testRoot + "/invalid.cbuild.yml")
		assert.Error(err)
	})

	t.Run("test parsing cbuild files referenced by cbuild-idx", func(t *testing.T) {
		m.Params.InputFile = testRoot + "/run/generic/solutionName1.cbuild-idx.yml"
		err := m.ParseCbuildFiles()
		assert.Nil(err)
	})

	t.Run("test parsing cbuild files referenced by cbuild-idx with debug flag", func(t *testing.T) {
		m.Params.InputFile = testRoot + "/run/generic/solutionName1.cbuild-idx.yml"
		m.Params.Options.Debug = true
		err := m.ParseCbuildFiles()
		assert.Nil(err)
	})

	t.Run("test parsing with invalid input param", func(t *testing.T) {
		m.Params.InputFile = testRoot + "invalid.cbuild-idx.yml"
		err := m.ParseCbuildFiles()
		assert.Error(err)
	})

	t.Run("test parsing with non-existent cbuild reference", func(t *testing.T) {
		m.Params.InputFile = testRoot + "/run/generic/solutionName2.cbuild-idx.yml"
		err := m.ParseCbuildFiles()
		assert.Nil(err)
	})

	t.Run("test parsing with invalid cbuild content", func(t *testing.T) {
		m.Params.InputFile = testRoot + "/run/generic/solutionName3.cbuild-idx.yml"
		err := m.ParseCbuildFiles()
		assert.Error(err)
	})

	t.Run("test parsing cbuild-set.yml", func(t *testing.T) {
		data, err := m.ParseCbuildSetFile(testRoot + "/run/generic/solutionName0.cbuild-set.yml")
		assert.Nil(err)
		assert.Equal("csolution version 2.4.0", data.BuildSet.GeneratedBy)
		assert.Equal("AC6", data.BuildSet.Compiler)
		assert.Equal("projectName.BuildType0+TargetType0", data.BuildSet.Contexts[0].Context)
	})
}
