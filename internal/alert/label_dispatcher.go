package alert

import "context"

// NewLabelDispatcher wraps next with a Labeler that applies p's labels before
// forwarding every notification. Panics if next or labeler is nil.
func NewLabelDispatcher(labeler *Labeler, next Dispatcher) Dispatcher {
	if labeler == nil {
		panic("alert: NewLabelDispatcher: labeler must not be nil")
	}
	if next == nil {
		panic("alert: NewLabelDispatcher: next must not be nil")
	}
	return &labelDispatcher{labeler: labeler, next: next}
}

type labelDispatcher struct {
	labeler *Labeler
	next    Dispatcher
}

func (d *labelDispatcher) Send(ctx context.Context, n Notification) error {
	copy := n
	if copy.Labels != nil {
		cloned := make(map[string]string, len(copy.Labels))
		for k, v := range copy.Labels {
			cloned[k] = v
		}
		copy.Labels = cloned
	}
	d.labeler.Apply(&copy)
	return d.next.Send(ctx, copy)
}
