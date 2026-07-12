package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBenchmarksAndMedian(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bench.txt")
	data := "BenchmarkWidget/subcase-8  10  100 ns/op  80 frame-ns/frame\nBenchmarkWidget/subcase-8  10  120 ns/op  90 frame-ns/frame\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	benchmarks, err := parseBenchmarks(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := median(benchmarks["BenchmarkWidget/subcase"]["ns/op"]); !ok || got != 110 {
		t.Fatalf("median ns/op = %v, ok=%t", got, ok)
	}
	if got := regressionPercent(100, 110); got != 10 {
		t.Fatalf("regression = %v", got)
	}
	if normalizeBenchmarkName("BenchmarkThing-12") != "BenchmarkThing" {
		t.Fatal("CPU suffix was not removed")
	}
}
