package system

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type testDriver struct {
	caps CapabilitySet
}

func (d testDriver) capabilities() CapabilitySet {
	return d.caps
}

type countingCapabilityDriver struct {
	caps  CapabilitySet
	calls *int
}

func (d countingCapabilityDriver) capabilities() CapabilitySet {
	(*d.calls)++
	return d.caps
}

type testProbeDriver struct {
	testDriver
	results map[Capability]CapabilityAvailability
}

func (d testProbeDriver) probeCapability(cap Capability) CapabilityAvailability {
	if result, ok := d.results[cap]; ok {
		return result
	}
	return defaultCapabilityAvailability(d.caps, cap)
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
	called         bool
	detailedCalled bool
	opts           messageBoxOptions
	result         MessageBoxResult
	detailedResult MessageBoxDetailedResult
	err            error
}

func (d *testMessageBoxDriver) showMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	d.called = true
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return d.result, d.err
}

func (d *testMessageBoxDriver) showDetailedMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxDetailedResult, error) {
	d.detailedCalled = true
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}
	if d.detailedResult.Result == "" {
		d.detailedResult = MessageBoxDetailedResult{
			Result:   d.result,
			ButtonID: string(d.result),
		}
	}
	return d.detailedResult, d.err
}

type testNotificationDriver struct {
	testDriver
	called        bool
	opts          notificationOptions
	canceledGroup string
	probeBackend  NotificationBackend
	probeOpts     notificationOptions
	probeResult   NotificationBackendProbe
	err           error
}

type testDragAndDropDriver struct {
	testDriver
	probeResult DragAndDropProbe
}

func (d testDragAndDropDriver) probeDragAndDrop(ctx context.Context) DragAndDropProbe {
	if err := ctx.Err(); err != nil {
		return DragAndDropProbe{
			Status: CapabilityStatusUnavailable,
			Err:    err,
		}
	}
	return d.probeResult
}

func (d *testNotificationDriver) notify(ctx context.Context, opts notificationOptions) error {
	d.called = true
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testNotificationDriver) cancelNotificationGroup(group string) error {
	d.canceledGroup = group
	if d.err != nil {
		return d.err
	}
	return nil
}

func (d *testNotificationDriver) probeNotificationBackend(ctx context.Context, backend NotificationBackend, opts notificationOptions) NotificationBackendProbe {
	d.probeBackend = backend
	d.probeOpts = opts
	if err := ctx.Err(); err != nil {
		return NotificationBackendProbe{
			Backend: backend,
			Status:  CapabilityStatusUnavailable,
			Err:     err,
		}
	}
	return d.probeResult
}

type testToastActivatorDriver struct {
	testDriver
	called bool
	clsid  string
	fn     func(ToastActivationEvent)
	handle *testToastActivatorHandle
	err    error
}

func (d *testToastActivatorDriver) startToastActivator(ctx context.Context, clsid string, fn func(ToastActivationEvent)) (toastActivatorHandle, error) {
	d.called = true
	d.clsid = clsid
	d.fn = fn
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d.err != nil {
		return nil, d.err
	}
	if d.handle == nil {
		d.handle = &testToastActivatorHandle{done: make(chan struct{})}
	}
	return d.handle, nil
}

type testToastActivatorHandle struct {
	closed bool
	done   chan struct{}
	once   sync.Once
}

func (h *testToastActivatorHandle) Close() error {
	h.closed = true
	h.once.Do(func() {
		close(h.done)
	})
	return nil
}

func (h *testToastActivatorHandle) Done() <-chan struct{} {
	return h.done
}

type testTrayDriver struct {
	testDriver
	called bool
	opts   trayOptions
	handle *testTrayHandle
	err    error
}

type testSystemEventDriver struct {
	testDriver
	called bool
	kinds  []SystemEventKind
	handle *testSystemEventHandle
	err    error
}

type testClipboardDriver struct {
	testDriver
	readCalled       bool
	writeCalled      bool
	readFilesCalled  bool
	writeFilesCalled bool
	readImageCalled  bool
	writeImageCalled bool
	text             string
	writtenText      string
	files            []string
	writtenFiles     []string
	image            []byte
	writtenImage     []byte
	err              error
}

func (d *testClipboardDriver) readClipboardText(ctx context.Context) (string, error) {
	d.readCalled = true
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return d.text, d.err
}

func (d *testClipboardDriver) writeClipboardText(ctx context.Context, text string) error {
	d.writeCalled = true
	d.writtenText = text
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testClipboardDriver) readClipboardFiles(ctx context.Context) ([]string, error) {
	d.readFilesCalled = true
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append([]string(nil), d.files...), d.err
}

func (d *testClipboardDriver) writeClipboardFiles(ctx context.Context, paths []string) error {
	d.writeFilesCalled = true
	d.writtenFiles = append([]string(nil), paths...)
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testClipboardDriver) readClipboardImagePNG(ctx context.Context) ([]byte, error) {
	d.readImageCalled = true
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return append([]byte(nil), d.image...), d.err
}

