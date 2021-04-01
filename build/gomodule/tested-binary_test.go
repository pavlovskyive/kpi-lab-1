package gomodule

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

func TestSimpleTestFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
 			go_tested_binary {
 			  name: "test-out",
 			  srcs: ["test-src.go"],
               testSrcs: ["**/*_test.go"],
 			  pkg: "./...",
 	          vendorFirst: true
 			}
 		`),
		"test-src.go": nil,
	})

	ctx.RegisterModuleType("go_tested_binary", SimpleTestBinFactory)

	cfg := bood.NewConfig()

	_, errs := ctx.ParseBlueprintsFiles(".", cfg)
	if len(errs) != 0 {
		t.Fatalf("Syntax errors in the test blueprint file: %s", errs)
	}

	_, errs = ctx.PrepareBuildActions(cfg)
	if len(errs) != 0 {
		t.Errorf("Unexpected errors while preparing test & build actions: %s", errs)
	}
	buffer := new(bytes.Buffer)
	if err := ctx.WriteBuildFile(buffer); err != nil {
		t.Errorf("Error writing ninja file: %s", err)
	} else {
		text := buffer.String()
		t.Logf("Gennerated ninja build file:\n%s", text)
		if !strings.Contains(text, "build out/bin/test-out: g.gomodule.binaryBuild") {
			t.Errorf("Generated ninja file does not have build of the test module")
		}
		if !strings.Contains(text, " test-src.go") {
			t.Errorf("Generated ninja file does not have source dependency")
		}
		if !strings.Contains(text, "build vendor: g.gomodule.vendor | ../go.mod") {
			t.Errorf("Generated ninja file does not have vendor build rule")
		}
		if !strings.Contains(text, "build out/tests/test.txt: g.gomodule.binaryTest") {
			t.Errorf("Generated ninja file does not have test build rule")
		}
	}
}
