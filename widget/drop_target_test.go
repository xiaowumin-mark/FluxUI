package widget

import (
	"errors"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gioui.org/io/transfer"
)

func TestNormalizeDropTypes(t *testing.T) {
	got := normalizeDropTypes([]string{" text/plain ", "TEXT/PLAIN", "", "text/uri-list"})
	want := []string{"text/plain", "text/uri-list"}
	if len(got) != len(want) {
		t.Fatalf("unexpected type count: got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("type %d: got %q want %q", i, got[i], want[i])
		}
	}
	if len(normalizeDropTypes(nil)) == 0 {
		t.Fatal("nil drop types should restore defaults")
	}
}

func TestParseDropPathsFromURIList(t *testing.T) {
	input := strings.Join([]string{
		"# comment",
		"file:///C:/Users/me/report%20one.txt",
		"file://localhost/C:/Users/me/report%20two.txt",
		"https://example.com/not-a-file",
		"",
	}, "\r\n")
	paths := parseDropPaths("text/uri-list", []byte(input))
	if len(paths) != 2 {
		t.Fatalf("expected two paths, got %#v", paths)
	}
	if runtime.GOOS == "windows" {
		if paths[0] != filepath.Clean(`C:\Users\me\report one.txt`) {
			t.Fatalf("unexpected first Windows path: %q", paths[0])
		}
	} else if !strings.HasSuffix(paths[0], filepath.FromSlash("C:/Users/me/report one.txt")) {
		t.Fatalf("unexpected first path: %q", paths[0])
	}
}

func TestParseDropPathsFromPlainText(t *testing.T) {
	abs, err := filepath.Abs("fluxui-drop.txt")
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	paths := parseDropPaths("text/plain", []byte(abs+"\nnot-absolute"))
	if len(paths) != 1 || paths[0] != filepath.Clean(abs) {
		t.Fatalf("unexpected text paths: %#v", paths)
	}
}

func TestReadDropDataLimit(t *testing.T) {
	data, err := readDropData(strings.NewReader("hello"), 8)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected data: %q", data)
	}
	if _, err := readDropData(strings.NewReader("toolong"), 3); err == nil {
		t.Fatal("expected limit error")
	}
}

func TestDropEventFromTransfer(t *testing.T) {
	event := transfer.DataEvent{
		Type: "text/plain",
		Open: func() io.ReadCloser {
			return io.NopCloser(strings.NewReader("hello"))
		},
	}
	got := dropEventFromTransfer(event, 16, DragOperationMove)
	if got.Err != nil ||
		got.Type != "text/plain" ||
		got.Text != "hello" ||
		string(got.Data) != "hello" ||
		got.Operation != DragOperationMove {
		t.Fatalf("unexpected drop event: %#v", got)
	}

	errEvent := dropEventFromTransfer(transfer.DataEvent{Type: "text/plain"}, 16, DragOperationCopy)
	if errEvent.Err == nil {
		t.Fatal("expected missing Open to produce an error")
	}
	if errors.Is(errEvent.Err, io.EOF) {
		t.Fatalf("unexpected EOF classification: %v", errEvent.Err)
	}
}
