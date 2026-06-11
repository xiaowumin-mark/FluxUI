//go:build windows

package system

import (
	"os"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

func TestWindowsTaskDialogConfigLayout(t *testing.T) {
	ptrSize := unsafe.Sizeof(uintptr(0))
	var wantSize uintptr
	var wantParentOffset uintptr
	var wantExpandedInformationOffset uintptr
	var wantCallbackDataOffset uintptr
	switch ptrSize {
	case 8:
		wantSize = 160
		wantParentOffset = 4
		wantExpandedInformationOffset = 100
		wantCallbackDataOffset = 148
	case 4:
		wantSize = 96
		wantParentOffset = 4
		wantExpandedInformationOffset = 64
		wantCallbackDataOffset = 88
	default:
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	if got := unsafe.Sizeof(taskDialogConfig{}); got != wantSize {
		t.Fatalf("unexpected TASKDIALOGCONFIG size: got %d want %d", got, wantSize)
	}
	if got := unsafe.Offsetof(taskDialogConfig{}.parent); got != wantParentOffset {
		t.Fatalf("unexpected TASKDIALOGCONFIG parent offset: got %d want %d", got, wantParentOffset)
	}
	if got := unsafe.Offsetof(taskDialogConfig{}.expandedInformation); got != wantExpandedInformationOffset {
		t.Fatalf("unexpected TASKDIALOGCONFIG expanded info offset: got %d want %d", got, wantExpandedInformationOffset)
	}
	if got := unsafe.Offsetof(taskDialogConfig{}.callbackData); got != wantCallbackDataOffset {
		t.Fatalf("unexpected TASKDIALOGCONFIG callback data offset: got %d want %d", got, wantCallbackDataOffset)
	}
	if got := unsafe.Offsetof(taskDialogButton{}.text); got != 4 {
		t.Fatalf("unexpected TASKDIALOG_BUTTON text offset: got %d want 4", got)
	}
	if got, want := unsafe.Sizeof(taskDialogButton{}), uintptr(4)+ptrSize; got != want {
		t.Fatalf("unexpected TASKDIALOG_BUTTON size: got %d want %d", got, want)
	}
}

func TestWindowsTaskDialogConfig(t *testing.T) {
	tests := []struct {
		name          string
		opts          messageBoxOptions
		buttons       uint32
		defaultButton int32
		icon          uintptr
	}{
		{
			name:          "default ok info",
			opts:          defaultMessageBoxOptions(),
			buttons:       tdcbfOKButton,
			defaultButton: idOK,
			icon:          tdInformationIcon,
		},
		{
			name: "warning ok cancel default cancel",
			opts: messageBoxOptions{
				kind:          MessageBoxWarning,
				buttons:       MessageBoxOKCancel,
				defaultButton: MessageBoxResultCancel,
			},
			buttons:       tdcbfOKButton | tdcbfCancelButton,
			defaultButton: idCancel,
			icon:          tdWarningIcon,
		},
		{
			name: "question yes no default no",
			opts: messageBoxOptions{
				kind:          MessageBoxQuestion,
				buttons:       MessageBoxYesNo,
				defaultButton: MessageBoxResultNo,
			},
			buttons:       tdcbfYesButton | tdcbfNoButton,
			defaultButton: idNo,
			icon:          tdInformationIcon,
		},
		{
			name: "error yes no cancel default cancel",
			opts: messageBoxOptions{
				kind:          MessageBoxError,
				buttons:       MessageBoxYesNoCancel,
				defaultButton: MessageBoxResultCancel,
			},
			buttons:       tdcbfYesButton | tdcbfNoButton | tdcbfCancelButton,
			defaultButton: idCancel,
			icon:          tdErrorIcon,
		},
		{
			name: "retry cancel default retry",
			opts: messageBoxOptions{
				kind:          MessageBoxWarning,
				buttons:       MessageBoxRetryCancel,
				defaultButton: MessageBoxResultRetry,
			},
			buttons:       tdcbfRetryButton | tdcbfCancelButton,
			defaultButton: idRetry,
			icon:          tdWarningIcon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := windowsTaskDialogConfig(tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			config := data.config
			if config.size == 0 {
				t.Fatal("expected config size to be set")
			}
			if config.commonButtons != tt.buttons {
				t.Fatalf("expected buttons 0x%x, got 0x%x", tt.buttons, config.commonButtons)
			}
			if config.defaultButton != tt.defaultButton {
				t.Fatalf("expected default button %d, got %d", tt.defaultButton, config.defaultButton)
			}
			if config.mainIcon.value() != tt.icon {
				t.Fatalf("expected icon 0x%x, got 0x%x", tt.icon, config.mainIcon.value())
			}
		})
	}
}

func TestWindowsTaskDialogConfigRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		opts messageBoxOptions
	}{
		{
			name: "unknown kind",
			opts: messageBoxOptions{
				kind:          MessageBoxKind("custom"),
				buttons:       MessageBoxOK,
				defaultButton: MessageBoxResultOK,
			},
		},
		{
			name: "unknown buttons",
			opts: messageBoxOptions{
				kind:          MessageBoxInfo,
				buttons:       MessageBoxButtons("custom"),
				defaultButton: MessageBoxResultOK,
			},
		},
		{
			name: "default not in button set",
			opts: messageBoxOptions{
				kind:          MessageBoxInfo,
				buttons:       MessageBoxYesNo,
				defaultButton: MessageBoxResultCancel,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := windowsTaskDialogConfig(tt.opts); err == nil {
				t.Fatal("expected invalid task dialog options to fail")
			}
		})
	}
}

