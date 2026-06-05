package ui

import (
	"context"

	"github.com/xiaowumin-mark/FluxUI/system"
)

// CurrentWindowNativeHandle returns the current FluxUI window's native handle.
//
// On Windows this is the HWND. Unsupported platforms, windows that have not
// received a native view event, and closed windows return false.
func CurrentWindowNativeHandle(ctx *Context) (uintptr, bool) {
	if ctx == nil {
		return 0, false
	}
	handle, ok := GetWindow(CurrentWindowID(ctx))
	if !ok {
		return 0, false
	}
	return handle.NativeHandle()
}

// OpenFileDialog opens a native file dialog and automatically binds it to the
// current FluxUI window when a native owner handle is available.
func OpenFileDialog(ctx *Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return OpenFileDialogContext(ctx, context.Background(), opts...)
}

// OpenFileDialogContext is OpenFileDialog with an explicit cancellation context.
func OpenFileDialogContext(ctx *Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return system.OpenFileDialog(callCtx, withFileDialogOwner(ctx, opts)...)
}

// OpenFilesDialog opens a native multi-file dialog and automatically binds it
// to the current FluxUI window when a native owner handle is available.
func OpenFilesDialog(ctx *Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return OpenFilesDialogContext(ctx, context.Background(), opts...)
}

// OpenFilesDialogContext is OpenFilesDialog with an explicit cancellation context.
func OpenFilesDialogContext(ctx *Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return system.OpenFilesDialog(callCtx, withFileDialogOwner(ctx, opts)...)
}

// SaveFileDialog opens a native save dialog and automatically binds it to the
// current FluxUI window when a native owner handle is available.
func SaveFileDialog(ctx *Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return SaveFileDialogContext(ctx, context.Background(), opts...)
}

// SaveFileDialogContext is SaveFileDialog with an explicit cancellation context.
func SaveFileDialogContext(ctx *Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return system.SaveFileDialog(callCtx, withFileDialogOwner(ctx, opts)...)
}

// PickFolderDialog opens a native folder picker and automatically binds it to
// the current FluxUI window when a native owner handle is available.
func PickFolderDialog(ctx *Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return PickFolderDialogContext(ctx, context.Background(), opts...)
}

// PickFolderDialogContext is PickFolderDialog with an explicit cancellation context.
func PickFolderDialogContext(ctx *Context, callCtx context.Context, opts ...system.FileDialogOption) (system.FileDialogResult, error) {
	return system.PickFolderDialog(callCtx, withFileDialogOwner(ctx, opts)...)
}

// ShowMessageBox opens a native message box and automatically binds it to the
// current FluxUI window when a native owner handle is available.
func ShowMessageBox(ctx *Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error) {
	return ShowMessageBoxContext(ctx, context.Background(), opts...)
}

// ShowMessageBoxContext is ShowMessageBox with an explicit cancellation context.
func ShowMessageBoxContext(ctx *Context, callCtx context.Context, opts ...system.MessageBoxOption) (system.MessageBoxResult, error) {
	return system.ShowMessageBox(callCtx, withMessageBoxOwner(ctx, opts)...)
}

func withFileDialogOwner(ctx *Context, opts []system.FileDialogOption) []system.FileDialogOption {
	owner, ok := CurrentWindowNativeHandle(ctx)
	if !ok {
		return opts
	}
	withOwner := make([]system.FileDialogOption, 0, len(opts)+1)
	withOwner = append(withOwner, system.FileDialogOwner(owner))
	withOwner = append(withOwner, opts...)
	return withOwner
}

func withMessageBoxOwner(ctx *Context, opts []system.MessageBoxOption) []system.MessageBoxOption {
	owner, ok := CurrentWindowNativeHandle(ctx)
	if !ok {
		return opts
	}
	withOwner := make([]system.MessageBoxOption, 0, len(opts)+1)
	withOwner = append(withOwner, system.MessageBoxOwner(owner))
	withOwner = append(withOwner, opts...)
	return withOwner
}
