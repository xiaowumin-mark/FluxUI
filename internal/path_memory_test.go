package internal

import "testing"

func TestPathAndMemoryKeysAreStableAndReadable(t *testing.T) {
	runtime := NewRuntime(nil)
	child := runtime.childPath(0, 2)
	if child == 0 || child != runtime.childPath(rootPathID, 2) {
		t.Fatalf("child path was not stable: %d", child)
	}
	scope := runtime.scopePath(child, "panel")
	if scope == 0 || scope != runtime.scopePath(child, "panel") {
		t.Fatalf("scope path was not stable: %d", scope)
	}
	if child == runtime.childPath(0, 3) || scope == runtime.scopePath(child, "other") {
		t.Fatal("distinct path segments reused an identity")
	}
	if got := runtime.DebugPath(0); got != "root" {
		t.Fatalf("root path = %q, want root", got)
	}
	if got := runtime.DebugPath(scope); got != "root/2/panel" {
		t.Fatalf("scope path = %q, want root/2/panel", got)
	}
	if got := runtime.DebugPath(999); got != "path#999" {
		t.Fatalf("unknown path = %q", got)
	}

	structured := MemoryKey{Path: scope, Namespace: "state", Slot: 3}
	if !structured.valid() || runtime.DebugMemoryKey(structured) != "root/2/panel/state:3" {
		t.Fatalf("unexpected structured key: %#v / %q", structured, runtime.DebugMemoryKey(structured))
	}
	noSlot := MemoryKey{Path: scope, Namespace: "cache", NoSlot: true}
	if got := runtime.DebugMemoryKey(noSlot); got != "root/2/panel/cache" {
		t.Fatalf("no-slot key = %q", got)
	}
	if got := runtime.DebugMemoryKey(MemoryKey{Opaque: "legacy"}); got != "legacy" {
		t.Fatalf("opaque key = %q", got)
	}
	if got := runtime.DebugMemoryKey(MemoryKey{}); got != "" {
		t.Fatalf("invalid key = %q", got)
	}
	if memoryKeyString("").valid() || !memoryKeyString("legacy").valid() {
		t.Fatal("legacy memory key validity was incorrect")
	}
	if got := debugMemoryKeyWithoutRuntime(MemoryKey{Path: rootPathID, Namespace: "state", Slot: 1}, ""); got != "root/state:1" {
		t.Fatalf("fallback structured key = %q", got)
	}
	if got := debugMemoryKeyWithoutRuntime(MemoryKey{Path: rootPathID, Namespace: "cache", NoSlot: true}, "custom"); got != "custom/cache" {
		t.Fatalf("fallback no-slot key = %q", got)
	}
	if got := debugMemoryKeyWithoutRuntime(MemoryKey{Opaque: "legacy"}, "custom"); got != "legacy" {
		t.Fatalf("fallback opaque key = %q", got)
	}

	if normalizePathID(0) != rootPathID || normalizePathID(child) != child {
		t.Fatal("path normalization was incorrect")
	}
	if joinPath("", "child") != "child" || joinPath("parent", "") != "parent" || joinPath("parent", "child") != "parent/child" {
		t.Fatal("path joining was incorrect")
	}

	var nilRuntime *Runtime
	if nilRuntime.childPath(rootPathID, 0) != 0 || nilRuntime.scopePath(rootPathID, "scope") != 0 || nilRuntime.DebugPath(rootPathID) != "" {
		t.Fatal("nil runtime path helpers should be inert")
	}
	if got := nilRuntime.DebugMemoryKey(MemoryKey{Path: rootPathID, Namespace: "state", Slot: 0}); got != "root/state:0" {
		t.Fatalf("nil runtime memory key = %q", got)
	}
}
