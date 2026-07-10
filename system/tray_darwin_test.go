//go:build darwin

package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDarwinTrayScriptBuildsStatusItemMenu(t *testing.T) {
	script := darwinTrayScript(trayOptions{
		tooltip: "FluxUI tray",
		icon:    "/tmp/fluxui.png",
		menu: TrayMenu{
			TrayMenuItem{ID: "open", Label: "Open"},
			TrayMenuSeparator(),
			TrayMenuItem{ID: "sync", Label: "Sync", Checked: true},
			TrayMenuItem{ID: "disabled", Label: "Disabled", Disabled: true},
			TrayMenuItem{
				ID:    "tools",
				Label: "Tools",
				Children: TrayMenu{
					TrayMenuItem{ID: "settings", Label: "Settings"},
				},
			},
		},
	}, "/tmp/fluxui-events.log", "/tmp/fluxui.png")

	for _, want := range []string{
		`NSStatusBar`,
		`NSApplication's sharedApplication()`,
		`NSApplicationActivationPolicyAccessory`,
		`property eventPath : "/tmp/fluxui-events.log"`,
		`buttonRef's setToolTip:"FluxUI tray"`,
		`initWithContentsOfFile:"/tmp/fluxui.png"`,
		`menuRef's setAutoenablesItems:false`,
		`initWithTitle:"Open" action:"menuClicked:"`,
		`setRepresentedObject:"open"`,
		`separatorItem()`,
		`setState:1`,
		`setEnabled:false`,
		`setSubmenu:`,
		`initWithTitle:"Settings" action:"menuClicked:"`,
		`my writeEvent("menu:" & itemID)`,
		`quoted form of (theEvent & linefeed)`,
		`appRef's run()`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("expected script to contain %q, got:\n%s", want, script)
		}
	}
}

func TestDarwinTrayScriptEnablesClickWithoutMenu(t *testing.T) {
	script := darwinTrayScript(trayOptions{
		onClick: func(TrayEvent) {},
	}, "/tmp/fluxui-events.log", "")

	if !strings.Contains(script, `buttonRef's setAction:"statusClicked:"`) {
		t.Fatalf("expected status click action, got:\n%s", script)
	}
	if strings.Contains(script, `set menuRef`) {
		t.Fatalf("unexpected menu for click-only tray script:\n%s", script)
	}
}

func TestDarwinTrayTextSanitizesNewlines(t *testing.T) {
	if got := darwinTrayText("hello\nworld"); got != "hello world" {
		t.Fatalf("unexpected text sanitization: %q", got)
	}
	if got := darwinTrayText(""); got != "FluxUI" {
		t.Fatalf("unexpected empty text fallback: %q", got)
	}
}

func TestDarwinTrayResourceIconUnsupported(t *testing.T) {
	handle := &darwinTrayHandle{}
	if err := handle.setIconResource(1); !IsUnsupported(err) {
		t.Fatalf("expected resource icon unsupported, got %v", err)
	}
}

func TestDarwinTrayScriptDoesNotEvaluateMenuProvider(t *testing.T) {
	called := false
	darwinTrayScript(trayOptions{
		menuProvider: func() TrayMenu {
			called = true
			return TrayMenu{{ID: "unexpected"}}
		},
	}, "/tmp/fluxui-events.log", "")
	if called {
		t.Fatal("script generation must only use an already-resolved menu")
	}
}

func TestDarwinTrayMenuProviderCanReenterHandle(t *testing.T) {
	clicked := make(chan struct{}, 1)
	handle := &darwinTrayHandle{}
	handle.opts.menuProvider = func() TrayMenu {
		if err := handle.setTooltip("from provider"); err != nil {
			t.Errorf("unexpected reentrant SetTooltip error: %v", err)
		}
		return TrayMenu{{
			ID: "open",
			OnClick: func(TrayEvent) {
				clicked <- struct{}{}
			},
		}}
	}

	done := make(chan struct{})
	go func() {
		handle.dispatchMenuItem("open")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("menu provider deadlocked while reentering the tray handle")
	}
	select {
	case <-clicked:
	default:
		t.Fatal("expected provider menu item callback")
	}
	if handle.opts.tooltip != "from provider" {
		t.Fatalf("unexpected tooltip after provider reentry: %q", handle.opts.tooltip)
	}
}

