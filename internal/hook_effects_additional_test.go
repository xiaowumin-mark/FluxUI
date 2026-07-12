package internal

import "testing"

func TestHookStoreEffectsDependenciesAndCleanup(t *testing.T) {
	if !ShouldRunHookEffect(nil, true, nil) {
		t.Fatal("nil effect slot should run")
	}
	if !ShouldRunHookEffect(&HookSlot{}, true, nil) {
		t.Fatal("uninitialized effect slot should run")
	}
	if !ShouldRunHookEffect(&HookSlot{Initialized: true, HasDeps: true, Deps: []any{1}}, false, []any{1}) {
		t.Fatal("effect without dependency tracking should run")
	}
	if !ShouldRunHookEffect(&HookSlot{Initialized: true}, true, []any{1}) {
		t.Fatal("effect that previously had no dependencies should run")
	}
	if ShouldRunHookEffect(&HookSlot{Initialized: true, HasDeps: true, Deps: []any{1}}, true, []any{1}) {
		t.Fatal("effect with unchanged dependencies should not run")
	}
	if !ShouldRunHookEffect(&HookSlot{Initialized: true, HasDeps: true, Deps: []any{1}}, true, []any{2}) {
		t.Fatal("effect with changed dependencies should run")
	}
	if !DepsEqual([]any{map[string]int{"one": 1}}, []any{map[string]int{"one": 1}}) || DepsEqual([]any{1}, []any{1, 2}) || DepsEqual([]any{1}, []any{2}) {
		t.Fatal("dependency equality was incorrect")
	}
	deps := []any{1, "two"}
	cloned := CloneDeps(deps)
	if len(cloned) != 2 || cloned[0] != 1 || cloned[1] != "two" {
		t.Fatalf("cloned dependencies = %#v", cloned)
	}
	cloned[0] = 3
	if deps[0] != 1 || CloneDeps(nil) != nil || CloneDeps([]any{}) != nil {
		t.Fatal("dependency cloning was not isolated")
	}

	store := NewHookStore()
	slot := &HookSlot{}
	runs := 0
	cleanups := 0
	setup := func() func() {
		runs++
		return func() { cleanups++ }
	}
	store.UseEffect(slot, true, []any{1}, setup)
	store.EndFrame()
	if runs != 1 || cleanups != 0 || !slot.Initialized || !slot.HasDeps {
		t.Fatalf("initial hook effect state: runs=%d cleanups=%d slot=%#v", runs, cleanups, slot)
	}

	store.BeginFrame()
	store.UseEffect(slot, true, []any{1}, setup)
	store.EndFrame()
	if runs != 1 || cleanups != 0 {
		t.Fatalf("unchanged hook effect restarted: runs=%d cleanups=%d", runs, cleanups)
	}

	store.BeginFrame()
	store.UseEffect(slot, true, []any{2}, setup)
	store.EndFrame()
	if runs != 2 || cleanups != 1 {
		t.Fatalf("changed hook effect result: runs=%d cleanups=%d", runs, cleanups)
	}
	store.Dispose()
	if cleanups != 1 {
		t.Fatalf("untracked hook slot was cleaned by store dispose: %d", cleanups)
	}
	runHookCleanups(&ComponentInstance{hooks: []HookSlot{{Cleanup: func() { cleanups++ }}, {}}})
	if cleanups != 2 {
		t.Fatalf("component hook cleanup count = %d, want 2", cleanups)
	}

	store.UseEffect(nil, true, nil, setup)
	store.UseEffect(slot, true, nil, nil)
	var nilStore *HookStore
	nilStore.BeginFrame()
	nilStore.EndFrame()
	nilStore.Dispose()
	nilStore.UseEffect(slot, true, nil, setup)
}
