package internal

import (
	"image/color"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/theme"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func newTestContext() (*Runtime, *Context) {
	rt := NewRuntime(nil)
	var ops op.Ops
	gtx := gioLayout.Context{Ops: &ops}
	rt.BeginFrame()
	return rt, NewContext(gtx, rt)
}

func TestNextKeyFormat(t *testing.T) {
	_, ctx := newTestContext()
	k0 := ctx.NextKey("state")
	k1 := ctx.NextKey("state")

	if k0 != "root/state:0" {
		t.Fatalf("expected root/state:0, got %s", k0)
	}
	if k1 != "root/state:1" {
		t.Fatalf("expected root/state:1, got %s", k1)
	}
}

func TestNextKeyIncrementsIndex(t *testing.T) {
	_, ctx := newTestContext()
	for i := range 5 {
		k := ctx.NextKey("ns")
		expected := "root/ns:" + itoa(i)
		if k != expected {
			t.Fatalf("key %d: expected %s, got %s", i, expected, k)
		}
	}
}

func TestChildResetsHookIndex(t *testing.T) {
	_, ctx := newTestContext()
	ctx.NextKey("state") // root hookIndex=1

	child := ctx.Child(0)
	k := child.NextKey("state")
	if k != "root/0/state:0" {
		t.Fatalf("expected root/0/state:0, got %s", k)
	}
}

func TestScopeResetsHookIndex(t *testing.T) {
	_, ctx := newTestContext()
	ctx.NextKey("state")

	scoped := ctx.Scope("myWidget")
	k := scoped.NextKey("effect")
	if k != "root/myWidget/effect:0" {
		t.Fatalf("expected root/myWidget/effect:0, got %s", k)
	}
}

func TestChildPathIndependence(t *testing.T) {
	_, ctx := newTestContext()
	c0 := ctx.Child(0)
	c1 := ctx.Child(1)

	k0 := c0.NextKey("state")
	k1 := c1.NextKey("state")

	if k0 == k1 {
		t.Fatalf("children should have different paths: %s == %s", k0, k1)
	}
}

func TestPersistentCaches(t *testing.T) {
	_, ctx := newTestContext()
	callCount := 0

	v1 := ctx.Persistent("pkey", func() any {
		callCount++
		return "hello"
	})
	v2 := ctx.Persistent("pkey", func() any {
		callCount++
		return "world"
	})

	if callCount != 1 {
		t.Fatalf("expected factory called once, got %d", callCount)
	}
	if v1 != v2 {
		t.Fatal("expected same cached value")
	}
	if v1.(string) != "hello" {
		t.Fatalf("expected 'hello', got %v", v1)
	}
}

func TestMemoUsesNextKey(t *testing.T) {
	_, ctx := newTestContext()

	ctx.NextKey("state") // hookIndex 0 → 1
	ctx.Memo("memo", func() any { return 42 }) // hookIndex 1 → 2
	k := ctx.NextKey("state") // should be index 2

	if k != "root/state:2" {
		t.Fatalf("expected root/state:2, got %s — Memo should consume a hookIndex", k)
	}
}

func TestWithForegroundCopy(t *testing.T) {
	_, ctx := newTestContext()
	original := ctx.Foreground()

	newColor := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	modified := ctx.WithForeground(newColor)

	if modified.Foreground() != newColor {
		t.Fatal("modified context should have new foreground")
	}
	if ctx.Foreground() != original {
		t.Fatal("original context should be unchanged")
	}
}

func TestWithFontCopy(t *testing.T) {
	_, ctx := newTestContext()
	originalFont := ctx.Font()

	newSpec := theme.FontSpec{Family: "CustomFont"}
	modified := ctx.WithFont(newSpec)

	if modified.Font().Family != "CustomFont" {
		t.Fatalf("expected CustomFont, got %s", modified.Font().Family)
	}
	if ctx.Font() != originalFont {
		t.Fatal("original context font should be unchanged")
	}
}

func TestFontFallbackChain(t *testing.T) {
	_, ctx := newTestContext()
	f := ctx.Font()
	if f == (theme.FontSpec{}) {
		t.Fatal("expected non-zero font from theme default")
	}
}

func TestWindowMethodsNilController(t *testing.T) {
	_, ctx := newTestContext()

	if ctx.WindowID() != 0 {
		t.Fatal("expected 0 WindowID with nil controller")
	}
	if ctx.WindowClose() {
		t.Fatal("expected false")
	}
	if ctx.WindowMinimize() {
		t.Fatal("expected false")
	}
	if ctx.WindowMaximize() {
		t.Fatal("expected false")
	}
	if ctx.WindowRestore() {
		t.Fatal("expected false")
	}
	if ctx.WindowFullscreen() {
		t.Fatal("expected false")
	}
	if ctx.WindowRaise() {
		t.Fatal("expected false")
	}
	if ctx.WindowCenter() {
		t.Fatal("expected false")
	}
	if ctx.WindowSetTitle("test") {
		t.Fatal("expected false")
	}
	if ctx.WindowSetSize(100, 100) {
		t.Fatal("expected false")
	}
	if ctx.WindowInvalidate() {
		t.Fatal("expected false")
	}
	if ctx.WindowIsAlive() {
		t.Fatal("expected false")
	}
}

func TestTreePath(t *testing.T) {
	_, ctx := newTestContext()
	if ctx.TreePath() != "root" {
		t.Fatalf("expected 'root', got %s", ctx.TreePath())
	}
	child := ctx.Child(3)
	if child.TreePath() != "root/3" {
		t.Fatalf("expected 'root/3', got %s", child.TreePath())
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}
