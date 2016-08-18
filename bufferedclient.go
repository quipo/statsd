package statsd

import (
	"expvar"
	"log"
	"math/rand"
	"os"
	"github.com/wyndhblb/statsd/event"
	"strings"
	"time"
)

// request to close the buffered statsd collector
type closeRequest struct {
	reply chan error
}

var (
	ExpMapCounts *expvar.Map
	ExpMapGauges *expvar.Map
	ExpMapTimers *expvar.Map
)

func init() {

	ExpMapCounts = expvar.NewMap("counters")
	ExpMapGauges = expvar.NewMap("gauges")
	ExpMapTimers = expvar.NewMap("timers")
}

// StatsdBuffer is a client library to aggregate events in memory before
// flushing aggregates to StatsD, useful if the frequency of events is extremely high
// and sampling is not desirable
type StatsdBuffer struct {
	statsd            *StatsdClient
	name              string
	flushInterval     time.Duration
	eventChannel      chan event.Event
	events            map[string]event.Event
	closeChannel      chan closeRequest
	Logger            *log.Logger
	RetainKeys        bool
	AddExpvar         bool // add stats to expvar list of vars
	ReCycleConnection bool //close the Statsd connection each time after send (helps with picking up new hosts)
	BufferLength      int  //buffer length before we send a multi line stat
	SampleRate        float32
	TimerSampleRate   float32
}

// NewStatsdBuffer Factory
func NewStatsdBuffer(name string, interval time.Duration, client *StatsdClient) *StatsdBuffer {
	sb := &StatsdBuffer{
		flushInterval:     interval,
		name:              name,
		statsd:            client,
		eventChannel:      make(chan event.Event, 100),
		events:            make(map[string]event.Event, 0),
		closeChannel:      make(chan closeRequest, 0),
		Logger:            log.New(os.Stdout, "[BufferedStatsdClient] ", log.Ldate|log.Ltime),
		RetainKeys:        false,
		AddExpvar:         true,
		ReCycleConnection: true,
		BufferLength:      512,
		SampleRate:        1.0,
		TimerSampleRate:   1.0,
	}
	rand.Seed(time.Now().UnixNano())
	go sb.collector()
	return sb
}

func (sb *StatsdBuffer) String() string {
	return sb.statsd.String()
}

// CreateSocket creates a UDP connection to a StatsD server
func (sb *StatsdBuffer) CreateSocket() error {
	return sb.statsd.CreateSocket()
}

func (sb *StatsdBuffer) registerStat() bool {
	if sb.SampleRate >= 1.0 {
		return true
	}
	return rand.Float32() < sb.SampleRate
}

func (sb *StatsdBuffer) registerTimerStat(samplerate float32) bool {
	if samplerate >= 1.0 {
		return true
	}
	return rand.Float32() < samplerate
}

// Incr - Increment a counter metric. Often used to note a particular event
func (sb *StatsdBuffer) Incr(stat string, count int64) error {
	if !sb.registerStat() {
		return nil
	}
	if 0 != count {
		//ct := int64(float32(count) / sb.SampleRate)
		sb.eventChannel <- &event.Increment{Name: stat, Value: count}
	}
	return nil
}

// Decr - Decrement a counter metric. Often used to note a particular event
func (sb *StatsdBuffer) Decr(stat string, count int64) error {
	if !sb.registerStat() {
		return nil
	}
	if 0 != count {
		//ct := int64(float32(count) / sb.SampleRate)
		sb.eventChannel <- &event.Increment{Name: stat, Value: -count}
	}
	return nil
}

// Timing - Track a duration event
func (sb *StatsdBuffer) Timing(stat string, delta int64) error {
	if !sb.registerTimerStat(sb.TimerSampleRate) {
		return nil
	}
	sb.eventChannel <- event.NewTiming(stat, delta, float64(sb.TimerSampleRate))
	return nil
}

// TimingSampling - Track a duration event at a sampling rate
func (sb *StatsdBuffer) TimingSampling(stat string, delta int64, samplerate float32) error {
	if !sb.registerTimerStat(samplerate) {
		return nil
	}
	sb.eventChannel <- event.NewTiming(stat, delta, float64(sb.TimerSampleRate))
	return nil
}