func (d *testClipboardDriver) writeClipboardImagePNG(ctx context.Context, data []byte) error {
	d.writeImageCalled = true
	d.writtenImage = append([]byte(nil), data...)
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

type testShellDriver struct {
	testDriver
	openURLCalled  bool
	openPathCalled bool
	revealCalled   bool
	target         string
	err            error
}

func (d *testShellDriver) openURL(ctx context.Context, target string) error {
	d.openURLCalled = true
	d.target = target
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

type testSingleInstanceDriver struct {
	testDriver
	called bool
	opts   singleInstanceOptions
	handle *testSingleInstanceHandle
	err    error
}

func (d *testSingleInstanceDriver) acquireSingleInstance(ctx context.Context, opts singleInstanceOptions) (singleInstanceHandle, error) {
	d.called = true
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d.err != nil {
		return nil, d.err
	}
	if d.handle == nil {
		d.handle = &testSingleInstanceHandle{ch: make(chan SingleInstanceEvent)}
	}
	return d.handle, nil
}

type testRegistrationDriver struct {
	testDriver
	protocolRegistered    bool
	protocolUnregistered  bool
	fileRegistered        bool
	fileUnregistered      bool
	startupRegistered     bool
	startupUnregistered   bool
	toastRegistered       bool
	toastUnregistered     bool
	activatorRegistered   bool
	activatorUnregistered bool
	scheme                string
	extension             string
	clsid                 string
	progID                string
	name                  string
	command               string
	opts                  registrationOptions
	appID                 string
	shortcutName          string
	executable            string
	toastOpts             toastShortcutOptions
	err                   error
}

func (d *testRegistrationDriver) registerProtocolHandler(ctx context.Context, scheme, command string, opts registrationOptions) error {
	d.protocolRegistered = true
	d.scheme = scheme
	d.command = command
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

type testGlobalShortcutDriver struct {
	testDriver
	called bool
	spec   GlobalShortcutSpec
	handle *testGlobalShortcutHandle
	err    error
}

func (d *testGlobalShortcutDriver) registerGlobalShortcut(ctx context.Context, spec GlobalShortcutSpec, fn func(GlobalShortcutEvent)) (globalShortcutHandle, error) {
	d.called = true
	d.spec = spec
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d.err != nil {
		return nil, d.err
	}
	if d.handle == nil {
		d.handle = &testGlobalShortcutHandle{ch: make(chan GlobalShortcutEvent)}
	}
	if fn != nil {
		fn(GlobalShortcutEvent{ID: spec.ID, Key: spec.Key, Modifiers: spec.Modifiers})
	}
	return d.handle, nil
}

func (d *testRegistrationDriver) unregisterProtocolHandler(ctx context.Context, scheme string) error {
	d.protocolUnregistered = true
	d.scheme = scheme
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) registerFileAssociation(ctx context.Context, extension, progID, command string, opts registrationOptions) error {
	d.fileRegistered = true
	d.extension = extension
	d.progID = progID
	d.command = command
	d.opts = opts
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) unregisterFileAssociation(ctx context.Context, extension, progID string) error {
	d.fileUnregistered = true
	d.extension = extension
	d.progID = progID
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) registerStartupTask(ctx context.Context, name, command string) error {
	d.startupRegistered = true
	d.name = name
	d.command = command
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) unregisterStartupTask(ctx context.Context, name string) error {
	d.startupUnregistered = true
	d.name = name
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) registerToastShortcut(ctx context.Context, appID, shortcutName, executable string, opts toastShortcutOptions) error {
	d.toastRegistered = true
	d.appID = appID
	d.shortcutName = shortcutName
	d.executable = executable
	d.toastOpts = opts
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) unregisterToastShortcut(ctx context.Context, shortcutName string) error {
	d.toastUnregistered = true
	d.shortcutName = shortcutName
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) registerToastActivator(ctx context.Context, clsid, command string) error {
	d.activatorRegistered = true
	d.clsid = clsid
	d.command = command
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testRegistrationDriver) unregisterToastActivator(ctx context.Context, clsid string) error {
	d.activatorUnregistered = true
	d.clsid = clsid
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testShellDriver) openPath(ctx context.Context, path string) error {
	d.openPathCalled = true
	d.target = path
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testShellDriver) revealPath(ctx context.Context, path string) error {
	d.revealCalled = true
	d.target = path
	if err := ctx.Err(); err != nil {
		return err
	}
	return d.err
}

func (d *testSystemEventDriver) subscribeSystemEvents(ctx context.Context, kinds []SystemEventKind) (systemEventHandle, error) {
	d.called = true
	d.kinds = cloneSystemEventKinds(kinds)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d.err != nil {
		return nil, d.err
	}
	if d.handle == nil {
		d.handle = &testSystemEventHandle{ch: make(chan SystemEvent, 1)}
	}
	return d.handle, nil
}

type testSystemEventHandle struct {
	ch     chan SystemEvent
	closed bool
}

func (h *testSystemEventHandle) events() <-chan SystemEvent {
	return h.ch
}

func (h *testSystemEventHandle) close() error {
	if h.closed {
		return ErrClosed
	}
	h.closed = true
	close(h.ch)
	return nil
}

func (d *testTrayDriver) newTray(opts trayOptions) (trayHandle, error) {
	d.called = true
	d.opts = opts
	if d.err != nil {
		return nil, d.err
	}
	if d.handle == nil {
		d.handle = &testTrayHandle{}
	}
	return d.handle, nil
}

type testTrayHandle struct {
	icon         string
	iconData     []byte
	iconResource uint16
	tooltip      string
	menu         TrayMenu
	menuProvider func() TrayMenu
	showCount    int
	hideCount    int
	closeCount   int
	err          error
}

type testSingleInstanceHandle struct {
	ch     chan SingleInstanceEvent
	closed bool
}

type testGlobalShortcutHandle struct {
	ch     chan GlobalShortcutEvent
	closed bool
}

func (h *testGlobalShortcutHandle) events() <-chan GlobalShortcutEvent {
	return h.ch
}

func (h *testGlobalShortcutHandle) close() error {
	if h.closed {
		return ErrClosed
	}
	h.closed = true
	close(h.ch)
	return nil
}

func (h *testSingleInstanceHandle) events() <-chan SingleInstanceEvent {
	return h.ch
}

func (h *testSingleInstanceHandle) close() error {
	if h.closed {
		return ErrClosed
	}
	h.closed = true
	close(h.ch)
	return nil
}

func (h *testTrayHandle) setIcon(path string) error {
	if h.err != nil {
		return h.err
	}
	h.icon = path
	return nil
}

func (h *testTrayHandle) setIconData(data []byte) error {
	if h.err != nil {
		return h.err
	}
	h.iconData = append([]byte(nil), data...)
	return nil
}

func (h *testTrayHandle) setIconResource(id uint16) error {
	if h.err != nil {
		return h.err
	}
	h.iconResource = id
	return nil
}

func (h *testTrayHandle) setTooltip(text string) error {
	if h.err != nil {
		return h.err
	}
	h.tooltip = text
	return nil
}

func (h *testTrayHandle) setMenu(menu TrayMenu) error {
	if h.err != nil {
		return h.err
	}
	h.menu = cloneTrayMenu(menu)
	return nil
}

func (h *testTrayHandle) setMenuProvider(fn func() TrayMenu) error {
	if h.err != nil {
		return h.err
	}
	h.menuProvider = fn
	return nil
}

func (h *testTrayHandle) show() error {
	if h.err != nil {
		return h.err
	}
	h.showCount++
	return nil
}

func (h *testTrayHandle) hide() error {
	if h.err != nil {
		return h.err
	}
	h.hideCount++
	return nil
}

func (h *testTrayHandle) close() error {
	if h.err != nil {
		return h.err
	}
	h.closeCount++
	return nil
}

func withTestDriver(t *testing.T, d driver) {
	t.Helper()

	driverMu.Lock()
	previous := activeDriver
	previousCaps := activeCapabilities.clone()
	driverMu.Unlock()
	setDriver(d)

	t.Cleanup(func() {
		_ = CloseTrays()
		driverMu.Lock()
		activeDriver = previous
		activeCapabilities = previousCaps
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
	if !IsClosed(fmt.Errorf("wrapped: %w", ErrClosed)) {
		t.Fatal("expected wrapped ErrClosed to be detected")
	}
	if IsClosed(errors.New("other")) {
		t.Fatal("unrelated error should not be closed")
	}
	if !IsAlreadyRunning(fmt.Errorf("wrapped: %w", ErrAlreadyRunning)) {
		t.Fatal("expected wrapped ErrAlreadyRunning to be detected")
	}
	if IsAlreadyRunning(errors.New("other")) {
		t.Fatal("unrelated error should not be already-running")
	}
	if !IsInvalidTarget(fmt.Errorf("wrapped: %w", ErrInvalidTarget)) {
		t.Fatal("expected wrapped ErrInvalidTarget to be detected")
	}
	if IsInvalidTarget(errors.New("other")) {
		t.Fatal("unrelated error should not be invalid-target")
	}
	if !IsTargetNotFound(fmt.Errorf("wrapped: %w", ErrTargetNotFound)) {
		t.Fatal("expected wrapped ErrTargetNotFound to be detected")
	}
	if IsTargetNotFound(errors.New("other")) {
		t.Fatal("unrelated error should not be target-not-found")
	}
	if !IsNoDefaultHandler(fmt.Errorf("wrapped: %w", ErrNoDefaultHandler)) {
		t.Fatal("expected wrapped ErrNoDefaultHandler to be detected")
	}
	if IsNoDefaultHandler(errors.New("other")) {
		t.Fatal("unrelated error should not be no-default-handler")
	}
	if !IsAccessDenied(fmt.Errorf("wrapped: %w", ErrAccessDenied)) {
		t.Fatal("expected wrapped ErrAccessDenied to be detected")
	}
	if IsAccessDenied(errors.New("other")) {
		t.Fatal("unrelated error should not be access-denied")
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

func TestSupportsUsesCachedCapabilities(t *testing.T) {
	calls := 0
	withTestDriver(t, countingCapabilityDriver{
		caps: CapabilitySet{
			CapabilityShell: true,
		},
		calls: &calls,
	})
	if calls != 1 {
		t.Fatalf("expected setDriver to read capabilities once, got %d", calls)
	}
	if !Supports(CapabilityShell) {
		t.Fatal("expected shell support")
	}
	if Supports(CapabilityTray) {
		t.Fatal("unexpected tray support")
	}
	if calls != 1 {
		t.Fatalf("Supports should use cached capabilities, driver calls=%d", calls)
	}
}

func TestProbeUsesCapabilitySetByDefault(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{
		CapabilityWindow: true,
	}})

	available := Probe(CapabilityWindow)
	if available.Capability != CapabilityWindow || available.Status != CapabilityStatusAvailable {
		t.Fatalf("unexpected available probe: %#v", available)
	}
	if !available.Supported() || !available.Available() {
		t.Fatalf("expected supported and available probe: %#v", available)
	}

	unsupported := Availability(CapabilityTray)
	if unsupported.Status != CapabilityStatusUnsupported {
		t.Fatalf("expected unsupported probe, got %#v", unsupported)
	}
	if unsupported.Supported() || unsupported.Available() || !IsUnsupported(unsupported.Err) {
		t.Fatalf("unexpected unsupported probe semantics: %#v", unsupported)
	}
}

func TestProbeUsesDriverSpecificAvailability(t *testing.T) {
	withTestDriver(t, testProbeDriver{
		testDriver: testDriver{caps: CapabilitySet{
			CapabilityNotification: true,
		}},
		results: map[Capability]CapabilityAvailability{
			CapabilityNotification: {
				Status: CapabilityStatusUnavailable,
				Err:    fmt.Errorf("notification shell disabled: %w", ErrUnavailable),
			},
		},
	})

	probe := Probe(CapabilityNotification)
	if probe.Capability != CapabilityNotification {
		t.Fatalf("expected capability to be filled, got %#v", probe)
	}
	if probe.Status != CapabilityStatusUnavailable {
		t.Fatalf("expected unavailable status, got %#v", probe)
	}
	if !probe.Supported() || probe.Available() || !IsUnavailable(probe.Err) {
		t.Fatalf("unexpected unavailable probe semantics: %#v", probe)
	}
}

func TestProbeDragAndDropRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	probe := ProbeDragAndDrop(context.Background())
	if probe.Status != CapabilityStatusUnsupported || probe.Supported() || probe.Available() || !IsUnsupported(probe.Err) {
		t.Fatalf("unexpected unsupported drag-and-drop probe: %#v", probe)
	}
}

func TestProbeDragAndDropDefaultAvailability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityDragAndDrop: true}})

	probe := ProbeDragAndDrop(context.Background())
	if probe.Status != CapabilityStatusAvailable || !probe.Supported() || !probe.Available() {
		t.Fatalf("expected available drag-and-drop probe, got %#v", probe)
	}
	if !probe.SupportsDropTarget || !probe.SupportsDragSource || !probe.SupportsText || !probe.SupportsFiles || !probe.SupportsCustomMIME {
		t.Fatalf("expected default drag-and-drop payload support, got %#v", probe)
	}
	if !probe.SupportsExternalDragIn || probe.SupportsExternalDragOut {
		t.Fatalf("expected default drag-and-drop probe to support external drag-in only, got %#v", probe)
	}
	if len(probe.SupportedOperations) != 3 {
		t.Fatalf("expected copy/move/link operations, got %#v", probe.SupportedOperations)
	}
}

func TestProbeDragAndDropUsesDriverSpecificProbe(t *testing.T) {
	withTestDriver(t, testDragAndDropDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityDragAndDrop: true}},
		probeResult: DragAndDropProbe{
			Status:              CapabilityStatusAvailable,
			SupportsDropTarget:  true,
			SupportsText:        true,
			SupportedOperations: []DragDropOperation{DragDropOperationMove, DragDropOperationMove, DragDropOperation("bad")},
		},
	})

	probe := ProbeDragAndDrop(context.Background())
	if !probe.Available() || !probe.SupportsDropTarget || probe.SupportsDragSource {
		t.Fatalf("unexpected custom drag-and-drop probe: %#v", probe)
	}
	if len(probe.SupportedOperations) != 1 || probe.SupportedOperations[0] != DragDropOperationMove {
		t.Fatalf("expected normalized custom operations, got %#v", probe.SupportedOperations)
	}
}

func TestProbeDragAndDropClearsSupportFlagsWhenUnavailable(t *testing.T) {
	withTestDriver(t, testDragAndDropDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityDragAndDrop: true}},
		probeResult: DragAndDropProbe{
			Status:                  CapabilityStatusUnavailable,
			SupportsDropTarget:      true,
			SupportsDragSource:      true,
			SupportsText:            true,
			SupportsFiles:           true,
			SupportsCustomMIME:      true,
			SupportsExternalDragIn:  true,
			SupportsExternalDragOut: true,
			SupportedOperations:     []DragDropOperation{DragDropOperationCopy},
		},
	})

	probe := ProbeDragAndDrop(context.Background())
	if probe.Status != CapabilityStatusUnavailable || !probe.Supported() || probe.Available() || !IsUnavailable(probe.Err) {
		t.Fatalf("unexpected unavailable drag-and-drop probe: %#v", probe)
	}
	if probe.SupportsDropTarget || probe.SupportsDragSource || probe.SupportsText || probe.SupportsFiles || probe.SupportsCustomMIME ||
		probe.SupportsExternalDragIn || probe.SupportsExternalDragOut || len(probe.SupportedOperations) != 0 {
		t.Fatalf("unavailable probe should clear support flags, got %#v", probe)
	}
}

