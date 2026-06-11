//go:build windows

package system

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	tdcbfOKButton     = 0x0001
	tdcbfYesButton    = 0x0002
	tdcbfNoButton     = 0x0004
	tdcbfCancelButton = 0x0008
	tdcbfRetryButton  = 0x0010

	tdfAllowDialogCancellation = 0x0008
	tdfUseCommandLinks         = 0x0010
	tdfUseCommandLinksNoIcon   = 0x0020
	tdfExpandFooterArea        = 0x0040
	tdfExpandedByDefault       = 0x0080
	tdfVerificationFlagChecked = 0x0100

	iccStandardClasses = 0x00004000

	tdnCreated       = 0
	tdnButtonClicked = 2

	tdWarningIcon     = 0xffff
	tdErrorIcon       = 0xfffe
	tdInformationIcon = 0xfffd

	taskDialogCustomButtonBase = 1000

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

	mbWMClose     = 0x0010
	mbWMKeyDown   = 0x0100
	mbWMNCDestroy = 0x0082
	mbVKEscape    = 0x1B

	gwOwner = 4

	idOK     = 1
	idCancel = 2
	idRetry  = 4
	idYes    = 6
	idNo     = 7

	taskDialogCloseReasonNone   = 0
	taskDialogCloseReasonClose  = 1
	taskDialogCloseReasonEscape = 2

	taskDialogCollapsedControlText = "Show details"
	taskDialogExpandedControlText  = "Hide details"
)

var (
	comctl32               = windows.NewLazySystemDLL("comctl32.dll")
	messageBoxKernel32     = windows.NewLazySystemDLL("kernel32.dll")
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procTaskDialog         = comctl32.NewProc("TaskDialog")
	procTaskDialogIndirect = comctl32.NewProc("TaskDialogIndirect")
	procInitCommonControls = comctl32.NewProc("InitCommonControlsEx")
	procMessageBox         = user32.NewProc("MessageBoxW")
	procMessageBoxPost     = user32.NewProc("PostMessageW")
	procMessageBoxSend     = user32.NewProc("SendMessageW")
	procTaskDialogCallProc = user32.NewProc("CallWindowProcW")
	procTaskDialogDefProc  = user32.NewProc("DefWindowProcW")
	procTaskDialogSetProc  = user32.NewProc("SetWindowLongPtrW")
	procEnumThreadWindows  = user32.NewProc("EnumThreadWindows")
	procGetClassName       = user32.NewProc("GetClassNameW")
	procGetCurrentThreadID = messageBoxKernel32.NewProc("GetCurrentThreadId")
	procGetWindow          = user32.NewProc("GetWindow")
	procGetWindowText      = user32.NewProc("GetWindowTextW")

	windowsTaskDialogCallback   = syscall.NewCallback(windowsTaskDialogCallbackProc)
	windowsTaskDialogWndProc    = syscall.NewCallback(windowsTaskDialogWindowProc)
	windowsTaskDialogCallbackMu sync.Mutex
	windowsTaskDialogCallbackID uint64
	windowsTaskDialogCallbacks  = make(map[uint64]*taskDialogCallbackState)
	windowsTaskDialogSubclassMu sync.RWMutex
	windowsTaskDialogSubclasses = make(map[uintptr]*taskDialogCallbackState)
	windowsModalDialogEnumProc  = syscall.NewCallback(windowsModalDialogEnumCallback)
	windowsModalDialogSearchMu  sync.Mutex
	windowsModalDialogSearchID  uintptr
	windowsModalDialogSearches  = make(map[uintptr]*windowsModalDialogSearch)
)

type taskDialogConfig struct {
	size                 uint32
	parent               packedUintptr
	instance             packedUintptr
	flags                uint32
	commonButtons        uint32
	windowTitle          packedUintptr
	mainIcon             packedUintptr
	mainInstruction      packedUintptr
	content              packedUintptr
	buttonCount          uint32
	buttons              packedUintptr
	defaultButton        int32
	radioButtonCount     uint32
	radioButtons         packedUintptr
	defaultRadioButton   int32
	verificationText     packedUintptr
	expandedInformation  packedUintptr
	expandedControlText  packedUintptr
	collapsedControlText packedUintptr
	footerIcon           packedUintptr
	footer               packedUintptr
	callback             packedUintptr
	callbackData         packedUintptr
	width                uint32
}

