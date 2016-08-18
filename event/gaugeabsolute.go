package event

import (
	"fmt"
	"time"
)

// GaugeAbsolute - Gauges are a constant data type. They are not subject to averaging,
// and they don’t change unless you change them. That is, once you set a gauge value,
// it will be a flat line on the graph until you change it again
type GaugeAbsolute struct {
	Name  string
	Value int64
}

func (e *GaugeAbsolute) StatClass() string {
	return "gauge"
}

// Update the event with metrics coming from a new one of the same type and with the same key
func (e *GaugeAbsolute) Update(e2 Event) error {
	if e.Type() != e2.Type() {
		return fmt.Errorf("statsd event type conflict: %s vs %s ", e.String(), e2.String())
	}
	//just RESET the value
	e.Value = e2.Payload().(int64)
	return nil
}

// Payload returns the aggregated value for this event
func (e GaugeAbsolute) Payload() interface{} {
	return e.Value
}

//Reset the value GAUGES TO NOT RESET
func (e *GaugeAbsolute) Reset() {

}

// Stats returns an array of StatsD events as they travel over UDP
func (e GaugeAbsolute) Stats(tick time.Duration) []string {
	if e.Value < 0 {
		// because a leading '+' or '-' in the value of a gauge denotes a delta, to send
		// a negative gauge value we first set the gauge absolutely to 0, then send the
		// negative value as a delta from 0 (that's just how the spec works :-)
		return []string{
			fmt.Sprintf("%s:%d|g", e.Name, 0),
			fmt.Sprintf("%s:%d|g", e.Name, e.Value),
		}
	}
	return []string{fmt.Sprintf("%s:%d|g", e.Name, e.Value)}
}

// Key returns the name of this metric
func (e GaugeAbsolute) Key() string {
	return e.Name
}

// SetKey sets the name of this metric
func (e *GaugeAbsolute) SetKey(key string) {
	e.Name = key
}

// Type returns an integer identifier for this type of metric
func (e GaugeAbsolute) Type() int {
	return EventGaugeAbsolute
}

// TypeString returns a name for this type of metric
func (e GaugeAbsolute) TypeString() string {
	return "GaugeAbsolute"
}

// String returns a debug-friendly representation of this metric
func (e GaugeAbsolute) String() string {
	return fmt.Sprintf("{Type: %s, Key: %s, Value: %d}", e.TypeString(), e.Name, e.Value)
}
