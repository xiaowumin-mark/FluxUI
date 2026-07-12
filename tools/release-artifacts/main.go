// Command release-artifacts writes and verifies SHA-256 checksums for release
// artifacts. The same verifier runs before a GitHub release is published.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const checksumFile = "SHA256SUMS"

func main() {
	directory := flag.String("dir", "dist", "artifact directory")
	write := flag.Bool("write", false, "write SHA256SUMS")
	verify := flag.Bool("verify", false, "verify SHA256SUMS")
	flag.Parse()
	if *write == *verify {
		fatal(errors.New("specify exactly one of -write or -verify"))
	}
	if *write {
		if err := writeChecksums(*directory); err != nil {
			fatal(err)
		}
		return
	}
	if err := verifyChecksums(*directory); err != nil {
		fatal(err)
	}
}

func writeChecksums(directory string) error {
	artifacts, err := artifactsIn(directory)
	if err != nil {
		return err
	}
	if len(artifacts) == 0 {
		return fmt.Errorf("no release artifacts found in %s", directory)
	}
	lines := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		hash, err := sha256File(filepath.Join(directory, filepath.FromSlash(artifact)))
		if err != nil {
			return err
		}
		lines = append(lines, hex.EncodeToString(hash)+"  "+artifact)
	}
	return os.WriteFile(filepath.Join(directory, checksumFile), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func verifyChecksums(directory string) error {
	data, err := os.ReadFile(filepath.Join(directory, checksumFile))
	if err != nil {
		return fmt.Errorf("read %s: %w", checksumFile, err)
	}
	expected := make(map[string]string)
	for lineNumber, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != sha256.Size*2 || parts[1] == "" {
			return fmt.Errorf("%s line %d is malformed", checksumFile, lineNumber+1)
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return fmt.Errorf("%s line %d has an invalid hash: %w", checksumFile, lineNumber+1, err)
		}
		if _, exists := expected[parts[1]]; exists {
			return fmt.Errorf("%s lists %q twice", checksumFile, parts[1])
		}
		expected[parts[1]] = parts[0]
	}
	artifacts, err := artifactsIn(directory)
	if err != nil {
		return err
	}
	if len(artifacts) != len(expected) {
		return fmt.Errorf("%s does not match the artifact set", checksumFile)
	}
	for _, artifact := range artifacts {
		hash, err := sha256File(filepath.Join(directory, filepath.FromSlash(artifact)))
		if err != nil {
			return err
		}
		expectedHash, ok := expected[artifact]
		if !ok {
			return fmt.Errorf("%s is missing checksum for %q", checksumFile, artifact)
		}
		if hex.EncodeToString(hash) != expectedHash {
			return fmt.Errorf("checksum mismatch for %q", artifact)
		}
	}
	return nil
}

func artifactsIn(directory string) ([]string, error) {
	artifacts := make([]string, 0)
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == checksumFile {
			return nil
		}
		artifacts = append(artifacts, relative)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan artifacts in %s: %w", directory, err)
	}
	sort.Strings(artifacts)
	return artifacts, nil
}

func sha256File(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "release-artifacts:", err)
	os.Exit(1)
}
