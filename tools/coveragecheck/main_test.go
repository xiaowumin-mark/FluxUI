package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProfileAndChangedCoverage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "package.out")
	data := "mode: atomic\ngithub.com/example/project/widget/sample.go:10.1,12.2 3 1\ngithub.com/example/project/widget/sample.go:20.1,21.2 2 0\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	blocks, err := parseProfile(path)
	if err != nil {
		t.Fatal(err)
	}
	metric := changedCoverage(map[string]map[int]struct{}{
		"widget/sample.go": {11: {}, 20: {}},
	}, blocks)
	if metric.Statements != 5 || metric.Covered != 3 || metric.Percent != 60 {
		t.Fatalf("changed metric = %#v", metric)
	}
	if !profileMatchesPath(blocks[0].File, "widget/sample.go") {
		t.Fatal("profile path did not match source path")
	}
}