type packedUintptr [int(unsafe.Sizeof(uintptr(0)))]byte

func (p *packedUintptr) set(value uintptr) {
	for i := range p {
		p[i] = byte(value >> (uint(i) * 8))
	}
}

func (p packedUintptr) value() uintptr {
	value := uintptr(0)
	for i := range p {
		value |= uintptr(p[i]) << (uint(i) * 8)
	}
	return value
}

type initCommonControlsEx struct {
	size uint32
	icc  uint32
}

type taskDialogButton struct {
	id   int32
	text packedUintptr
}

type windowsTaskDialogConfigData struct {
	config       taskDialogConfig
	buttons      []taskDialogButton
	buttonText   [][]uint16
	customByID   map[int32]MessageBoxButton
	stringFields [][]uint16
}

type taskDialogCallbackState struct {
	clickedButton int32
	closeReason   int32
	hwnd          atomic.Uintptr
	oldWndProc    atomic.Uintptr
}

type windowsModalDialogSearch struct {
	owner uintptr
	title string
	found uintptr
}

func (windowsDriver) showMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxResult, error) {
	result, err := windowsDriver{}.showDetailedMessageBox(ctx, opts)
	if err != nil {
		return "", err
	}
	return result.Result, nil
}

func (windowsDriver) showDetailedMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxDetailedResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}

	result, err := showWindowsTaskDialogIndirect(ctx, opts)
	if err == nil {
		return result, nil
	}
	if opts.requiresTaskDialogIndirect() {
		return MessageBoxDetailedResult{}, err
	}
	if IsUnavailable(err) {
		result, err := showWindowsTaskDialog(ctx, opts)
		if err == nil {
			return result, nil
		}
		if IsUnavailable(err) {
			return showWindowsMessageBox(ctx, opts)
		}
	}
	return MessageBoxDetailedResult{}, err
}

func showWindowsTaskDialogIndirect(ctx context.Context, opts messageBoxOptions) (MessageBoxDetailedResult, error) {
	if err := procTaskDialogIndirect.Find(); err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: task dialog indirect unavailable: %w", ErrUnavailable)
	}
	if err := initWindowsCommonControls(); err != nil {
		return MessageBoxDetailedResult{}, err
	}

	data, err := windowsTaskDialogConfig(opts)
	if err != nil {
		return MessageBoxDetailedResult{}, err
	}
	callbackID, callbackState := registerWindowsTaskDialogCallback()
	defer unregisterWindowsTaskDialogCallback(callbackID)
	data.config.callback.set(windowsTaskDialogCallback)
	data.config.callbackData.set(uintptr(callbackID))
	config := data.config

	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}

	done := make(chan struct{})
	defer close(done)
	watchWindowsTaskDialogContext(ctx, callbackState, done)

	var button int32
	var verificationChecked int32
	var verificationCheckedPtr uintptr
	if data.config.verificationText.value() != 0 {
		verificationCheckedPtr = uintptr(unsafe.Pointer(&verificationChecked))
	}
	r, _, callErr := procTaskDialogIndirect.Call(
		uintptr(unsafe.Pointer(&config)),
		uintptr(unsafe.Pointer(&button)),
		0,
		verificationCheckedPtr,
	)
	keepAliveWindowsTaskDialogData(data)
	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}
	if err := hresultError(r); err != nil {
		if callErr != syscall.Errno(0) {
			return MessageBoxDetailedResult{}, fmt.Errorf("system: show task dialog indirect: %w: %v", ErrUnavailable, callErr)
		}
		return MessageBoxDetailedResult{}, fmt.Errorf("system: show task dialog indirect: %w: %v", ErrUnavailable, err)
	}

	clickedButton := atomic.LoadInt32(&callbackState.clickedButton)
	closeReason := atomic.LoadInt32(&callbackState.closeReason)
	result, err := windowsMessageBoxDetailedResult(button, verificationChecked != 0, clickedButton, closeReason, data)
	if err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: show task dialog indirect: %w", err)
	}
	return result, nil
}