func TestWindowsTaskDialogConfigRichOptions(t *testing.T) {
	data, err := windowsTaskDialogConfig(messageBoxOptions{
		kind:                     MessageBoxInfo,
		details:                  "Long details",
		footer:                   "Footer",
		verificationText:         "Do not ask again",
		verificationChecked:      true,
		commandLinks:             true,
		expandedDetailsByDefault: true,
		customButtons: []MessageBoxButton{
			{ID: "save", Label: "Save\nSave changes", Result: MessageBoxResultYes},
			{ID: "discard", Label: "Discard", Result: MessageBoxResultNo},
		},
		defaultButtonID: "discard",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	config := data.config
	if config.commonButtons != 0 {
		t.Fatalf("custom buttons should replace common buttons, got 0x%x", config.commonButtons)
	}
	if config.buttonCount != 2 || config.buttons.value() == 0 {
		t.Fatalf("expected two custom buttons, got count=%d ptr=%d", config.buttonCount, config.buttons.value())
	}
	if config.defaultButton != taskDialogCustomButtonBase+1 {
		t.Fatalf("expected discard button to be default, got %d", config.defaultButton)
	}
	wantFlags := uint32(tdfAllowDialogCancellation | tdfUseCommandLinks | tdfExpandedByDefault | tdfVerificationFlagChecked)
	if config.flags&wantFlags != wantFlags {
		t.Fatalf("expected flags 0x%x to include 0x%x", config.flags, wantFlags)
	}
	if config.expandedInformation.value() == 0 || config.footer.value() == 0 || config.verificationText.value() == 0 {
		t.Fatalf("expected rich text fields to be configured: %#v", config)
	}
	if config.collapsedControlText.value() == 0 || config.expandedControlText.value() == 0 {
		t.Fatalf("expected details toggle text fields to be configured: %#v", config)
	}
	if config.windowTitle.value() == 0 || config.mainInstruction.value() == 0 {
		t.Fatalf("expected title and main instruction fields to stay alive: %#v", config)
	}
	if len(data.stringFields) != 7 {
		t.Fatalf("expected title, text, details controls, footer, and verification strings to be retained, got %d", len(data.stringFields))
	}
}

func TestWindowsTaskDialogConfigRejectsInvalidCustomButtons(t *testing.T) {
	tests := []struct {
		name string
		opts messageBoxOptions
	}{
		{
			name: "empty id",
			opts: messageBoxOptions{
				kind:          MessageBoxInfo,
				customButtons: []MessageBoxButton{{Label: "Save"}},
			},
		},
		{
			name: "empty label",
			opts: messageBoxOptions{
				kind:          MessageBoxInfo,
				customButtons: []MessageBoxButton{{ID: "save"}},
			},
		},
		{
			name: "duplicate id",
			opts: messageBoxOptions{
				kind: MessageBoxInfo,
				customButtons: []MessageBoxButton{
					{ID: "save", Label: "Save"},
					{ID: "save", Label: "Save again"},
				},
			},
		},
		{
			name: "missing default id",
			opts: messageBoxOptions{
				kind:            MessageBoxInfo,
				customButtons:   []MessageBoxButton{{ID: "save", Label: "Save"}},
				defaultButtonID: "discard",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := windowsTaskDialogConfig(tt.opts); err == nil {
				t.Fatal("expected invalid custom button options to fail")
			}
		})
	}
}

func TestWindowsMessageBoxFlags(t *testing.T) {
	tests := []struct {
		name string
		opts messageBoxOptions
		want uint32
	}{
		{
			name: "default ok info",
			opts: defaultMessageBoxOptions(),
			want: mbOK | mbIconInformation,
		},
		{
			name: "warning ok cancel default cancel",
			opts: messageBoxOptions{
				kind:          MessageBoxWarning,
				buttons:       MessageBoxOKCancel,
				defaultButton: MessageBoxResultCancel,
			},
			want: mbOKCancel | mbIconWarning | mbDefButton2,
		},
		{
			name: "question yes no default no",
			opts: messageBoxOptions{
				kind:          MessageBoxQuestion,
				buttons:       MessageBoxYesNo,
				defaultButton: MessageBoxResultNo,
			},
			want: mbYesNo | mbIconQuestion | mbDefButton2,
		},
		{
			name: "error yes no cancel default cancel",
			opts: messageBoxOptions{
				kind:          MessageBoxError,
				buttons:       MessageBoxYesNoCancel,
				defaultButton: MessageBoxResultCancel,
			},
			want: mbYesNoCancel | mbIconError | mbDefButton3,
		},
		{
			name: "retry cancel default retry",
			opts: messageBoxOptions{
				kind:          MessageBoxWarning,
				buttons:       MessageBoxRetryCancel,
				defaultButton: MessageBoxResultRetry,
			},
			want: mbRetryCancel | mbIconWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := windowsMessageBoxFlags(tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected flags 0x%x, got 0x%x", tt.want, got)
			}
		})
	}
}

func TestWindowsMessageBoxFlagsRejectInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		opts messageBoxOptions
	}{
		{
			name: "unknown kind",
			opts: messageBoxOptions{
				kind:          MessageBoxKind("custom"),
				buttons:       MessageBoxOK,
				defaultButton: MessageBoxResultOK,
			},
		},
		{
			name: "unknown buttons",
			opts: messageBoxOptions{
				kind:          MessageBoxInfo,
				buttons:       MessageBoxButtons("custom"),
				defaultButton: MessageBoxResultOK,
			},
		},
		{
			name: "default not in button set",
			opts: messageBoxOptions{
				kind:          MessageBoxInfo,
				buttons:       MessageBoxYesNo,
				defaultButton: MessageBoxResultCancel,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := windowsMessageBoxFlags(tt.opts); err == nil {
				t.Fatal("expected invalid message box options to fail")
			}
		})
	}
}

