package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndVerifyChecksums(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "linux.bin"), []byte("linux"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(directory, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "nested", "windows.exe"), []byte("windows"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeChecksums(directory); err != nil {
		t.Fatalf("write checksums: %v", err)
	}
	if err := verifyChecksums(directory); err != nil {
		t.Fatalf("verify checksums: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "linux.bin"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verifyChecksums(directory); err == nil {
		t.Fatal("expected a checksum mismatch")
	}
}