func showWindowsTaskDialog(ctx context.Context, opts messageBoxOptions) (MessageBoxDetailedResult, error) {
	if err := procTaskDialog.Find(); err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: task dialog unavailable: %w", ErrUnavailable)
	}
	if err := initWindowsCommonControls(); err != nil {
		return MessageBoxDetailedResult{}, err
	}

	commonButtons, err := windowsTaskDialogCommonButtons(opts.buttons)
	if err != nil {
		return MessageBoxDetailedResult{}, err
	}
	icon, err := windowsTaskDialogIcon(opts.kind)
	if err != nil {
		return MessageBoxDetailedResult{}, err
	}

	defaultButton := opts.defaultButton
	if defaultButton == "" {
		defaultButton = firstMessageBoxButton(opts.buttons)
	}
	if _, ok := messageBoxButtonID(opts.buttons, defaultButton); !ok {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: default button %q is not in button set %q", defaultButton, opts.buttons)
	}

	title16, err := windows.UTF16PtrFromString(opts.title)
	if err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: configure task dialog title: %w", err)
	}
	text16, err := windows.UTF16PtrFromString(opts.text)
	if err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: configure task dialog text: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	done := make(chan struct{})
	defer close(done)
	watchWindowsModalContext(ctx, uint32(currentWindowsThreadID()), opts.owner, opts.title, done)

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
	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}
	if err := hresultError(r); err != nil {
		if callErr != syscall.Errno(0) {
			return MessageBoxDetailedResult{}, fmt.Errorf("system: show task dialog: %w: %v", ErrUnavailable, callErr)
		}
		return MessageBoxDetailedResult{}, fmt.Errorf("system: show task dialog: %w: %v", ErrUnavailable, err)
	}

	result, err := windowsMessageBoxResult(uintptr(button))
	if err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: show task dialog: %w", err)
	}
	return MessageBoxDetailedResult{Result: result, ButtonID: string(result)}, nil
}

func showWindowsMessageBox(ctx context.Context, opts messageBoxOptions) (MessageBoxDetailedResult, error) {
	flags, err := windowsMessageBoxFlags(opts)
	if err != nil {
		return MessageBoxDetailedResult{}, err
	}

	title16, err := windows.UTF16PtrFromString(opts.title)
	if err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: configure message box title: %w", err)
	}
	text16, err := windows.UTF16PtrFromString(opts.text)
	if err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: configure message box text: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	done := make(chan struct{})
	defer close(done)
	watchWindowsModalContext(ctx, uint32(currentWindowsThreadID()), opts.owner, opts.title, done)

	r, _, callErr := procMessageBox.Call(
		opts.owner,
		uintptr(unsafe.Pointer(text16)),
		uintptr(unsafe.Pointer(title16)),
		uintptr(flags),
	)
	if err := ctx.Err(); err != nil {
		return MessageBoxDetailedResult{}, err
	}
	if r == 0 {
		if callErr != syscall.Errno(0) {
			return MessageBoxDetailedResult{}, fmt.Errorf("system: show message box: %w", callErr)
		}
		return MessageBoxDetailedResult{}, fmt.Errorf("system: show message box: %w", ErrUnavailable)
	}

	result, err := windowsMessageBoxResult(r)
	if err != nil {
		return MessageBoxDetailedResult{}, fmt.Errorf("system: show message box: %w", err)
	}
	return MessageBoxDetailedResult{Result: result, ButtonID: string(result)}, nil
}

