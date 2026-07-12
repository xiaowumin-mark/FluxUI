package internal

import (
	"testing"

	"github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestContextScopedValuesProvidersAndMemory(t *testing.T) {
	var ops op.Ops
	runtime := NewRuntime(nil)
	context := NewContext(gioLayout.Context{Ops: &ops}, runtime)

	intKey := ProviderKeyFor[int](1, "number")
	otherIntKey := ProviderKeyFor[int](2, "number")
	provided := WithProviderKeyValue(context, intKey, 42)
	if got := ProviderKeyValue(provided, intKey, 0); got != 42 {
		t.Fatalf("provider value = %d, want 42", got)
	}
	if got := ProviderKeyValue(context, intKey, 7); got != 7 {
		t.Fatalf("parent provider value = %d, want fallback 7", got)
	}
	if got := ProviderKeyValue(provided, otherIntKey, 7); got != 7 {
		t.Fatalf("different provider key = %d, want fallback 7", got)
	}
	legacy := WithProviderValue(provided, "hello")
	if got := ProviderValue(legacy, "fallback"); got != "hello" {
		t.Fatalf("legacy provider value = %q", got)
	}
	if got := ProviderKeyValue[string](provided, ProviderKeyFor[string](0, ""), "fallback"); got != "fallback" {
		t.Fatalf("missing provider type returned %q", got)
	}

	key := context.ScopeMemoryKey("cache")
	if got := context.DebugMemoryKey(key); got != "root/cache" {
		t.Fatalf("scope memory key = %q", got)
	}
	if got := context.PersistentKey(key, func() any { return "value" }); got != "value" {
		t.Fatalf("persistent value = %#v", got)
	}
	if got, ok := context.PersistentValueKey(key); !ok || got != "value" {
		t.Fatalf("persistent lookup = %#v, %t", got, ok)
	}
	context.ForgetPersistentKey(key)
	if _, ok := context.PersistentValueKey(key); ok {
		t.Fatal("forgotten persistent value remained cached")
	}
	context.Persistent("legacy", func() any { return 9 })
	if got, ok := context.PersistentValue("legacy"); !ok || got != 9 {
		t.Fatalf("legacy persistent lookup = %#v, %t", got, ok)
	}
	context.ForgetPersistent("legacy")

	instance := &ComponentInstance{ID: "instance"}
	componentContext := context.WithComponentInstance(instance)
	if componentContext.ComponentInstance() != instance {
		t.Fatal("component instance was not scoped")
	}
	if slot := componentContext.NextHookSlot("test"); slot == nil || instance.HookCount() != 1 {
		t.Fatalf("component hook slot = %#v, count=%d", slot, instance.HookCount())
	}

	if WithProviderValue[int](nil, 1) != nil || WithProviderKeyValue[int](nil, intKey, 1) != nil {
		t.Fatal("nil context accepted provider values")
	}
	if ProviderValue[int](nil, 5) != 5 || ProviderKeyValue[int](nil, intKey, 6) != 6 {
		t.Fatal("nil context provider fallback was incorrect")
	}
	var nilContext *Context
	if nilContext.ComponentInstance() != nil || nilContext.NextHookSlot("test") != nil {
		t.Fatal("nil context state helpers should be inert")
	}
	if _, ok := nilContext.PersistentValue("x"); ok {
		t.Fatal("nil context persistent lookup reported a value")
	}
}

func TestContextScopedThemeTextAndPointerOptions(t *testing.T) {
	var ops op.Ops
	context := NewContext(gioLayout.Context{Ops: &ops}, NewRuntime(nil))
	dark := theme.DarkTheme()
	scopedTheme := context.WithTheme(dark)
	if scopedTheme.Theme() != dark || context.Theme() == dark {
		t.Fatal("theme override leaked into the original context")
	}
	context.SetThemeOverride(dark)
	if context.Theme() != dark {
		t.Fatal("direct theme override was not applied")
	}
	context.SetThemeOverride(nil)
	if context.Theme() == nil {
		t.Fatal("clearing theme override removed the runtime theme")
	}

	if _, ok := context.WithTextStyle(theme.TextStyle{}).TextStyle(); ok {
		t.Fatal("zero text style was marked present")
	}
	style := theme.TextStyle{Size: 14, LineHeight: 20}
	if got, ok := context.WithTextStyle(style).TextStyle(); !ok || got != style {
		t.Fatalf("scoped text style = %#v, %t", got, ok)
	}
	passThrough := context.WithPointerPassThrough(true)
	if !passThrough.PointerPassThrough() || context.PointerPassThrough() {
		t.Fatal("pointer pass-through scope was not isolated")
	}
	if pop := passThrough.pushPointerPassThrough(&ops); pop == nil {
		t.Fatal("pointer pass-through did not create an operation")
	} else {
		pop()
	}
	if context.pushPointerPassThrough(nil) != nil {
		t.Fatal("context without pass-through returned an operation")
	}

	var nilContext *Context
	if nilContext.WithTheme(dark) != nil || nilContext.WithTextStyle(style) != nil || nilContext.WithPointerPassThrough(true) != nil || nilContext.PointerPassThrough() {
		t.Fatal("nil context presentation helpers should be inert")
	}
}
