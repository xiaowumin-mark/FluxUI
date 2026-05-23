package internal

import (
	"strings"
	"testing"
)

func TestComponentIdentityStableIDUsesKey(t *testing.T) {
	id0 := ComponentIdentity{ParentID: "root", TypeID: "Todo", Key: "a", Position: 0}.StableID()
	id1 := ComponentIdentity{ParentID: "root", TypeID: "Todo", Key: "a", Position: 99}.StableID()
	id2 := ComponentIdentity{ParentID: "root", TypeID: "Todo", Key: "b", Position: 0}.StableID()

	if id0 != id1 {
		t.Fatalf("keyed identity should ignore position: %s != %s", id0, id1)
	}
	if id0 == id2 {
		t.Fatalf("different keys should produce different ids: %s", id0)
	}
}

func TestComponentIdentityStableIDFallsBackToPosition(t *testing.T) {
	id0 := ComponentIdentity{ParentID: "root", TypeID: "Row", Position: 0}.StableID()
	id1 := ComponentIdentity{ParentID: "root", TypeID: "Row", Position: 1}.StableID()

	if id0 == id1 {
		t.Fatalf("unkeyed siblings should use position fallback: %s", id0)
	}
}

func TestHookStoreReusesInstanceAndSlots(t *testing.T) {
	store := NewHookStore()
	identity := ComponentIdentity{ParentID: "root", TypeID: "Counter", Key: "main"}

	store.BeginFrame()
	first := store.BeginInstance(identity)
	stateSlot := first.NextHook(HookState)
	stateSlot.Value = 41
	stateSlot.Initialized = true
	if first.HookCount() != 1 {
		t.Fatalf("expected one consumed hook, got %d", first.HookCount())
	}
	store.EndFrame()

	store.BeginFrame()
	second := store.BeginInstance(identity)
	if second != first {
		t.Fatal("expected same instance for same identity")
	}
	stateSlot = second.NextHook(HookState)
	if got := stateSlot.Value.(int); got != 41 {
		t.Fatalf("expected persisted slot value 41, got %d", got)
	}
	store.EndFrame()
}

func TestComponentInstanceHookKindMismatchPanics(t *testing.T) {
	store := NewHookStore()
	identity := ComponentIdentity{ParentID: "root", TypeID: "Switcher", Key: "main"}

	store.BeginFrame()
	inst := store.BeginInstance(identity)
	inst.NextHook(HookState)
	store.EndFrame()

	store.BeginFrame()
	inst = store.BeginInstance(identity)
	defer func() {
		r := recover()
		store.EndFrame()
		if r == nil {
			t.Fatal("expected panic when hook kind changes")
		}
		if msg := r.(string); !strings.Contains(msg, "rendered hook #0 as") {
			t.Fatalf("unexpected panic message: %s", msg)
		}
	}()
	inst.NextHook(HookAnimValue)
}

func TestHookStoreUnmountRunsCleanups(t *testing.T) {
	store := NewHookStore()
	identity := ComponentIdentity{ParentID: "root", TypeID: "Panel", Key: "settings"}
	cleanupCount := 0

	store.BeginFrame()
	inst := store.BeginInstance(identity)
	inst.NextHook(HookEffect).Cleanup = func() { cleanupCount++ }
	store.EndFrame()

	store.BeginFrame()
	store.EndFrame()

	if cleanupCount != 1 {
		t.Fatalf("expected cleanup once on unmount, got %d", cleanupCount)
	}
	if store.Instance(identity.StableID()) != nil {
		t.Fatal("expected unmounted instance to be removed")
	}
}

func TestHookStoreDisposeRunsCleanupsOnce(t *testing.T) {
	store := NewHookStore()
	cleanupCount := 0

	store.BeginFrame()
	inst := store.BeginInstance(ComponentIdentity{ParentID: "root", TypeID: "Dialog", Key: "confirm"})
	inst.NextHook(HookEffect).Cleanup = func() { cleanupCount++ }
	store.EndFrame()

	store.Dispose()
	store.Dispose()

	if cleanupCount != 1 {
		t.Fatalf("expected cleanup once after repeated dispose, got %d", cleanupCount)
	}
}

func TestRuntimeOwnsHookStoreFrameLifecycle(t *testing.T) {
	rt := NewRuntime(nil)
	identity := ComponentIdentity{ParentID: "root", TypeID: "Toast", Key: "one"}
	cleanupCount := 0

	rt.BeginFrame()
	inst := rt.HookStore().BeginInstance(identity)
	inst.NextHook(HookEffect).Cleanup = func() { cleanupCount++ }
	rt.EndFrame()

	rt.BeginFrame()
	rt.EndFrame()

	if cleanupCount != 1 {
		t.Fatalf("expected runtime EndFrame to unmount inactive hook instance, got %d", cleanupCount)
	}
}