func windowsTaskDialogConfig(opts messageBoxOptions) (windowsTaskDialogConfigData, error) {
	icon, err := windowsTaskDialogIcon(opts.kind)
	if err != nil {
		return windowsTaskDialogConfigData{}, err
	}

	data := windowsTaskDialogConfigData{}
	title, err := data.addString(opts.title)
	if err != nil {
		return windowsTaskDialogConfigData{}, fmt.Errorf("system: configure task dialog title: %w", err)
	}
	text, err := data.addString(opts.text)
	if err != nil {
		return windowsTaskDialogConfigData{}, fmt.Errorf("system: configure task dialog text: %w", err)
	}

	config := taskDialogConfig{
		size: uint32(unsafe.Sizeof(taskDialogConfig{})),
	}
	config.parent.set(opts.owner)
	config.windowTitle.set(title)
	config.mainIcon.set(icon)
	config.mainInstruction.set(text)
	if opts.details != "" {
		details, err := data.addString(opts.details)
		if err != nil {
			return windowsTaskDialogConfigData{}, fmt.Errorf("system: configure task dialog details: %w", err)
		}
		collapsedText, err := data.addString(taskDialogCollapsedControlText)
		if err != nil {
			return windowsTaskDialogConfigData{}, fmt.Errorf("system: configure task dialog collapsed details text: %w", err)
		}
		expandedText, err := data.addString(taskDialogExpandedControlText)
		if err != nil {
			return windowsTaskDialogConfigData{}, fmt.Errorf("system: configure task dialog expanded details text: %w", err)
		}
		config.expandedInformation.set(details)
		config.collapsedControlText.set(collapsedText)
		config.expandedControlText.set(expandedText)
	}
	if opts.footer != "" {
		footer, err := data.addString(opts.footer)
		if err != nil {
			return windowsTaskDialogConfigData{}, fmt.Errorf("system: configure task dialog footer: %w", err)
		}
		config.footer.set(footer)
	}
	if opts.verificationText != "" {
		verification, err := data.addString(opts.verificationText)
		if err != nil {
			return windowsTaskDialogConfigData{}, fmt.Errorf("system: configure task dialog verification: %w", err)
		}
		config.verificationText.set(verification)
	}
	if opts.verificationChecked {
		config.flags |= tdfVerificationFlagChecked
	}
	if opts.commandLinks {
		config.flags |= tdfUseCommandLinks
	}
	if opts.commandLinksNoIcon {
		config.flags |= tdfUseCommandLinksNoIcon
	}
	if opts.expandedDetailsByDefault {
		config.flags |= tdfExpandedByDefault
	}
	if opts.expandDetailsInFooterArea {
		config.flags |= tdfExpandFooterArea
	}

	if len(opts.customButtons) > 0 {
		buttons, customByID, err := windowsTaskDialogButtons(opts.customButtons)
		if err != nil {
			return windowsTaskDialogConfigData{}, err
		}
		data.buttons = buttons.buttons
		data.buttonText = buttons.text
		data.customByID = customByID
		config.buttonCount = uint32(len(data.buttons))
		config.buttons.set(uintptr(unsafe.Pointer(&data.buttons[0])))
		defaultID, err := windowsTaskDialogDefaultCustomButtonID(opts, data.customByID)
		if err != nil {
			return windowsTaskDialogConfigData{}, err
		}
		config.defaultButton = defaultID
		config.flags |= tdfAllowDialogCancellation
		data.config = config
		return data, nil
	}

	commonButtons, err := windowsTaskDialogCommonButtons(opts.buttons)
	if err != nil {
		return windowsTaskDialogConfigData{}, err
	}
	defaultButton := opts.defaultButton
	if defaultButton == "" {
		defaultButton = firstMessageBoxButton(opts.buttons)
	}
	defaultID, ok := messageBoxButtonID(opts.buttons, defaultButton)
	if !ok {
		return windowsTaskDialogConfigData{}, fmt.Errorf("system: default button %q is not in button set %q", defaultButton, opts.buttons)
	}
	config.commonButtons = commonButtons
	config.defaultButton = int32(defaultID)
	data.config = config
	return data, nil
}

func (data *windowsTaskDialogConfigData) addString(value string) (uintptr, error) {
	text16, err := windows.UTF16FromString(value)
	if err != nil {
		return 0, err
	}
	data.stringFields = append(data.stringFields, text16)
	return uintptr(unsafe.Pointer(&data.stringFields[len(data.stringFields)-1][0])), nil
}

