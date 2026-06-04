package system

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"
)

type testDriver struct {
	caps CapabilitySet
}

func (d testDriver) capabilities() CapabilitySet {
	return d.caps
}

type testFileDialogDriver struct {
	testDriver
	called bool
	mode   FileDialogMode
	opts   fileDialogOptions
	result FileDialogResult
	err    error
}

func (d *testFileDialogDriver) openFileDialog(ctx context.Context, mode FileDialogMode, opts fileDialogOptions) (FileDialogResult, error) {
	d.called = true
	d.mode = mode
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return FileDialogResult{}, err
	}
	return d.result, d.err
}

type testMessageBoxDriver struct {
	testDriver
	called bool
	opts   messageBoxOptions
	result MessageBoxResult
	err    error
}

func (d *testMessageBoxDriver) showMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	d.called = true
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return d.result, d.err
}

type testNotificationDriver struct {
	testDriver
	called bool
	opts   notificationOptions
	err    error
}

func (d *testNotificationDriver) notify(ctx context.Context, opts notificationOptions) error {
	d.called = true
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func withTestDriver(t *testing.T, d driver) {
	t.Helper()

	driverMu.Lock()
	previous := activeDriver
	activeDriver = d
	driverMu.Unlock()

	t.Cleanup(func() {
		driverMu.Lock()
		activeDriver = previous
		driverMu.Unlock()
	})
}

func TestErrorHelpers(t *testing.T) {
	if !IsUnsupported(fmt.Errorf("wrapped: %w", ErrUnsupported)) {
		t.Fatal("expected wrapped ErrUnsupported to be detected")
	}
	if IsUnsupported(nil) {
		t.Fatal("nil error should not be unsupported")
	}
	if !IsUnavailable(fmt.Errorf("wrapped: %w", ErrUnavailable)) {
		t.Fatal("expected wrapped ErrUnavailable to be detected")
	}
	if IsUnavailable(errors.New("other")) {
		t.Fatal("unrelated error should not be unavailable")
	}
}

func TestCapabilitySetSupports(t *testing.T) {
	set := CapabilitySet{
		CapabilityWindow:     true,
		CapabilityFileDialog: false,
	}

	if !set.Supports(CapabilityWindow) {
		t.Fatal("expected window capability to be supported")
	}
	if set.Supports(CapabilityFileDialog) {
		t.Fatal("explicit false capability should not be supported")
	}
	if set.Supports(Capability("unknown")) {
		t.Fatal("unknown capability should not be supported")
	}
	if set.Supports("") {
		t.Fatal("empty capability should not be supported")
	}
	if (CapabilitySet{}).Supports(CapabilityWindow) {
		t.Fatal("empty set should not support capabilities")
	}
}

func TestCapabilitiesReturnsCopy(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{
		CapabilityWindow: true,
	}})

	first := Capabilities()
	first[CapabilityWindow] = false
	first[CapabilityTray] = true

	second := Capabilities()
	if !second.Supports(CapabilityWindow) {
		t.Fatal("mutating returned capabilities should not affect driver capabilities")
	}
	if second.Supports(CapabilityTray) {
		t.Fatal("mutating returned capabilities should not add capabilities")
	}
}

func TestSupportsUsesActiveDriver(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{
		CapabilityMessageBox: true,
	}})

	if !Supports(CapabilityMessageBox) {
		t.Fatal("expected Supports to use active driver")
	}
	if Supports(CapabilityNotification) {
		t.Fatal("unexpected unsupported capability")
	}
}

func TestNilDriverFallsBackToPlatformDriver(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{
		CapabilityTray: true,
	}})

	setDriver(nil)
	caps := Capabilities()

	if runtime.GOOS == "windows" {
		if !caps.Supports(CapabilityWindow) {
			t.Fatal("windows platform driver should support window capability")
		}
		return
	}

	if len(caps) != 0 {
		t.Fatalf("unsupported platform driver should expose no capabilities, got %#v", caps)
	}
}

func TestCurrentPlatformCapabilities(t *testing.T) {
	caps := Capabilities()

	if runtime.GOOS == "windows" {
		if !caps.Supports(CapabilityWindow) {
			t.Fatal("windows should expose window capability")
		}
		if !caps.Supports(CapabilityFileDialog) {
			t.Fatal("windows should expose file dialog capability")
		}
		if !caps.Supports(CapabilityMessageBox) {
			t.Fatal("windows should expose message box capability")
		}
		if !caps.Supports(CapabilityNotification) {
			t.Fatal("windows should expose notification capability")
		}
		return
	}

	if caps.Supports(CapabilityWindow) {
		t.Fatal("unsupported platform should not expose window capability")
	}
	if caps.Supports(CapabilityFileDialog) {
		t.Fatal("unsupported platform should not expose file dialog capability")
	}
	if caps.Supports(CapabilityMessageBox) {
		t.Fatal("unsupported platform should not expose message box capability")
	}
	if caps.Supports(CapabilityNotification) {
		t.Fatal("unsupported platform should not expose notification capability")
	}
}

func TestOpenFileDialogRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	_, err := OpenFileDialog(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestOpenFileDialogChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
	}
	withTestDriver(t, fd)

	_, err := OpenFileDialog(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if fd.called {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestFileDialogEntrypointsDispatchModes(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, ...FileDialogOption) (FileDialogResult, error)
		mode FileDialogMode
	}{
		{name: "open file", call: OpenFileDialog, mode: FileDialogOpenFile},
		{name: "open files", call: OpenFilesDialog, mode: FileDialogOpenFiles},
		{name: "save file", call: SaveFileDialog, mode: FileDialogSaveFile},
		{name: "pick folder", call: PickFolderDialog, mode: FileDialogPickFolder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := FileDialogResult{Paths: []string{"C:\\tmp\\demo.txt"}}
			fd := &testFileDialogDriver{
				testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
				result:     expected,
			}
			withTestDriver(t, fd)

			got, err := tt.call(context.Background(), FileDialogTitle("Choose"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !fd.called {
				t.Fatal("expected driver to be called")
			}
			if fd.mode != tt.mode {
				t.Fatalf("expected mode %q, got %q", tt.mode, fd.mode)
			}
			if fd.opts.title != "Choose" {
				t.Fatalf("expected title option to reach driver, got %q", fd.opts.title)
			}
			if len(got.Paths) != 1 || got.Paths[0] != expected.Paths[0] {
				t.Fatalf("unexpected result: %#v", got)
			}
		})
	}
}

func TestFileDialogOptions(t *testing.T) {
	filters := []FileFilter{{
		Name:     "Images",
		Patterns: []string{"*.png", "*.jpg"},
	}}
	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
	}
	withTestDriver(t, fd)

	_, err := SaveFileDialog(context.Background(),
		FileDialogTitle("Save"),
		FileDialogDefaultDir("C:\\tmp"),
		FileDialogDefaultName("demo.txt"),
		FileDialogFilters(filters...),
		FileDialogOwner(12345),
		FileDialogAllowCreateDirs(false),
		FileDialogAllowMissingPath(true),
		FileDialogOverwritePrompt(false),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filters[0].Name = "Mutated"
	filters[0].Patterns[0] = "*.gif"

	if fd.opts.title != "Save" {
		t.Fatalf("unexpected title: %q", fd.opts.title)
	}
	if fd.opts.defaultDir != "C:\\tmp" {
		t.Fatalf("unexpected default dir: %q", fd.opts.defaultDir)
	}
	if fd.opts.defaultName != "demo.txt" {
		t.Fatalf("unexpected default name: %q", fd.opts.defaultName)
	}
	if fd.opts.owner != 12345 {
		t.Fatalf("unexpected owner: %d", fd.opts.owner)
	}
	if fd.opts.allowCreateDirs {
		t.Fatal("expected allowCreateDirs=false")
	}
	if !fd.opts.allowMissingPath {
		t.Fatal("expected allowMissingPath=true")
	}
	if fd.opts.overwritePrompt {
		t.Fatal("expected overwritePrompt=false")
	}
	if len(fd.opts.filters) != 1 || fd.opts.filters[0].Name != "Images" || fd.opts.filters[0].Patterns[0] != "*.png" {
		t.Fatalf("filters were not cloned correctly: %#v", fd.opts.filters)
	}
}

func TestFileDialogDefaultOptions(t *testing.T) {
	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
	}
	withTestDriver(t, fd)

	_, err := OpenFileDialog(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fd.opts.allowCreateDirs {
		t.Fatal("expected allowCreateDirs to default true")
	}
	if fd.opts.allowMissingPath {
		t.Fatal("expected allowMissingPath to default false")
	}
	if !fd.opts.overwritePrompt {
		t.Fatal("expected overwritePrompt to default true")
	}
}

func TestShowMessageBoxRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	_, err := ShowMessageBox(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestShowMessageBoxChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mb := &testMessageBoxDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityMessageBox: true}},
	}
	withTestDriver(t, mb)

	_, err := ShowMessageBox(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if mb.called {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestShowMessageBoxOptions(t *testing.T) {
	expected := MessageBoxResultYes
	mb := &testMessageBoxDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityMessageBox: true}},
		result:     expected,
	}
	withTestDriver(t, mb)

	got, err := ShowMessageBox(context.Background(),
		MessageBoxTitle("Confirm"),
		MessageBoxText("Continue?"),
		MessageBoxStyle(MessageBoxQuestion),
		MessageBoxButtonSet(MessageBoxYesNo),
		MessageBoxDefaultButton(MessageBoxResultNo),
		MessageBoxOwner(12345),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mb.called {
		t.Fatal("expected driver to be called")
	}
	if got != expected {
		t.Fatalf("expected result %q, got %q", expected, got)
	}
	if mb.opts.title != "Confirm" {
		t.Fatalf("unexpected title: %q", mb.opts.title)
	}
	if mb.opts.text != "Continue?" {
		t.Fatalf("unexpected text: %q", mb.opts.text)
	}
	if mb.opts.kind != MessageBoxQuestion {
		t.Fatalf("unexpected kind: %q", mb.opts.kind)
	}
	if mb.opts.buttons != MessageBoxYesNo {
		t.Fatalf("unexpected buttons: %q", mb.opts.buttons)
	}
	if mb.opts.defaultButton != MessageBoxResultNo {
		t.Fatalf("unexpected default button: %q", mb.opts.defaultButton)
	}
	if mb.opts.owner != 12345 {
		t.Fatalf("unexpected owner: %d", mb.opts.owner)
	}
}