func TestWindowsMessageBoxResultMapping(t *testing.T) {
	tests := []struct {
		id   uintptr
		want MessageBoxResult
	}{
		{id: idOK, want: MessageBoxResultOK},
		{id: idCancel, want: MessageBoxResultCancel},
		{id: idRetry, want: MessageBoxResultRetry},
		{id: idYes, want: MessageBoxResultYes},
		{id: idNo, want: MessageBoxResultNo},
	}

	for _, tt := range tests {
		got, err := windowsMessageBoxResult(tt.id)
		if err != nil {
			t.Fatalf("unexpected error for id %d: %v", tt.id, err)
		}
		if got != tt.want {
			t.Fatalf("expected result %q for id %d, got %q", tt.want, tt.id, got)
		}
	}

	if _, err := windowsMessageBoxResult(999); err == nil {
		t.Fatal("expected unknown result id to fail")
	}
}

func TestWindowsMessageBoxDetailedResultMapping(t *testing.T) {
	data, err := windowsTaskDialogConfig(messageBoxOptions{
		kind: MessageBoxInfo,
		customButtons: []MessageBoxButton{
			{ID: "save", Label: "Save", Result: MessageBoxResultYes},
		},
	})
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}
	got, err := windowsMessageBoxDetailedResult(taskDialogCustomButtonBase, true, taskDialogCustomButtonBase, taskDialogCloseReasonNone, data)
	if err != nil {
		t.Fatalf("unexpected result error: %v", err)
	}
	if got.Result != MessageBoxResultYes || got.ButtonID != "save" || !got.VerificationChecked {
		t.Fatalf("unexpected detailed result: %#v", got)
	}

	standard, err := windowsMessageBoxDetailedResult(idOK, false, idOK, taskDialogCloseReasonNone, windowsTaskDialogConfigData{})
	if err != nil {
		t.Fatalf("unexpected standard result error: %v", err)
	}
	if standard.Result != MessageBoxResultOK || standard.ButtonID != "ok" || standard.VerificationChecked {
		t.Fatalf("unexpected standard detailed result: %#v", standard)
	}
}