func TestProbeDragAndDropChecksContext(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityDragAndDrop: true}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	probe := ProbeDragAndDrop(ctx)
	if probe.Status != CapabilityStatusUnavailable || !errors.Is(probe.Err, context.Canceled) {
		t.Fatalf("expected cancelled drag-and-drop probe, got %#v", probe)
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
		if !caps.Supports(CapabilityTray) {
			t.Fatal("windows platform driver should support tray capability")
		}
		if !caps.Supports(CapabilitySystemEvents) {
			t.Fatal("windows platform driver should support system events capability")
		}
		if !caps.Supports(CapabilitySingleInstance) {
			t.Fatal("windows platform driver should support single instance capability")
		}
		if !caps.Supports(CapabilitySystemRegistration) {
			t.Fatal("windows platform driver should support system registration capability")
		}
		if !caps.Supports(CapabilityGlobalShortcut) {
			t.Fatal("windows platform driver should support global shortcut capability")
		}
		if !caps.Supports(CapabilityDragAndDrop) {
			t.Fatal("windows platform driver should support drag and drop capability")
		}
		return
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityClipboard) {
			t.Fatal("macOS/Linux platform driver should support clipboard capability")
		}
		if !caps.Supports(CapabilityFileDialog) {
			t.Fatal("macOS/Linux platform driver should support file dialog capability")
		}
		if !caps.Supports(CapabilityMessageBox) {
			t.Fatal("macOS/Linux platform driver should support message box capability")
		}
		if !caps.Supports(CapabilityNotification) {
			t.Fatal("macOS/Linux platform driver should support notification capability")
		}
		if !caps.Supports(CapabilitySystemEvents) {
			t.Fatal("macOS/Linux platform driver should support system events capability")
		}
		if !caps.Supports(CapabilityShell) {
			t.Fatal("macOS/Linux platform driver should support shell capability")
		}
		if !caps.Supports(CapabilitySingleInstance) {
			t.Fatal("macOS/Linux platform driver should support single instance capability")
		}
		if !caps.Supports(CapabilityDragAndDrop) {
			t.Fatal("macOS/Linux platform driver should support drag and drop capability")
		}
		expectedLen := 9
		if !caps.Supports(CapabilityTray) {
			t.Fatal("macOS/Linux platform driver should support tray capability")
		}
		if runtime.GOOS == "linux" {
			expectedLen = 9
		}
		if len(caps) != expectedLen {
			t.Fatalf("macOS/Linux platform driver exposed unexpected capabilities, got %#v", caps)
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
		if !caps.Supports(CapabilityTray) {
			t.Fatal("windows should expose tray capability")
		}
		if !caps.Supports(CapabilityClipboard) {
			t.Fatal("windows should expose clipboard capability")
		}
		if !caps.Supports(CapabilityShell) {
			t.Fatal("windows should expose shell capability")
		}
		if !caps.Supports(CapabilitySingleInstance) {
			t.Fatal("windows should expose single instance capability")
		}
		if !caps.Supports(CapabilitySystemRegistration) {
			t.Fatal("windows should expose system registration capability")
		}
		if !caps.Supports(CapabilityGlobalShortcut) {
			t.Fatal("windows should expose global shortcut capability")
		}
		if !caps.Supports(CapabilityDragAndDrop) {
			t.Fatal("windows should expose drag and drop capability")
		}
		return
	}

	if caps.Supports(CapabilityWindow) {
		t.Fatal("unsupported platform should not expose window capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityFileDialog) {
			t.Fatal("macOS/Linux should expose file dialog capability")
		}
	} else if caps.Supports(CapabilityFileDialog) {
		t.Fatal("unsupported platform should not expose file dialog capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityMessageBox) {
			t.Fatal("macOS/Linux should expose message box capability")
		}
	} else if caps.Supports(CapabilityMessageBox) {
		t.Fatal("unsupported platform should not expose message box capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityNotification) {
			t.Fatal("macOS/Linux should expose notification capability")
		}
	} else if caps.Supports(CapabilityNotification) {
		t.Fatal("unsupported platform should not expose notification capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityTray) {
			t.Fatal("macOS/Linux should expose tray capability")
		}
	} else if caps.Supports(CapabilityTray) {
		t.Fatal("unsupported platform should not expose tray capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilitySystemEvents) {
			t.Fatal("macOS/Linux should expose system events capability")
		}
	} else if caps.Supports(CapabilitySystemEvents) {
		t.Fatal("unsupported platform should not expose system events capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityClipboard) {
			t.Fatal("macOS/Linux should expose clipboard capability")
		}
	} else if caps.Supports(CapabilityClipboard) {
		t.Fatal("unsupported platform should not expose clipboard capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityShell) {
			t.Fatal("macOS/Linux should expose shell capability")
		}
	} else if caps.Supports(CapabilityShell) {
		t.Fatal("unsupported platform should not expose shell capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilitySingleInstance) {
			t.Fatal("macOS/Linux should expose single instance capability")
		}
	} else if caps.Supports(CapabilitySingleInstance) {
		t.Fatal("unsupported platform should not expose single instance capability")
	}
	if caps.Supports(CapabilitySystemRegistration) {
		t.Fatal("non-Windows platform should not expose system registration capability")
	}
	if caps.Supports(CapabilityGlobalShortcut) {
		t.Fatal("non-Windows platform should not expose global shortcut capability")
	}
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if !caps.Supports(CapabilityDragAndDrop) {
			t.Fatal("macOS/Linux should expose drag and drop capability")
		}
	} else if caps.Supports(CapabilityDragAndDrop) {
		t.Fatal("unsupported platform should not expose drag and drop capability")
	}
}

func TestReadClipboardTextRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	_, err := ReadClipboardText(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestReadClipboardTextRequiresDriverImplementation(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityClipboard: true}})

	_, err := ReadClipboardText(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestReadClipboardTextChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
	}
	withTestDriver(t, cd)

	_, err := ReadClipboardText(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if cd.readCalled {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestClipboardTextDispatch(t *testing.T) {
	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
		text:       "hello",
	}
	withTestDriver(t, cd)

	text, err := ReadClipboardText(context.Background())
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if !cd.readCalled || text != "hello" {
		t.Fatalf("unexpected read result: called=%v text=%q", cd.readCalled, text)
	}

	if err := WriteClipboardText(context.Background(), "next"); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	if !cd.writeCalled || cd.writtenText != "next" {
		t.Fatalf("unexpected write result: called=%v text=%q", cd.writeCalled, cd.writtenText)
	}
}

func TestWriteClipboardTextPropagatesUnavailable(t *testing.T) {
	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
		err:        fmt.Errorf("wrapped: %w", ErrUnavailable),
	}
	withTestDriver(t, cd)

	if err := WriteClipboardText(context.Background(), "text"); !IsUnavailable(err) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestClipboardFilesRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	if _, err := ReadClipboardFiles(context.Background()); !IsUnsupported(err) {
		t.Fatalf("expected unsupported read files error, got %v", err)
	}
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := WriteClipboardFiles(context.Background(), []string{file}); !IsUnsupported(err) {
		t.Fatalf("expected unsupported write files error, got %v", err)
	}
}

func TestClipboardFilesRequiresDriverImplementation(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityClipboard: true}})

	if _, err := ReadClipboardFiles(context.Background()); !IsUnsupported(err) {
		t.Fatalf("expected unsupported read files error, got %v", err)
	}
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := WriteClipboardFiles(context.Background(), []string{file}); !IsUnsupported(err) {
		t.Fatalf("expected unsupported write files error, got %v", err)
	}
}

func TestClipboardFilesChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
	}
	withTestDriver(t, cd)

	if _, err := ReadClipboardFiles(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled read files, got %v", err)
	}
	if err := WriteClipboardFiles(ctx, []string{"ignored"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled write files, got %v", err)
	}
	if cd.readFilesCalled || cd.writeFilesCalled {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestClipboardFilesDispatchAndNormalization(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("ok"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
		files:      []string{file},
	}
	withTestDriver(t, cd)

	files, err := ReadClipboardFiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected read files error: %v", err)
	}
	if !cd.readFilesCalled || len(files) != 1 || files[0] != file {
		t.Fatalf("unexpected read files result: called=%v files=%#v", cd.readFilesCalled, files)
	}

	if err := WriteClipboardFiles(context.Background(), []string{file}); err != nil {
		t.Fatalf("unexpected write files error: %v", err)
	}
	if !cd.writeFilesCalled || len(cd.writtenFiles) != 1 || cd.writtenFiles[0] != file {
		t.Fatalf("unexpected write files result: called=%v files=%#v", cd.writeFilesCalled, cd.writtenFiles)
	}
}

func TestWriteClipboardFilesRejectsInvalidPaths(t *testing.T) {
	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
	}
	withTestDriver(t, cd)

	if err := WriteClipboardFiles(context.Background(), nil); err == nil {
		t.Fatal("expected empty file list to fail")
	}
	if err := WriteClipboardFiles(context.Background(), []string{""}); err == nil {
		t.Fatal("expected empty file path to fail")
	}
	if err := WriteClipboardFiles(context.Background(), []string{filepath.Join(t.TempDir(), "missing.txt")}); !IsUnavailable(err) {
		t.Fatalf("expected missing path to be unavailable, got %v", err)
	}
	if cd.writeFilesCalled {
		t.Fatal("driver should not be called for invalid file paths")
	}
}

func TestClipboardImageRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})
	data := testClipboardPNG(t)

	if _, err := ReadClipboardImagePNG(context.Background()); !IsUnsupported(err) {
		t.Fatalf("expected unsupported read image error, got %v", err)
	}
	if err := WriteClipboardImagePNG(context.Background(), data); !IsUnsupported(err) {
		t.Fatalf("expected unsupported write image error, got %v", err)
	}
}

func TestClipboardImageRequiresDriverImplementation(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityClipboard: true}})
	data := testClipboardPNG(t)

	if _, err := ReadClipboardImagePNG(context.Background()); !IsUnsupported(err) {
		t.Fatalf("expected unsupported read image error, got %v", err)
	}
	if err := WriteClipboardImagePNG(context.Background(), data); !IsUnsupported(err) {
		t.Fatalf("expected unsupported write image error, got %v", err)
	}
}

func TestClipboardImageChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
	}
	withTestDriver(t, cd)

	if _, err := ReadClipboardImagePNG(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled read image, got %v", err)
	}
	if err := WriteClipboardImagePNG(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled write image, got %v", err)
	}
	if cd.readImageCalled || cd.writeImageCalled {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestClipboardImageDispatchAndCloning(t *testing.T) {
	data := testClipboardPNG(t)
	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
		image:      data,
	}
	withTestDriver(t, cd)

	read, err := ReadClipboardImagePNG(context.Background())
	if err != nil {
		t.Fatalf("unexpected read image error: %v", err)
	}
	if !cd.readImageCalled || !bytes.Equal(read, data) {
		t.Fatalf("unexpected read image result: called=%v len=%d", cd.readImageCalled, len(read))
	}
	read[0] ^= 0xff
	if bytes.Equal(read, cd.image) {
		t.Fatal("read image result should be cloned")
	}

	input := append([]byte(nil), data...)
	if err := WriteClipboardImagePNG(context.Background(), input); err != nil {
		t.Fatalf("unexpected write image error: %v", err)
	}
	input[0] ^= 0xff
	if !cd.writeImageCalled || !bytes.Equal(cd.writtenImage, data) {
		t.Fatalf("unexpected write image result: called=%v len=%d", cd.writeImageCalled, len(cd.writtenImage))
	}
}

func TestWriteClipboardImageRejectsInvalidPNG(t *testing.T) {
	cd := &testClipboardDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityClipboard: true}},
	}
	withTestDriver(t, cd)

	if err := WriteClipboardImagePNG(context.Background(), nil); !IsInvalidTarget(err) {
		t.Fatalf("expected empty image to be invalid target, got %v", err)
	}
	if err := WriteClipboardImagePNG(context.Background(), []byte("not png")); !IsInvalidTarget(err) {
		t.Fatalf("expected invalid image to be invalid target, got %v", err)
	}
	if cd.writeImageCalled {
		t.Fatal("driver should not be called for invalid image data")
	}
}

func TestOpenURLRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	err := OpenURL(context.Background(), "https://example.com")
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestOpenURLRequiresDriverImplementation(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityShell: true}})

	err := OpenURL(context.Background(), "https://example.com")
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestOpenURLValidation(t *testing.T) {
	sd := &testShellDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityShell: true}},
	}
	withTestDriver(t, sd)

	if err := OpenURL(context.Background(), "example.com"); !IsInvalidTarget(err) {
		t.Fatalf("expected URL without scheme to be invalid target, got %v", err)
	}
	if err := OpenURL(context.Background(), ""); !IsInvalidTarget(err) {
		t.Fatalf("expected empty URL to be invalid target, got %v", err)
	}
	if err := OpenPath(context.Background(), ""); !IsInvalidTarget(err) {
		t.Fatalf("expected empty path to be invalid target, got %v", err)
	}
	if err := RevealPath(context.Background(), ""); !IsInvalidTarget(err) {
		t.Fatalf("expected empty reveal path to be invalid target, got %v", err)
	}
	if sd.openURLCalled {
		t.Fatal("driver should not be called for invalid URL")
	}
}

func TestShellChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sd := &testShellDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityShell: true}},
	}
	withTestDriver(t, sd)

	err := OpenPath(ctx, "C:\\tmp")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if sd.openPathCalled {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestShellDispatch(t *testing.T) {
	sd := &testShellDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityShell: true}},
	}
	withTestDriver(t, sd)

	if err := OpenURL(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("unexpected OpenURL error: %v", err)
	}
	if !sd.openURLCalled || sd.target != "https://example.com" {
		t.Fatalf("unexpected OpenURL dispatch: called=%v target=%q", sd.openURLCalled, sd.target)
	}
	if err := OpenPath(context.Background(), "C:\\tmp\\file.txt"); err != nil {
		t.Fatalf("unexpected OpenPath error: %v", err)
	}
	if !sd.openPathCalled || sd.target != "C:\\tmp\\file.txt" {
		t.Fatalf("unexpected OpenPath dispatch: called=%v target=%q", sd.openPathCalled, sd.target)
	}
	if err := RevealPath(context.Background(), "C:\\tmp\\file.txt"); err != nil {
		t.Fatalf("unexpected RevealPath error: %v", err)
	}
	if !sd.revealCalled || sd.target != "C:\\tmp\\file.txt" {
		t.Fatalf("unexpected RevealPath dispatch: called=%v target=%q", sd.revealCalled, sd.target)
	}
}

func TestShellPropagatesUnavailable(t *testing.T) {
	sd := &testShellDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityShell: true}},
		err:        fmt.Errorf("wrapped: %w", ErrUnavailable),
	}
	withTestDriver(t, sd)

	if err := RevealPath(context.Background(), "C:\\tmp\\file.txt"); !IsUnavailable(err) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func testClipboardPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0x30, G: 0x80, B: 0xc0, A: 0xff})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test PNG: %v", err)
	}
	return buf.Bytes()
}

func TestAcquireSingleInstanceRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	instance, err := AcquireSingleInstance(context.Background(), SingleInstanceID("com.example.test"))
	if instance != nil {
		t.Fatalf("expected nil instance, got %#v", instance)
	}
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestAcquireSingleInstanceChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sd := &testSingleInstanceDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySingleInstance: true}},
	}
	withTestDriver(t, sd)

	instance, err := AcquireSingleInstance(ctx, SingleInstanceID("com.example.test"))
	if instance != nil {
		t.Fatalf("expected nil instance, got %#v", instance)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if sd.called {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestAcquireSingleInstanceDispatch(t *testing.T) {
	sd := &testSingleInstanceDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySingleInstance: true}},
	}
	withTestDriver(t, sd)

	called := false
	instance, err := AcquireSingleInstance(context.Background(),
		SingleInstanceID("com.example.dispatch"),
		SingleInstanceArgs("--open", "demo.txt"),
		SingleInstancePayload("fluxui://open/demo.txt"),
		SingleInstancePort(53123),
		SingleInstanceOnSecondLaunch(func(SingleInstanceEvent) {
			called = true
		}),
	)
	if err != nil {
		t.Fatalf("unexpected acquire error: %v", err)
	}
	if instance == nil || instance.Events() == nil {
		t.Fatalf("expected instance and events channel, got %#v", instance)
	}
	if !sd.called {
		t.Fatal("expected driver to be called")
	}
	if sd.opts.id != "com.example.dispatch" {
		t.Fatalf("unexpected id: %q", sd.opts.id)
	}
	if len(sd.opts.args) != 2 || sd.opts.args[0] != "--open" || sd.opts.args[1] != "demo.txt" {
		t.Fatalf("unexpected args: %#v", sd.opts.args)
	}
	if sd.opts.payload != "fluxui://open/demo.txt" || sd.opts.port != 53123 {
		t.Fatalf("unexpected payload/port: payload=%q port=%d", sd.opts.payload, sd.opts.port)
	}
	if sd.opts.onSecondLaunch == nil {
		t.Fatal("expected callback to reach driver")
	}
	sd.opts.onSecondLaunch(SingleInstanceEvent{})
	if !called {
		t.Fatal("expected callback to be callable")
	}
	if err := instance.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
}

func TestLoopbackSingleInstanceForwardsSecondaryLaunch(t *testing.T) {
	id := fmt.Sprintf("com.example.fluxui.test.%d", time.Now().UnixNano())
	primaryOpts := singleInstanceOptions{id: id}
	if err := normalizeSingleInstanceOptions(&primaryOpts); err != nil {
		t.Fatalf("normalize primary options: %v", err)
	}

	handle, err := acquireLoopbackSingleInstance(context.Background(), primaryOpts)
	if err != nil {
		t.Fatalf("unexpected primary acquire error: %v", err)
	}
	defer handle.close()

	secondaryOpts := singleInstanceOptions{
		id:      id,
		args:    []string{"--file", "demo.txt"},
		payload: "fluxui://open/demo.txt",
	}
	if err := normalizeSingleInstanceOptions(&secondaryOpts); err != nil {
		t.Fatalf("normalize secondary options: %v", err)
	}
	_, err = acquireLoopbackSingleInstance(context.Background(), secondaryOpts)
	if !IsAlreadyRunning(err) {
		t.Fatalf("expected already-running error, got %v", err)
	}

	select {
	case event := <-handle.events():
		if event.ID != id || event.Payload != secondaryOpts.payload {
			t.Fatalf("unexpected forwarded event: %#v", event)
		}
		if len(event.Args) != 2 || event.Args[0] != "--file" || event.Args[1] != "demo.txt" {
			t.Fatalf("unexpected forwarded args: %#v", event.Args)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for forwarded single-instance event")
	}
}

func TestSystemRegistrationRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	if err := RegisterProtocolHandler(context.Background(), "demo", `"C:\app.exe"`); !IsUnsupported(err) {
		t.Fatalf("expected unsupported protocol registration error, got %v", err)
	}
	if err := UnregisterProtocolHandler(context.Background(), "demo"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported protocol unregister error, got %v", err)
	}
	if err := RegisterStartupTask(context.Background(), "Demo", `"C:\app.exe"`); !IsUnsupported(err) {
		t.Fatalf("expected unsupported startup registration error, got %v", err)
	}
	if err := RegisterToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}", `"C:\app.exe" --toast`); !IsUnsupported(err) {
		t.Fatalf("expected unsupported toast activator registration error, got %v", err)
	}
	if err := UnregisterToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported toast activator unregister error, got %v", err)
	}
}

func TestSystemRegistrationChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rd := &testRegistrationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySystemRegistration: true}},
	}
	withTestDriver(t, rd)

	if err := RegisterProtocolHandler(ctx, "demo", `"C:\app.exe"`); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if rd.protocolRegistered {
		t.Fatal("driver should not be called when context is already cancelled")
	}
	if err := RegisterToastActivator(ctx, "{01234567-89AB-CDEF-0123-456789ABCDEF}", `"C:\app.exe" --toast`); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled for toast activator registration, got %v", err)
	}
	if rd.activatorRegistered {
		t.Fatal("toast activator driver should not be called when context is already cancelled")
	}
}

func TestSystemRegistrationDispatchAndNormalization(t *testing.T) {
	rd := &testRegistrationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySystemRegistration: true}},
	}
	withTestDriver(t, rd)

	if err := RegisterProtocolHandler(context.Background(),
		"FluxUI:",
		`"C:\Program Files\FluxUI\app.exe"`,
		RegistrationDisplayName("FluxUI Protocol"),
		RegistrationIcon(`C:\Program Files\FluxUI\app.exe,0`),
	); err != nil {
		t.Fatalf("unexpected protocol registration error: %v", err)
	}
	if !rd.protocolRegistered || rd.scheme != "fluxui" {
		t.Fatalf("unexpected protocol registration state: %#v", rd)
	}
	if !strings.Contains(rd.command, "%1") {
		t.Fatalf("protocol command should include %%1 placeholder, got %q", rd.command)
	}
	if rd.opts.displayName != "FluxUI Protocol" || rd.opts.icon == "" {
		t.Fatalf("unexpected registration options: %#v", rd.opts)
	}

	if err := RegisterFileAssociation(context.Background(), "flux", "FluxUI.Document", `"C:\app.exe" "%1"`); err != nil {
		t.Fatalf("unexpected file association error: %v", err)
	}
	if !rd.fileRegistered || rd.extension != ".flux" || rd.progID != "FluxUI.Document" || rd.command != `"C:\app.exe" "%1"` {
		t.Fatalf("unexpected file association state: %#v", rd)
	}

	if err := RegisterStartupTask(context.Background(), "FluxUI", `"C:\app.exe" --tray`); err != nil {
		t.Fatalf("unexpected startup registration error: %v", err)
	}
	if !rd.startupRegistered || rd.name != "FluxUI" || rd.command != `"C:\app.exe" --tray` {
		t.Fatalf("unexpected startup registration state: %#v", rd)
	}

	if err := RegisterToastShortcut(context.Background(),
		"com.example.FluxUI",
		"FluxUI.lnk",
		`C:\Program Files\FluxUI\app.exe`,
		ToastShortcutArguments("--tray", "value with space"),
		ToastShortcutIcon(`C:\Program Files\FluxUI\app.exe,0`),
		ToastShortcutActivatorCLSID("{01234567-89AB-CDEF-0123-456789ABCDEF}"),
	); err != nil {
		t.Fatalf("unexpected toast shortcut registration error: %v", err)
	}
	if !rd.toastRegistered ||
		rd.appID != "com.example.FluxUI" ||
		rd.shortcutName != "FluxUI.lnk" ||
		rd.executable != `C:\Program Files\FluxUI\app.exe` ||
		len(rd.toastOpts.arguments) != 2 ||
		rd.toastOpts.arguments[1] != "value with space" ||
		rd.toastOpts.icon == "" ||
		rd.toastOpts.activatorCLSID != "{01234567-89AB-CDEF-0123-456789ABCDEF}" {
		t.Fatalf("unexpected toast shortcut registration state: %#v", rd)
	}

	if err := UnregisterToastShortcut(context.Background(), "FluxUI"); err != nil {
		t.Fatalf("unexpected toast shortcut unregister error: %v", err)
	}
	if !rd.toastUnregistered || rd.shortcutName != "FluxUI.lnk" {
		t.Fatalf("unexpected toast shortcut unregister state: %#v", rd)
	}

	if err := RegisterToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}", `"C:\app.exe" --toast`); err != nil {
		t.Fatalf("unexpected toast activator registration error: %v", err)
	}
	if !rd.activatorRegistered || rd.clsid != "{01234567-89AB-CDEF-0123-456789ABCDEF}" || rd.command != `"C:\app.exe" --toast` {
		t.Fatalf("unexpected toast activator registration state: %#v", rd)
	}
	if err := UnregisterToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}"); err != nil {
		t.Fatalf("unexpected toast activator unregister error: %v", err)
	}
	if !rd.activatorUnregistered || rd.clsid != "{01234567-89AB-CDEF-0123-456789ABCDEF}" {
		t.Fatalf("unexpected toast activator unregister state: %#v", rd)
	}
}

func TestSystemRegistrationValidation(t *testing.T) {
	rd := &testRegistrationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySystemRegistration: true}},
	}
	withTestDriver(t, rd)

	if err := RegisterProtocolHandler(context.Background(), "bad scheme", `"C:\app.exe"`); err == nil {
		t.Fatal("expected invalid protocol scheme to fail")
	}
	if err := RegisterFileAssociation(context.Background(), ".", "FluxUI.Document", `"C:\app.exe"`); err == nil {
		t.Fatal("expected invalid extension to fail")
	}
	if err := RegisterFileAssociation(context.Background(), ".flux", "", `"C:\app.exe"`); err == nil {
		t.Fatal("expected empty progID to fail")
	}
	if err := RegisterStartupTask(context.Background(), "FluxUI", ""); err == nil {
		t.Fatal("expected empty startup command to fail")
	}
	if err := RegisterToastShortcut(context.Background(), "", "FluxUI", `C:\app.exe`); err == nil {
		t.Fatal("expected empty AppUserModelID to fail")
	}
	if err := RegisterToastShortcut(context.Background(), "com.example.FluxUI", `bad/name`, `C:\app.exe`); err == nil {
		t.Fatal("expected invalid toast shortcut name to fail")
	}
	if err := RegisterToastShortcut(context.Background(), "com.example.FluxUI", "FluxUI", ""); err == nil {
		t.Fatal("expected empty toast shortcut executable to fail")
	}
	if err := RegisterToastActivator(context.Background(), "", `"C:\app.exe" --toast`); err == nil {
		t.Fatal("expected empty toast activator CLSID to fail")
	}
	if err := RegisterToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}", ""); err == nil {
		t.Fatal("expected empty toast activator command to fail")
	}
	if rd.protocolRegistered || rd.fileRegistered || rd.startupRegistered || rd.toastRegistered || rd.activatorRegistered {
		t.Fatal("driver should not be called for invalid registration inputs")
	}
}

func TestGlobalShortcutRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	shortcut, err := RegisterGlobalShortcut(context.Background(), GlobalShortcutSpec{
		ID:        "open",
		Key:       "F9",
		Modifiers: GlobalShortcutControl,
	}, nil)
	if shortcut != nil {
		t.Fatalf("expected nil shortcut, got %#v", shortcut)
	}
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestGlobalShortcutChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	gd := &testGlobalShortcutDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityGlobalShortcut: true}},
	}
	withTestDriver(t, gd)

	_, err := RegisterGlobalShortcut(ctx, GlobalShortcutSpec{
		ID:        "open",
		Key:       "F9",
		Modifiers: GlobalShortcutControl,
	}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if gd.called {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestGlobalShortcutDispatchAndNormalization(t *testing.T) {
	gd := &testGlobalShortcutDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityGlobalShortcut: true}},
	}
	withTestDriver(t, gd)

	var callbackEvent GlobalShortcutEvent
	shortcut, err := RegisterGlobalShortcut(context.Background(), GlobalShortcutSpec{
		ID:        "open",
		Key:       "f9",
		Modifiers: GlobalShortcutControl | GlobalShortcutShift,
	}, func(event GlobalShortcutEvent) {
		callbackEvent = event
	})
	if err != nil {
		t.Fatalf("unexpected global shortcut error: %v", err)
	}
	if shortcut == nil || shortcut.Events() == nil {
		t.Fatalf("expected shortcut and event channel, got %#v", shortcut)
	}
	if !gd.called || gd.spec.ID != "open" || gd.spec.Key != "F9" || gd.spec.Modifiers != GlobalShortcutControl|GlobalShortcutShift {
		t.Fatalf("unexpected global shortcut dispatch: called=%v spec=%#v", gd.called, gd.spec)
	}
	if callbackEvent.ID != "open" || callbackEvent.Key != "F9" {
		t.Fatalf("unexpected callback event: %#v", callbackEvent)
	}
	if err := shortcut.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if err := shortcut.Close(); !IsClosed(err) {
		t.Fatalf("expected closed error on second close, got %v", err)
	}
}

