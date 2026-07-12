package main

import (
	"testing"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func TestFieldStateKeepsHostPendingWithAsyncError(t *testing.T) {
	state := fieldState(
		"email",
		"邮箱",
		validationResult{errors: map[string]string{}},
		true,
		true,
		"宿主异步校验发现该邮箱已注册。",
	)
	if state.Status != ui.FieldInvalid {
		t.Fatalf("status = %v, want invalid", state.Status)
	}
	if !state.Pending {
		t.Fatal("host pending state was lost when an async error was supplied")
	}
	if state.ErrorText == "" || state.PendingText == "" {
		t.Fatalf("state = %#v, want both error and pending messages", state)
	}
}