func initWindowsCommonControls() error {
	data := initCommonControlsEx{
		size: uint32(unsafe.Sizeof(initCommonControlsEx{})),
		icc:  iccStandardClasses,
	}
	r, _, callErr := procInitCommonControls.Call(uintptr(unsafe.Pointer(&data)))
	if r == 0 {
		if callErr != syscall.Errno(0) {
			return fmt.Errorf("system: init common controls: %w: %v", ErrUnavailable, callErr)
		}
		return fmt.Errorf("system: init common controls: %w", ErrUnavailable)
	}
	return nil
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

type windowsTaskDialogButtonsData struct {
	buttons []taskDialogButton
	text    [][]uint16
}

func windowsTaskDialogButtons(buttons []MessageBoxButton) (windowsTaskDialogButtonsData, map[int32]MessageBoxButton, error) {
	if len(buttons) == 0 {
		return windowsTaskDialogButtonsData{}, nil, nil
	}
	data := windowsTaskDialogButtonsData{
		buttons: make([]taskDialogButton, 0, len(buttons)),
		text:    make([][]uint16, 0, len(buttons)),
	}
	seen := make(map[string]bool, len(buttons))
	customByID := make(map[int32]MessageBoxButton, len(buttons))
	for i, button := range buttons {
		if button.ID == "" {
			return windowsTaskDialogButtonsData{}, nil, fmt.Errorf("system: custom message box button %d has empty ID", i)
		}
		if button.Label == "" {
			return windowsTaskDialogButtonsData{}, nil, fmt.Errorf("system: custom message box button %q has empty label", button.ID)
		}
		if seen[button.ID] {
			return windowsTaskDialogButtonsData{}, nil, fmt.Errorf("system: duplicate custom message box button ID %q", button.ID)
		}
		seen[button.ID] = true

		label16, err := windows.UTF16FromString(button.Label)
		if err != nil {
			return windowsTaskDialogButtonsData{}, nil, fmt.Errorf("system: configure custom message box button %q: %w", button.ID, err)
		}
		id := int32(taskDialogCustomButtonBase + i)
		data.text = append(data.text, label16)
		data.buttons = append(data.buttons, taskDialogButton{id: id})
		data.buttons[len(data.buttons)-1].text.set(uintptr(unsafe.Pointer(&data.text[len(data.text)-1][0])))
		if button.Result == "" {
			button.Result = MessageBoxResultCustom
		}
		customByID[id] = button
	}
	return data, customByID, nil
}

func windowsTaskDialogDefaultCustomButtonID(opts messageBoxOptions, customByID map[int32]MessageBoxButton) (int32, error) {
	if opts.defaultButtonID != "" {
		for id, button := range customByID {
			if button.ID == opts.defaultButtonID {
				return id, nil
			}
		}
		return 0, fmt.Errorf("system: default custom button ID %q is not in custom button set", opts.defaultButtonID)
	}
	if opts.defaultButton != "" {
		for id, button := range customByID {
			if button.Result == opts.defaultButton {
				return id, nil
			}
		}
	}
	for id := int32(taskDialogCustomButtonBase); id < int32(taskDialogCustomButtonBase+len(customByID)); id++ {
		if _, ok := customByID[id]; ok {
			return id, nil
		}
	}
	return 0, fmt.Errorf("system: custom message box buttons are empty")
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

func windowsMessageBoxDetailedResult(id int32, verificationChecked bool, clickedButton int32, closeReason int32, data windowsTaskDialogConfigData) (MessageBoxDetailedResult, error) {
	if button, ok := data.customByID[id]; ok {
		result := button.Result
		if result == "" {
			result = MessageBoxResultCustom
		}
		return MessageBoxDetailedResult{
			Result:              result,
			ButtonID:            button.ID,
			VerificationChecked: verificationChecked,
		}, nil
	}

	if id == idCancel {
		result := MessageBoxResult("")
		switch closeReason {
		case taskDialogCloseReasonEscape:
			result = MessageBoxResultEscape
		case taskDialogCloseReasonClose:
			result = MessageBoxResultClose
		default:
			if clickedButton != idCancel {
				result = MessageBoxResultClose
			}
		}
		if result == "" {
			result = MessageBoxResultCancel
		}
		return MessageBoxDetailedResult{
			Result:              result,
			ButtonID:            string(result),
			VerificationChecked: verificationChecked,
		}, nil
	}

	result, err := windowsMessageBoxResult(uintptr(id))
	if err != nil {
		return MessageBoxDetailedResult{}, err
	}
	return MessageBoxDetailedResult{
		Result:              result,
		ButtonID:            string(result),
		VerificationChecked: verificationChecked,
	}, nil
}

func registerWindowsTaskDialogCallback() (uint64, *taskDialogCallbackState) {
	state := &taskDialogCallbackState{}
	windowsTaskDialogCallbackMu.Lock()
	windowsTaskDialogCallbackID++
	id := windowsTaskDialogCallbackID
	windowsTaskDialogCallbacks[id] = state
	windowsTaskDialogCallbackMu.Unlock()
	return id, state
}

func unregisterWindowsTaskDialogCallback(id uint64) {
	windowsTaskDialogCallbackMu.Lock()
	delete(windowsTaskDialogCallbacks, id)
	windowsTaskDialogCallbackMu.Unlock()
}

func lookupWindowsTaskDialogCallback(id uint64) *taskDialogCallbackState {
	windowsTaskDialogCallbackMu.Lock()
	state := windowsTaskDialogCallbacks[id]
	windowsTaskDialogCallbackMu.Unlock()
	return state
}

func windowsTaskDialogCallbackProc(hwnd uintptr, msg uintptr, wParam uintptr, _ uintptr, callbackData uintptr) uintptr {
	if callbackData == 0 {
		return 0
	}
	state := lookupWindowsTaskDialogCallback(uint64(callbackData))
	if state == nil {
		return 0
	}
	switch uint32(msg) {
	case tdnCreated:
		state.hwnd.Store(hwnd)
		installWindowsTaskDialogSubclass(hwnd, state)
	case tdnButtonClicked:
		atomic.StoreInt32(&state.clickedButton, int32(wParam))
	}
	return 0
}

func installWindowsTaskDialogSubclass(hwnd uintptr, state *taskDialogCallbackState) bool {
	if hwnd == 0 || state == nil {
		return false
	}
	oldProc, _, err := procTaskDialogSetProc.Call(hwnd, ^uintptr(3), windowsTaskDialogWndProc)
	if oldProc == 0 && err != syscall.Errno(0) {
		return false
	}
	state.oldWndProc.Store(oldProc)

	windowsTaskDialogSubclassMu.Lock()
	windowsTaskDialogSubclasses[hwnd] = state
	windowsTaskDialogSubclassMu.Unlock()
	return true
}

func lookupWindowsTaskDialogSubclass(hwnd uintptr) *taskDialogCallbackState {
	windowsTaskDialogSubclassMu.RLock()
	state := windowsTaskDialogSubclasses[hwnd]
	windowsTaskDialogSubclassMu.RUnlock()
	return state
}

func forgetWindowsTaskDialogSubclass(hwnd uintptr, state *taskDialogCallbackState) {
	windowsTaskDialogSubclassMu.Lock()
	if current := windowsTaskDialogSubclasses[hwnd]; current == state {
		delete(windowsTaskDialogSubclasses, hwnd)
	}
	windowsTaskDialogSubclassMu.Unlock()
}

func windowsTaskDialogWindowProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	state := lookupWindowsTaskDialogSubclass(hwnd)
	oldProc := uintptr(0)
	if state != nil {
		oldProc = state.oldWndProc.Load()
		switch msg {
		case mbWMKeyDown:
			if wParam == mbVKEscape {
				markWindowsTaskDialogCloseReason(state, taskDialogCloseReasonEscape)
			}
		case mbWMClose:
			markWindowsTaskDialogCloseReason(state, taskDialogCloseReasonClose)
		case mbWMNCDestroy:
			forgetWindowsTaskDialogSubclass(hwnd, state)
			if oldProc != 0 {
				procTaskDialogSetProc.Call(hwnd, ^uintptr(3), oldProc)
			}
		}
	}

	if oldProc != 0 {
		result, _, _ := procTaskDialogCallProc.Call(oldProc, hwnd, msg, wParam, lParam)
		return result
	}
	result, _, _ := procTaskDialogDefProc.Call(hwnd, msg, wParam, lParam)
	return result
}

func markWindowsTaskDialogCloseReason(state *taskDialogCallbackState, reason int32) {
	if state == nil || reason == taskDialogCloseReasonNone {
		return
	}
	atomic.CompareAndSwapInt32(&state.closeReason, taskDialogCloseReasonNone, reason)
}

func watchWindowsTaskDialogContext(ctx context.Context, state *taskDialogCallbackState, done <-chan struct{}) {
	if ctx == nil || state == nil {
		return
	}
	go func() {
		select {
		case <-ctx.Done():
		case <-done:
			return
		}

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			if hwnd := state.hwnd.Load(); hwnd != 0 {
				procMessageBoxSend.Call(hwnd, mbWMClose, 0, 0)
				return
			}
			select {
			case <-done:
				return
			case <-ticker.C:
			}
		}
	}()
}

