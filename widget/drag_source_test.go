package widget

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNormalizeDragPayloads(t *testing.T) {
	payloads := normalizeDragPayloads([]DragPayload{
		{Type: " TEXT/PLAIN ", Data: []byte("hello")},
		{Type: "text/plain", Data: []byte("duplicate")},
		{Type: "", Data: []byte("ignored")},
		{Type: "application/text"},
	})
	if len(payloads) != 1 {
		t.Fatalf("expected one payload, got %#v", payloads)
	}
	if payloads[0].Type != "text/plain" || string(payloads[0].Data) != "hello" {
		t.Fatalf("unexpected payload: %#v", payloads[0])
	}

	payloads[0].Data[0] = 'X'
	second := normalizeDragPayloads([]DragPayload{{Type: "text/plain", Data: []byte("hello")}})
	if string(second[0].Data) != "hello" {
		t.Fatalf("payload data should be copied, got %q", second[0].Data)
	}
}

func TestDragSourceTextOption(t *testing.T) {
	var cfg dragSourceConfig
	DragSourceText("hello")(&cfg)
	payloads := normalizeDragPayloads(cfg.payloads)
	if len(payloads) != 3 {
		t.Fatalf("expected three text payloads, got %#v", payloads)
	}
	for _, payload := range payloads {
		if string(payload.Data) != "hello" {
			t.Fatalf("unexpected text payload: %#v", payload)
		}
	}
}

func TestDragSourceFilesOption(t *testing.T) {
	file := filepath.Join(t.TempDir(), "report one.txt")
	var cfg dragSourceConfig
	DragSourceFiles(file, file, " ")(&cfg)
	payloads := normalizeDragPayloads(cfg.payloads)
	if len(payloads) != 2 {
		t.Fatalf("expected uri-list and plain text payloads, got %#v", payloads)
	}
	if payloads[0].Type != "text/uri-list" {
		t.Fatalf("expected uri-list first, got %#v", payloads[0])
	}
	if !strings.HasPrefix(string(payloads[0].Data), "file://") {
		t.Fatalf("expected file URI payload, got %q", payloads[0].Data)
	}
	if strings.Count(string(payloads[1].Data), file) > 1 {
		t.Fatalf("expected duplicate file paths to be removed, got %q", payloads[1].Data)
	}
}

func TestNormalizeDragOperations(t *testing.T) {
	got := normalizeDragOperations([]DragOperation{
		DragOperationMove,
		DragOperationCopy,
		DragOperationMove,
		DragOperation("invalid"),
		"",
		DragOperationLink,
	})
	want := []DragOperation{DragOperationMove, DragOperationCopy, DragOperationLink}
	if len(got) != len(want) {
		t.Fatalf("unexpected operation count: got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("operation %d: got %q want %q", i, got[i], want[i])
		}
	}
	if got := firstDragOperation(nil); got != DragOperationCopy {
		t.Fatalf("expected default copy operation, got %q", got)
	}
}

func TestDragSourceOperationsOption(t *testing.T) {
	var cfg dragSourceConfig
	DragSourceOperations(DragOperationMove, DragOperationMove, DragOperationLink)(&cfg)
	if got := normalizeDragOperations(cfg.operations); len(got) != 2 || got[0] != DragOperationMove || got[1] != DragOperationLink {
		t.Fatalf("unexpected configured operations: %#v", got)
	}
}

func TestPathToFileURI(t *testing.T) {
	if runtime.GOOS == "windows" {
		got := pathToFileURI(`C:\Users\me\report one.txt`)
		if got != "file:///C:/Users/me/report%20one.txt" {
			t.Fatalf("unexpected Windows file URI: %q", got)
		}
		return
	}
	got := pathToFileURI("/tmp/report one.txt")
	if got != "file:///tmp/report%20one.txt" {
		t.Fatalf("unexpected file URI: %q", got)
	}
}

func TestDragFileURIList(t *testing.T) {
	list := dragFileURIList([]string{"/tmp/a.txt", "/tmp/b.txt"})
	if !strings.Contains(list, "\r\n") {
		t.Fatalf("expected CRLF separated URI list, got %q", list)
	}
}
