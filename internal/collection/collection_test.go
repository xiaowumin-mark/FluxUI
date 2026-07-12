package collection_test

import (
	"reflect"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal/collection"
	"github.com/xiaowumin-mark/FluxUI/internal/testkit"
)

func TestModelRejectsMissingAndDuplicateKeys(t *testing.T) {
	if _, err := collection.New([]collection.Item{{}}); err == nil {
		t.Fatal("expected an empty key error")
	}
	if _, err := collection.New([]collection.Item{{Key: "a"}, {Key: "a"}}); err == nil {
		t.Fatal("expected a duplicate key error")
	}
}

func TestModelAccessorsAndNavigationBoundaries(t *testing.T) {
	model := testkit.CollectionFixture{Items: []collection.Item{
		{Key: "disabled", Disabled: true},
		{Key: "middle"},
		{Key: "last"},
	}}.Model(t)
	if model.Len() != 3 || !model.Has("middle") || model.Has("missing") {
		t.Fatalf("unexpected model contents: %#v", model)
	}
	items := model.Items()
	items[0].Key = "mutated"
	if got, _ := model.Item("disabled"); got.Key != "disabled" {
		t.Fatalf("items returned an aliased slice: %#v", got)
	}
	if _, ok := model.Item("missing"); ok {
		t.Fatal("missing item was found")
	}
	if got, ok := model.Index("middle"); !ok || got != 1 {
		t.Fatalf("middle index = %d, ok=%t", got, ok)
	}
	if _, ok := model.Index("missing"); ok {
		t.Fatal("missing index was found")
	}
	if got, ok := model.FirstEnabled(); !ok || got != "middle" {
		t.Fatalf("first enabled = %q, ok=%t", got, ok)
	}
	if got, ok := model.LastEnabled(); !ok || got != "last" {
		t.Fatalf("last enabled = %q, ok=%t", got, ok)
	}
	if _, ok := model.NextEnabled("last", 1, false); ok {
		t.Fatal("non-wrapping next moved past the end")
	}
	if got, ok := model.NextEnabled("middle", -1, false); ok || got != "" {
		t.Fatalf("backward navigation selected disabled item: %q, ok=%t", got, ok)
	}
	if got, ok := model.NextEnabled("missing", -1, false); !ok || got != "last" {
		t.Fatalf("missing-key backward navigation = %q, ok=%t", got, ok)
	}
	if _, ok := model.NextEnabled("middle", 0, true); ok {
		t.Fatal("zero navigation direction moved focus")
	}

	disabled, err := collection.New([]collection.Item{{Key: "only", Disabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := disabled.FirstEnabled(); ok {
		t.Fatal("disabled-only collection has an enabled first item")
	}
}

func TestSelectionSurvivesReorderAndDropsRemovedKeys(t *testing.T) {
	model := testkit.CollectionFixture{Items: []collection.Item{
		{Key: "third"},
		{Key: "first"},
		{Key: "second", Disabled: true},
	}}.Model(t)
	selection := collection.NewSelection("first", "second", "removed")
	selection = selection.Reconcile(model)

	if got, want := selection.Ordered(model), []collection.Key{"first", "second"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("ordered selection = %v, want %v", got, want)
	}
	if selection.Contains("removed") {
		t.Fatal("removed key remained selected")
	}
	if got, want := selection.Toggle("first").Keys(), []collection.Key{"second"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("toggled keys = %v, want %v", got, want)
	}
	if selection.Empty() || !selection.With("").Contains("first") || selection.Without("missing").Empty() {
		t.Fatal("selection helpers did not preserve expected membership")
	}
	if got := collection.NewSelection().Toggle(""); !got.Empty() {
		t.Fatalf("empty key created a selection: %#v", got)
	}
}

func TestRovingFocusSkipsDisabledAndReconcilesByKey(t *testing.T) {
	model := testkit.CollectionFixture{Items: []collection.Item{
		{Key: "a"},
		{Key: "b", Disabled: true},
		{Key: "c"},
	}}.Model(t)

	focus := collection.RovingFocus{Key: "a", Wrap: true}
	next, changed := focus.Move(model, 1)
	if !changed || next.Key != "c" {
		t.Fatalf("forward focus = %#v, changed=%t; want c", next, changed)
	}
	next, changed = next.Move(model, 1)
	if !changed || next.Key != "a" {
		t.Fatalf("wrapped focus = %#v, changed=%t; want a", next, changed)
	}

	reordered := testkit.CollectionFixture{Items: []collection.Item{{Key: "c"}, {Key: "a"}}}.Model(t)
	if got := focus.Reconcile(reordered); got.Key != "a" {
		t.Fatalf("reconciled focus = %#v, want key a", got)
	}
	if got, changed := (collection.RovingFocus{Key: "removed"}).End(reordered); !changed || got.Key != "a" {
		t.Fatalf("end focus = %#v, changed=%t; want a", got, changed)
	}
	if got, changed := (collection.RovingFocus{Key: "a"}).Home(reordered); !changed || got.Key != "c" {
		t.Fatalf("home focus = %#v, changed=%t; want c", got, changed)
	}
	empty, err := collection.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := (collection.RovingFocus{Key: "missing", Wrap: true}).Reconcile(empty); got.Key != "" || !got.Wrap {
		t.Fatalf("empty reconciliation = %#v", got)
	}
	if got, changed := (collection.RovingFocus{Key: "old"}).Home(empty); !changed || got.Key != "" {
		t.Fatalf("empty home = %#v, changed=%t", got, changed)
	}
}

func TestRovingFocusInvalidActivePrefersNextThenPreviousThenClear(t *testing.T) {
	previous := testkit.CollectionFixture{Items: []collection.Item{
		{Key: "a"},
		{Key: "b"},
		{Key: "c"},
		{Key: "d"},
	}}.Model(t)

	for _, test := range []struct {
		name    string
		current []collection.Item
		want    collection.Key
	}{
		{
			name: "removed active chooses its next stable key",
			current: []collection.Item{
				{Key: "a"},
				{Key: "c"},
				{Key: "d"},
			},
			want: "c",
		},
		{
			name: "removed active skips a removed next key",
			current: []collection.Item{
				{Key: "a"},
				{Key: "d"},
			},
			want: "d",
		},
		{
			name: "removed active skips a disabled next key",
			current: []collection.Item{
				{Key: "a"},
				{Key: "c", Disabled: true},
				{Key: "d"},
			},
			want: "d",
		},
		{
			name: "removed active falls back to its previous stable key",
			current: []collection.Item{
				{Key: "a"},
			},
			want: "a",
		},
		{
			name:    "removed active clears when no stable neighbor remains",
			current: nil,
			want:    "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			focus := (collection.RovingFocus{Key: "b", Wrap: true}).Reconcile(previous)
			current := testkit.CollectionFixture{Items: test.current}.Model(t)
			if got := focus.Reconcile(current); got.Key != test.want || got.Wrap != focus.Wrap {
				t.Fatalf("reconciled focus = %#v, want key %q with wrap=%t", got, test.want, focus.Wrap)
			}
		})
	}
}

func TestRovingFocusDisabledActivePrefersCurrentNextThenPreviousThenClear(t *testing.T) {
	for _, test := range []struct {
		name  string
		items []collection.Item
		want  collection.Key
	}{
		{
			name: "next",
			items: []collection.Item{
				{Key: "a"},
				{Key: "b", Disabled: true},
				{Key: "c"},
			},
			want: "c",
		},
		{
			name: "previous",
			items: []collection.Item{
				{Key: "a"},
				{Key: "b", Disabled: true},
				{Key: "c", Disabled: true},
			},
			want: "a",
		},
		{
			name: "clear",
			items: []collection.Item{
				{Key: "b", Disabled: true},
			},
			want: "",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := testkit.CollectionFixture{Items: test.items}.Model(t)
			got := (collection.RovingFocus{Key: "b", Wrap: true}).Reconcile(model)
			if got.Key != test.want || !got.Wrap {
				t.Fatalf("reconciled focus = %#v, want key %q with wrap=true", got, test.want)
			}
		})
	}
}

func TestRovingFocusInitializesEmptyKeyAndClearsUnknownKey(t *testing.T) {
	model := testkit.CollectionFixture{Items: []collection.Item{{Key: "a"}, {Key: "b"}}}.Model(t)
	if got := (collection.RovingFocus{}).Reconcile(model); got.Key != "a" {
		t.Fatalf("empty active key initialized to %q, want a", got.Key)
	}
	if got := (collection.RovingFocus{Key: "removed", Wrap: true}).Reconcile(model); got.Key != "" || !got.Wrap {
		t.Fatalf("unknown active key = %#v, want cleared key with wrap=true", got)
	}
	observed := (collection.RovingFocus{Key: "a", Wrap: true}).Reconcile(model)
	observed.Key = "removed"
	if got := observed.Reconcile(model); got.Key != "" || !got.Wrap {
		t.Fatalf("externally changed active key reused stale neighbors: %#v", got)
	}
}

func TestViewportVisibleRangeAndScrollOffset(t *testing.T) {
	viewport := collection.Viewport{Offset: 12, Extent: 12}
	if got, want := viewport.VisibleRange([]int{0, 10, 20, 35, 50}, 1), (collection.Range{Start: 0, End: 4}); got != want {
		t.Fatalf("visible range = %#v, want %#v", got, want)
	}
	if got := viewport.VisibleRange([]int{0, 10, 5}, 0); got.Len() != 0 {
		t.Fatalf("malformed boundaries produced %#v", got)
	}
	if got := viewport.ScrollOffsetFor(30, 40, 80); got != 28 {
		t.Fatalf("scroll offset = %d, want 28", got)
	}
	if got := (collection.Viewport{Offset: 50, Extent: 20}).ScrollOffsetFor(0, 10, 60); got != 0 {
		t.Fatalf("scroll-to-start offset = %d, want 0", got)
	}
	if got := (collection.Viewport{Offset: -2, Extent: 0}).ScrollOffsetFor(5, 4, 10); got != 5 {
		t.Fatalf("zero-extent offset = %d, want 5", got)
	}
	if got := (collection.Viewport{Offset: 45, Extent: 20}).ScrollOffsetFor(70, 80, 80); got != 60 {
		t.Fatalf("clamped end offset = %d, want 60", got)
	}
}