func watchWindowsModalContext(ctx context.Context, threadID uint32, owner uintptr, title string, done <-chan struct{}) {
	if ctx == nil || threadID == 0 {
		return
	}
	go func() {
		select {
		case <-ctx.Done():
		case <-done:
			return
		}

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			if hwnd := findWindowsModalDialog(threadID, owner, title); hwnd != 0 {
				procMessageBoxSend.Call(hwnd, mbWMClose, 0, 0)
				return
			}
			select {
			case <-done:
				return
			case <-ticker.C:
			}
		}
	}()
}

func findWindowsModalDialog(threadID uint32, owner uintptr, title string) uintptr {
	if threadID == 0 {
		return 0
	}
	search := &windowsModalDialogSearch{owner: owner, title: title}
	searchID := registerWindowsModalDialogSearch(search)
	defer unregisterWindowsModalDialogSearch(searchID)
	procEnumThreadWindows.Call(uintptr(threadID), windowsModalDialogEnumProc, searchID)
	return search.found
}

func registerWindowsModalDialogSearch(search *windowsModalDialogSearch) uintptr {
	if search == nil {
		return 0
	}
	windowsModalDialogSearchMu.Lock()
	windowsModalDialogSearchID++
	if windowsModalDialogSearchID == 0 {
		windowsModalDialogSearchID = 1
	}
	id := windowsModalDialogSearchID
	windowsModalDialogSearches[id] = search
	windowsModalDialogSearchMu.Unlock()
	return id
}

