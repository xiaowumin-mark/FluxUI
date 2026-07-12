// Command coveragecheck enforces FluxUI's coverage non-regression baseline and
// reports coverage for changed executable lines when a Git base revision is
// provided.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type baseline struct {
	Version      int                        `json:"version"`
	Scope        string                     `json:"scope"`
	Packages     map[string]packageBaseline `json:"packages"`
	TotalMinimum float64                    `json:"total_minimum"`
	Target       float64                    `json:"target"`
	Tolerance    float64                    `json:"tolerance"`
}

type packageBaseline struct {
	Profile string  `json:"profile"`
	Minimum float64 `json:"minimum"`
}

type coverageMetric struct {
	Covered    int     `json:"covered_statements"`
	Statements int     `json:"statements"`
	Percent    float64 `json:"percent"`
}

type report struct {
	Packages map[string]coverageMetric `json:"packages"`
	Total    coverageMetric            `json:"total"`
	Changed  *coverageMetric           `json:"changed,omitempty"`
}

type profileBlock struct {
	File       string
	StartLine  int
	EndLine    int
	Statements int
	Count      int
}

var profileLocation = regexp.MustCompile(`^(.+):(\d+)\.\d+,(\d+)\.\d+$`)
var addedHunk = regexp.MustCompile(`\+(\d+)(?:,(\d+))?`)

func main() {
	profilesDir := flag.String("profiles", "coverage", "directory containing profile files")
	baselinePath := flag.String("baseline", "ci/coverage-baseline.json", "coverage baseline file")
	baseRevision := flag.String("base", "", "Git base revision for changed-line coverage")
	changedMinimum := flag.Float64("changed-minimum", 80, "minimum changed executable-line coverage percentage")
	reportPath := flag.String("report", "", "optional JSON report path")
	write := flag.Bool("write", false, "update the committed non-regression baseline")
	flag.Parse()

	policy, err := readBaseline(*baselinePath)
	if err != nil {
		fatal(err)
	}
	if policy.Version != 1 {
		fatal(fmt.Errorf("unsupported coverage baseline version %d", policy.Version))
	}

	result, blocks, err := collect(*profilesDir, policy)
	if err != nil {
		fatal(err)
	}
	var failures []string
	if *write {
		for name, metric := range result.Packages {
			entry := policy.Packages[name]
			entry.Minimum = roundPercent(metric.Percent)
			policy.Packages[name] = entry
		}
		policy.TotalMinimum = roundPercent(result.Total.Percent)
		if err := writeBaseline(*baselinePath, policy); err != nil {
			fatal(err)
		}
	} else {
		failures = append(failures, nonRegressionFailures(policy, result)...)
	}

	if *baseRevision != "" {
		changed, err := changedLines(*baseRevision)
		if err != nil {
			fatal(err)
		}
		metric := changedCoverage(changed, blocks)
		result.Changed = &metric
		if metric.Statements > 0 && metric.Percent+policy.Tolerance < *changedMinimum {
			failures = append(failures, fmt.Sprintf("changed executable-line coverage %.2f%% is below %.2f%%", metric.Percent, *changedMinimum))
		}
	}

	if *reportPath != "" {
		if err := writeReport(*reportPath, result); err != nil {
			fatal(err)
		}
	}
	if len(failures) > 0 {
		for _, failure := range failures {
			fmt.Fprintln(os.Stderr, "coveragecheck:", failure)
		}
		os.Exit(1)
	}
}

func readBaseline(path string) (baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return baseline{}, fmt.Errorf("read baseline %s: %w", path, err)
	}
	var policy baseline
	if err := json.Unmarshal(data, &policy); err != nil {
		return baseline{}, fmt.Errorf("parse baseline %s: %w", path, err)
	}
	if len(policy.Packages) == 0 {
		return baseline{}, errors.New("coverage baseline has no packages")
	}
	return policy, nil
}

func writeBaseline(path string, policy baseline) error {
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal baseline: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write baseline %s: %w", path, err)
	}
	return nil
}

