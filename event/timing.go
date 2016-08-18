package event

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// Timing keeps min/max/avg information about a timer over a certain interval
type Timing struct {
	Name string

	mu               sync.Mutex
	Min              int64
	Max              int64
	Value            int64
	Values           []int64
	Count            int64
	Sample           float64
	PercentThreshold []float64
}

func (e *Timing) StatClass() string {
	return "timer"
}

// for sorting
type statdInt64arr []int64

func (a statdInt64arr) Len() int           { return len(a) }
func (a statdInt64arr) Swap(i int, j int)  { a[i], a[j] = a[j], a[i] }
func (a statdInt64arr) Less(i, j int) bool { return (a[i] - a[j]) < 0 } //this is the sorting statsd uses for its timings

func round(a float64) float64 {
	if a < 0 {
		return math.Ceil(a - 0.5)
	}
	return math.Floor(a + 0.5)
}

// NewTiming is a factory for a Timing event, setting the Count to 1 to prevent div_by_0 errors
func NewTiming(k string, delta int64, sample float64) *Timing {
	return &Timing{Name: k, Min: delta, Max: delta, Value: delta, Count: 1, Sample: 1.0 / sample, PercentThreshold: []float64{0.9, 0.99}}
}

// Update the event with metrics coming from a new one of the same type and with the same key
func (e *Timing) Update(e2 Event) error {
	if e.Type() != e2.Type() {
		return fmt.Errorf("statsd event type conflict: %s vs %s ", e.String(), e2.String())
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	p := e2.Payload().(map[string]interface{})
	e.Sample += p["sample"].(float64)
	e.Count += p["cnt"].(int64)
	e.Value += p["val"].(int64)
	e.Min = minInt64(e.Min, p["min"].(int64))
	e.Max = maxInt64(e.Max, p["max"].(int64))
	e.Values = append(e.Values, p["val"].(int64))
	return nil
}

//Reset the value
func (e *Timing) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Value = 0
	e.Count = 1
	e.Min = 0
	e.Max = 0
	e.Values = []int64{}
}

// Payload returns the aggregated value for this event
func (e Timing) Payload() interface{} {
	return map[string]interface{}{
		"min":    e.Min,
		"sample": e.Sample,
		"max":    e.Max,
		"val":    e.Value,
		"cnt":    e.Count,
		"vals":   e.Values,
	}
}

// Stats returns an array of StatsD events
func (e Timing) Stats(tick time.Duration) []string {
	e.mu.Lock()
	defer e.mu.Unlock()

	std := float64(0)
	avg := float64(e.Value / e.Count)
	cumulativeValues := []int64{e.Min}

	sort.Sort(statdInt64arr(e.Values))

	for idx, v := range e.Values {
		std += math.Pow((float64(v) - avg), 2.0)
		if idx > 0 {
			cumulativeValues = append(cumulativeValues, v+cumulativeValues[idx-1])
		}
	}

	base := []string{
		fmt.Sprintf("%s.count:%d|c", e.Name, e.Count),
		fmt.Sprintf("%s.min:%d|ms", e.Name, e.Min),
		fmt.Sprintf("%s.max:%d|ms", e.Name, e.Max),
		fmt.Sprintf("%s.sum:%d|c", e.Name, int64(e.Value)),
	}

	if e.Count > 0 {
		mid := int(math.Floor(float64(e.Count) / 2.0))
		median := int64(0)
		if mid < len(e.Values) {
			if math.Mod(float64(mid), 2.0) == 0 {
				median = e.Values[mid]
			} else if len(e.Values) > 1 {
				median = (e.Values[mid-1] + e.Values[mid]) / 2.0
			}
		}
		std = math.Sqrt(std / float64(e.Count))
		base = append(base,
			fmt.Sprintf("%s.median:%d|ms", e.Name, int64(median)),
			fmt.Sprintf("%s.std:%d|ms", e.Name, int64(e.Value)),
		)
	}

	return base
}

// Key returns the name of this metric
func (e Timing) Key() string {
	return e.Name
}

// SetKey sets the name of this metric
func (e *Timing) SetKey(key string) {
	e.Name = key
}

// Type returns an integer identifier for this type of metric
func (e Timing) Type() int {
	return EventTiming
}

// TypeString returns a name for this type of metric
func (e Timing) TypeString() string {
	return "Timing"
}

// String returns a debug-friendly representation of this metric
func (e Timing) String() string {
	return fmt.Sprintf("{Type: %s, Key: %s, Value: %+v}", e.TypeString(), e.Name, e.Payload())
}

func minInt64(v1, v2 int64) int64 {
	if v1 <= v2 {
		return v1
	}
	return v2
}
func maxInt64(v1, v2 int64) int64 {
	if v1 >= v2 {
		return v1
	}
	return v2
}
