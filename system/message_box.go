package system

import (
	"context"
	"fmt"
)

// MessageBoxKind identifies the native message box icon and semantic style.
type MessageBoxKind string

const (
	MessageBoxInfo     MessageBoxKind = "info"
	MessageBoxWarning  MessageBoxKind = "warning"
	MessageBoxError    MessageBoxKind = "error"
	MessageBoxQuestion MessageBoxKind = "question"
)

// MessageBoxButtons identifies the visible native button set.
type MessageBoxButtons string

const (
	MessageBoxOK          MessageBoxButtons = "ok"
	MessageBoxOKCancel    MessageBoxButtons = "ok_cancel"
	MessageBoxYesNo       MessageBoxButtons = "yes_no"
	MessageBoxYesNoCancel MessageBoxButtons = "yes_no_cancel"
	MessageBoxRetryCancel MessageBoxButtons = "retry_cancel"
)

// MessageBoxResult identifies the native button or close action chosen by the user.
type MessageBoxResult string

const (
	MessageBoxResultOK     MessageBoxResult = "ok"
	MessageBoxResultCancel MessageBoxResult = "cancel"
	MessageBoxResultYes    MessageBoxResult = "yes"
	MessageBoxResultNo     MessageBoxResult = "no"
	MessageBoxResultRetry  MessageBoxResult = "retry"
	// MessageBoxResultClose is reserved for drivers that can distinguish a window close action.
	//
	// Windows MessageBoxW reports Escape and close as Cancel when a cancel action is available,
	// so the Windows v1 driver returns MessageBoxResultCancel for those cases.
	MessageBoxResultClose MessageBoxResult = "close"
)

// MessageBoxOption configures native message boxes.
type MessageBoxOption func(*messageBoxOptions)

type messageBoxOptions struct {
	title         string
	text          string
	kind          MessageBoxKind
	buttons       MessageBoxButtons
	defaultButton MessageBoxResult
	owner         uintptr
}

type messageBoxDriver interface {
	showMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error)
}

func defaultMessageBoxOptions() messageBoxOptions {
	return messageBoxOptions{
		kind:          MessageBoxInfo,
		buttons:       MessageBoxOK,
		defaultButton: MessageBoxResultOK,
	}
}

// MessageBoxTitle sets the native message box title.
func MessageBoxTitle(value string) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.title = value
	}
}

// MessageBoxText sets the native message box body text.
func MessageBoxText(value string) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.text = value
	}
}

// MessageBoxStyle sets the native message box semantic style.
func MessageBoxStyle(kind MessageBoxKind) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.kind = kind
	}
}

// MessageBoxButtonSet sets the native message box button set.
func MessageBoxButtonSet(buttons MessageBoxButtons) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.buttons = buttons
	}
}

// MessageBoxDefaultButton sets the native message box default action.
func MessageBoxDefaultButton(result MessageBoxResult) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.defaultButton = result
	}
}

// MessageBoxOwner sets the native owner window handle when the caller already has one.
//
// On Windows this value is interpreted as HWND. Passing 0 keeps the message box ownerless.
func MessageBoxOwner(owner uintptr) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.owner = owner
	}
}

// ShowMessageBox opens a blocking native message box.
func ShowMessageBox(ctx context.Context, options ...MessageBoxOption) (MessageBoxResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	opts := defaultMessageBoxOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	driverMu.RLock()
	d := activeDriver
	driverMu.RUnlock()

	mb, ok := d.(messageBoxDriver)
	if !ok || d == nil || !d.capabilities().Supports(CapabilityMessageBox) {
		return "", fmt.Errorf("system: %s: %w", CapabilityMessageBox, ErrUnsupported)
	}
	return mb.showMessageBox(ctx, opts)
}
