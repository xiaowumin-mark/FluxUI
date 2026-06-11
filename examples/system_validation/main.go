package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xiaowumin-mark/FluxUI/system"
)

var coreCapabilities = []system.Capability{
	system.CapabilityWindow,
	system.CapabilityFileDialog,
	system.CapabilityMessageBox,
	system.CapabilityNotification,
	system.CapabilityTray,
	system.CapabilitySystemEvents,
	system.CapabilityClipboard,
	system.CapabilityShell,
	system.CapabilitySingleInstance,
	system.CapabilitySystemRegistration,
	system.CapabilityGlobalShortcut,
	system.CapabilityDragAndDrop,
}

const validationToastActivatorCLSID = "{01234567-89AB-CDEF-0123-456789ABCDEF}"

func main() {
	var (
		failUnavailable   = flag.Bool("fail-unavailable", runtime.GOOS == "windows", "fail when Windows PA-F capabilities are unavailable")
		requireToast      = flag.Bool("require-toast", false, "fail when the Toast backend probe is unavailable")
		dialogs           = flag.Bool("dialogs", false, "show native file dialog/message box and close them via context cancellation")
		messageBoxSources = flag.Bool("messagebox-sources", false, "show native TaskDialog message boxes and verify Cancel/Escape/Close result sources")
		ownerModal        = flag.Bool("owner-modal", false, "show owner-bound file dialog/message box and verify owner modal disabling")
		notify            = flag.Bool("notify", false, "send and cancel a real notification group")
		toast             = flag.Bool("toast", false, "send and cancel a real Toast notification with action buttons")
		toastShortcut     = flag.Bool("toast-shortcut", false, "create and remove a Start Menu AppUserModelID shortcut for Toast")
		toastActivator    = flag.Bool("toast-activator", false, "create and remove a current-user Toast COM LocalServer registration")
		tray              = flag.Bool("tray", false, "create, show, update, hide, and close a real tray icon")
		clipboard         = flag.Bool("clipboard", false, "write/read/restore system clipboard text")
		clipboardFiles    = flag.Bool("clipboard-files", false, "write/read system clipboard file list, then restore clipboard text when possible")
		clipboardImage    = flag.Bool("clipboard-image", false, "write/read system clipboard image as PNG, then restore image or text when possible")
		shell             = flag.Bool("shell", false, "open/reveal a stable workspace path with the system shell")
		shellErrors       = flag.Bool("shell-errors", false, "verify shell invalid target and missing target error classification")
		singleInstance    = flag.Bool("single-instance", false, "acquire primary single-instance channel and simulate a secondary launch")
		dragDrop          = flag.Bool("drag-drop", false, "print the drag-and-drop capability probe")
		trayResourceID    = flag.Int("tray-resource-id", 0, "optional process icon resource id to test with TrayIconResource")
		events            = flag.Duration("events", 0, "listen for system events for the given duration, for example 10s")
		timeout           = flag.Duration("timeout", 60*time.Second, "overall timeout for validation operations")
		toastAppID        = flag.String("toast-app-id", "FluxUI", "Windows Toast AppUserModelID used for Toast probe/notification tests")
		toastShortcutName = flag.String("toast-shortcut-name", "FluxUI Validation", "Start Menu shortcut name used by -toast-shortcut")
		toastActivatorID  = flag.String("toast-activator-clsid", "", "optional Windows Toast COM activator CLSID for -toast-shortcut or -toast-activator")
	)
	flag.Parse()
	activatorCLSID := strings.TrimSpace(*toastActivatorID)
	if *toastActivator && activatorCLSID == "" {
		activatorCLSID = validationToastActivatorCLSID
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	failures := 0
	fmt.Printf("FluxUI System API validation (%s)\n", runtime.GOOS)
	for _, cap := range coreCapabilities {
		availability := system.Probe(cap)
		required := *failUnavailable && runtime.GOOS == "windows"
		if reportAvailability("probe."+string(cap), availability, required) {
			failures++
		}
	}
	if reportBackendProbe("notification.balloon", system.ProbeNotificationBackend(ctx, system.NotificationBackendBalloon), *failUnavailable && runtime.GOOS == "windows") {
		failures++
	}
	if reportBackendProbe("notification.toast", system.ProbeNotificationBackend(ctx, system.NotificationBackendToast, system.NotificationAppID(*toastAppID)), *requireToast) {
		failures++
	}

	if *notify {
		if err := validateNotificationGroup(ctx); err != nil {
			fmt.Printf("FAIL notification.group: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS notification.group: replacement and cancel submitted")
		}
	}
	if *toast {
		if err := validateToastNotification(ctx, *toastAppID); err != nil {
			fmt.Printf("FAIL notification.toast: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS notification.toast: toast with action buttons submitted and cancelled")
		}
	}
	if *toastShortcut {
		if err := validateToastShortcut(ctx, *toastAppID, *toastShortcutName, activatorCLSID); err != nil {
			fmt.Printf("FAIL notification.toast_shortcut: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS notification.toast_shortcut: Start Menu AppUserModelID shortcut created and removed")
		}
	}
	if *toastActivator {
		if err := validateToastActivatorRegistration(ctx, activatorCLSID); err != nil {
			fmt.Printf("FAIL notification.toast_activator: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS notification.toast_activator: COM LocalServer registration created and removed")
		}
	}
	if *dialogs {
		if err := validateDialogCancellation(); err != nil {
			fmt.Printf("FAIL dialogs.cancel_context: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS dialogs.cancel_context: file dialog and message box auto-cancel completed")
		}
	}
	if *messageBoxSources {
		if err := validateMessageBoxSources(); err != nil {
			fmt.Printf("FAIL message_box.sources: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS message_box.sources: Cancel/Escape/Close sources completed")
		}
	}
	if *ownerModal {
		if err := validateOwnerModal(); err != nil {
			fmt.Printf("FAIL owner_modal: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS owner_modal: message box and file dialog disabled their owner window")
		}
	}
	if *tray {
		if err := validateTray(*trayResourceID); err != nil {
			fmt.Printf("FAIL tray.lifecycle: %v\n", err)
			failures++
		} else if *trayResourceID > 0 {
			fmt.Println("PASS tray.lifecycle: show/hide/bytes-icon/path-icon/resource-icon/state/close completed")
		} else {
			fmt.Println("PASS tray.lifecycle: show/hide/bytes-icon/path-icon/state/close completed")
		}
	}
	if *clipboard {
		if err := validateClipboard(ctx); err != nil {
			fmt.Printf("FAIL clipboard.text: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS clipboard.text: write/read/restore completed")
		}
	}
	if *clipboardFiles {
		if err := validateClipboardFiles(ctx); err != nil {
			fmt.Printf("FAIL clipboard.files: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS clipboard.files: write/read file list completed")
		}
	}
	if *clipboardImage {
		if err := validateClipboardImage(ctx); err != nil {
			fmt.Printf("FAIL clipboard.image: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS clipboard.image: write/read PNG image completed")
		}
	}
	if *shell {
		if err := validateShell(ctx); err != nil {
			fmt.Printf("FAIL shell.open: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS shell.open: open URL/path and reveal path submitted")
		}
	}
	if *shellErrors {
		if err := validateShellErrors(ctx); err != nil {
			fmt.Printf("FAIL shell.errors: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS shell.errors: invalid and missing target classification completed")
		}
	}
	if *singleInstance {
		if err := validateSingleInstance(ctx); err != nil {
			fmt.Printf("FAIL single_instance.forward: %v\n", err)
			failures++
		} else {
			fmt.Println("PASS single_instance.forward: secondary launch payload delivered")
		}
	}
	if *dragDrop {
		if reportDragAndDropProbe("drag_drop.probe", system.ProbeDragAndDrop(ctx), *failUnavailable && runtime.GOOS == "windows") {
			failures++
		}
	}
	if *events > 0 {
		if err := validateSystemEvents(*events); err != nil {
			fmt.Printf("FAIL system_events.listen: %v\n", err)
			failures++
		}
	}

	if failures > 0 {
		os.Exit(1)
	}
}

func reportAvailability(name string, availability system.CapabilityAvailability, required bool) bool {
	detail := availability.Status
	if availability.Err != nil {
		detail = system.CapabilityStatus(fmt.Sprintf("%s (%v)", detail, availability.Err))
	}
	if required && !availability.Available() {
		fmt.Printf("FAIL %s: %s\n", name, detail)
		return true
	}
	fmt.Printf("%s %s: %s\n", passLabel(availability.Available()), name, detail)
	return false
}

func reportBackendProbe(name string, probe system.NotificationBackendProbe, required bool) bool {
	detail := fmt.Sprintf(
		"%s actions=%v click=%v dismiss=%v actionCallback=%v protocol=%v durable=%v",
		probe.Status,
		probe.SupportsActionButtons,
		probe.SupportsClickCallback,
		probe.SupportsDismissCallback,
		probe.SupportsActionCallback,
		probe.SupportsProtocolActivation,
		probe.SupportsDurableActivation,
	)
	if probe.Err != nil {
		detail += " (" + probe.Err.Error() + ")"
	}
	if required && !probe.Available() {
		fmt.Printf("FAIL %s: %s\n", name, detail)
		return true
	}
	fmt.Printf("%s %s: %s\n", passLabel(probe.Available()), name, detail)
	return false
}

func reportDragAndDropProbe(name string, probe system.DragAndDropProbe, required bool) bool {
	detail := fmt.Sprintf(
		"%s drop=%v source=%v text=%v files=%v custom=%v external_in=%v external_out=%v operations=%s",
		probe.Status,
		probe.SupportsDropTarget,
		probe.SupportsDragSource,
		probe.SupportsText,
		probe.SupportsFiles,
		probe.SupportsCustomMIME,
		probe.SupportsExternalDragIn,
		probe.SupportsExternalDragOut,
		formatDragDropOperations(probe.SupportedOperations),
	)
	if probe.Err != nil {
		detail += " (" + probe.Err.Error() + ")"
	}
	if required && !probe.Available() {
		fmt.Printf("FAIL %s: %s\n", name, detail)
		return true
	}
	fmt.Printf("%s %s: %s\n", passLabel(probe.Available()), name, detail)
	return false
}

func validateDialogCancellation() error {
	if err := validateMessageBoxContextCancel(); err != nil {
		return fmt.Errorf("message_box: %w", err)
	}
	if err := validateFileDialogContextCancel(); err != nil {
		return fmt.Errorf("file_dialog: %w", err)
	}
	return nil
}

func validateMessageBoxContextCancel() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	result, err := system.ShowMessageBox(ctx,
		system.MessageBoxTitle("FluxUI validation"),
		system.MessageBoxText("This message box should close automatically when validation cancels its context."),
		system.MessageBoxStyle(system.MessageBoxInfo),
		system.MessageBoxButtonSet(system.MessageBoxOKCancel),
	)
	if err == nil {
		return fmt.Errorf("expected context cancellation, got result=%s", result)
	}
	if !isContextCancellation(err) {
		return err
	}
	return nil
}