func TestDarwinVisibleTrayUnstableProviderFallsBackWithoutRecursion(t *testing.T) {
	handle := &darwinTrayHandle{
		activeMenu: TrayMenu{{ID: "committed", Label: "Committed"}},
		visible:    true,
	}
	var provider func() TrayMenu
	provider = func() TrayMenu {
		if err := handle.setMenuProvider(provider); err != nil {
			t.Errorf("unexpected reentrant SetMenuProvider error: %v", err)
		}
		return TrayMenu{{ID: "open", Label: "Open"}}
	}
	handle.opts.menuProvider = provider

	done := make(chan TrayMenu, 1)
	go func() {
		done <- handle.currentMenu()
	}()
	select {
	case menu := <-done:
		if len(menu) != 1 || menu[0].ID != "committed" {
			t.Fatalf("unstable provider menu = %#v, want committed fallback", menu)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("visible provider recursively refreshed itself")
	}
}

func TestDarwinTrayProviderReplacementUsesLatestMenu(t *testing.T) {
	handle := &darwinTrayHandle{visible: true}
	latest := func() TrayMenu {
		return TrayMenu{{ID: "new", Label: "New"}}
	}
	middle := func() TrayMenu {
		if err := handle.setMenuProvider(latest); err != nil {
			t.Errorf("unexpected second provider replacement error: %v", err)
		}
		return TrayMenu{{ID: "middle", Label: "Middle"}}
	}
	handle.opts.menuProvider = func() TrayMenu {
		if err := handle.setMenuProvider(middle); err != nil {
			t.Errorf("unexpected provider replacement error: %v", err)
		}
		return TrayMenu{{ID: "old", Label: "Old"}}
	}

	menu := handle.currentMenu()
	if len(menu) != 1 || menu[0].ID != "new" {
		t.Fatalf("resolved menu = %#v, want latest provider menu", menu)
	}
}

func TestDarwinTrayConcurrentRefreshQueuesWithoutApplyingFallback(t *testing.T) {
	commandDir := t.TempDir()
	commandPath := filepath.Join(commandDir, darwinTrayProviderCommand)
	if err := os.WriteFile(commandPath, []byte("#!/bin/sh\nexec /bin/sleep 30\n"), 0o755); err != nil {
		t.Fatalf("write fake tray command: %v", err)
	}
	t.Setenv("PATH", commandDir)
	tempDir := t.TempDir()
	fallback := TrayMenu{{ID: "old", Label: "Old"}}
	latest := TrayMenu{{ID: "new", Label: "New"}}
	providerStarted := make(chan struct{})
	releaseProvider := make(chan struct{})
	providerCalls := 0
	handle := &darwinTrayHandle{
		activeMenu: cloneTrayMenu(fallback),
		tempDir:    tempDir,
		eventPath:  tempDir + "/events",
		visible:    true,
	}
	handle.opts.menuProvider = func() TrayMenu {
		providerCalls++
		if providerCalls == 1 {
			close(providerStarted)
			<-releaseProvider
		}
		return latest
	}
	defer func() {
		handle.mu.Lock()
		handle.stopProcessLocked()
		handle.mu.Unlock()
	}()

	ownerDone := make(chan error, 1)
	go func() {
		ownerDone <- handle.refreshMenu()
	}()
	select {
	case <-providerStarted:
	case <-time.After(2 * time.Second):
		close(releaseProvider)
		t.Fatal("menu provider did not start")
	}

	refreshDone := make(chan error, 1)
	go func() {
		refreshDone <- handle.refreshMenu()
	}()
	var refreshErr error
	select {
	case refreshErr = <-refreshDone:
	case <-time.After(2 * time.Second):
		close(releaseProvider)
		t.Fatal("concurrent menu refresh did not queue behind the active resolution")
	}

	handle.mu.Lock()
	pending := handle.menuRefreshPending
	committedBeforeRelease := cloneTrayMenu(handle.activeMenu)
	handle.mu.Unlock()
	close(releaseProvider)

	var ownerErr error
	select {
	case ownerErr = <-ownerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("owner menu refresh did not finish")
	}
	if refreshErr != nil {
		t.Fatalf("concurrent refresh tried to apply the fallback menu: %v", refreshErr)
	}
	if ownerErr != nil {
		t.Fatalf("owner menu refresh failed: %v", ownerErr)
	}
	if !pending {
		t.Fatal("concurrent refresh did not record a pending refresh")
	}
	if len(committedBeforeRelease) != 1 || committedBeforeRelease[0].ID != "old" {
		t.Fatalf("active menu changed during provider resolution: %#v", committedBeforeRelease)
	}
	if providerCalls != 2 {
		t.Fatalf("provider calls = %d, want 2 after pending refresh", providerCalls)
	}

	handle.mu.Lock()
	committedAfterOwner := cloneTrayMenu(handle.activeMenu)
	handle.mu.Unlock()
	if len(committedAfterOwner) != 1 || committedAfterOwner[0].ID != "new" {
		t.Fatalf("active menu after owner commit = %#v, want latest menu", committedAfterOwner)
	}
}

func TestDarwinTrayDispatchUsesActiveMenuWhenProviderDoesNotStabilize(t *testing.T) {
	activeClicks := 0
	transientClicks := 0
	providerCalls := 0
	handle := &darwinTrayHandle{
		activeMenu: TrayMenu{{
			ID: "open",
			OnClick: func(TrayEvent) {
				activeClicks++
			},
		}},
	}
	var provider func() TrayMenu
	provider = func() TrayMenu {
		providerCalls++
		if err := handle.setMenuProvider(provider); err != nil {
			t.Errorf("unexpected provider replacement error: %v", err)
		}
		return TrayMenu{{
			ID: "open",
			OnClick: func(TrayEvent) {
				transientClicks++
			},
		}}
	}
	handle.opts.menuProvider = provider

	handle.dispatchMenuItem("open")

	if activeClicks != 1 {
		t.Fatalf("active menu callback calls = %d, want 1", activeClicks)
	}
	if transientClicks != 0 {
		t.Fatalf("unstable provider callback calls = %d, want 0", transientClicks)
	}
	if providerCalls != maxTrayMenuResolutionRetries+1 {
		t.Fatalf("provider calls = %d, want %d", providerCalls, maxTrayMenuResolutionRetries+1)
	}
}

func TestDarwinTrayScheduledRefreshPreservesFollowup(t *testing.T) {
	handle := &darwinTrayHandle{
		menuRefreshScheduled: true,
		visible:              true,
		cmd:                  &exec.Cmd{},
	}
	handle.mu.Lock()
	handle.scheduleMenuRefreshLocked()
	queued := handle.menuRefreshFollowup
	continued := handle.continueScheduledMenuRefreshLocked()
	stillScheduled := handle.menuRefreshScheduled
	followupConsumed := !handle.menuRefreshFollowup
	handle.cmd = nil
	finished := !handle.continueScheduledMenuRefreshLocked()
	cleared := !handle.menuRefreshScheduled && !handle.menuRefreshFollowup
	handle.mu.Unlock()

	if !queued {
		t.Fatal("active scheduled refresh did not retain the follow-up request")
	}
	if !continued || !stillScheduled || !followupConsumed {
		t.Fatalf("follow-up transition = continued:%v scheduled:%v consumed:%v", continued, stillScheduled, followupConsumed)
	}
	if !finished || !cleared {
		t.Fatalf("terminal transition = finished:%v cleared:%v", finished, cleared)
	}
}
