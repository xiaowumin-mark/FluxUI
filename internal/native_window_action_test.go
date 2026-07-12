package internal

import (
	"image"
	"testing"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestNativeWindowActionFrameRegistration(t *testing.T) {
	runtime := NewRuntime(nil)
	runtime.BeginFrame()
	if runtime.NativeWindowActionRegionsActive() || runtime.NativeWindowActionRegions() != nil || runtime.NativeWindowActionRouterActive() {
		t.Fatal("fresh frame unexpectedly has native action state")
	}

	runtime.RegisterNativeWindowActionRegion(0, image.Rect(0, 0, 10, 10))
	runtime.RegisterNativeWindowActionRegion(NativeWindowActionMaximizeButton, image.Rectangle{})
	region := image.Rect(2, 3, 12, 14)
	runtime.RegisterNativeWindowActionRegion(NativeWindowActionMaximizeButton, region)
	runtime.RegisterNativeWindowActionRouter()
	regions := runtime.NativeWindowActionRegions()
	if !runtime.NativeWindowActionRegionsActive() || !runtime.NativeWindowActionRouterActive() || len(regions) != 1 || regions[0] != (NativeWindowActionRegion{Action: NativeWindowActionMaximizeButton, Rect: region}) {
		t.Fatalf("native action state = %#v, router=%t", regions, runtime.NativeWindowActionRouterActive())
	}

	runtime.RegisterWindowDragArea()
	if !runtime.WindowDragAreaActive() {
		t.Fatal("window drag area was not recorded")
	}
	runtime.BeginFrame()
	if runtime.NativeWindowActionRegionsActive() || runtime.NativeWindowActionRouterActive() || runtime.WindowDragAreaActive() {
		t.Fatal("BeginFrame did not reset native action state")
	}

	var ops op.Ops
	context := NewContext(gioLayout.Context{Ops: &ops, Constraints: gioLayout.Exact(image.Pt(20, 10))}, runtime)
	context.RegisterWindowMaximizeButton(image.Rect(-5, 2, 8, 20))
	context.RegisterWindowActionInput()
	regions = runtime.NativeWindowActionRegions()
	if len(regions) != 1 || regions[0].Rect != image.Rect(0, 2, 8, 10) || !runtime.NativeWindowActionRouterActive() {
		t.Fatalf("context native action registration = %#v, router=%t", regions, runtime.NativeWindowActionRouterActive())
	}
	context.RegisterWindowMaximizeButton(image.Rect(30, 30, 40, 40))
	if len(runtime.NativeWindowActionRegions()) != 1 {
		t.Fatal("empty viewport intersection registered a native action")
	}

	var nilRuntime *Runtime
	nilRuntime.RegisterNativeWindowActionRegion(NativeWindowActionMaximizeButton, region)
	nilRuntime.RegisterNativeWindowActionRouter()
	if nilRuntime.NativeWindowActionRegions() != nil || nilRuntime.NativeWindowActionRegionsActive() || nilRuntime.NativeWindowActionRouterActive() || nilRuntime.WindowDragAreaActive() {
		t.Fatal("nil runtime native action helpers should be inert")
	}
	var nilContext *Context
	nilContext.RegisterWindowMaximizeButton(region)
	nilContext.RegisterWindowActionInput()
}
