//go:build windows

package system

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	tdcbfOKButton     = 0x0001
	tdcbfYesButton    = 0x0002
	tdcbfNoButton     = 0x0004
	tdcbfCancelButton = 0x0008
	tdcbfRetryButton  = 0x0010

	tdWarningIcon     = 0xffff
	tdErrorIcon       = 0xfffe
	tdInformationIcon = 0xfffd

	mbOK          = 0x00000000
	mbOKCancel    = 0x00000001
	mbYesNoCancel = 0x00000003
	mbYesNo       = 0x00000004
	mbRetryCancel = 0x00000005

	mbIconError       = 0x00000010
	mbIconQuestion    = 0x00000020
	mbIconWarning     = 0x00000030
	mbIconInformation = 0x00000040

	mbDefButton2 = 0x00000100
	mbDefButton3 = 0x00000200

	idOK     = 1
	idCancel = 2
	idRetry  = 4
	idYes    = 6
	idNo     = 7
)

var (
	comctl32               = windows.NewLazySystemDLL("comctl32.dll")
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procTaskDialog         = comctl32.NewProc("TaskDialog")
	procTaskDialogIndirect = comctl32.NewProc("TaskDialogIndirect")
	procMessageBox         = user32.NewProc("MessageBoxW")
)

type taskDialogConfig struct {
	size                 uint32
	parent               uintptr
	instance             uintptr
	flags                uint32
	commonButtons        uint32
	windowTitle          *uint16
	mainIcon             uintptr
	mainInstruction      *uint16
	content              *uint16
	buttonCount          uint32
	buttons              uintptr
	defaultButton        int32
	radioButtonCount     uint32
	radioButtons         uintptr
	defaultRadioButton   int32
	verificationText     *uint16
	expandedInformation  *uint16
	expandedControlText  *uint16
	collapsedControlText *uint16
	footerIcon           uintptr
	footer               *uint16
	callback             uintptr
	callbackData         uintptr
	width                uint32
}

func (windowsDriver) showMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	result, err := showWindowsTaskDialog(ctx, opts)
	if err == nil {
		return result, nil
	}
	if IsUnavailable(err) {
		return showWindowsMessageBox(ctx, opts)
	}
	return "", err
}