func TestGlobalShortcutValidation(t *testing.T) {
	gd := &testGlobalShortcutDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityGlobalShortcut: true}},
	}
	withTestDriver(t, gd)

	if _, err := RegisterGlobalShortcut(context.Background(), GlobalShortcutSpec{Key: "F9", Modifiers: GlobalShortcutControl}, nil); err == nil {
		t.Fatal("expected empty shortcut id to fail")
	}
	if _, err := RegisterGlobalShortcut(context.Background(), GlobalShortcutSpec{ID: "open", Modifiers: GlobalShortcutControl}, nil); err == nil {
		t.Fatal("expected empty shortcut key to fail")
	}
	if _, err := RegisterGlobalShortcut(context.Background(), GlobalShortcutSpec{ID: "open", Key: "F9"}, nil); err == nil {
		t.Fatal("expected empty shortcut modifiers to fail")
	}
	if gd.called {
		t.Fatal("driver should not be called for invalid shortcut specs")
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
		result:     FileDialogResult{Paths: []string{"C:\\tmp\\demo.txt"}},
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
		result:     FileDialogResult{Paths: []string{"C:\\tmp\\demo.txt"}},
	}
	withTestDriver(t, fd)

	_, err := SaveFileDialog(context.Background(),
		FileDialogTitle("Save"),
		FileDialogDefaultDir("C:\\tmp"),
		FileDialogDefaultName("demo.txt"),
		FileDialogDefaultExtension("txt"),
		FileDialogFilters(filters...),
		FileDialogOwner(12345),
		FileDialogAllowCreateDirs(false),
		FileDialogAllowMissingPath(true),
		FileDialogOverwritePrompt(false),
		FileDialogRememberDir("save-doc"),
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
	if fd.opts.defaultExtension != "txt" {
		t.Fatalf("unexpected default extension: %q", fd.opts.defaultExtension)
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
	if fd.opts.rememberDirKey != "save-doc" {
		t.Fatalf("unexpected remember dir key: %q", fd.opts.rememberDirKey)
	}
	if len(fd.opts.filters) != 1 || fd.opts.filters[0].Name != "Images" || fd.opts.filters[0].Patterns[0] != "*.png" {
		t.Fatalf("filters were not cloned correctly: %#v", fd.opts.filters)
	}
}

func TestFileDialogRemembersSuccessfulDirectory(t *testing.T) {
	key := "recent-test"
	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
		result:     FileDialogResult{Paths: []string{"C:\\projects\\demo\\first.txt"}},
	}
	withTestDriver(t, fd)

	if _, err := OpenFileDialog(context.Background(), FileDialogRememberDir(key)); err != nil {
		t.Fatalf("unexpected first dialog error: %v", err)
	}

	second := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
		result:     FileDialogResult{Paths: []string{"C:\\projects\\demo\\second.txt"}},
	}
	setDriver(second)
	if _, err := OpenFileDialog(context.Background(), FileDialogRememberDir(key)); err != nil {
		t.Fatalf("unexpected second dialog error: %v", err)
	}
	if second.opts.defaultDir != "C:\\projects\\demo" {
		t.Fatalf("expected remembered default dir, got %q", second.opts.defaultDir)
	}
}

func TestSaveFileDialogAppendsDefaultExtension(t *testing.T) {
	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
		result:     FileDialogResult{Paths: []string{"C:\\projects\\demo\\report"}},
	}
	withTestDriver(t, fd)

	result, err := SaveFileDialog(context.Background(), FileDialogDefaultExtension(".txt"))
	if err != nil {
		t.Fatalf("unexpected save dialog error: %v", err)
	}
	if len(result.Paths) != 1 || result.Paths[0] != "C:\\projects\\demo\\report.txt" {
		t.Fatalf("expected default extension to be appended, got %#v", result.Paths)
	}
	if fd.result.Paths[0] != "C:\\projects\\demo\\report" {
		t.Fatalf("driver result should not be mutated, got %#v", fd.result.Paths)
	}
}

func TestSaveFileDialogKeepsExistingExtension(t *testing.T) {
	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
		result:     FileDialogResult{Paths: []string{"C:\\projects\\demo\\report.md"}},
	}
	withTestDriver(t, fd)

	result, err := SaveFileDialog(context.Background(), FileDialogDefaultExtension("txt"))
	if err != nil {
		t.Fatalf("unexpected save dialog error: %v", err)
	}
	if len(result.Paths) != 1 || result.Paths[0] != "C:\\projects\\demo\\report.md" {
		t.Fatalf("expected existing extension to be preserved, got %#v", result.Paths)
	}
}

func TestFileDialogNormalizesRelativeResultPath(t *testing.T) {
	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
		result:     FileDialogResult{Paths: []string{"relative-file.txt"}},
	}
	withTestDriver(t, fd)

	result, err := OpenFileDialog(context.Background())
	if err != nil {
		t.Fatalf("unexpected dialog error: %v", err)
	}
	if len(result.Paths) != 1 || !filepath.IsAbs(result.Paths[0]) {
		t.Fatalf("expected absolute result path, got %#v", result.Paths)
	}
}

func TestFileDialogStructuredPathErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		result FileDialogResult
	}{
		{name: "missing paths", result: FileDialogResult{}},
		{name: "empty path", result: FileDialogResult{Paths: []string{""}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fd := &testFileDialogDriver{
				testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
				result:     tt.result,
			}
			withTestDriver(t, fd)

			if _, err := OpenFileDialog(context.Background()); !IsFileDialogErrorKind(err, FileDialogErrorPath) {
				t.Fatalf("expected structured path error, got %v", err)
			}
		})
	}
}

func TestFileDialogStructuredErrorKind(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &FileDialogError{
		Kind: FileDialogErrorDefaultDir,
		Path: "C:\\missing",
		Err:  ErrUnavailable,
	})
	if !IsFileDialogErrorKind(err, FileDialogErrorDefaultDir) {
		t.Fatalf("expected default dir error kind, got %v", err)
	}
	if IsFileDialogErrorKind(err, FileDialogErrorSelectedPath) {
		t.Fatalf("unexpected selected path error kind, got %v", err)
	}
}

func TestFileDialogDefaultOptions(t *testing.T) {
	fd := &testFileDialogDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityFileDialog: true}},
		result:     FileDialogResult{Paths: []string{"C:\\tmp\\demo.txt"}},
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
		MessageBoxDetails("Unsaved changes will be lost."),
		MessageBoxFooter("FluxUI"),
		MessageBoxVerification("Do not ask again", true),
		MessageBoxStyle(MessageBoxQuestion),
		MessageBoxButtonSet(MessageBoxYesNo),
		MessageBoxDefaultButton(MessageBoxResultNo),
		MessageBoxCommandLinks(true),
		MessageBoxExpandedDetailsByDefault(true),
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
	if mb.opts.details != "Unsaved changes will be lost." {
		t.Fatalf("unexpected details: %q", mb.opts.details)
	}
	if mb.opts.footer != "FluxUI" {
		t.Fatalf("unexpected footer: %q", mb.opts.footer)
	}
	if mb.opts.verificationText != "Do not ask again" || !mb.opts.verificationChecked {
		t.Fatalf("unexpected verification option: text=%q checked=%v", mb.opts.verificationText, mb.opts.verificationChecked)
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
	if !mb.opts.commandLinks || !mb.opts.expandedDetailsByDefault {
		t.Fatalf("expected rich task dialog flags, got commandLinks=%v expanded=%v", mb.opts.commandLinks, mb.opts.expandedDetailsByDefault)
	}
	if mb.opts.owner != 12345 {
		t.Fatalf("unexpected owner: %d", mb.opts.owner)
	}
}

