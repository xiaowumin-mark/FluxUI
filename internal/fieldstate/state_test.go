package fieldstate

import "testing"

func TestMessagePrecedenceAndStatusNormalization(t *testing.T) {
	base := State{SupportingText: "helper", ErrorText: "invalid", PendingText: "checking"}
	if got := base.Message(); got != "helper" {
		t.Fatalf("valid message = %q, want helper", got)
	}
	if got := (State{SupportingText: base.SupportingText, ErrorText: base.ErrorText, Status: Invalid}).Message(); got != "invalid" {
		t.Fatalf("invalid message = %q, want invalid", got)
	}
	if got := (State{SupportingText: base.SupportingText, PendingText: base.PendingText, Status: Pending}).Message(); got != "checking" {
		t.Fatalf("pending message = %q, want checking", got)
	}
	unknown := State{SupportingText: "helper", Status: Status(99)}
	if normalized := unknown.Normalized(); normalized.Status != Valid || normalized.Message() != "helper" {
		t.Fatalf("unknown status normalized to %#v", normalized)
	}
}
