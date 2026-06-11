//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/xiaowumin-mark/FluxUI/system"
)

const (
	validationWSOverlappedWindow = 0x00CF0000
	validationSWShowNoActivate   = 8
)

var (
	validationUser32              = syscall.NewLazyDLL("user32.dll")
	validationKernel32            = syscall.NewLazyDLL("kernel32.dll")
	validationDefWindowProc       = validationUser32.NewProc("DefWindowProcW")
	validationRegisterClassEx     = validationUser32.NewProc("RegisterClassExW")
	validationCreateWindowEx      = validationUser32.NewProc("CreateWindowExW")
	validationDestroyWindow       = validationUser32.NewProc("DestroyWindow")
	validationIsWindowEnabled     = validationUser32.NewProc("IsWindowEnabled")
	validationShowWindow          = validationUser32.NewProc("ShowWindow")
	validationGetModuleHandle     = validationKernel32.NewProc("GetModuleHandleW")
	validationOwnerWindowCallback = syscall.NewCallback(validationOwnerWindowProc)
)

type validationWndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

func validateOwnerModal() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	owner, cleanup, err := createValidationOwnerWindow()
	if err != nil {
		return err
	}
	defer cleanup()

	if !validationWindowEnabled(owner) {
		return fmt.Errorf("owner window started disabled")
	}
	if err := validateMessageBoxOwnerModal(owner); err != nil {
		return fmt.Errorf("message_box: %w", err)
	}
	if err := validateFileDialogOwnerModal(owner); err != nil {
		return fmt.Errorf("file_dialog: %w", err)
	}
	return nil
}

func validateMessageBoxOwnerModal(owner uintptr) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	disabledCh := make(chan bool, 1)
	go func() {
		disabled := waitForValidationOwnerDisabled(owner, 2*time.Second)
		disabledCh <- disabled
		if disabled {
			cancel()
		}
	}()

	_, err := system.ShowMessageBox(ctx,
		system.MessageBoxOwner(owner),
		system.MessageBoxTitle("FluxUI owner validation"),
		system.MessageBoxText("This owner-bound message box should close automatically."),
		system.MessageBoxButtonSet(system.MessageBoxOKCancel),
	)

	if !<-disabledCh {
		return fmt.Errorf("owner window was not disabled while message box was open")
	}
	if err == nil {
		return fmt.Errorf("expected context cancellation")
	}
	if !isContextCancellation(err) {
		return err
	}
	return nil
}

func validateFileDialogOwnerModal(owner uintptr) error {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	disabledCh := make(chan bool, 1)
	go func() {
		disabled := waitForValidationOwnerDisabled(owner, 3*time.Second)
		disabledCh <- disabled
		if disabled {
			cancel()
		}
	}()

	_, err := system.OpenFileDialog(ctx,
		system.FileDialogOwner(owner),
		system.FileDialogTitle("FluxUI owner validation"),
	)

	if !<-disabledCh {
		return fmt.Errorf("owner window was not disabled while file dialog was open")
	}
	if err == nil {
		return fmt.Errorf("expected context cancellation")
	}
	if !isContextCancellation(err) {
		return err
	}
	return nil
}

func waitForValidationOwnerDisabled(owner uintptr, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !validationWindowEnabled(owner) {
			return true
		}
		time.Sleep(25 * time.Millisecond)
	}
	return false
}

func validationWindowEnabled(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	enabled, _, _ := validationIsWindowEnabled.Call(hwnd)
	return enabled != 0
}

func createValidationOwnerWindow() (uintptr, func(), error) {
	hInstance, _, callErr := validationGetModuleHandle.Call(0)
	if hInstance == 0 {
		if callErr != syscall.Errno(0) {
			return 0, func() {}, callErr
		}
		return 0, func() {}, fmt.Errorf("get module handle: unavailable")
	}

	className, err := windows.UTF16PtrFromString(fmt.Sprintf("FluxUIValidationOwner-%d-%x", os.Getpid(), uintptr(validationOwnerWindowCallback)))
	if err != nil {
		return 0, func() {}, err
	}
	wc := validationWndClassEx{
		cbSize:        uint32(unsafe.Sizeof(validationWndClassEx{})),
		lpfnWndProc:   validationOwnerWindowCallback,
		hInstance:     hInstance,
		lpszClassName: className,
	}
	registered, _, callErr := validationRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if registered == 0 && callErr != syscall.Errno(1410) {
		if callErr != syscall.Errno(0) {
			return 0, func() {}, callErr
		}
		return 0, func() {}, fmt.Errorf("register owner window class: unavailable")
	}

	title, err := windows.UTF16PtrFromString("FluxUI validation owner")
	if err != nil {
		return 0, func() {}, err
	}
	hwnd, _, callErr := validationCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		validationWSOverlappedWindow,
		0,
		0,
		160,
		120,
		0,
		0,
		hInstance,
		0,
	)
	if hwnd == 0 {
		if callErr != syscall.Errno(0) {
			return 0, func() {}, callErr
		}
		return 0, func() {}, fmt.Errorf("create owner window: unavailable")
	}
	validationShowWindow.Call(hwnd, validationSWShowNoActivate)

	return hwnd, func() {
		validationDestroyWindow.Call(hwnd)
	}, nil
}

func validationOwnerWindowProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	result, _, _ := validationDefWindowProc.Call(hwnd, msg, wParam, lParam)
	return result
}