func TestShowMessageBoxDetailedOptions(t *testing.T) {
	expected := MessageBoxDetailedResult{
		Result:              MessageBoxResultCustom,
		ButtonID:            "archive",
		VerificationChecked: true,
	}
	mb := &testMessageBoxDriver{
		testDriver:     testDriver{caps: CapabilitySet{CapabilityMessageBox: true}},
		detailedResult: expected,
	}
	withTestDriver(t, mb)

	buttons := []MessageBoxButton{
		{ID: "archive", Label: "Archive", Result: MessageBoxResultCustom},
		{ID: "delete", Label: "Delete", Result: MessageBoxResultNo},
	}
	got, err := ShowMessageBoxDetailed(context.Background(),
		MessageBoxCustomButtons(buttons...),
		MessageBoxDefaultButtonID("delete"),
		MessageBoxCommandLinksNoIcon(true),
		MessageBoxExpandDetailsInFooterArea(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mb.detailedCalled {
		t.Fatal("expected detailed driver to be called")
	}
	if got != expected {
		t.Fatalf("expected detailed result %#v, got %#v", expected, got)
	}

	buttons[0].Label = "Mutated"
	if len(mb.opts.customButtons) != 2 || mb.opts.customButtons[0].Label != "Archive" {
		t.Fatalf("custom buttons should be cloned, got %#v", mb.opts.customButtons)
	}
	if mb.opts.defaultButtonID != "delete" {
		t.Fatalf("unexpected default button ID: %q", mb.opts.defaultButtonID)
	}
	if !mb.opts.commandLinksNoIcon || !mb.opts.expandDetailsInFooterArea {
		t.Fatalf("expected command link/footer flags, got noIcon=%v footer=%v", mb.opts.commandLinksNoIcon, mb.opts.expandDetailsInFooterArea)
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

func TestShowMessageBoxAsync(t *testing.T) {
	mb := &testMessageBoxDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityMessageBox: true}},
		result:     MessageBoxResultRetry,
	}
	withTestDriver(t, mb)

	ch := ShowMessageBoxAsync(context.Background(), MessageBoxTitle("Async"))
	select {
	case response := <-ch:
		if response.Err != nil {
			t.Fatalf("unexpected async error: %v", response.Err)
		}
		if response.Result != MessageBoxResultRetry {
			t.Fatalf("unexpected async result: %q", response.Result)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async message box response")
	}
	if !mb.called || mb.opts.title != "Async" {
		t.Fatalf("expected async call to reach driver, called=%v opts=%#v", mb.called, mb.opts)
	}
}

func TestShowMessageBoxDetailedAsync(t *testing.T) {
	expected := MessageBoxDetailedResult{
		Result:   MessageBoxResultCustom,
		ButtonID: "details",
	}
	mb := &testMessageBoxDriver{
		testDriver:     testDriver{caps: CapabilitySet{CapabilityMessageBox: true}},
		detailedResult: expected,
	}
	withTestDriver(t, mb)

	ch := ShowMessageBoxDetailedAsync(context.Background(), MessageBoxTitle("Detailed Async"))
	select {
	case response := <-ch:
		if response.Err != nil {
			t.Fatalf("unexpected async error: %v", response.Err)
		}
		if response.Result != expected {
			t.Fatalf("unexpected async detailed result: %#v", response.Result)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for async detailed message box response")
	}
	if !mb.detailedCalled || mb.opts.title != "Detailed Async" {
		t.Fatalf("expected async detailed call to reach driver, called=%v opts=%#v", mb.detailedCalled, mb.opts)
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
	dismissed := false
	actioned := false
	var gotEvent NotificationEvent
	var gotDismissEvent NotificationEvent
	var gotActionEvent NotificationEvent
	actions := []NotificationAction{
		{ID: "open", Label: "Open"},
		{ID: "dismiss", Label: "Dismiss"},
	}
	err := Notify(context.Background(),
		NotificationTitle("Done"),
		NotificationBody("Export finished."),
		NotificationKindStyle(NotificationSuccess),
		NotificationIcon("C:\\tmp\\icon.ico"),
		NotificationGroup("exports"),
		NotificationBackendPath(NotificationBackendToast),
		NotificationAppID("com.example.FluxUI"),
		NotificationActions(actions...),
		NotificationTimeout(5*time.Second),
		NotificationOnClick(func(event NotificationEvent) {
			clicked = true
			gotEvent = event
		}),
		NotificationOnDismiss(func(event NotificationEvent) {
			dismissed = true
			gotDismissEvent = event
		}),
		NotificationOnAction(func(event NotificationEvent) {
			actioned = true
			gotActionEvent = event
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
	if nd.opts.backend != NotificationBackendToast {
		t.Fatalf("unexpected backend: %q", nd.opts.backend)
	}
	if nd.opts.appID != "com.example.FluxUI" {
		t.Fatalf("unexpected appID: %q", nd.opts.appID)
	}
	actions[0].Label = "Mutated"
	if len(nd.opts.actions) != 2 || nd.opts.actions[0].Label != "Open" || nd.opts.actions[1].ID != "dismiss" {
		t.Fatalf("notification actions should be cloned, got %#v", nd.opts.actions)
	}
	if nd.opts.timeout != 5*time.Second {
		t.Fatalf("unexpected timeout: %s", nd.opts.timeout)
	}
	if nd.opts.onClick == nil {
		t.Fatal("expected onClick option to reach driver")
	}
	if nd.opts.onDismiss == nil {
		t.Fatal("expected onDismiss option to reach driver")
	}
	if nd.opts.onAction == nil {
		t.Fatal("expected onAction option to reach driver")
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
	nd.opts.onDismiss(NotificationEvent{
		Kind:  NotificationEventDismissed,
		Group: "exports",
	})
	if !dismissed {
		t.Fatal("expected dismiss callback to run when driver invokes it")
	}
	if gotDismissEvent.Kind != NotificationEventDismissed || gotDismissEvent.Group != "exports" {
		t.Fatalf("unexpected dismiss event: %#v", gotDismissEvent)
	}
	nd.opts.onAction(NotificationEvent{
		Kind:   NotificationEventAction,
		Group:  "exports",
		Action: "open",
	})
	if !actioned {
		t.Fatal("expected action callback to run when driver invokes it")
	}
	if gotActionEvent.Kind != NotificationEventAction || gotActionEvent.Group != "exports" || gotActionEvent.Action != "open" {
		t.Fatalf("unexpected action event: %#v", gotActionEvent)
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

func TestProbeNotificationBackendRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	probe := ProbeNotificationBackend(context.Background(), NotificationBackendToast)
	if probe.Status != CapabilityStatusUnsupported || probe.Supported() || probe.Available() || !IsUnsupported(probe.Err) {
		t.Fatalf("unexpected unsupported backend probe: %#v", probe)
	}
}

func TestProbeNotificationBackendRequiresDriverImplementation(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityNotification: true}})

	probe := ProbeNotificationBackend(context.Background(), NotificationBackendBalloon)
	if probe.Status != CapabilityStatusUnsupported || probe.Supported() || probe.Available() || !IsUnsupported(probe.Err) {
		t.Fatalf("unexpected missing driver backend probe: %#v", probe)
	}
}

func TestProbeNotificationBackendDispatchesDriver(t *testing.T) {
	nd := &testNotificationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
		probeResult: NotificationBackendProbe{
			Backend:                    NotificationBackendToast,
			Status:                     CapabilityStatusAvailable,
			SupportsActionButtons:      true,
			SupportsProtocolActivation: true,
		},
	}
	withTestDriver(t, nd)

	probe := ProbeNotificationBackend(context.Background(),
		NotificationBackendToast,
		NotificationAppID("com.example.FluxUI"),
		NotificationActions(NotificationAction{ID: "open", Label: "Open"}),
	)
	if !probe.Supported() || !probe.Available() || !probe.SupportsActionButtons || !probe.SupportsProtocolActivation {
		t.Fatalf("unexpected available backend probe: %#v", probe)
	}
	if nd.probeBackend != NotificationBackendToast {
		t.Fatalf("expected toast backend to reach driver, got %q", nd.probeBackend)
	}
	if nd.probeOpts.appID != "com.example.FluxUI" {
		t.Fatalf("expected appID option to reach probe, got %q", nd.probeOpts.appID)
	}
	if len(nd.probeOpts.actions) != 1 || nd.probeOpts.actions[0].ID != "open" {
		t.Fatalf("expected action option to reach probe, got %#v", nd.probeOpts.actions)
	}
}

func TestProbeNotificationBackendContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	nd := &testNotificationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, nd)

	probe := ProbeNotificationBackend(ctx, NotificationBackendToast)
	if probe.Status != CapabilityStatusUnavailable || probe.Available() || !errors.Is(probe.Err, context.Canceled) {
		t.Fatalf("unexpected canceled backend probe: %#v", probe)
	}
	if nd.probeBackend != "" {
		t.Fatal("driver should not be called when context is already canceled")
	}
}

func TestStartToastActivatorRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	activator, err := StartToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}", func(ToastActivationEvent) {})
	if activator != nil {
		t.Fatalf("expected nil activator, got %#v", activator)
	}
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestStartToastActivatorValidation(t *testing.T) {
	driver := &testToastActivatorDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, driver)

	if _, err := StartToastActivator(context.Background(), "", func(ToastActivationEvent) {}); err == nil {
		t.Fatal("expected empty CLSID to fail")
	}
	if _, err := StartToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}", nil); err == nil {
		t.Fatal("expected nil callback to fail")
	}
	if driver.called {
		t.Fatal("driver should not be called for invalid StartToastActivator inputs")
	}
}

func TestStartToastActivatorDispatchesDriver(t *testing.T) {
	driver := &testToastActivatorDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, driver)

	activator, err := StartToastActivator(context.Background(), "{01234567-89AB-CDEF-0123-456789ABCDEF}", func(ToastActivationEvent) {})
	if err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	if activator == nil {
		t.Fatal("expected activator handle")
	}
	if !driver.called || driver.clsid != "{01234567-89AB-CDEF-0123-456789ABCDEF}" || driver.fn == nil {
		t.Fatalf("unexpected toast activator driver state: %#v", driver)
	}
	if err := activator.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !driver.handle.closed {
		t.Fatal("expected underlying handle to close")
	}
	select {
	case <-activator.Done():
	default:
		t.Fatal("expected Done channel to be closed after Close")
	}
}

func TestStartToastActivatorChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	driver := &testToastActivatorDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, driver)

	_, err := StartToastActivator(ctx, "{01234567-89AB-CDEF-0123-456789ABCDEF}", func(ToastActivationEvent) {})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if driver.called {
		t.Fatal("driver should not be called when context is cancelled")
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

func TestCancelNotificationGroup(t *testing.T) {
	nd := &testNotificationDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityNotification: true}},
	}
	withTestDriver(t, nd)

	if err := CancelNotificationGroup(context.Background(), "exports"); err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}
	if nd.canceledGroup != "exports" {
		t.Fatalf("expected canceled group exports, got %q", nd.canceledGroup)
	}
}

func TestCancelNotificationGroupRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	if err := CancelNotificationGroup(context.Background(), "exports"); !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestSubscribeSystemEventsRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	_, err := SubscribeSystemEvents(context.Background())
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestSubscribeSystemEventsChecksContextBeforeDriver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	sd := &testSystemEventDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySystemEvents: true}},
	}
	withTestDriver(t, sd)

	_, err := SubscribeSystemEvents(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if sd.called {
		t.Fatal("driver should not be called when context is already cancelled")
	}
}

func TestSubscribeSystemEventsRejectsUnknownKind(t *testing.T) {
	sd := &testSystemEventDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySystemEvents: true}},
	}
	withTestDriver(t, sd)

	_, err := SubscribeSystemEvents(context.Background(), SystemEventKind("not-real"))
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported unknown system event kind, got %v", err)
	}
	if sd.called {
		t.Fatal("driver should not be called for unknown system event kind")
	}
}

func TestSubscribeSystemEvents(t *testing.T) {
	handle := &testSystemEventHandle{ch: make(chan SystemEvent, 1)}
	sd := &testSystemEventDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilitySystemEvents: true}},
		handle:     handle,
	}
	withTestDriver(t, sd)

	sub, err := SubscribeSystemEvents(context.Background(), SystemEventDisplayChanged, SystemEventThemeChanged)
	if err != nil {
		t.Fatalf("unexpected subscribe error: %v", err)
	}
	if !sd.called {
		t.Fatal("expected driver to be called")
	}
	if len(sd.kinds) != 2 || sd.kinds[0] != SystemEventDisplayChanged || sd.kinds[1] != SystemEventThemeChanged {
		t.Fatalf("unexpected subscribed kinds: %#v", sd.kinds)
	}
	handle.ch <- SystemEvent{Kind: SystemEventDisplayChanged, Detail: "1920x1080"}
	select {
	case event := <-sub.Events():
		if event.Kind != SystemEventDisplayChanged || event.Detail != "1920x1080" {
			t.Fatalf("unexpected event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for system event")
	}
	if err := sub.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if !handle.closed {
		t.Fatal("expected handle to be closed")
	}
	if err := sub.Close(); !IsClosed(err) {
		t.Fatalf("expected second close to report closed, got %v", err)
	}
}

func TestNewTrayRequiresCapability(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{}})

	_, err := NewTray()
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestNewTrayRequiresDriverImplementation(t *testing.T) {
	withTestDriver(t, testDriver{caps: CapabilitySet{CapabilityTray: true}})

	_, err := NewTray()
	if !IsUnsupported(err) {
		t.Fatalf("expected unsupported error, got %v", err)
	}
}

