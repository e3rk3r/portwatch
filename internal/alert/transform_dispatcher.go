package alert

// TransformDispatcher wraps a downstream Dispatcher and applies a Transformer
// to each Notification before forwarding it. If the Transformer mutates the
// notification (e.g. re-titles or adds labels) the downstream always receives
// the enriched copy.
type TransformDispatcher struct {
	next        Dispatcher
	transformer *Transformer
}

// NewTransformDispatcher returns a TransformDispatcher that applies t before
// delegating to next. Both arguments must be non-nil.
func NewTransformDispatcher(next Dispatcher, t *Transformer) *TransformDispatcher {
	if next == nil {
		panic("alert: NewTransformDispatcher: next must not be nil")
	}
	if t == nil {
		panic("alert: NewTransformDispatcher: transformer must not be nil")
	}
	return &TransformDispatcher{next: next, transformer: t}
}

// Send transforms n and forwards the result to the wrapped Dispatcher.
func (d *TransformDispatcher) Send(n Notification) error {
	transformed, err := d.transformer.Apply(n)
	if err != nil {
		return err
	}
	return d.next.Send(transformed)
}
