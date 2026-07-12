// Package fieldstate defines the presentation-neutral field state contract
// shared by future form controls.
package fieldstate

// Status describes host-supplied validation progress. Components render this
// synchronous snapshot; they never start validation work during layout.
type Status uint8

const (
	// Valid means the host has no current validation error.
	Valid Status = iota
	// Invalid means ErrorText, when present, takes precedence over supporting
	// text.
	Invalid
	// Pending means validation is in progress and PendingText, when present,
	// takes precedence over supporting text.
	Pending
)

// State contains field presentation semantics. It owns neither a business
// value nor the validation operation that produced Status.
type State struct {
	Label          string
	SupportingText string
	ErrorText      string
	PendingText    string
	Required       bool
	Disabled       bool
	ReadOnly       bool
	Status         Status
}

// Normalized maps unknown statuses to Valid so a malformed host snapshot does
// not create an undocumented visual state.
func (s State) Normalized() State {
	switch s.Status {
	case Valid, Invalid, Pending:
		return s
	default:
		s.Status = Valid
		return s
	}
}

// IsInvalid reports whether the normalized state is invalid.
func (s State) IsInvalid() bool {
	return s.Normalized().Status == Invalid
}

// IsPending reports whether the normalized state is pending validation.
func (s State) IsPending() bool {
	return s.Normalized().Status == Pending
}

// Message returns the single supporting message that should be announced and
// displayed below a field. Error and pending states have explicit precedence.
func (s State) Message() string {
	s = s.Normalized()
	switch s.Status {
	case Invalid:
		if s.ErrorText != "" {
			return s.ErrorText
		}
	case Pending:
		if s.PendingText != "" {
			return s.PendingText
		}
	}
	return s.SupportingText
}