func showWindowsTaskDialog(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	if err := procTaskDialog.Find(); err != nil {
		return "", fmt.Errorf("system: task dialog unavailable: %w", ErrUnavailable)
	}

	commonButtons, err := windowsTaskDialogCommonButtons(opts.buttons)
	if err != nil {
		return "", err
	}
	icon, err := windowsTaskDialogIcon(opts.kind)
	if err != nil {
		return "", err
	}

	defaultButton := opts.defaultButton
	if defaultButton == "" {
		defaultButton = firstMessageBoxButton(opts.buttons)
	}
	if _, ok := messageBoxButtonID(opts.buttons, defaultButton); !ok {
		return "", fmt.Errorf("system: default button %q is not in button set %q", defaultButton, opts.buttons)
	}

	title16, err := windows.UTF16PtrFromString(opts.title)
	if err != nil {
		return "", fmt.Errorf("system: configure task dialog title: %w", err)
	}
	text16, err := windows.UTF16PtrFromString(opts.text)
	if err != nil {
		return "", fmt.Errorf("system: configure task dialog text: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	var button int32
	r, _, callErr := procTaskDialog.Call(
		opts.owner,
		0,
		uintptr(unsafe.Pointer(title16)),
		uintptr(unsafe.Pointer(text16)),
		0,
		uintptr(commonButtons),
		icon,
		uintptr(unsafe.Pointer(&button)),
	)
	if err := hresultError(r); err != nil {
		if callErr != syscall.Errno(0) {
			return "", fmt.Errorf("system: show task dialog: %w: %v", ErrUnavailable, callErr)
		}
		return "", fmt.Errorf("system: show task dialog: %w: %v", ErrUnavailable, err)
	}

	result, err := windowsMessageBoxResult(uintptr(button))
	if err != nil {
		return "", fmt.Errorf("system: show task dialog: %w", err)
	}
	return result, nil
}

func showWindowsMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	flags, err := windowsMessageBoxFlags(opts)
	if err != nil {
		return "", err
	}

	title16, err := windows.UTF16PtrFromString(opts.title)
	if err != nil {
		return "", fmt.Errorf("system: configure message box title: %w", err)
	}
	text16, err := windows.UTF16PtrFromString(opts.text)
	if err != nil {
		return "", fmt.Errorf("system: configure message box text: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	r, _, callErr := procMessageBox.Call(
		opts.owner,
		uintptr(unsafe.Pointer(text16)),
		uintptr(unsafe.Pointer(title16)),
		uintptr(flags),
	)
	if r == 0 {
		if callErr != syscall.Errno(0) {
			return "", fmt.Errorf("system: show message box: %w", callErr)
		}
		return "", fmt.Errorf("system: show message box: %w", ErrUnavailable)
	}

	result, err := windowsMessageBoxResult(r)
	if err != nil {
		return "", fmt.Errorf("system: show message box: %w", err)
	}
	return result, nil
}

func windowsTaskDialogConfig(opts messageBoxOptions) (taskDialogConfig, error) {
	commonButtons, err := windowsTaskDialogCommonButtons(opts.buttons)
	if err != nil {
		return taskDialogConfig{}, err
	}

	icon, err := windowsTaskDialogIcon(opts.kind)
	if err != nil {
		return taskDialogConfig{}, err
	}

	defaultButton := opts.defaultButton
	if defaultButton == "" {
		defaultButton = firstMessageBoxButton(opts.buttons)
	}
	defaultID, ok := messageBoxButtonID(opts.buttons, defaultButton)
	if !ok {
		return taskDialogConfig{}, fmt.Errorf("system: default button %q is not in button set %q", defaultButton, opts.buttons)
	}

	title16, err := windows.UTF16PtrFromString(opts.title)
	if err != nil {
		return taskDialogConfig{}, fmt.Errorf("system: configure task dialog title: %w", err)
	}
	text16, err := windows.UTF16PtrFromString(opts.text)
	if err != nil {
		return taskDialogConfig{}, fmt.Errorf("system: configure task dialog text: %w", err)
	}

	return taskDialogConfig{
		size:            uint32(unsafe.Sizeof(taskDialogConfig{})),
		parent:          opts.owner,
		commonButtons:   commonButtons,
		windowTitle:     title16,
		mainIcon:        icon,
		mainInstruction: text16,
		defaultButton:   int32(defaultID),
	}, nil
}

func windowsTaskDialogCommonButtons(buttons MessageBoxButtons) (uint32, error) {
	switch buttons {
	case "", MessageBoxOK:
		return tdcbfOKButton, nil
	case MessageBoxOKCancel:
		return tdcbfOKButton | tdcbfCancelButton, nil
	case MessageBoxYesNo:
		return tdcbfYesButton | tdcbfNoButton, nil
	case MessageBoxYesNoCancel:
		return tdcbfYesButton | tdcbfNoButton | tdcbfCancelButton, nil
	case MessageBoxRetryCancel:
		return tdcbfRetryButton | tdcbfCancelButton, nil
	default:
		return 0, fmt.Errorf("system: unknown message box button set %q", buttons)
	}
}

func windowsTaskDialogIcon(kind MessageBoxKind) (uintptr, error) {
	switch kind {
	case "", MessageBoxInfo:
		return tdInformationIcon, nil
	case MessageBoxWarning:
		return tdWarningIcon, nil
	case MessageBoxError:
		return tdErrorIcon, nil
	case MessageBoxQuestion:
		return tdInformationIcon, nil
	default:
		return 0, fmt.Errorf("system: unknown message box kind %q", kind)
	}
}

func windowsMessageBoxFlags(opts messageBoxOptions) (uint32, error) {
	var flags uint32

	switch opts.buttons {
	case "", MessageBoxOK:
		flags |= mbOK
	case MessageBoxOKCancel:
		flags |= mbOKCancel
	case MessageBoxYesNo:
		flags |= mbYesNo
	case MessageBoxYesNoCancel:
		flags |= mbYesNoCancel
	case MessageBoxRetryCancel:
		flags |= mbRetryCancel
	default:
		return 0, fmt.Errorf("system: unknown message box button set %q", opts.buttons)
	}

	switch opts.kind {
	case "", MessageBoxInfo:
		flags |= mbIconInformation
	case MessageBoxWarning:
		flags |= mbIconWarning
	case MessageBoxError:
		flags |= mbIconError
	case MessageBoxQuestion:
		flags |= mbIconQuestion
	default:
		return 0, fmt.Errorf("system: unknown message box kind %q", opts.kind)
	}

	defaultButton := opts.defaultButton
	if defaultButton == "" {
		defaultButton = firstMessageBoxButton(opts.buttons)
	}
	index, ok := messageBoxButtonIndex(opts.buttons, defaultButton)
	if !ok {
		return 0, fmt.Errorf("system: default button %q is not in button set %q", defaultButton, opts.buttons)
	}
	switch index {
	case 1:
	case 2:
		flags |= mbDefButton2
	case 3:
		flags |= mbDefButton3
	default:
		return 0, fmt.Errorf("system: unsupported default button index %d", index)
	}

	return flags, nil
}

func firstMessageBoxButton(buttons MessageBoxButtons) MessageBoxResult {
	switch buttons {
	case MessageBoxOKCancel:
		return MessageBoxResultOK
	case MessageBoxYesNo, MessageBoxYesNoCancel:
		return MessageBoxResultYes
	case MessageBoxRetryCancel:
		return MessageBoxResultRetry
	default:
		return MessageBoxResultOK
	}
}

func messageBoxButtonIndex(buttons MessageBoxButtons, result MessageBoxResult) (int, bool) {
	switch buttons {
	case "", MessageBoxOK:
		if result == MessageBoxResultOK {
			return 1, true
		}
	case MessageBoxOKCancel:
		switch result {
		case MessageBoxResultOK:
			return 1, true
		case MessageBoxResultCancel:
			return 2, true
		}
	case MessageBoxYesNo:
		switch result {
		case MessageBoxResultYes:
			return 1, true
		case MessageBoxResultNo:
			return 2, true
		}
	case MessageBoxYesNoCancel:
		switch result {
		case MessageBoxResultYes:
			return 1, true
		case MessageBoxResultNo:
			return 2, true
		case MessageBoxResultCancel:
			return 3, true
		}
	case MessageBoxRetryCancel:
		switch result {
		case MessageBoxResultRetry:
			return 1, true
		case MessageBoxResultCancel:
			return 2, true
		}
	}
	return 0, false
}

func messageBoxButtonID(buttons MessageBoxButtons, result MessageBoxResult) (uintptr, bool) {
	if _, ok := messageBoxButtonIndex(buttons, result); !ok {
		return 0, false
	}
	switch result {
	case MessageBoxResultOK:
		return idOK, true
	case MessageBoxResultCancel:
		return idCancel, true
	case MessageBoxResultRetry:
		return idRetry, true
	case MessageBoxResultYes:
		return idYes, true
	case MessageBoxResultNo:
		return idNo, true
	default:
		return 0, false
	}
}

func windowsMessageBoxResult(id uintptr) (MessageBoxResult, error) {
	switch id {
	case idOK:
		return MessageBoxResultOK, nil
	case idCancel:
		return MessageBoxResultCancel, nil
	case idRetry:
		return MessageBoxResultRetry, nil
	case idYes:
		return MessageBoxResultYes, nil
	case idNo:
		return MessageBoxResultNo, nil
	default:
		return "", fmt.Errorf("unexpected message box result %d", id)
	}
}
