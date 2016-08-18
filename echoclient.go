package statsd

import (
	"log"
	"time"
)

//impliment a "echo" statsd for debuging statsd
type StatsdEcho struct{}

func prit(stat string, val interface{}) error {
	log.Printf("%s:%v", stat, val)
	return nil
}

func (s StatsdEcho) String() string                                         { return "EchoClient" }
func (s StatsdEcho) CreateSocket() error                                    { return nil }
func (s StatsdEcho) Close() error                                           { return nil }
func (s StatsdEcho) Incr(stat string, count int64) error                    { return prit(stat, count) }
func (s StatsdEcho) Decr(stat string, count int64) error                    { return prit(stat, count) }
func (s StatsdEcho) Timing(stat string, count int64) error                  { return prit(stat, count) }
func (s StatsdEcho) PrecisionTiming(stat string, delta time.Duration) error { return prit(stat, delta) }
func (s StatsdEcho) Gauge(stat string, value int64) error                   { return prit(stat, value) }
func (s StatsdEcho) GaugeAbsolute(stat string, value int64) error           { return prit(stat, value) }
func (s StatsdEcho) GaugeAvg(stat string, value int64) error                { return prit(stat, value) }
func (s StatsdEcho) GaugeDelta(stat string, value int64) error              { return prit(stat, value) }
func (s StatsdEcho) Absolute(stat string, value int64) error                { return prit(stat, value) }
func (s StatsdEcho) Total(stat string, value int64) error                   { return prit(stat, value) }
func (s StatsdEcho) FGauge(stat string, value float64) error                { return prit(stat, value) }
func (s StatsdEcho) FGaugeDelta(stat string, value float64) error           { return prit(stat, value) }
func (s StatsdEcho) FAbsolute(stat string, value float64) error             { return prit(stat, value) }