func TestShowMessageBoxDefaultOptions(t *testing.T) {
	mb := &testMessageBoxDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityMessageBox: true}},
		result:     MessageBoxResultOK,
	}
	withTestDriver(t, mb)

	_, err := ShowMessageBox(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mb.opts.kind != MessageBoxInfo {
		t.Fatalf("expected default info kind, got %q", mb.opts.kind)
	}
	if mb.opts.buttons != MessageBoxOK {
		t.Fatalf("expected default OK buttons, got %q", mb.opts.buttons)
	}
	if mb.opts.defaultButton != MessageBoxResultOK {
		t.Fatalf("expected default OK button, got %q", mb.opts.defaultButton)
	}
}

func TestNotifyRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	err := Notify(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestNotifyChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	nd := &testNotificationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, nd)

	err := Notify(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if nd.called {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestNotifyOptions(t *testing.T) {
	nd := &testNotificationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, nd)

	clicked := false
	var gotEvent NotificationEvent
	err := Notify(context.Background(),
		NotificationTitle("Done"),
		NotificationBody("Export finished."),
		NotificationKindStyle(NotificationSuccess),
		NotificationIcon("C:\\tmp\\icon.ico"),
		NotificationGroup("exports"),
		NotificationTimeout(5*time.Second),
		NotificationOnClick(func(event NotificationEvent) {
			clicked = true
			gotEvent = event
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !nd.called {
		t.Fatal("expected driver to be called")
	}
	if nd.opts.title != "Done" {
		t.Fatalf("unexpected title: %q", nd.opts.title)
	}
	if nd.opts.body != "Export finished." {
		t.Fatalf("unexpected body: %q", nd.opts.body)
	}
	if nd.opts.kind != NotificationSuccess {
		t.Fatalf("unexpected kind: %q", nd.opts.kind)
	}
	if nd.opts.icon != "C:\\tmp\\icon.ico" {
		t.Fatalf("unexpected icon: %q", nd.opts.icon)
	}
	if nd.opts.group != "exports" {
		t.Fatalf("unexpected group: %q", nd.opts.group)
	}
	if nd.opts.timeout != 5*time.Second {
		t.Fatalf("unexpected timeout: %s", nd.opts.timeout)
	}
	if nd.opts.onClick == nil {
		t.Fatal("expected onClick option to reach driver")
	}
	nd.opts.onClick(NotificationEvent{
		Kind:   NotificationEventClicked,
		Group:  "exports",
		Action: "open",
	})
	if !clicked {
		t.Fatal("expected click callback to run when driver invokes it")
	}
	if gotEvent.Kind != NotificationEventClicked || gotEvent.Group != "exports" || gotEvent.Action != "open" {
		t.Fatalf("unexpected click event: %#v", gotEvent)
	}
}

func TestNotifyDefaultOptions(t *testing.T) {
	nd := &testNotificationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, nd)

	err := Notify(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nd.opts.kind != NotificationInfo {
		t.Fatalf("expected default info kind, got %q", nd.opts.kind)
	}
	if nd.opts.timeout != 0 {
		t.Fatalf("expected default timeout 0, got %s", nd.opts.timeout)
	}
}

func TestNotifyPropagatesUnavailable(t *testing.T) {
	nd := &testNotificationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
		err:        fmt.Errorf("wrapped: %w", ErrUnavailable),
	}
	withTestDriver(t, nd)

	err := Notify(context.Background())
	if !IsUnavailable(err) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}
