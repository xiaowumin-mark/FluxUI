package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestP7ExampleDocsMatchRunElementEntries(t *testing.T) {
	cases := []struct {
		dir      string
		smokeID  string
		coverage string
	}{
		{dir: "advanced_components", smokeID: "AC-01", coverage: "Select"},
		{dir: "form_validation", smokeID: "FV-01", coverage: "TextFieldElement"},
		{dir: "horizontal_scroll", smokeID: "HS-01", coverage: "ScrollHorizontal(true)"},
		{dir: "virtual_scroll", smokeID: "VS-01", coverage: "ListViewElement"},
		{dir: "drag_drop_showcase", smokeID: "DD-01", coverage: "DragSourceElement"},
	}

	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			exampleDir := filepath.Join("..", tc.dir)
			mainText := readP7DocFile(t, filepath.Join(exampleDir, "main.go"))
			if !strings.Contains(mainText, "ui.RunElement(") {
				t.Fatalf("%s main.go should use ui.RunElement", tc.dir)
			}

			readmeText := readP7DocFile(t, filepath.Join(exampleDir, "README.md"))
			for _, token := range []string{
				"go run ./examples/" + tc.dir,
				"## P7 Smoke",
				tc.smokeID,
				tc.coverage,
			} {
				if !strings.Contains(readmeText, token) {
					t.Fatalf("%s README.md missing %q", tc.dir, token)
				}
			}
			for _, stale := range []string{
				"remains a legacy `Run` / `Widget`",
				"Do not rewrite this example",
			} {
				if strings.Contains(readmeText, stale) {
					t.Fatalf("%s README.md still contains stale legacy wording %q", tc.dir, stale)
				}
			}
		})
	}
}

func TestP7RegressionDocsExposeManualAndBenchmarkEntrypoints(t *testing.T) {
	checks := []struct {
		path   string
		tokens []string
	}{
		{
			path: "README.md",
			tokens: []string{
				"## P7 Smoke",
				"DB-03",
				"example popup",
			},
		},
		{
			path: filepath.Join("..", "component_lab", "README.md"),
			tokens: []string{
				"## P7 Smoke",
				"CL-04",
				"cursor",
			},
		},
		{
			path: filepath.Join("..", "event_system_testbench", "README.md"),
			tokens: []string{
				"## P0-P7 手工验收",
				"P7",
				"event type",
				"default/cancel",
			},
		},
		{
			path: filepath.Join("..", "..", "README.md"),
			tokens: []string{
				"P7 回归重点入口",
				"Benchmark(WheelScrollViewVertical|HorizontalWheelDelta)",
				"Benchmark(LayoutStaticTree|MouseMoveInteractiveTree|ListVirtualized|StaticSurfaceCache)",
			},
		},
		{
			path: filepath.Join("..", "..", "docs", "examples-inventory.md"),
			tokens: []string{
				"| `examples/advanced_components` | Higher-level component integration showcase | React-style `RunElement`",
				"| `examples/form_validation` | Complex input workflow | React-style `RunElement`",
				"| `examples/horizontal_scroll` | Scroll behavior demo | React-style `RunElement`",
				"| `examples/virtual_scroll` | Virtual list/grid performance reference | React-style `RunElement`",
			},
		},
	}

	for _, check := range checks {
		t.Run(filepath.ToSlash(check.path), func(t *testing.T) {
			text := readP7DocFile(t, check.path)
			for _, token := range check.tokens {
				if !strings.Contains(text, token) {
					t.Fatalf("%s missing %q", check.path, token)
				}
			}
		})
	}
}

func readP7DocFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
