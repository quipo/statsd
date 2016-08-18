package statsd

import (
	"time"
)

//impliment a "noop" statsd in case there is no statsd
type StatsdNoop struct{}

func (s StatsdNoop) String() string                                         { return "NoopClient" }
func (s StatsdNoop) CreateSocket() error                                    { return nil }
func (s StatsdNoop) Close() error                                           { return nil }
func (s StatsdNoop) Incr(stat string, count int64) error                    { return nil }
func (s StatsdNoop) Decr(stat string, count int64) error                    { return nil }
func (s StatsdNoop) Timing(stat string, count int64) error                  { return nil }
func (s StatsdNoop) PrecisionTiming(stat string, delta time.Duration) error { return nil }
func (s StatsdNoop) Gauge(stat string, value int64) error                   { return nil }
func (s StatsdNoop) GaugeAbsolute(stat string, value int64) error           { return nil }
func (s StatsdNoop) GaugeAvg(stat string, value int64) error                { return nil }
func (s StatsdNoop) GaugeDelta(stat string, value int64) error              { return nil }
func (s StatsdNoop) Absolute(stat string, value int64) error                { return nil }
func (s StatsdNoop) Total(stat string, value int64) error                   { return nil }
func (s StatsdNoop) FGauge(stat string, value float64) error                { return nil }
func (s StatsdNoop) FGaugeDelta(stat string, value float64) error           { return nil }
func (s StatsdNoop) FAbsolute(stat string, value float64) error             { return nil }
