//go:build windows

package system

import "testing"

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
			config, err := windowsTaskDialogConfig(tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if config.size == 0 {
				t.Fatal("expected config size to be set")
			}
			if config.commonButtons != tt.buttons {
				t.Fatalf("expected buttons 0x%x, got 0x%x", tt.buttons, config.commonButtons)
			}
			if config.defaultButton != tt.defaultButton {
				t.Fatalf("expected default button %d, got %d", tt.defaultButton, config.defaultButton)
			}
			if config.mainIcon != tt.icon {
				t.Fatalf("expected icon 0x%x, got 0x%x", tt.icon, config.mainIcon)
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

func TestWindowsMessageBoxCloseAndEscapeMapToCancel(t *testing.T) {
	got, err := windowsMessageBoxResult(idCancel)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != MessageBoxResultCancel {
		t.Fatalf("expected Windows IDCANCEL to map to cancel, got %q", got)
	}
}
