// Command benchgate compares benchmark samples from the base and candidate
// revisions and applies the committed regression and frame-budget policy.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type policy struct {
	Version                  int                        `json:"version"`
	MaximumRegressionPercent float64                    `json:"maximum_regression_percent"`
	Benchmarks               map[string]benchmarkPolicy `json:"benchmarks"`
}

type benchmarkPolicy struct {
	Metric        string  `json:"metric"`
	FrameBudgetNS float64 `json:"frame_budget_ns"`
}

func main() {
	configPath := flag.String("config", "ci/benchmark-gate.json", "benchmark gate policy")
	basePath := flag.String("base", "", "base benchmark output")
	candidatePath := flag.String("candidate", "", "candidate benchmark output")
	flag.Parse()
	if *basePath == "" || *candidatePath == "" {
		fatal("-base and -candidate are required")
	}
	config, err := readPolicy(*configPath)
	if err != nil {
		fatal(err.Error())
	}
	base, err := parseBenchmarks(*basePath)
	if err != nil {
		fatal(err.Error())
	}
	candidate, err := parseBenchmarks(*candidatePath)
	if err != nil {
		fatal(err.Error())
	}

	names := make([]string, 0, len(config.Benchmarks))
	for name := range config.Benchmarks {
		names = append(names, name)
	}
	sort.Strings(names)
	failures := make([]string, 0)
	for _, name := range names {
		entry := config.Benchmarks[name]
		baseValue, baseOK := median(base[name][entry.Metric])
		candidateValue, candidateOK := median(candidate[name][entry.Metric])
		if !baseOK || !candidateOK {
			failures = append(failures, fmt.Sprintf("%s did not produce metric %q in both samples", name, entry.Metric))
			continue
		}
		regression := regressionPercent(baseValue, candidateValue)
		fmt.Printf("%s %s: base=%.3f candidate=%.3f regression=%+.2f%%\n", name, entry.Metric, baseValue, candidateValue, regression)
		if regression > config.MaximumRegressionPercent {
			failures = append(failures, fmt.Sprintf("%s regressed %.2f%% (limit %.2f%%)", name, regression, config.MaximumRegressionPercent))
		}
		if entry.FrameBudgetNS > 0 && candidateValue > entry.FrameBudgetNS {
			failures = append(failures, fmt.Sprintf("%s %s %.3f ns exceeds %.3f ns frame budget", name, entry.Metric, candidateValue, entry.FrameBudgetNS))
		}
	}
	if len(failures) > 0 {
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, "benchgate:", failure)
		}
		os.Exit(1)
	}
}

func readPolicy(path string) (policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return policy{}, fmt.Errorf("read policy %s: %w", path, err)
	}
	var config policy
	if err := json.Unmarshal(data, &config); err != nil {
		return policy{}, fmt.Errorf("parse policy %s: %w", path, err)
	}
	if config.Version != 1 || config.MaximumRegressionPercent <= 0 || len(config.Benchmarks) == 0 {
		return policy{}, fmt.Errorf("policy %s is incomplete", path)
	}
	return config, nil
}

func parseBenchmarks(path string) (map[string]map[string][]float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open benchmark output %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	benchmarks := make(map[string]map[string][]float64)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 || !strings.HasPrefix(fields[0], "Benchmark") {
			continue
		}
		name := normalizeBenchmarkName(fields[0])
		metrics := benchmarks[name]
		if metrics == nil {
			metrics = make(map[string][]float64)
			benchmarks[name] = metrics
		}
		for index := 2; index+1 < len(fields); index += 2 {
			value, err := strconv.ParseFloat(fields[index], 64)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
				continue
			}
			metrics[fields[index+1]] = append(metrics[fields[index+1]], value)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read benchmark output %s: %w", path, err)
	}
	return benchmarks, nil
}

func normalizeBenchmarkName(name string) string {
	index := strings.LastIndex(name, "-")
	if index < 0 {
		return name
	}
	if _, err := strconv.Atoi(name[index+1:]); err == nil {
		return name[:index]
	}
	return name
}

func median(values []float64) (float64, bool) {
	if len(values) == 0 {
		return 0, false
	}
	values = append([]float64(nil), values...)
	sort.Float64s(values)
	middle := len(values) / 2
	if len(values)%2 == 1 {
		return values[middle], true
	}
	return (values[middle-1] + values[middle]) / 2, true
}

func regressionPercent(base, candidate float64) float64 {
	if base == 0 {
		if candidate == 0 {
			return 0
		}
		return math.Inf(1)
	}
	return (candidate - base) * 100 / base
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, "benchgate:", message)
	os.Exit(1)
}