func collect(profilesDir string, policy baseline) (report, []profileBlock, error) {
	result := report{Packages: make(map[string]coverageMetric, len(policy.Packages))}
	blocks := make([]profileBlock, 0)
	names := make([]string, 0, len(policy.Packages))
	for name := range policy.Packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		entry := policy.Packages[name]
		if entry.Profile == "" {
			return report{}, nil, fmt.Errorf("baseline package %q has no profile", name)
		}
		profileBlocks, err := parseProfile(filepath.Join(profilesDir, entry.Profile))
		if err != nil {
			return report{}, nil, err
		}
		metric := metricFor(profileBlocks)
		result.Packages[name] = metric
		result.Total.Covered += metric.Covered
		result.Total.Statements += metric.Statements
		blocks = append(blocks, profileBlocks...)
	}
	result.Total.Percent = percent(result.Total.Covered, result.Total.Statements)
	return result, blocks, nil
}

func parseProfile(path string) ([]profileBlock, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open profile %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	blocks := make([]profileBlock, 0)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if lineNumber == 1 {
			if !strings.HasPrefix(line, "mode: ") {
				return nil, fmt.Errorf("profile %s has no mode header", path)
			}
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("profile %s line %d is malformed", path, lineNumber)
		}
		matches := profileLocation.FindStringSubmatch(fields[0])
		if len(matches) != 4 {
			return nil, fmt.Errorf("profile %s line %d has an invalid location", path, lineNumber)
		}
		startLine, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, err
		}
		endLine, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, err
		}
		statements, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, err
		}
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, profileBlock{
			File:       filepath.ToSlash(matches[1]),
			StartLine:  startLine,
			EndLine:    endLine,
			Statements: statements,
			Count:      count,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read profile %s: %w", path, err)
	}
	return blocks, nil
}

func metricFor(blocks []profileBlock) coverageMetric {
	metric := coverageMetric{}
	for _, block := range blocks {
		metric.Statements += block.Statements
		if block.Count > 0 {
			metric.Covered += block.Statements
		}
	}
	metric.Percent = percent(metric.Covered, metric.Statements)
	return metric
}

func percent(covered, statements int) float64 {
	if statements == 0 {
		return 100
	}
	return float64(covered) * 100 / float64(statements)
}

func roundPercent(value float64) float64 {
	return math.Round(value*100) / 100
}

func nonRegressionFailures(policy baseline, result report) []string {
	failures := make([]string, 0)
	for name, entry := range policy.Packages {
		metric := result.Packages[name]
		if metric.Percent+policy.Tolerance < entry.Minimum {
			failures = append(failures, fmt.Sprintf("%s coverage %.2f%% is below baseline %.2f%%", name, metric.Percent, entry.Minimum))
		}
	}
	if result.Total.Percent+policy.Tolerance < policy.TotalMinimum {
		failures = append(failures, fmt.Sprintf("total core coverage %.2f%% is below baseline %.2f%%", result.Total.Percent, policy.TotalMinimum))
	}
	return failures
}

func changedLines(base string) (map[string]map[int]struct{}, error) {
	command := exec.Command("git", "diff", "--unified=0", base+"...HEAD", "--", "*.go")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read changed Go lines from %s: %w", base, err)
	}
	changed := make(map[string]map[int]struct{})
	path := ""
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "+++ ") {
			path = strings.TrimPrefix(line, "+++ b/")
			if path == "/dev/null" {
				path = ""
			}
			path = filepath.ToSlash(path)
			continue
		}
		if path == "" || !strings.HasPrefix(line, "@@") {
			continue
		}
		match := addedHunk.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		start, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, err
		}
		count := 1
		if match[2] != "" {
			count, err = strconv.Atoi(match[2])
			if err != nil {
				return nil, err
			}
		}
		if count == 0 {
			continue
		}
		if changed[path] == nil {
			changed[path] = make(map[int]struct{})
		}
		for lineNumber := start; lineNumber < start+count; lineNumber++ {
			changed[path][lineNumber] = struct{}{}
		}
	}
	return changed, nil
}

func changedCoverage(changed map[string]map[int]struct{}, blocks []profileBlock) coverageMetric {
	selected := make([]profileBlock, 0)
	for _, block := range blocks {
		for path, lines := range changed {
			if !profileMatchesPath(block.File, path) {
				continue
			}
			for line := range lines {
				if line >= block.StartLine && line <= block.EndLine {
					selected = append(selected, block)
					break
				}
			}
			break
		}
	}
	return metricFor(selected)
}

func profileMatchesPath(profilePath, changedPath string) bool {
	profilePath = filepath.ToSlash(profilePath)
	changedPath = filepath.ToSlash(changedPath)
	return profilePath == changedPath || strings.HasSuffix(profilePath, "/"+changedPath)
}

func writeReport(path string, result report) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "coveragecheck:", err)
	os.Exit(1)
}
