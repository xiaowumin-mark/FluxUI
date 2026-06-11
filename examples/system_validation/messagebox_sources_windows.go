//go:build windows

package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/xiaowumin-mark/FluxUI/system"
)

const (
	validationWMClose        = 0x0010
	validationWMKeyDown      = 0x0100
	validationVKEscape       = 0x1B
	validationTDMClickButton = 0x0400 + 102
	validationIDCancel       = 2
	validationCustomButton   = 1000
)

var (
	validationEnumWindows        = validationUser32.NewProc("EnumWindows")
	validationGetWindowText      = validationUser32.NewProc("GetWindowTextW")
	validationGetClassName       = validationUser32.NewProc("GetClassNameW")
	validationPostMessage        = validationUser32.NewProc("PostMessageW")
	validationSendMessage        = validationUser32.NewProc("SendMessageW")
	validationEnumWindowCallback = syscall.NewCallback(validationEnumWindowProc)

	validationFindWindowMu     sync.Mutex
	validationFindWindowTitle  string
	validationFindWindowResult uintptr
)

func validateMessageBoxSources() error {
	tests := []struct {
		name       string
		drive      func(hwnd uintptr)
		wantResult system.MessageBoxResult
	}{
		{
			name: "cancel",
			drive: func(hwnd uintptr) {
				validationSendMessage.Call(hwnd, validationTDMClickButton, validationIDCancel, 0)
			},
			wantResult: system.MessageBoxResultCancel,
		},
		{
			name: "escape",
			drive: func(hwnd uintptr) {
				validationSendMessage.Call(hwnd, validationWMKeyDown, validationVKEscape, 0)
				time.AfterFunc(100*time.Millisecond, func() {
					validationSendMessage.Call(hwnd, validationWMClose, 0, 0)
				})
			},
			wantResult: system.MessageBoxResultEscape,
		},
		{
			name: "close",
			drive: func(hwnd uintptr) {
				validationSendMessage.Call(hwnd, validationWMClose, 0, 0)
			},
			wantResult: system.MessageBoxResultClose,
		},
	}

	for _, tt := range tests {
		title := fmt.Sprintf("FluxUI source validation %s %d", tt.name, time.Now().UnixNano())
		resultCh := make(chan system.MessageBoxDetailedResponse, 1)
		go func() {
			result, err := system.ShowMessageBoxDetailed(
				nil,
				system.MessageBoxTitle(title),
				system.MessageBoxText("This TaskDialog is closed automatically by validation."),
				system.MessageBoxDetails("TaskDialogIndirect source validation."),
				system.MessageBoxStyle(system.MessageBoxInfo),
				system.MessageBoxButtonSet(system.MessageBoxOKCancel),
				system.MessageBoxDefaultButton(system.MessageBoxResultCancel),
			)
			resultCh <- system.MessageBoxDetailedResponse{Result: result, Err: err}
		}()

		hwnd, response, returned := waitForValidationWindowOrResult(title, resultCh, 3*time.Second)
		if returned {
			if response.Err != nil {
				return fmt.Errorf("%s: %w", tt.name, response.Err)
			}
			return fmt.Errorf("%s: message box returned before validation could find it: result=%q", tt.name, response.Result.Result)
		}
		if hwnd == 0 {
			return fmt.Errorf("%s: timed out waiting for native message box", tt.name)
		}
		tt.drive(hwnd)

		select {
		case response := <-resultCh:
			if response.Err != nil {
				return fmt.Errorf("%s: %w", tt.name, response.Err)
			}
			if response.Result.Result != tt.wantResult {
				return fmt.Errorf("%s: got result %q, want %q", tt.name, response.Result.Result, tt.wantResult)
			}
		case <-time.After(5 * time.Second):
			validationSendMessage.Call(hwnd, validationWMClose, 0, 0)
			return fmt.Errorf("%s: timed out waiting for message box result", tt.name)
		}
	}
	return validateRichMessageBox()
}

