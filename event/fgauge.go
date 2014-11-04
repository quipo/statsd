package event

import "fmt"

// Gauge - Gauges are a constant data type. They are not subject to averaging,
// and they don’t change unless you change them. That is, once you set a gauge value,
// it will be a flat line on the graph until you change it again
type FGauge struct {
	Name  string
	Value float64
}

// Update the event with metrics coming from a new one of the same type and with the same key
func (e *FGauge) Update(e2 Event) error {
	if e.Type() != e2.Type() {
		return fmt.Errorf("statsd event type conflict: %s vs %s ", e.String(), e2.String())
	}
	e.Value += e2.Payload().(float64)
	return nil
}

// Payload returns the aggregated value for this event
func (e FGauge) Payload() interface{} {
	return e.Value
}

// Stats returns an array of StatsD events as they travel over UDP
func (e FGauge) Stats() []string {
	return []string{fmt.Sprintf("%s:%f|t", e.Name, e.Value)}
}

// Key returns the name of this metric
func (e FGauge) Key() string {
	return e.Name
}

// Type returns an integer identifier for this type of metric
func (e FGauge) Type() int {
	return EventGauge
}

// TypeString returns a name for this type of metric
func (e FGauge) TypeString() string {
	return "FGauge"
}

// String returns a debug-friendly representation of this metric
func (e FGauge) String() string {
	return fmt.Sprintf("{Type: %s, Key: %s, Value: %f}", e.TypeString(), e.Name, e.Value)
}
