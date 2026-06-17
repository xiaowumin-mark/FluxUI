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
	MessageBoxResultCustom MessageBoxResult = "custom"
	// MessageBoxResultEscape is returned by drivers that can distinguish the Escape key.
	//
	// Windows TaskDialogIndirect can distinguish Escape from a clicked Cancel button
	// and from WM_CLOSE/titlebar close in the common path. Traditional MessageBoxW
	// still reports Escape and close as Cancel.
	MessageBoxResultEscape MessageBoxResult = "escape"
	// MessageBoxResultClose is reserved for drivers that can distinguish a window close action.
	//
	// Windows TaskDialogIndirect can distinguish a clicked Cancel button from WM_CLOSE
	// titlebar close in the common path. Traditional MessageBoxW still reports Escape
	// and close as Cancel.
	MessageBoxResultClose MessageBoxResult = "close"
)

// MessageBoxButton describes a TaskDialogIndirect custom button.
type MessageBoxButton struct {
	ID     string
	Label  string
	Result MessageBoxResult
}

// MessageBoxOption configures native message boxes.
type MessageBoxOption func(*messageBoxOptions)

// MessageBoxResponse is returned by asynchronous message box calls.
type MessageBoxResponse struct {
	Result MessageBoxResult
	Err    error
}

// MessageBoxDetailedResult includes result data only richer native dialogs can report.
type MessageBoxDetailedResult struct {
	Result              MessageBoxResult
	ButtonID            string
	VerificationChecked bool
}

// MessageBoxDetailedResponse is returned by asynchronous detailed message box calls.
type MessageBoxDetailedResponse struct {
	Result MessageBoxDetailedResult
	Err    error
}

type messageBoxOptions struct {
	title                     string
	text                      string
	details                   string
	footer                    string
	verificationText          string
	verificationChecked       bool
	kind                      MessageBoxKind
	buttons                   MessageBoxButtons
	customButtons             []MessageBoxButton
	defaultButton             MessageBoxResult
	defaultButtonID           string
	commandLinks              bool
	commandLinksNoIcon        bool
	expandedDetailsByDefault  bool
	expandDetailsInFooterArea bool
	owner                     uintptr
}

type messageBoxDriver interface {
	showMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error)
}

type detailedMessageBoxDriver interface {
	showDetailedMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxDetailedResult, error)
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

// MessageBoxDetails sets expandable details for drivers that support rich task dialogs.
func MessageBoxDetails(value string) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.details = value
	}
}

// MessageBoxFooter sets footer text for drivers that support rich task dialogs.
func MessageBoxFooter(value string) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.footer = value
	}
}

// MessageBoxVerification sets a checkbox label and initial checked state.
func MessageBoxVerification(label string, checked bool) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.verificationText = label
		opts.verificationChecked = checked
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

// MessageBoxCustomButtons replaces the standard button set with custom buttons.
//
// Use ShowMessageBoxDetailed when you need the selected custom button ID.
func MessageBoxCustomButtons(buttons ...MessageBoxButton) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.customButtons = cloneMessageBoxButtons(buttons)
	}
}

// MessageBoxDefaultButton sets the native message box default action.
func MessageBoxDefaultButton(result MessageBoxResult) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.defaultButton = result
	}
}

// MessageBoxDefaultButtonID sets the default custom button by ID.
func MessageBoxDefaultButtonID(id string) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.defaultButtonID = id
	}
}

// MessageBoxCommandLinks displays custom buttons as command links when supported.
func MessageBoxCommandLinks(enabled bool) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.commandLinks = enabled
		opts.commandLinksNoIcon = false
	}
}

// MessageBoxCommandLinksNoIcon displays custom buttons as command links without icons when supported.
func MessageBoxCommandLinksNoIcon(enabled bool) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.commandLinksNoIcon = enabled
		if enabled {
			opts.commandLinks = false
		}
	}
}

// MessageBoxExpandedDetailsByDefault controls whether details start expanded.
func MessageBoxExpandedDetailsByDefault(enabled bool) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.expandedDetailsByDefault = enabled
	}
}

// MessageBoxExpandDetailsInFooterArea places expanded details in the footer area when supported.
func MessageBoxExpandDetailsInFooterArea(enabled bool) MessageBoxOption {
	return func(opts *messageBoxOptions) {
		opts.expandDetailsInFooterArea = enabled
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

	d, supported := currentDriverFor(CapabilityMessageBox)
	mb, ok := d.(messageBoxDriver)
	if !ok || !supported {
		return "", fmt.Errorf("system: %s: %w", CapabilityMessageBox, ErrUnsupported)
	}
	return mb.showMessageBox(ctx, opts)
}

// ShowMessageBoxDetailed opens a blocking native message box and returns rich result data.
func ShowMessageBoxDetailed(ctx context.Context, options ...MessageBoxOption) (MessageBoxDetailedResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}

	opts := defaultMessageBoxOptions()
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}

	d, supported := currentDriverFor(CapabilityMessageBox)
	if !supported {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: %s: %w", CapabilityMessageBox, ErrUnsupported)
	}
	if detailed, ok := d.(detailedMessageBoxDriver); ok {
		return detailed.showDetailedMessageBox(ctx, opts)
	}
	if mb, ok := d.(messageBoxDriver); ok {
		result, err := mb.showMessageBox(ctx, opts)
		return MessageBoxDetailedResult{
			Result:   result,
			ButtonID: string(result),
		}, err
	}
	return MessageBoxDetailedResult{}, fmt.Errorf("system: %s: %w", CapabilityMessageBox, ErrUnsupported)
}

// ShowMessageBoxAsync opens a native message box in a goroutine.
func ShowMessageBoxAsync(ctx context.Context, options ...MessageBoxOption) <-chan MessageBoxResponse {
	ch := make(chan MessageBoxResponse, 1)
	go func() {
		result, err := ShowMessageBox(ctx, options...)
		ch <- MessageBoxResponse{Result: result, Err: err}
	}()
	return ch
}

// ShowMessageBoxDetailedAsync opens a detailed native message box in a goroutine.
func ShowMessageBoxDetailedAsync(ctx context.Context, options ...MessageBoxOption) <-chan MessageBoxDetailedResponse {
	ch := make(chan MessageBoxDetailedResponse, 1)
	go func() {
		result, err := ShowMessageBoxDetailed(ctx, options...)
		ch <- MessageBoxDetailedResponse{Result: result, Err: err}
	}()
	return ch
}

func cloneMessageBoxButtons(buttons []MessageBoxButton) []MessageBoxButton {
	if len(buttons) == 0 {
		return nil
	}
	cloned := make([]MessageBoxButton, len(buttons))
	copy(cloned, buttons)
	return cloned
}