func validateFileDialogContextCancel() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	result, err := system.OpenFileDialog(ctx,
		system.FileDialogTitle("FluxUI validation"),
		system.FileDialogFilters(system.FileFilter{
			Name:     "Text files",
			Patterns: []string{"*.txt"},
		}),
	)
	if err == nil {
		return fmt.Errorf("expected context cancellation, got cancelled=%v paths=%v", result.Cancelled, result.Paths)
	}
	if !isContextCancellation(err) {
		return err
	}
	return nil
}

func isContextCancellation(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func passLabel(ok bool) string {
	if ok {
		return "PASS"
	}
	return "INFO"
}

func validateNotificationGroup(ctx context.Context) error {
	group := "fluxui-validation"
	if err := system.Notify(ctx,
		system.NotificationTitle("FluxUI validation"),
		system.NotificationBody("First notification in the validation group."),
		system.NotificationGroup(group),
		system.NotificationKindStyle(system.NotificationInfo),
	); err != nil {
		return err
	}
	time.Sleep(400 * time.Millisecond)
	if err := system.Notify(ctx,
		system.NotificationTitle("FluxUI validation"),
		system.NotificationBody("Replacement notification in the same group."),
		system.NotificationGroup(group),
		system.NotificationKindStyle(system.NotificationSuccess),
	); err != nil {
		return err
	}
	time.Sleep(400 * time.Millisecond)
	return system.CancelNotificationGroup(ctx, group)
}

func validateToastNotification(ctx context.Context, appID string) error {
	group := "fluxui-validation-toast"
	if err := system.Notify(ctx,
		system.NotificationBackendPath(system.NotificationBackendToast),
		system.NotificationAppID(appID),
		system.NotificationTitle("FluxUI validation"),
		system.NotificationBody("Toast notification validation with an action button."),
		system.NotificationGroup(group),
		system.NotificationKindStyle(system.NotificationInfo),
		system.NotificationActions(system.NotificationAction{
			ID:    "ack",
			Label: "OK",
		}),
	); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	return system.CancelNotificationGroup(ctx, group)
}

func validateToastShortcut(ctx context.Context, appID, shortcutName, activatorCLSID string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve validation executable: %w", err)
	}
	options := []system.ToastShortcutOption{
		system.ToastShortcutArguments("--fluxui-toast-validation"),
		system.ToastShortcutIcon(executable),
	}
	if strings.TrimSpace(activatorCLSID) != "" {
		options = append(options, system.ToastShortcutActivatorCLSID(activatorCLSID))
	}
	if err := system.RegisterToastShortcut(ctx,
		appID,
		shortcutName,
		executable,
		options...,
	); err != nil {
		return err
	}
	defer system.UnregisterToastShortcut(context.Background(), shortcutName)

	if err := system.UnregisterToastShortcut(ctx, shortcutName); err != nil {
		return err
	}
	return nil
}

