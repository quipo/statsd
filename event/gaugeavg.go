package event

import (
	"fmt"
	"sync/atomic"
	"time"
)

// GaugeAvg - Since this is buffered, we keep a list of the gauge values
// sent, then at send time average them for the "gauge" value to statsd
type GaugeAvg struct {
	Name  string
	Value int64
	Times int64
}

func (e *GaugeAvg) StatClass() string {
	return "gauge"
}

// Update the event with metrics coming from a new one of the same type and with the same key
func (e *GaugeAvg) Update(e2 Event) error {
	if e.Type() != e2.Type() {
		return fmt.Errorf("statsd event type conflict: %s vs %s ", e.String(), e2.String())
	}
	//add it to the current
	if e.Value == -1 {
		e.Value = 0
	}
	atomic.AddInt64(&e.Value, e2.Payload().(int64))
	atomic.AddInt64(&e.Times, 1)
	return nil
}

// Payload returns the aggregated value for this event
func (e GaugeAvg) Payload() interface{} {
	return e.Value
}

//this gauge
func (e *GaugeAvg) Reset() {
	e.Times = 1
	e.Value = -1
}

// Stats returns an array of StatsD events as they travel over UDP
func (e GaugeAvg) Stats(tick time.Duration) []string {

	//return NOTHING if value is -1
	if e.Value == -1 {
		return []string{}
	}
	true_val := e.Value
	if e.Times > 0 {
		true_val = e.Value / e.Times
	}
	if true_val < 0 {
		// because a leading '+' or '-' in the value of a gauge denotes a delta, to send
		// a negative gauge value we first set the gauge absolutely to 0, then send the
		// negative value as a delta from 0 (that's just how the spec works :-)
		return []string{
			fmt.Sprintf("%s:%d|g", e.Name, 0),
			fmt.Sprintf("%s:%d|g", e.Name, true_val),
		}
	}
	return []string{fmt.Sprintf("%s:%d|g", e.Name, true_val)}
}

// Key returns the name of this metric
func (e GaugeAvg) Key() string {
	return e.Name
}

// SetKey sets the name of this metric
func (e *GaugeAvg) SetKey(key string) {
	e.Name = key
}

// Type returns an integer identifier for this type of metric
func (e GaugeAvg) Type() int {
	return EventGaugeAvg
}

// TypeString returns a name for this type of metric
func (e GaugeAvg) TypeString() string {
	return "Gauge"
}

// String returns a debug-friendly representation of this metric
func (e GaugeAvg) String() string {
	return fmt.Sprintf("{Type: %s, Key: %s, Value: %d}", e.TypeString(), e.Name, e.Value)
}
