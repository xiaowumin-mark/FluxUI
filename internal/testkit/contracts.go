// Package testkit provides internal-only fixtures for component contract tests.
package testkit

import (
	"image"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal/collection"
	"github.com/xiaowumin-mark/FluxUI/internal/overlay"
)

// CollectionFixture creates a validated collection model for behavior tests.
type CollectionFixture struct {
	Items []collection.Item
}

// Model returns the validated model or fails the calling test.
func (f CollectionFixture) Model(t testing.TB) collection.Model {
	t.Helper()
	model, err := collection.New(f.Items)
	if err != nil {
		t.Fatalf("build collection fixture: %v", err)
	}
	return model
}

// AnchoredOverlayFixture describes a trigger and its popup in window
// coordinates.
type AnchoredOverlayFixture struct {
	Anchor  image.Rectangle
	Content image.Rectangle
}

// Region returns the shared anchored-overlay protection contract.
func (f AnchoredOverlayFixture) Region() overlay.AnchoredRegion {
	return overlay.AnchoredRegion{Anchor: f.Anchor, Content: f.Content}
}
