package gomodule

import (
	"fmt"
	"path"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

type testCoveragedModule struct {
	blueprint.SimpleName

	properties struct {
		Pkg      string
		TestPkg  string
		Srcs     []string
		TestSrcs []string
		Optional bool
	}
}

var (
	goTestCoverage = pctx.StaticRule("testCoverage", blueprint.RuleParams{
		Command:     "cd $workDir && go test $pkg -coverprofile=coverage.out && go tool cover -html=coverage.out -o $htmlFile && rm coverage.out",
		Description: "test go package $pkg",
	}, "workDir", "pkg", "htmlFile")
)

func (tc *testCoveragedModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Info.Printf("Adding test coverage analysis actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "reports", name+".html")

	var buildInputs []string
	var testInputs []string
	inputErrors := false

	for _, src := range tc.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, tc.properties.TestSrcs); err == nil {
			buildInputs = append(buildInputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErrors = true
		}
	}

	for _, src := range tc.properties.TestSrcs {
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

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Test and analyze coverage %s as Go binary", name),
		Rule:        goTestCoverage,
		Outputs:     []string{outputPath},
		Implicits:   testInputs,
		Args: map[string]string{
			"workDir":  ctx.ModuleDir(),
			"pkg":      tc.properties.TestPkg,
			"htmlFile": outputPath,
		},
	})
}

func SimpleTestCoverageFactory() (blueprint.Module, []interface{}) {
	cType := &testCoveragedModule{}
	return cType, []interface{}{&cType.SimpleName.Properties, &cType.properties}
}