func TestWindowsTaskDialogDistinguishesCancelButtonFromClose(t *testing.T) {
	cancel, err := windowsMessageBoxDetailedResult(idCancel, false, idCancel, taskDialogCloseReasonNone, windowsTaskDialogConfigData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancel.Result != MessageBoxResultCancel || cancel.ButtonID != "cancel" {
		t.Fatalf("expected clicked cancel to map to cancel, got %#v", cancel)
	}

	closeResult, err := windowsMessageBoxDetailedResult(idCancel, false, 0, taskDialogCloseReasonClose, windowsTaskDialogConfigData{})
	if err != nil {
		t.Fatalf("unexpected close result error: %v", err)
	}
	if closeResult.Result != MessageBoxResultClose || closeResult.ButtonID != "close" {
		t.Fatalf("expected non-button IDCANCEL to map to close, got %#v", closeResult)
	}

	closeWithCancelCallback, err := windowsMessageBoxDetailedResult(idCancel, false, idCancel, taskDialogCloseReasonClose, windowsTaskDialogConfigData{})
	if err != nil {
		t.Fatalf("unexpected close/cancel result error: %v", err)
	}
	if closeWithCancelCallback.Result != MessageBoxResultClose || closeWithCancelCallback.ButtonID != "close" {
		t.Fatalf("expected close source to win over IDCANCEL callback, got %#v", closeWithCancelCallback)
	}

	escapeResult, err := windowsMessageBoxDetailedResult(idCancel, false, 0, taskDialogCloseReasonEscape, windowsTaskDialogConfigData{})
	if err != nil {
		t.Fatalf("unexpected escape result error: %v", err)
	}
	if escapeResult.Result != MessageBoxResultEscape || escapeResult.ButtonID != "escape" {
		t.Fatalf("expected escape IDCANCEL to map to escape, got %#v", escapeResult)
	}

	escapeWithCancelCallback, err := windowsMessageBoxDetailedResult(idCancel, false, idCancel, taskDialogCloseReasonEscape, windowsTaskDialogConfigData{})
	if err != nil {
		t.Fatalf("unexpected escape/cancel result error: %v", err)
	}
	if escapeWithCancelCallback.Result != MessageBoxResultEscape || escapeWithCancelCallback.ButtonID != "escape" {
		t.Fatalf("expected escape source to win over IDCANCEL callback, got %#v", escapeWithCancelCallback)
	}
}

func TestWindowsTaskDialogIndirectRealCloseSources(t *testing.T) {
	if os.Getenv("FLUXUI_RUN_NATIVE_DIALOG_TESTS") != "1" {
		t.Skip("set FLUXUI_RUN_NATIVE_DIALOG_TESTS=1 to show and drive native TaskDialog windows")
	}
	if err := procTaskDialogIndirect.Find(); err != nil {
		t.Skipf("TaskDialogIndirect unavailable: %v", err)
	}

	const tdmClickButton = 0x0400 + 102
	tests := []struct {
		name       string
		drive      func(hwnd uintptr)
		wantResult MessageBoxResult
		wantButton string
	}{
		{
			name: "cancel button",
			drive: func(hwnd uintptr) {
				procMessageBoxPost.Call(hwnd, tdmClickButton, idCancel, 0)
			},
			wantResult: MessageBoxResultCancel,
			wantButton: "cancel",
		},
		{
			name: "escape key",
			drive: func(hwnd uintptr) {
				procMessageBoxPost.Call(hwnd, mbWMKeyDown, mbVKEscape, 0)
				time.AfterFunc(100*time.Millisecond, func() {
					procMessageBoxPost.Call(hwnd, mbWMClose, 0, 0)
				})
			},
			wantResult: MessageBoxResultEscape,
			wantButton: "escape",
		},
		{
			name: "window close",
			drive: func(hwnd uintptr) {
				procMessageBoxPost.Call(hwnd, mbWMClose, 0, 0)
			},
			wantResult: MessageBoxResultClose,
			wantButton: "close",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := showWindowsTaskDialogIndirectDriven(t, messageBoxOptions{
				title:         "FluxUI native TaskDialog test: " + tt.name,
				text:          "This dialog is closed automatically by the test.",
				kind:          MessageBoxInfo,
				buttons:       MessageBoxOKCancel,
				defaultButton: MessageBoxResultCancel,
			}, tt.drive)
			if got.Result != tt.wantResult || got.ButtonID != tt.wantButton {
				t.Fatalf("unexpected native TaskDialog result: got %#v want result=%q button=%q", got, tt.wantResult, tt.wantButton)
			}
		})
	}
}