// PrecisionTiming - Track a duration event
// the time delta has to be a duration
func (sb *StatsdBuffer) PrecisionTimingSampling(stat string, delta time.Duration, samplerate float32) error {
	if !sb.registerTimerStat(samplerate) {
		return nil
	}
	sb.eventChannel <- event.NewPrecisionTiming(stat, time.Duration(float64(delta)/float64(time.Millisecond)))
	return nil
}

// PrecisionTiming - Track a duration event
// the time delta has to be a duration
func (sb *StatsdBuffer) PrecisionTiming(stat string, delta time.Duration) error {
	if !sb.registerTimerStat(sb.TimerSampleRate) {
		return nil
	}
	sb.eventChannel <- event.NewPrecisionTiming(stat, time.Duration(float64(delta)/float64(time.Millisecond)))
	return nil
}

// Gauge - Gauges are a constant data type. They are not subject to averaging,
// and they don’t change unless you change them. That is, once you set a gauge value,
// it will be a flat line on the graph until you change it again
func (sb *StatsdBuffer) Gauge(stat string, value int64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.Gauge{Name: stat, Value: value}
	return nil
}

// GaugeAbsolute which DOES NOT "add" values in the buffer, just
// resets the value to the current value, here in the non-buffered client it is
// the same as a Gauge
func (sb *StatsdBuffer) GaugeAbsolute(stat string, value int64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.GaugeAbsolute{Name: stat, Value: value}
	return nil
}

func (sb *StatsdBuffer) GaugeAvg(stat string, value int64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.GaugeAvg{Name: stat, Value: value}
	return nil
}

func (sb *StatsdBuffer) GaugeDelta(stat string, value int64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.GaugeDelta{Name: stat, Value: value}
	return nil
}

func (sb *StatsdBuffer) FGauge(stat string, value float64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.FGauge{Name: stat, Value: value}
	return nil
}

func (sb *StatsdBuffer) FGaugeDelta(stat string, value float64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.FGaugeDelta{Name: stat, Value: value}
	return nil
}

// Absolute - Send absolute-valued metric (not averaged/aggregated)
func (sb *StatsdBuffer) Absolute(stat string, value int64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.Absolute{Name: stat, Values: []int64{value}}
	return nil
}

// FAbsolute - Send absolute-valued metric (not averaged/aggregated)
func (sb *StatsdBuffer) FAbsolute(stat string, value float64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.FAbsolute{Name: stat, Values: []float64{value}}
	return nil
}

// Total - Send a metric that is continously increasing, e.g. read operations since boot
func (sb *StatsdBuffer) Total(stat string, value int64) error {
	if !sb.registerStat() {
		return nil
	}
	sb.eventChannel <- &event.Total{Name: stat, Value: value}
	return nil
}

func (sb *StatsdBuffer) expCollect(ev event.Event) {
	payload := ev.Payload()
	k := ev.Key()

	var use_map *expvar.Map

	switch ev.StatClass() {
	case "counter":
		use_map = ExpMapCounts
		switch payload.(type) {
		case int64:
			use_map.Add(k, payload.(int64))

		case float64:
			use_map.AddFloat(k, payload.(float64))

		case []int64:
			p_len := len(payload.([]int64))
			if p_len > 0 {
				use_map.Add(k, payload.([]int64)[p_len-1])
			}
		case []float64:
			p_len := len(payload.([]float64))
			if p_len > 0 {
				use_map.AddFloat(k, payload.([]float64)[p_len-1])
			}
		case map[string]interface{}:
			use_map.Add(k, payload.(map[string]interface{})["val"].(int64))

		}
	case "gauge":
		use_map = ExpMapGauges
		switch payload.(type) {
		case int64:
			v := new(expvar.Int)
			v.Set(payload.(int64))
			use_map.Set(k, v)
		case float64:
			v := new(expvar.Float)
			v.Set(payload.(float64))
			use_map.Set(k, v)

		case []int64:
			p_len := len(payload.([]int64))
			if p_len > 0 {
				v := new(expvar.Int)
				v.Set(payload.([]int64)[p_len-1])
				use_map.Set(k, v)
			}
		case []float64:
			p_len := len(payload.([]float64))
			if p_len > 0 {
				v := new(expvar.Float)
				v.Set(payload.([]float64)[p_len-1])
				use_map.Set(k, v)
			}
		case map[string]interface{}:
			v := new(expvar.Int)
			v.Set(payload.(map[string]interface{})["val"].(int64))
			use_map.Set(k, v)
		}
	case "timer":
		use_map = ExpMapTimers
		switch payload.(type) {
		case int64:
			v := new(expvar.Int)
			v.Set(payload.(int64))
			use_map.Set(k, v)
		case float64:
			v := new(expvar.Float)
			v.Set(payload.(float64))
			use_map.Set(k, v)

		case []int64:
			p_len := len(payload.([]int64))
			if p_len > 0 {
				v := new(expvar.Int)
				v.Set(payload.([]int64)[p_len-1])
				use_map.Set(k, v)
			}
		case []float64:
			p_len := len(payload.([]float64))
			if p_len > 0 {
				v := new(expvar.Float)
				v.Set(payload.([]float64)[p_len-1])
				use_map.Set(k, v)
			}
		case map[string]interface{}:
			v := new(expvar.Int)
			v.Set(payload.(map[string]interface{})["val"].(int64))
			use_map.Set(k, v)
		}
	}

}