func TestNewTrayOptionsAndCallbacks(t *testing.T) {
	td := &testTrayDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityTray: true}},
	}
	withTestDriver(t, td)

	clicked := make(chan TrayEvent, 1)
	doubleClicked := make(chan TrayEvent, 1)
	menuClicked := make(chan TrayEvent, 1)
	menu := TrayMenu{
		TrayMenuAction("open", "Open", func(event TrayEvent) {
			menuClicked <- event
		}),
		TrayMenuSeparator(),
		{ID: "enabled", Label: "Enabled"},
		{ID: "disabled", Label: "Disabled", Disabled: true},
		{ID: "checked", Label: "Checked", Checked: true},
	}

	tray, err := NewTray(
		TrayIcon("C:\\tmp\\tray.ico"),
		TrayIconBytes([]byte{1, 2, 3}),
		TrayIconResource(101),
		TrayTooltip("FluxUI"),
		TrayMenuItems(menu...),
		TrayOnClick(func(event TrayEvent) {
			clicked <- event
		}),
		TrayOnDoubleClick(func(event TrayEvent) {
			doubleClicked <- event
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tray == nil {
		t.Fatal("expected tray")
	}
	if !td.called {
		t.Fatal("expected tray driver to be called")
	}
	if td.opts.icon != "" || len(td.opts.iconData) != 0 || td.opts.iconResource != 101 {
		t.Fatalf("unexpected icon source: path=%q bytes=%v resource=%d", td.opts.icon, td.opts.iconData, td.opts.iconResource)
	}
	if td.opts.tooltip != "FluxUI" {
		t.Fatalf("unexpected tooltip: %q", td.opts.tooltip)
	}
	if len(td.opts.menu) != len(menu) {
		t.Fatalf("expected %d menu items, got %d", len(menu), len(td.opts.menu))
	}
	if !td.opts.menu[1].Separator || !td.opts.menu[3].Disabled || !td.opts.menu[4].Checked {
		t.Fatalf("menu item flags were not preserved: %#v", td.opts.menu)
	}

	menu[0].Label = "Mutated"
	if td.opts.menu[0].Label != "Open" {
		t.Fatalf("tray menu option should be cloned, got %q", td.opts.menu[0].Label)
	}

	td.opts.onClick(TrayEvent{Kind: TrayEventClicked})
	assertTrayEvent(t, clicked, TrayEventClicked, "")
	td.opts.onDoubleClick(TrayEvent{Kind: TrayEventDoubleClick})
	assertTrayEvent(t, doubleClicked, TrayEventDoubleClick, "")
	td.opts.menu[0].OnClick(TrayEvent{Kind: TrayEventMenuItem, ItemID: "open"})
	assertTrayEvent(t, menuClicked, TrayEventMenuItem, "open")
}

func TestTrayLifecycleMethods(t *testing.T) {
	handle := &testTrayHandle{}
	td := &testTrayDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityTray: true}},
		handle:     handle,
	}
	withTestDriver(t, td)

	tray, err := NewTray()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := tray.SetIcon("C:\\tmp\\next.ico"); err != nil {
		t.Fatalf("unexpected SetIcon error: %v", err)
	}
	if handle.icon != "C:\\tmp\\next.ico" {
		t.Fatalf("unexpected icon: %q", handle.icon)
	}
	if err := tray.SetIconBytes([]byte{4, 5, 6}); err != nil {
		t.Fatalf("unexpected SetIconBytes error: %v", err)
	}
	if len(handle.iconData) != 3 || handle.iconData[0] != 4 {
		t.Fatalf("unexpected icon data: %#v", handle.iconData)
	}
	if err := tray.SetIconResource(202); err != nil {
		t.Fatalf("unexpected SetIconResource error: %v", err)
	}
	if handle.iconResource != 202 {
		t.Fatalf("unexpected icon resource: %d", handle.iconResource)
	}
	if err := tray.SetTooltip("Ready"); err != nil {
		t.Fatalf("unexpected SetTooltip error: %v", err)
	}
	if handle.tooltip != "Ready" {
		t.Fatalf("unexpected tooltip: %q", handle.tooltip)
	}

	menu := TrayMenu{{ID: "quit", Label: "Quit"}}
	if err := tray.SetMenu(menu); err != nil {
		t.Fatalf("unexpected SetMenu error: %v", err)
	}
	menu[0].Label = "Mutated"
	if len(handle.menu) != 1 || handle.menu[0].Label != "Quit" {
		t.Fatalf("expected menu to be cloned, got %#v", handle.menu)
	}

	if err := tray.Show(); err != nil {
		t.Fatalf("unexpected Show error: %v", err)
	}
	if err := tray.Hide(); err != nil {
		t.Fatalf("unexpected Hide error: %v", err)
	}
	if handle.showCount != 1 || handle.hideCount != 1 {
		t.Fatalf("unexpected show/hide counts: %d/%d", handle.showCount, handle.hideCount)
	}
	if err := tray.Close(); err != nil {
		t.Fatalf("unexpected Close error: %v", err)
	}
	if handle.closeCount != 1 {
		t.Fatalf("expected close count 1, got %d", handle.closeCount)
	}
	if err := tray.SetIcon("C:\\tmp\\closed.ico"); !IsClosed(err) {
		t.Fatalf("expected closed error after Close, got %v", err)
	}
	if !tray.Closed() {
		t.Fatal("expected Closed to report true after Close")
	}
	if err := tray.Show(); !IsClosed(err) {
		t.Fatalf("expected closed Show error, got %v", err)
	}
	if err := tray.Close(); !IsClosed(err) {
		t.Fatalf("expected second Close to report closed, got %v", err)
	}
	if handle.closeCount != 1 {
		t.Fatalf("closed tray should not call driver again, got close count %d", handle.closeCount)
	}
}

func TestTrayVisibleClosedAndMenuProvider(t *testing.T) {
	handle := &testTrayHandle{}
	td := &testTrayDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityTray: true}},
		handle:     handle,
	}
	withTestDriver(t, td)

	tray, err := NewTray(TrayMenuProvider(func() TrayMenu {
		return TrayMenu{
			{
				ID:      "root",
				Label:   "Root",
				Default: true,
				Children: TrayMenu{
					TrayMenuAction("child", "Child", nil),
				},
			},
		}
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tray.Closed() {
		t.Fatal("new tray should not be closed")
	}
	if tray.Visible() {
		t.Fatal("new tray should not be visible until Show")
	}
	if err := tray.Show(); err != nil {
		t.Fatalf("unexpected Show error: %v", err)
	}
	if !tray.Visible() {
		t.Fatal("tray should be visible after Show")
	}

	if td.opts.menuProvider == nil {
		t.Fatal("expected menu provider to be installed")
	}
	menu := td.opts.menuProvider()
	if len(menu) != 1 || !menu[0].Default || len(menu[0].Children) != 1 || menu[0].Children[0].ID != "child" {
		t.Fatalf("unexpected provider menu: %#v", menu)
	}
	menu[0].Children[0].ID = "mutated"
	nextMenu := td.opts.menuProvider()
	if nextMenu[0].Children[0].ID != "child" {
		t.Fatalf("menu provider result should be cloned, got %#v", nextMenu)
	}

	if err := tray.SetMenuProvider(func() TrayMenu {
		return TrayMenu{TrayMenuAction("runtime", "Runtime", nil)}
	}); err != nil {
		t.Fatalf("unexpected SetMenuProvider error: %v", err)
	}
	if handle.menuProvider == nil {
		t.Fatal("expected runtime menu provider to be installed on handle")
	}
	runtimeMenu := handle.menuProvider()
	if len(runtimeMenu) != 1 || runtimeMenu[0].ID != "runtime" {
		t.Fatalf("unexpected runtime provider menu: %#v", runtimeMenu)
	}

	if err := tray.Hide(); err != nil {
		t.Fatalf("unexpected Hide error: %v", err)
	}
	if tray.Visible() {
		t.Fatal("tray should not be visible after Hide")
	}
}

func TestCloseTraysClosesRegisteredTrays(t *testing.T) {
	firstHandle := &testTrayHandle{}
	firstDriver := &testTrayDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityTray: true}},
		handle:     firstHandle,
	}
	withTestDriver(t, firstDriver)
	first, err := NewTray()
	if err != nil {
		t.Fatalf("unexpected first tray error: %v", err)
	}

	secondHandle := &testTrayHandle{}
	secondDriver := &testTrayDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityTray: true}},
		handle:     secondHandle,
	}
	setDriver(secondDriver)
	second, err := NewTray()
	if err != nil {
		t.Fatalf("unexpected second tray error: %v", err)
	}

	if err := CloseTrays(); err != nil {
		t.Fatalf("unexpected CloseTrays error: %v", err)
	}
	if firstHandle.closeCount != 1 || secondHandle.closeCount != 1 {
		t.Fatalf("expected both trays to close, got %d and %d", firstHandle.closeCount, secondHandle.closeCount)
	}
	if !first.Closed() || !second.Closed() {
		t.Fatal("CloseTrays should mark trays closed")
	}
}

func TestTrayNilReceiverReportsClosed(t *testing.T) {
	var tray *Tray
	if !tray.Closed() {
		t.Fatal("nil tray should report closed")
	}
	if tray.Visible() {
		t.Fatal("nil tray should not report visible")
	}
	if err := tray.Show(); !IsClosed(err) {
		t.Fatalf("expected nil tray Show to report closed, got %v", err)
	}
	if err := tray.Close(); !IsClosed(err) {
		t.Fatalf("expected nil tray Close to report closed, got %v", err)
	}
}

func TestNewTrayPropagatesUnavailable(t *testing.T) {
	td := &testTrayDriver{
		testDriver: testDriver{caps: CapabilitySet{CapabilityTray: true}},
		err:        fmt.Errorf("wrapped: %w", ErrUnavailable),
	}
	withTestDriver(t, td)

	_, err := NewTray()
	if !IsUnavailable(err) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func assertTrayEvent(t *testing.T, ch <-chan TrayEvent, kind TrayEventKind, itemID string) {
	t.Helper()

	select {
	case event := <-ch:
		if event.Kind != kind || event.ItemID != itemID {
			t.Fatalf("unexpected tray event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for tray event %q", kind)
	}
}