func validateRichMessageBox() error {
	title := fmt.Sprintf("FluxUI rich validation %d", time.Now().UnixNano())
	resultCh := make(chan system.MessageBoxDetailedResponse, 1)
	go func() {
		result, err := system.ShowMessageBoxDetailed(
			nil,
			system.MessageBoxTitle(title),
			system.MessageBoxText("This rich TaskDialog is closed automatically by validation."),
			system.MessageBoxDetails("TaskDialogIndirect rich details validation."),
			system.MessageBoxFooter("TaskDialogIndirect rich footer validation."),
			system.MessageBoxVerification("Remember this choice", false),
			system.MessageBoxStyle(system.MessageBoxInfo),
			system.MessageBoxCustomButtons(
				system.MessageBoxButton{ID: "save", Label: "Save and close\nSave changes before closing.", Result: system.MessageBoxResultCustom},
				system.MessageBoxButton{ID: "discard", Label: "Discard\nClose without saving.", Result: system.MessageBoxResultCustom},
				system.MessageBoxButton{ID: "cancel", Label: "Cancel", Result: system.MessageBoxResultCancel},
			),
			system.MessageBoxDefaultButtonID("cancel"),
			system.MessageBoxCommandLinks(true),
		)
		resultCh <- system.MessageBoxDetailedResponse{Result: result, Err: err}
	}()

	hwnd, response, returned := waitForValidationWindowOrResult(title, resultCh, 3*time.Second)
	if returned {
		if response.Err != nil {
			return fmt.Errorf("rich: %w", response.Err)
		}
		return fmt.Errorf("rich: message box returned before validation could find it: result=%q", response.Result.Result)
	}
	if hwnd == 0 {
		return fmt.Errorf("rich: timed out waiting for native message box")
	}
	validationSendMessage.Call(hwnd, validationTDMClickButton, validationCustomButton, 0)

	select {
	case response := <-resultCh:
		if response.Err != nil {
			return fmt.Errorf("rich: %w", response.Err)
		}
		if response.Result.Result != system.MessageBoxResultCustom || response.Result.ButtonID != "save" {
			return fmt.Errorf("rich: got result=%q button=%q, want custom/save", response.Result.Result, response.Result.ButtonID)
		}
		if response.Result.VerificationChecked {
			return fmt.Errorf("rich: verification unexpectedly checked")
		}
	case <-time.After(5 * time.Second):
		validationSendMessage.Call(hwnd, validationWMClose, 0, 0)
		return fmt.Errorf("rich: timed out waiting for message box result")
	}
	return nil
}

func waitForValidationWindowOrResult(title string, resultCh <-chan system.MessageBoxDetailedResponse, timeout time.Duration) (uintptr, system.MessageBoxDetailedResponse, bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case response := <-resultCh:
			return 0, response, true
		default:
		}
		if hwnd := findValidationWindowByTitle(title); hwnd != 0 {
			return hwnd, system.MessageBoxDetailedResponse{}, false
		}
		time.Sleep(25 * time.Millisecond)
	}
	return 0, system.MessageBoxDetailedResponse{}, false
}

func waitForValidationWindowTitle(title string, timeout time.Duration) uintptr {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if hwnd := findValidationWindowByTitle(title); hwnd != 0 {
			return hwnd
		}
		time.Sleep(25 * time.Millisecond)
	}
	return 0
}

func findValidationWindowByTitle(title string) uintptr {
	validationFindWindowMu.Lock()
	validationFindWindowTitle = title
	validationFindWindowResult = 0
	validationFindWindowMu.Unlock()

	validationEnumWindows.Call(validationEnumWindowCallback, 0)

	validationFindWindowMu.Lock()
	hwnd := validationFindWindowResult
	validationFindWindowTitle = ""
	validationFindWindowResult = 0
	validationFindWindowMu.Unlock()
	return hwnd
}

func validationEnumWindowProc(hwnd uintptr, _ uintptr) uintptr {
	validationFindWindowMu.Lock()
	title := validationFindWindowTitle
	validationFindWindowMu.Unlock()
	if title == "" {
		return 0
	}
	if validationWindowTitle(hwnd) == title && validationWindowClassLooksNative(hwnd) {
		validationFindWindowMu.Lock()
		validationFindWindowResult = hwnd
		validationFindWindowMu.Unlock()
		return 0
	}
	return 1
}

func validationWindowTitle(hwnd uintptr) string {
	var buf [256]uint16
	validationGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return windows.UTF16ToString(buf[:])
}

func validationWindowClassLooksNative(hwnd uintptr) bool {
	var buf [128]uint16
	validationGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	className := windows.UTF16ToString(buf[:])
	return className == "#32770" || strings.Contains(className, "TaskDialog")
}
