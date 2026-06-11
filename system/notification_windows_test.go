//go:build windows

package system

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestWindowsNotificationData(t *testing.T) {
	data, err := newWindowsNotificationData(123, 7, 456, notificationOptions{
		title:   "Done",
		body:    "Export finished.",
		kind:    NotificationSuccess,
		group:   "exports",
		timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.cbSize == 0 {
		t.Fatal("expected notify icon data size to be set")
	}
	if data.hWnd != 123 {
		t.Fatalf("expected hwnd 123, got %d", data.hWnd)
	}
	if data.uID != 7 {
		t.Fatalf("expected icon id 7, got %d", data.uID)
	}
	if data.hIcon != 456 {
		t.Fatalf("expected icon 456, got %d", data.hIcon)
	}
	if data.uFlags != nifMessage|nifIcon|nifTip|nifInfo {
		t.Fatalf("unexpected flags 0x%x", data.uFlags)
	}
	if data.uCallbackMessage != wmFluxUINotification {
		t.Fatalf("unexpected callback message 0x%x", data.uCallbackMessage)
	}
	if data.uTimeout != 5000 {
		t.Fatalf("expected timeout 5000ms, got %d", data.uTimeout)
	}
	if data.dwInfoFlags != niifInfo {
		t.Fatalf("expected success to map to info icon, got 0x%x", data.dwInfoFlags)
	}
	if fixedUTF16String(data.szInfoTitle[:]) != "Done" {
		t.Fatalf("unexpected title %q", fixedUTF16String(data.szInfoTitle[:]))
	}
	if fixedUTF16String(data.szInfo[:]) != "Export finished." {
		t.Fatalf("unexpected body %q", fixedUTF16String(data.szInfo[:]))
	}
}

func TestWindowsNotificationInfoFlags(t *testing.T) {
	tests := []struct {
		kind NotificationKind
		want uint32
	}{
		{kind: NotificationInfo, want: niifInfo},
		{kind: NotificationSuccess, want: niifInfo},
		{kind: NotificationWarning, want: niifWarning},
		{kind: NotificationError, want: niifError},
		{kind: NotificationKind("custom"), want: niifInfo},
	}

	for _, tt := range tests {
		if got := windowsNotificationInfoFlags(tt.kind); got != tt.want {
			t.Fatalf("expected kind %q to map to 0x%x, got 0x%x", tt.kind, tt.want, got)
		}
	}
}

func TestWindowsNotificationTimeoutMilliseconds(t *testing.T) {
	tests := []struct {
		timeout time.Duration
		want    uint32
	}{
		{timeout: 0, want: 10000},
		{timeout: 100 * time.Millisecond, want: 1000},
		{timeout: 5 * time.Second, want: 5000},
		{timeout: 2 * time.Minute, want: 60000},
	}

	for _, tt := range tests {
		if got := windowsNotificationTimeoutMilliseconds(tt.timeout); got != tt.want {
			t.Fatalf("expected timeout %s to map to %d, got %d", tt.timeout, tt.want, got)
		}
	}
}

func TestWindowsNotificationEvent(t *testing.T) {
	tests := []struct {
		name        string
		lParam      uint32
		wantKind    NotificationEventKind
		wantCleanup bool
	}{
		{name: "click", lParam: ninBalloonUserClick, wantKind: NotificationEventClicked, wantCleanup: true},
		{name: "hide", lParam: ninBalloonHide, wantKind: NotificationEventDismissed, wantCleanup: true},
		{name: "timeout", lParam: ninBalloonTimeout, wantKind: NotificationEventDismissed, wantCleanup: true},
		{name: "unknown", lParam: 123, wantCleanup: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKind, gotCleanup := windowsNotificationEvent(tt.lParam)
			if gotKind != tt.wantKind {
				t.Fatalf("expected kind %q, got %q", tt.wantKind, gotKind)
			}
			if gotCleanup != tt.wantCleanup {
				t.Fatalf("expected cleanup %v, got %v", tt.wantCleanup, gotCleanup)
			}
		})
	}
}

func TestWindowsNotificationFallbackCleanupDelay(t *testing.T) {
	tests := []struct {
		timeout time.Duration
		want    time.Duration
	}{
		{timeout: 0, want: 2 * time.Minute},
		{timeout: 5 * time.Second, want: 2 * time.Minute},
		{timeout: 2 * time.Minute, want: 2 * time.Minute},
		{timeout: 3 * time.Minute, want: 3*time.Minute + 30*time.Second},
	}

	for _, tt := range tests {
		if got := windowsNotificationFallbackCleanupDelay(tt.timeout); got != tt.want {
			t.Fatalf("expected timeout %s to map to %s, got %s", tt.timeout, tt.want, got)
		}
	}
}

func TestWindowsNotificationClassNameIsProcessScoped(t *testing.T) {
	name := windowsNotificationClassName()
	if name == "" {
		t.Fatal("expected non-empty class name")
	}
	if name == "FluxUINotificationWindow" {
		t.Fatal("expected class name to be process scoped")
	}
}

func TestWindowsNotificationDefaultTitle(t *testing.T) {
	data, err := newWindowsNotificationData(1, 2, 3, notificationOptions{
		body: "Body only",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fixedUTF16String(data.szInfoTitle[:]) != "FluxUI" {
		t.Fatalf("expected default title FluxUI, got %q", fixedUTF16String(data.szInfoTitle[:]))
	}
}

func TestWindowsToastXML(t *testing.T) {
	xml, err := windowsToastXML(notificationOptions{
		title: "Export <done>",
		body:  "Saved & indexed",
		icon:  "C:\\tmp\\icon.ico",
		group: `reports"daily`,
		actions: []NotificationAction{
			{ID: "open", Label: "Open report"},
			{ID: "dismiss", Label: "Dismiss"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected toast xml error: %v", err)
	}
	if !containsAll(xml,
		`<toast launch="fluxui:click:reports&quot;daily">`,
		`Export &lt;done&gt;`,
		`Saved &amp; indexed`,
		`<image placement="appLogoOverride" src="C:\tmp\icon.ico"/>`,
		`arguments="fluxui:action:open" content="Open report"`,
		`arguments="fluxui:action:dismiss" content="Dismiss"`,
	) {
		t.Fatalf("unexpected toast xml: %s", xml)
	}
}

func TestWindowsNotificationUsesToast(t *testing.T) {
	if !windowsNotificationUsesToast(notificationOptions{backend: NotificationBackendToast}) {
		t.Fatal("explicit toast backend should use toast")
	}
	if !windowsNotificationUsesToast(notificationOptions{actions: []NotificationAction{{ID: "open", Label: "Open"}}}) {
		t.Fatal("actions should imply toast backend")
	}
	if windowsNotificationUsesToast(notificationOptions{backend: NotificationBackendBalloon}) {
		t.Fatal("explicit balloon backend should not use toast")
	}
	if err := validateWindowsNotificationBackend(NotificationBackend("custom")); err == nil {
		t.Fatal("expected unknown notification backend to fail")
	}
}

func TestWindowsToastXMLProtocolActivation(t *testing.T) {
	xml, err := windowsToastXML(notificationOptions{
		title:     "Report ready",
		launchURI: "fluxui://notifications/reports",
		actions: []NotificationAction{
			{Label: "Open", URI: "fluxui://notifications/reports/open"},
			{ID: "dismiss", Label: "Dismiss"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected toast xml error: %v", err)
	}
	if !containsAll(xml,
		`<toast activationType="protocol" launch="fluxui://notifications/reports">`,
		`activationType="protocol" arguments="fluxui://notifications/reports/open" content="Open"`,
		`activationType="foreground" arguments="fluxui:action:dismiss" content="Dismiss"`,
	) {
		t.Fatalf("unexpected protocol toast xml: %s", xml)
	}
}

func TestWindowsToastValidationAllowsProtocolActions(t *testing.T) {
	err := validateWindowsToastOptions(notificationOptions{
		launchURI: "fluxui://notifications/reports",
		actions: []NotificationAction{
			{Label: "Open", URI: "fluxui://notifications/reports/open"},
		},
	})
	if err != nil {
		t.Fatalf("expected protocol action to be valid, got %v", err)
	}
	if err := validateWindowsToastOptions(notificationOptions{launchURI: "not a uri"}); err == nil {
		t.Fatal("expected invalid launch URI to fail")
	}
	if err := validateWindowsToastOptions(notificationOptions{
		actions: []NotificationAction{{Label: "Open", URI: "not a uri"}},
	}); err == nil {
		t.Fatal("expected invalid action URI to fail")
	}
}

func TestWindowsToastCallbacksSkipProtocolActions(t *testing.T) {
	callbacks := windowsToastCallbacksFromOptions(notificationOptions{
		actions: []NotificationAction{
			{ID: "open", Label: "Open", URI: "fluxui://notifications/open"},
			{ID: "dismiss", Label: "Dismiss"},
		},
	})
	if callbacks.actionIDs["open"] {
		t.Fatal("protocol actions should not be registered as foreground callbacks")
	}
	if !callbacks.actionIDs["dismiss"] {
		t.Fatal("foreground actions should be registered as callbacks")
	}
}

func TestWindowsToastCleanupMissingGroupIsNoop(t *testing.T) {
	state := newWindowsNotificationState()
	if err := state.cleanupToastGroup("missing"); err != nil {
		t.Fatalf("expected missing toast group cleanup to be no-op, got %v", err)
	}
	if err := state.replaceGroup("missing"); err != nil {
		t.Fatalf("expected missing group replacement to be no-op, got %v", err)
	}
	if err := state.replaceGroup(""); err != nil {
		t.Fatalf("expected empty group replacement to be no-op, got %v", err)
	}
}

func TestWindowsToastGroupTracksListenerStop(t *testing.T) {
	state := newWindowsNotificationState()
	stopped := false

	state.rememberToastGroup("reports", "FluxUI.App", func() {
		stopped = true
	})
	entry, ok := state.takeToastGroup("reports")
	if !ok {
		t.Fatal("expected toast group entry")
	}
	if entry.appID != "FluxUI.App" {
		t.Fatalf("unexpected app ID: %q", entry.appID)
	}
	if _, ok := state.takeToastGroup("reports"); ok {
		t.Fatal("expected toast group entry to be removed after take")
	}

	stopWindowsToastEntry(entry)
	if !stopped {
		t.Fatal("expected listener stop callback to run")
	}
}

func TestWindowsToastPowerShellScript(t *testing.T) {
	script := windowsToastPowerShellScript("FluxUI.App", "reports", "<toast/>", 0)
	if !containsAll(script,
		"ToastNotificationManager",
		"FromBase64String",
		"$toast.Group='reports'",
		"CreateToastNotifier('FluxUI.App')",
	) {
		t.Fatalf("unexpected toast PowerShell script: %s", script)
	}
	if powershellEncodedCommand(script) == "" {
		t.Fatal("expected encoded PowerShell command")
	}
}

func TestWindowsToastCancelPowerShellScript(t *testing.T) {
	script := windowsToastCancelPowerShellScript("FluxUI.App", "reports")
	if !containsAll(script,
		"ToastNotificationManager",
		"History.RemoveGroup('reports','FluxUI.App')",
	) {
		t.Fatalf("unexpected toast cancel PowerShell script: %s", script)
	}
}

func TestWindowsToastProbePowerShellScript(t *testing.T) {
	script := windowsToastProbePowerShellScript("FluxUI.App")
	if !containsAll(script,
		"ToastNotificationManager",
		"CreateToastNotifier('FluxUI.App')",
		"CreateToastNotifier returned null",
	) {
		t.Fatalf("unexpected toast probe PowerShell script: %s", script)
	}
}

func TestWindowsNotificationBackendProbeFlags(t *testing.T) {
	driver := windowsDriver{}

	balloon := driver.probeNotificationBackend(context.Background(), NotificationBackendBalloon, notificationOptions{})
	if balloon.Backend != NotificationBackendBalloon {
		t.Fatalf("expected balloon backend, got %#v", balloon)
	}
	if balloon.Status == CapabilityStatusAvailable {
		if !balloon.SupportsClickCallback || !balloon.SupportsDismissCallback {
			t.Fatalf("expected balloon callbacks to be marked when available: %#v", balloon)
		}
		if balloon.SupportsActionButtons || balloon.SupportsActionCallback || balloon.SupportsProtocolActivation || balloon.SupportsDurableActivation {
			t.Fatalf("unexpected balloon advanced flags: %#v", balloon)
		}
	}

	toast := driver.probeNotificationBackend(context.Background(), NotificationBackendToast, notificationOptions{})
	if toast.Backend != NotificationBackendToast {
		t.Fatalf("expected toast backend, got %#v", toast)
	}
	if toast.Status == CapabilityStatusAvailable {
		if !toast.SupportsActionButtons || !toast.SupportsClickCallback || !toast.SupportsDismissCallback || !toast.SupportsActionCallback || !toast.SupportsProtocolActivation || !toast.SupportsDurableActivation {
			t.Fatalf("expected toast real-time/protocol flags: %#v", toast)
		}
	}
}

func TestWindowsToastPowerShellListenerScript(t *testing.T) {
	script := windowsToastPowerShellScript("FluxUI.App", "reports", "<toast/>", time.Minute)
	if !containsAll(script,
		"Register-ObjectEvent",
		"Activated",
		"Dismissed",
		windowsToastReadyLine,
		windowsToastEventPrefix+"activated|",
		windowsToastEventPrefix+"dismissed|",
		"Wait-Event",
	) {
		t.Fatalf("unexpected toast listener PowerShell script: %s", script)
	}
}

func TestWindowsToastEventListenDuration(t *testing.T) {
	if got := windowsToastEventListenDuration(notificationOptions{}); got != 0 {
		t.Fatalf("expected no callbacks to skip listener, got %s", got)
	}
	if got := windowsToastEventListenDuration(notificationOptions{onClick: func(NotificationEvent) {}}); got != 40*time.Second {
		t.Fatalf("expected default listener duration 40s, got %s", got)
	}
	if got := windowsToastEventListenDuration(notificationOptions{
		timeout: 10 * time.Minute,
		onClick: func(NotificationEvent) {},
	}); got != 5*time.Minute {
		t.Fatalf("expected listener duration cap, got %s", got)
	}
}

func TestParseWindowsToastEventLine(t *testing.T) {
	line := windowsToastEventPrefix + "activated|" + base64.StdEncoding.EncodeToString([]byte(windowsToastActionPrefix+"open"))
	kind, arg, ok := parseWindowsToastEventLine(line)
	if !ok || kind != "activated" || arg != windowsToastActionPrefix+"open" {
		t.Fatalf("unexpected parsed event: kind=%q arg=%q ok=%v", kind, arg, ok)
	}
	if _, _, ok := parseWindowsToastEventLine("other"); ok {
		t.Fatal("expected unrelated line to be ignored")
	}
	if _, _, ok := parseWindowsToastEventLine(windowsToastEventPrefix + "activated|not-base64"); ok {
		t.Fatal("expected invalid base64 event to be ignored")
	}
}

func TestDispatchWindowsToastEventLine(t *testing.T) {
	actionCh := make(chan NotificationEvent, 1)
	clickCh := make(chan NotificationEvent, 1)
	dismissCh := make(chan NotificationEvent, 1)
	callbacks := windowsToastCallbacks{
		group:     "reports",
		actionIDs: map[string]bool{"open": true},
		onAction:  func(event NotificationEvent) { actionCh <- event },
		onClick:   func(event NotificationEvent) { clickCh <- event },
		onDismiss: func(event NotificationEvent) { dismissCh <- event },
	}

	dispatchWindowsToastEventLine(windowsToastEventLine("activated", windowsToastActionPrefix+"open"), callbacks)
	select {
	case event := <-actionCh:
		if event.Kind != NotificationEventAction || event.Group != "reports" || event.Action != "open" {
			t.Fatalf("unexpected action event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected action callback")
	}

	dispatchWindowsToastEventLine(windowsToastEventLine("activated", windowsToastClickPrefix+"reports"), callbacks)
	select {
	case event := <-clickCh:
		if event.Kind != NotificationEventClicked || event.Group != "reports" {
			t.Fatalf("unexpected click event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected click callback")
	}

	dispatchWindowsToastEventLine(windowsToastEventLine("dismissed", "user_canceled"), callbacks)
	select {
	case event := <-dismissCh:
		if event.Kind != NotificationEventDismissed || event.Group != "reports" {
			t.Fatalf("unexpected dismiss event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected dismiss callback")
	}
}

func windowsToastEventLine(kind, arg string) string {
	return windowsToastEventPrefix + kind + "|" + base64.StdEncoding.EncodeToString([]byte(arg))
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}

func fixedUTF16String(value []uint16) string {
	for i, c := range value {
		if c == 0 {
			return stringFromUTF16(value[:i])
		}
	}
	return stringFromUTF16(value)
}

func stringFromUTF16(value []uint16) string {
	runes := make([]rune, 0, len(value))
	for _, c := range value {
		runes = append(runes, rune(c))
	}
	return string(runes)
}
