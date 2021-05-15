package gomodule

import (
	"fmt"
	"path"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

type testedBinaryModule struct {
	blueprint.SimpleName

	properties struct {
		Pkg         string
		TestPkg     string
		Srcs        []string
		TestSrcs    []string
		VendorFirst bool
		Deps        []string
	}
}

var (
	goTest = pctx.StaticRule("binaryTest", blueprint.RuleParams{
		Command:     "cd $workDir && go test $pkg > $testFile",
		Description: "test go package $pkg",
	}, "workDir", "pkg", "testFile")
)

func (tb *testedBinaryModule) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return tb.properties.Deps
}

func (tb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Info.Printf("Adding build & test actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "bin", name)
	testsPath := path.Join(config.BaseOutputDir, "tests", name+".txt")

	var buildInputs []string
	var testInputs []string
	inputErrors := false

	for _, src := range tb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, tb.properties.TestSrcs); err == nil {
			buildInputs = append(buildInputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErrors = true
		}
	}

	for _, src := range tb.properties.TestSrcs {
		if matches, err := ctx.GlobWithDeps(src, nil); err == nil {
			testInputs = append(buildInputs, matches...)
		} else {
			ctx.PropertyErrorf("testSrcs", "Cannot resolve files that match pattern %s", src)
			inputErrors = true
		}
	}

	if inputErrors {
		return
	}

	if tb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "../go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		buildInputs = append(buildInputs, vendorDirPath)
		testInputs = append(testInputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{outputPath},
		Implicits:   buildInputs,
		Args: map[string]string{
			"outputPath": outputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        tb.properties.Pkg,
		},
	})

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Test %s as Go binary", name),
		Rule:        goTest,
		Outputs:     []string{testsPath},
		Implicits:   testInputs,
		Args: map[string]string{
			"workDir":  ctx.ModuleDir(),
			"pkg":      tb.properties.TestPkg,
			"testFile": testsPath,
		},
	})
}

func SimpleTestBinFactory() (blueprint.Module, []interface{}) {
	testType := &testedBinaryModule{}
	return testType, []interface{}{&testType.SimpleName.Properties, &testType.properties}
}
