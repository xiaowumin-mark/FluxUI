// Command gofmtcheck verifies that all repository Go files match gofmt. It is
// intentionally cross-platform so local Windows contributors and CI use the
// same check.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	write := flag.Bool("write", false, "rewrite files with gofmt")
	flag.Parse()
	paths, err := goFiles(".")
	if err != nil {
		fatal(err)
	}
	changed := make([]string, 0)
	for _, path := range paths {
		original, err := os.ReadFile(path)
		if err != nil {
			fatal(err)
		}
		formatted, err := format.Source(original)
		if err != nil {
			fatal(fmt.Errorf("format %s: %w", path, err))
		}
		if bytes.Equal(original, formatted) {
			continue
		}
		changed = append(changed, filepath.ToSlash(path))
		if *write {
			if err := os.WriteFile(path, formatted, 0o644); err != nil {
				fatal(err)
			}
		}
	}
	if len(changed) == 0 || *write {
		return
	}
	fmt.Fprintln(os.Stderr, "gofmtcheck: files are not formatted:")
	for _, path := range changed {
		fmt.Fprintln(os.Stderr, path)
	}
	fmt.Fprintln(os.Stderr, "run: go run ./tools/gofmtcheck -write")
	os.Exit(1)
}

func goFiles(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "coverage", "dist", "out":
				if path != root {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if filepath.Ext(path) == ".go" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "gofmtcheck:", err)
	os.Exit(1)
}