// handle flushes and updates in one single thread (instead of locking the events map)
func (sb *StatsdBuffer) collector() {
	// on a panic event, flush all the pending stats before panicking
	defer func(sb *StatsdBuffer) {
		if r := recover(); r != nil {
			sb.Logger.Println("Caught panic, flushing stats before throwing the panic again")
			sb.flush()
			panic(r)
		}
	}(sb)

	ticker := time.NewTicker(sb.flushInterval)

	for {
		select {
		case <-ticker.C:
			//sb.Logger.Println("Flushing stats")
			sb.flush()
		case e := <-sb.eventChannel:
			//sb.Logger.Println("Received ", e.String())
			// convert %HOST% in key
			k := strings.Replace(e.Key(), "%HOST%", Hostname, 1)
			e.SetKey(k)

			if e2, ok := sb.events[k]; ok {
				//sb.Logger.Println("Updating existing event")
				e2.Update(e)
				sb.events[k] = e2
			} else {
				//sb.Logger.Println("Adding new event")
				sb.events[k] = e
			}
			sb.expCollect(sb.events[k])

		case c := <-sb.closeChannel:
			sb.Logger.Println("Asked to terminate. Flushing stats before returning.")
			c.reply <- sb.flush()
			break
		}
	}
}

// Close sends a close event to the collector asking to stop & flush pending stats
// and closes the statsd client
func (sb *StatsdBuffer) Close() (err error) {
	// 1. send a close event to the collector
	req := closeRequest{reply: make(chan error, 0)}
	sb.closeChannel <- req
	// 2. wait for the collector to drain the queue and respond
	err = <-req.reply
	// 3. close the statsd client
	err2 := sb.statsd.Close()
	if err != nil {
		return err
	}
	return err2
}

// send the events to StatsD and reset them.
// This function is NOT thread-safe, so it must only be invoked synchronously
// from within the collector() goroutine
func (sb *StatsdBuffer) flush() (err error) {
	n := len(sb.events)
	if n == 0 {
		return nil
	}
	err = sb.statsd.CreateSocket()

	if sb.ReCycleConnection {
		defer sb.statsd.Close()
	}

	if nil != err {
		sb.Logger.Println("Error establishing UDP connection for sending statsd events:", err)
	}

	//buffer stats in BufferLength blocks save some frames
	var out_string = ""

	for k, v := range sb.events {
		var str string
		switch v.(type) {
		case *event.Timing, *event.PrecisionTiming:
			str, err = sb.statsd.EventStatsdStringTimerSample(v, sb.flushInterval, sb.TimerSampleRate)

		default:
			str, err = sb.statsd.EventStatsdString(v, sb.flushInterval, sb.SampleRate)
		}
		if nil != err {
			sb.Logger.Println(err)
			continue
		}
		if len([]byte(out_string+str)) >= sb.BufferLength {
			sb.statsd.SendRaw(out_string)
			out_string = ""
		}

		out_string += str

		// if we are "retaining" the names (so that we can send "0"s do NOT delete
		if sb.RetainKeys {
			sb.events[k].Reset()
		} else {
			delete(sb.events, k)
		}
	}

	if len(out_string) > 0 {
		sb.statsd.SendRaw(out_string)
	}

	return nil
}