func TestWindowsTaskDialogIndirectRealRichCustomButtons(t *testing.T) {
	if os.Getenv("FLUXUI_RUN_NATIVE_DIALOG_TESTS") != "1" {
		t.Skip("set FLUXUI_RUN_NATIVE_DIALOG_TESTS=1 to show and drive native TaskDialog windows")
	}
	if err := procTaskDialogIndirect.Find(); err != nil {
		t.Skipf("TaskDialogIndirect unavailable: %v", err)
	}

	const tdmClickButton = 0x0400 + 102
	got := showWindowsTaskDialogIndirectDriven(t, messageBoxOptions{
		title:            "FluxUI native rich TaskDialog test",
		text:             "This rich dialog is closed automatically by the test.",
		details:          "Expandable details exercise TaskDialogIndirect rich text fields.",
		footer:           "Footer text exercises the native footer pointer.",
		verificationText: "Remember this choice",
		kind:             MessageBoxInfo,
		commandLinks:     true,
		customButtons: []MessageBoxButton{
			{ID: "save", Label: "Save and close\nSave changes before closing.", Result: MessageBoxResultCustom},
			{ID: "discard", Label: "Discard\nClose without saving.", Result: MessageBoxResultCustom},
			{ID: "cancel", Label: "Cancel", Result: MessageBoxResultCancel},
		},
		defaultButtonID: "cancel",
	}, func(hwnd uintptr) {
		procMessageBoxPost.Call(hwnd, tdmClickButton, taskDialogCustomButtonBase, 0)
	})
	if got.Result != MessageBoxResultCustom || got.ButtonID != "save" || got.VerificationChecked {
		t.Fatalf("unexpected rich native TaskDialog result: %#v", got)
	}
}

func showWindowsTaskDialogIndirectDriven(t *testing.T, opts messageBoxOptions, drive func(hwnd uintptr)) MessageBoxDetailedResult {
	t.Helper()

	data, err := windowsTaskDialogConfig(opts)
	if err != nil {
		t.Fatalf("configure task dialog: %v", err)
	}
	callbackID, callbackState := registerWindowsTaskDialogCallback()
	defer unregisterWindowsTaskDialogCallback(callbackID)
	data.config.callback.set(windowsTaskDialogCallback)
	data.config.callbackData.set(uintptr(callbackID))
	config := data.config

	type taskDialogResult struct {
		result MessageBoxDetailedResult
		err    error
	}
	done := make(chan taskDialogResult, 1)
	go func() {
		var button int32
		var verificationChecked int32
		var verificationCheckedPtr uintptr
		if data.config.verificationText.value() != 0 {
			verificationCheckedPtr = uintptr(unsafe.Pointer(&verificationChecked))
		}
		r, _, _ := procTaskDialogIndirect.Call(
			uintptr(unsafe.Pointer(&config)),
			uintptr(unsafe.Pointer(&button)),
			0,
			verificationCheckedPtr,
		)
		keepAliveWindowsTaskDialogData(data)
		if err := hresultError(r); err != nil {
			done <- taskDialogResult{err: err}
			return
		}
		result, err := windowsMessageBoxDetailedResult(
			button,
			verificationChecked != 0,
			atomic.LoadInt32(&callbackState.clickedButton),
			atomic.LoadInt32(&callbackState.closeReason),
			data,
		)
		done <- taskDialogResult{result: result, err: err}
	}()

	hwnd := waitWindowsTaskDialogHWND(callbackState, 2*time.Second)
	if hwnd == 0 {
		t.Fatal("timed out waiting for TaskDialog HWND")
	}
	if drive != nil {
		drive(hwnd)
	}

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("TaskDialogIndirect returned error: %v", got.err)
		}
		return got.result
	case <-time.After(5 * time.Second):
		procMessageBoxPost.Call(hwnd, mbWMClose, 0, 0)
		t.Fatal("timed out waiting for TaskDialogIndirect to return")
	}
	return MessageBoxDetailedResult{}
}

