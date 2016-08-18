package event

import (
	"time"
)

// constant event type identifiers
const (
	EventIncr = iota
	EventTiming
	EventAbsolute
	EventTotal
	EventGauge
	EventGaugeDelta
	EventFGauge
	EventFGaugeDelta
	EventFAbsolute
	EventPrecisionTiming
	EventGaugeAbsolute
	EventGaugeAvg
)

// Event is an interface to a generic StatsD event, used by the buffered client collator
type Event interface {
	//tick duration for those in the buffered that need a "persecond" metrix as well
	Stats(tick time.Duration) []string
	Type() int
	TypeString() string
	Payload() interface{}
	Update(e2 Event) error
	String() string
	Key() string
	SetKey(string)
	Reset()
	StatClass() string // counter, gauge, timer
}