func unregisterWindowsModalDialogSearch(id uintptr) {
	if id == 0 {
		return
	}
	windowsModalDialogSearchMu.Lock()
	delete(windowsModalDialogSearches, id)
	windowsModalDialogSearchMu.Unlock()
}

func lookupWindowsModalDialogSearch(id uintptr) *windowsModalDialogSearch {
	if id == 0 {
		return nil
	}
	windowsModalDialogSearchMu.Lock()
	search := windowsModalDialogSearches[id]
	windowsModalDialogSearchMu.Unlock()
	return search
}

func windowsModalDialogEnumCallback(hwnd uintptr, lParam uintptr) uintptr {
	state := lookupWindowsModalDialogSearch(lParam)
	if state == nil {
		return 0
	}
	if windowsModalDialogMatches(hwnd, state.owner, state.title) {
		state.found = hwnd
		return 0
	}
	return 1
}

func windowsModalDialogMatches(hwnd uintptr, owner uintptr, title string) bool {
	if hwnd == 0 || !windowsModalDialogClassLooksNative(windowsWindowClass(hwnd)) {
		return false
	}
	if owner != 0 && windowsWindowOwner(hwnd) != owner {
		return false
	}
	if title != "" && windowsWindowText(hwnd) != title {
		return false
	}
	return true
}

func windowsModalDialogClassLooksNative(className string) bool {
	switch className {
	case "#32770", "TaskDialog":
		return true
	default:
		return false
	}
}

func windowsWindowOwner(hwnd uintptr) uintptr {
	owner, _, _ := procGetWindow.Call(hwnd, gwOwner)
	return owner
}

func windowsWindowClass(hwnd uintptr) string {
	var buf [128]uint16
	n, _, _ := procGetClassName.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:n])
}

func windowsWindowText(hwnd uintptr) string {
	var buf [256]uint16
	n, _, _ := procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:n])
}

func currentWindowsThreadID() uintptr {
	threadID, _, _ := procGetCurrentThreadID.Call()
	return threadID
}

func keepAliveWindowsTaskDialogData(data windowsTaskDialogConfigData) {
	runtime.KeepAlive(data)
}

func (opts messageBoxOptions) requiresTaskDialogIndirect() bool {
	return opts.details != "" ||
		opts.footer != "" ||
		opts.verificationText != "" ||
		len(opts.customButtons) > 0 ||
		opts.commandLinks ||
		opts.commandLinksNoIcon ||
		opts.expandedDetailsByDefault ||
		opts.expandDetailsInFooterArea
}