func waitWindowsTaskDialogHWND(state *taskDialogCallbackState, timeout time.Duration) uintptr {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if hwnd := state.hwnd.Load(); hwnd != 0 {
			return hwnd
		}
		time.Sleep(10 * time.Millisecond)
	}
	return 0
}

func TestWindowsTaskDialogCallbackRecordsWindowAndButton(t *testing.T) {
	callbackID, state := registerWindowsTaskDialogCallback()
	defer unregisterWindowsTaskDialogCallback(callbackID)

	const hwnd = uintptr(12345)
	windowsTaskDialogCallbackProc(hwnd, tdnCreated, 0, 0, uintptr(callbackID))
	if got := state.hwnd.Load(); got != hwnd {
		t.Fatalf("expected hwnd %d, got %d", hwnd, got)
	}

	windowsTaskDialogCallbackProc(hwnd, tdnButtonClicked, idCancel, 0, uintptr(callbackID))
	if got := atomic.LoadInt32(&state.clickedButton); got != idCancel {
		t.Fatalf("expected clicked button %d, got %d", idCancel, got)
	}
}

func TestWindowsTaskDialogCloseReasonKeepsFirstCause(t *testing.T) {
	state := &taskDialogCallbackState{}
	markWindowsTaskDialogCloseReason(state, taskDialogCloseReasonEscape)
	markWindowsTaskDialogCloseReason(state, taskDialogCloseReasonClose)
	if got := atomic.LoadInt32(&state.closeReason); got != taskDialogCloseReasonEscape {
		t.Fatalf("expected first close reason to win, got %d", got)
	}
}

func TestWindowsModalDialogHelpers(t *testing.T) {
	if !windowsModalDialogClassLooksNative("#32770") {
		t.Fatal("expected standard dialog class to be recognized")
	}
	if !windowsModalDialogClassLooksNative("TaskDialog") {
		t.Fatal("expected task dialog class to be recognized")
	}
	if windowsModalDialogClassLooksNative("NotADialog") {
		t.Fatal("unexpected non-dialog class match")
	}
	if windowsModalDialogMatches(0, 0, "") {
		t.Fatal("zero hwnd should never match")
	}
	if findWindowsModalDialog(0, 0, "") != 0 {
		t.Fatal("zero thread id should not find a dialog")
	}
	if currentWindowsThreadID() == 0 {
		t.Fatal("expected current Windows thread id")
	}

	search := &windowsModalDialogSearch{title: "FluxUI"}
	searchID := registerWindowsModalDialogSearch(search)
	if searchID == 0 {
		t.Fatal("expected search registration id")
	}
	if got := lookupWindowsModalDialogSearch(searchID); got != search {
		t.Fatal("expected registered search to be returned")
	}
	unregisterWindowsModalDialogSearch(searchID)
	if got := lookupWindowsModalDialogSearch(searchID); got != nil {
		t.Fatal("expected unregistered search to be removed")
	}
}

func TestWindowsMessageBoxFallbackCancelMapping(t *testing.T) {
	got, err := windowsMessageBoxResult(idCancel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != MessageBoxResultCancel {
		t.Fatalf("expected fallback Windows IDCANCEL to map to cancel, got %q", got)
	}
}
