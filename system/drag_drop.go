package system

import (
	"context"
	"fmt"
)

// DragDropOperation identifies the semantic operation an application associates
// with a drag-and-drop transfer.
type DragDropOperation string

const (
	DragDropOperationCopy DragDropOperation = "copy"
	DragDropOperationMove DragDropOperation = "move"
	DragDropOperationLink DragDropOperation = "link"
)

// DragAndDropProbe reports the active driver's drag-and-drop transfer surface.
//
// FluxUI's UI/widget drag-and-drop implementation is built on Gio transfer
// events. The probe describes whether that transfer path is expected to be
// usable on the current platform and which payload classes FluxUI normalizes.
type DragAndDropProbe struct {
	Status                  CapabilityStatus
	Err                     error
	SupportsDropTarget      bool
	SupportsDragSource      bool
	SupportsText            bool
	SupportsFiles           bool
	SupportsCustomMIME      bool
	SupportsExternalDragIn  bool
	SupportsExternalDragOut bool
	SupportedOperations     []DragDropOperation
}

// Supported reports whether drag-and-drop is implemented by the active driver.
func (p DragAndDropProbe) Supported() bool {
	return p.Status != CapabilityStatusUnsupported
}

// Available reports whether drag-and-drop appears usable right now.
func (p DragAndDropProbe) Available() bool {
	return p.Status == CapabilityStatusAvailable
}

type dragAndDropProbeDriver interface {
	probeDragAndDrop(ctx context.Context) DragAndDropProbe
}

// ProbeDragAndDrop returns a point-in-time probe for FluxUI drag-and-drop.
func ProbeDragAndDrop(ctx context.Context) DragAndDropProbe {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return DragAndDropProbe{
			Status: CapabilityStatusUnavailable,
			Err:    err,
		}
	}

	d, supported := currentDriverFor(CapabilityDragAndDrop)
	if !supported {
		return DragAndDropProbe{
			Status: CapabilityStatusUnsupported,
			Err:    fmt.Errorf("system: %s: %w", CapabilityDragAndDrop, ErrUnsupported),
		}
	}
	if pd, ok := d.(dragAndDropProbeDriver); ok {
		return normalizeDragAndDropProbe(pd.probeDragAndDrop(ctx))
	}
	return defaultDragAndDropProbe(CapabilityStatusAvailable, nil)
}

func defaultDragAndDropProbe(status CapabilityStatus, err error) DragAndDropProbe {
	probe := DragAndDropProbe{
		Status:                 status,
		Err:                    err,
		SupportsDropTarget:     status == CapabilityStatusAvailable,
		SupportsDragSource:     status == CapabilityStatusAvailable,
		SupportsText:           status == CapabilityStatusAvailable,
		SupportsFiles:          status == CapabilityStatusAvailable,
		SupportsCustomMIME:     status == CapabilityStatusAvailable,
		SupportsExternalDragIn: status == CapabilityStatusAvailable,
		SupportedOperations: []DragDropOperation{
			DragDropOperationCopy,
			DragDropOperationMove,
			DragDropOperationLink,
		},
	}
	if status == CapabilityStatusUnsupported && probe.Err == nil {
		probe.Err = ErrUnsupported
	}
	if status == CapabilityStatusUnavailable && probe.Err == nil {
		probe.Err = ErrUnavailable
	}
	if status != CapabilityStatusAvailable {
		probe.SupportsDropTarget = false
		probe.SupportsDragSource = false
		probe.SupportsText = false
		probe.SupportsFiles = false
		probe.SupportsCustomMIME = false
		probe.SupportsExternalDragIn = false
		probe.SupportsExternalDragOut = false
		probe.SupportedOperations = nil
	}
	return probe
}

func normalizeDragAndDropProbe(probe DragAndDropProbe) DragAndDropProbe {
	switch probe.Status {
	case CapabilityStatusAvailable:
		probe.Err = nil
	case CapabilityStatusUnavailable:
		if probe.Err == nil {
			probe.Err = ErrUnavailable
		}
	default:
		probe.Status = CapabilityStatusUnsupported
		if probe.Err == nil {
			probe.Err = ErrUnsupported
		}
	}
	if len(probe.SupportedOperations) == 0 && probe.Status == CapabilityStatusAvailable {
		probe.SupportedOperations = []DragDropOperation{DragDropOperationCopy}
	}
	probe.SupportedOperations = normalizeDragDropOperations(probe.SupportedOperations)
	if probe.Status != CapabilityStatusAvailable {
		probe.SupportsDropTarget = false
		probe.SupportsDragSource = false
		probe.SupportsText = false
		probe.SupportsFiles = false
		probe.SupportsCustomMIME = false
		probe.SupportsExternalDragIn = false
		probe.SupportsExternalDragOut = false
		probe.SupportedOperations = nil
	}
	return probe
}

func normalizeDragDropOperations(operations []DragDropOperation) []DragDropOperation {
	if len(operations) == 0 {
		return nil
	}
	seen := map[DragDropOperation]bool{}
	normalized := make([]DragDropOperation, 0, len(operations))
	for _, op := range operations {
		switch op {
		case "", DragDropOperationCopy:
			op = DragDropOperationCopy
		case DragDropOperationMove, DragDropOperationLink:
		default:
			continue
		}
		if seen[op] {
			continue
		}
		seen[op] = true
		normalized = append(normalized, op)
	}
	return normalized
}