func validateToastActivatorRegistration(ctx context.Context, clsid string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve validation executable: %w", err)
	}
	command := quoteValidationCommandArgument(executable) + " --fluxui-toast-activator"
	if err := system.RegisterToastActivator(ctx, clsid, command); err != nil {
		return err
	}
	defer system.UnregisterToastActivator(context.Background(), clsid)

	if err := system.UnregisterToastActivator(ctx, clsid); err != nil {
		return err
	}
	return nil
}

func quoteValidationCommandArgument(value string) string {
	if value == "" {
		return `""`
	}
	if !strings.ContainsAny(value, " \t\n\v\"") {
		return value
	}
	var b strings.Builder
	b.WriteByte('"')
	backslashes := 0
	for _, r := range value {
		if r == '\\' {
			backslashes++
			continue
		}
		if r == '"' {
			b.WriteString(strings.Repeat(`\`, backslashes*2+1))
			b.WriteRune(r)
			backslashes = 0
			continue
		}
		if backslashes > 0 {
			b.WriteString(strings.Repeat(`\`, backslashes))
			backslashes = 0
		}
		b.WriteRune(r)
	}
	if backslashes > 0 {
		b.WriteString(strings.Repeat(`\`, backslashes*2))
	}
	b.WriteByte('"')
	return b.String()
}

func validateTray(resourceID int) error {
	checked := false
	tray, err := system.NewTray(
		system.TrayTooltip("FluxUI validation"),
		system.TrayIconBytes(sampleTrayIconBytes()),
		system.TrayMenuProvider(func() system.TrayMenu {
			return system.TrayMenu{
				system.TrayMenuItem{ID: "show", Label: "Show", Default: true},
				system.TrayMenuItem{ID: "checked", Label: "Checked", Checked: checked},
				system.TrayMenuItem{ID: "disabled", Label: "Disabled", Disabled: !checked},
				system.TrayMenuItem{
					ID:    "tools",
					Label: "Tools",
					Children: system.TrayMenu{
						system.TrayMenuItem{ID: "nested", Label: "Nested item"},
					},
				},
			}
		}),
	)
	if err != nil {
		return err
	}
	defer tray.Close()

	if err := tray.Show(); err != nil {
		return err
	}
	if !tray.Visible() || tray.Closed() {
		return fmt.Errorf("unexpected state after Show: visible=%v closed=%v", tray.Visible(), tray.Closed())
	}
	checked = true
	if err := tray.SetMenuProvider(func() system.TrayMenu {
		return system.TrayMenu{
			system.TrayMenuItem{ID: "checked", Label: "Checked", Checked: checked, Default: true},
			system.TrayMenuItem{ID: "disabled", Label: "Disabled", Disabled: !checked},
		}
	}); err != nil {
		return err
	}
	if err := tray.SetIconBytes(sampleTrayIconBytes()); err != nil {
		return err
	}
	iconPath, cleanupIconPath, err := writeTempTrayIcon()
	if err != nil {
		return err
	}
	defer cleanupIconPath()
	if err := tray.SetIcon(iconPath); err != nil {
		return err
	}
	if resourceID > 0 {
		if err := tray.SetIconResource(uint16(resourceID)); err != nil {
			return err
		}
	}
	if err := tray.Hide(); err != nil {
		return err
	}
	if tray.Visible() || tray.Closed() {
		return fmt.Errorf("unexpected state after Hide: visible=%v closed=%v", tray.Visible(), tray.Closed())
	}
	if err := tray.Close(); err != nil {
		return err
	}
	if !tray.Closed() {
		return fmt.Errorf("expected tray to be closed")
	}
	return nil
}

func validateClipboard(ctx context.Context) error {
	previous, err := system.ReadClipboardText(ctx)
	if err != nil {
		return fmt.Errorf("read existing text: %w", err)
	}
	token := fmt.Sprintf("FluxUI clipboard validation %d", time.Now().UnixNano())
	if err := system.WriteClipboardText(ctx, token); err != nil {
		return fmt.Errorf("write token: %w", err)
	}
	restored := false
	defer func() {
		if !restored {
			_ = system.WriteClipboardText(context.Background(), previous)
		}
	}()
	current, err := system.ReadClipboardText(ctx)
	if err != nil {
		return fmt.Errorf("read token: %w", err)
	}
	if current != token {
		return fmt.Errorf("clipboard mismatch: got %q", current)
	}
	if err := system.WriteClipboardText(ctx, previous); err != nil {
		return fmt.Errorf("restore previous text: %w", err)
	}
	restored = true
	return nil
}

func validateClipboardFiles(ctx context.Context) error {
	previous, _ := system.ReadClipboardText(ctx)
	restoreText := previous != ""
	if restoreText {
		defer system.WriteClipboardText(context.Background(), previous)
	}

	file, cleanup, err := writeValidationClipboardFile()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := system.WriteClipboardFiles(ctx, []string{file}); err != nil {
		return fmt.Errorf("write file list: %w", err)
	}
	files, err := system.ReadClipboardFiles(ctx)
	if err != nil {
		return fmt.Errorf("read file list: %w", err)
	}
	for _, got := range files {
		if samePath(got, file) {
			return nil
		}
	}
	return fmt.Errorf("clipboard file list missing validation file %q: %v", file, files)
}

func validateClipboardImage(ctx context.Context) error {
	previousImage, _ := system.ReadClipboardImagePNG(ctx)
	previousText, _ := system.ReadClipboardText(ctx)
	defer func() {
		if len(previousImage) > 0 {
			_ = system.WriteClipboardImagePNG(context.Background(), previousImage)
			return
		}
		if previousText != "" {
			_ = system.WriteClipboardText(context.Background(), previousText)
		}
	}()

	data, err := sampleClipboardPNG()
	if err != nil {
		return err
	}
	if err := system.WriteClipboardImagePNG(ctx, data); err != nil {
		return fmt.Errorf("write image: %w", err)
	}
	current, err := system.ReadClipboardImagePNG(ctx)
	if err != nil {
		return fmt.Errorf("read image: %w", err)
	}
	if len(current) == 0 {
		return fmt.Errorf("clipboard image is empty after write")
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(current))
	if err != nil {
		return fmt.Errorf("decode returned image: %w", err)
	}
	if cfg.Width != 2 || cfg.Height != 2 {
		return fmt.Errorf("unexpected image dimensions: %dx%d", cfg.Width, cfg.Height)
	}
	return nil
}

func validateShell(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	file := filepath.Join(cwd, "examples", "system_validation", "main.go")
	if _, err := os.Stat(file); err != nil {
		fallback := filepath.Join(cwd, "main.go")
		if _, fallbackErr := os.Stat(fallback); fallbackErr != nil {
			return fmt.Errorf("locate validation file %q: %w", file, err)
		}
		file = fallback
	}

	fileURL := "file:///" + strings.TrimPrefix(filepath.ToSlash(file), "/")
	if err := system.OpenURL(ctx, fileURL); err != nil {
		return fmt.Errorf("open file URL: %w", err)
	}
	if err := system.OpenPath(ctx, cwd); err != nil {
		return fmt.Errorf("open working directory: %w", err)
	}
	if err := system.RevealPath(ctx, file); err != nil {
		return fmt.Errorf("reveal validation file: %w", err)
	}
	return nil
}

func validateShellErrors(ctx context.Context) error {
	if err := system.OpenURL(ctx, ""); !system.IsInvalidTarget(err) {
		return fmt.Errorf("expected empty URL to be invalid target, got %v", err)
	}
	missing := filepath.Join(os.TempDir(), fmt.Sprintf("fluxui-missing-%d.txt", time.Now().UnixNano()))
	if err := system.OpenPath(ctx, missing); !system.IsUnavailable(err) || !system.IsTargetNotFound(err) {
		return fmt.Errorf("expected missing path to be unavailable target-not-found, got %v", err)
	}
	return nil
}

func validateSingleInstance(ctx context.Context) error {
	id := fmt.Sprintf("com.fluxui.validation.%d", time.Now().UnixNano())
	instance, err := system.AcquireSingleInstance(ctx, system.SingleInstanceID(id))
	if err != nil {
		return fmt.Errorf("acquire primary: %w", err)
	}
	defer instance.Close()

	payload := "fluxui://validation/open"
	_, err = system.AcquireSingleInstance(ctx,
		system.SingleInstanceID(id),
		system.SingleInstanceArgs("--validation", "secondary"),
		system.SingleInstancePayload(payload),
	)
	if !system.IsAlreadyRunning(err) {
		return fmt.Errorf("expected ErrAlreadyRunning from secondary launch, got %v", err)
	}

	select {
	case event := <-instance.Events():
		if event.ID != id || event.Payload != payload {
			return fmt.Errorf("unexpected secondary event: %#v", event)
		}
		if len(event.Args) != 2 || event.Args[0] != "--validation" || event.Args[1] != "secondary" {
			return fmt.Errorf("unexpected secondary args: %#v", event.Args)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timeout waiting for secondary launch event")
	}
}

func writeTempTrayIcon() (string, func(), error) {
	file, err := os.CreateTemp("", "fluxui-validation-*.ico")
	if err != nil {
		return "", func() {}, fmt.Errorf("create temp tray icon: %w", err)
	}
	path := file.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}
	if _, err := file.Write(sampleTrayIconBytes()); err != nil {
		_ = file.Close()
		cleanup()
		return "", func() {}, fmt.Errorf("write temp tray icon: %w", err)
	}
	if err := file.Close(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("close temp tray icon: %w", err)
	}
	return path, cleanup, nil
}

func writeValidationClipboardFile() (string, func(), error) {
	file, err := os.CreateTemp("", "fluxui-clipboard-validation-*.txt")
	if err != nil {
		return "", func() {}, fmt.Errorf("create clipboard validation file: %w", err)
	}
	path := file.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}
	if _, err := file.WriteString("FluxUI clipboard file validation\n"); err != nil {
		_ = file.Close()
		cleanup()
		return "", func() {}, fmt.Errorf("write clipboard validation file: %w", err)
	}
	if err := file.Close(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("close clipboard validation file: %w", err)
	}
	return path, cleanup, nil
}

func samePath(a, b string) bool {
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA == nil {
		a = absA
	}
	if errB == nil {
		b = absB
	}
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func validateSystemEvents(duration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	sub, err := system.SubscribeSystemEvents(ctx)
	if err != nil {
		return err
	}
	defer sub.Close()

	count := 0
	kinds := map[system.SystemEventKind]int{}
	for {
		select {
		case event, ok := <-sub.Events():
			if !ok {
				fmt.Printf("PASS system_events.listen: received %d events (%s)\n", count, formatSystemEventKinds(kinds))
				return nil
			}
			count++
			kinds[event.Kind]++
		case <-ctx.Done():
			fmt.Printf("PASS system_events.listen: received %d events (%s)\n", count, formatSystemEventKinds(kinds))
			return nil
		}
	}
}

func formatSystemEventKinds(kinds map[system.SystemEventKind]int) string {
	if len(kinds) == 0 {
		return "no events observed"
	}
	parts := make([]string, 0, len(kinds))
	for kind, count := range kinds {
		parts = append(parts, fmt.Sprintf("%s=%d", kind, count))
	}
	return strings.Join(parts, ", ")
}

func formatDragDropOperations(operations []system.DragDropOperation) string {
	if len(operations) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(operations))
	for _, operation := range operations {
		parts = append(parts, string(operation))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func sampleTrayIconBytes() []byte {
	const (
		width      = 16
		height     = 16
		headerSize = 40
		pixelBytes = width * height * 4
		maskBytes  = height * 4
		imageBytes = headerSize + pixelBytes + maskBytes
		imageOff   = 6 + 16
	)

	data := make([]byte, imageOff+imageBytes)
	binary.LittleEndian.PutUint16(data[2:], 1)
	binary.LittleEndian.PutUint16(data[4:], 1)
	data[6] = width
	data[7] = height
	binary.LittleEndian.PutUint16(data[10:], 1)
	binary.LittleEndian.PutUint16(data[12:], 32)
	binary.LittleEndian.PutUint32(data[14:], imageBytes)
	binary.LittleEndian.PutUint32(data[18:], imageOff)

	image := data[imageOff:]
	binary.LittleEndian.PutUint32(image[0:], headerSize)
	binary.LittleEndian.PutUint32(image[4:], width)
	binary.LittleEndian.PutUint32(image[8:], height*2)
	binary.LittleEndian.PutUint16(image[12:], 1)
	binary.LittleEndian.PutUint16(image[14:], 32)
	binary.LittleEndian.PutUint32(image[20:], pixelBytes)
	pixels := image[headerSize : headerSize+pixelBytes]
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 4
			pixels[i+0] = byte(24 + y*7)
			pixels[i+1] = byte(110 + x*6)
			pixels[i+2] = 220
			pixels[i+3] = 255
		}
	}
	return data
}

func sampleClipboardPNG() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 0x20, G: 0x80, B: 0xc0, A: 0xff})
	img.Set(1, 0, color.RGBA{R: 0x20, G: 0xc0, B: 0x80, A: 0xff})
	img.Set(0, 1, color.RGBA{R: 0xc0, G: 0x80, B: 0x20, A: 0xff})
	img.Set(1, 1, color.RGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode validation PNG: %w", err)
	}
	return buf.Bytes(), nil
}
