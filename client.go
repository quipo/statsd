package statsd

import (
	"fmt"
	"net"
	"time"

	"github.com/quipo/statsd/event"
)

// StatsdClient is a client library to send events to StatsD
type StatsdClient struct {
	conn   net.Conn
	addr   string
	prefix string
}

// NewStatsdClient - Factory
func NewStatsdClient(addr string, prefix string) *StatsdClient {
	return &StatsdClient{
		addr:   addr,
		prefix: prefix,
	}
}

// String returns the StatsD server address
func (c *StatsdClient) String() string {
	return c.addr
}

// CreateSocket creates a UDP connection to a StatsD server
func (c *StatsdClient) CreateSocket() error {
	conn, err := net.DialTimeout("udp", c.addr, time.Second)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// Close the UDP connection
func (c *StatsdClient) Close() error {
	if nil == c.conn {
		return nil
	}
	return c.conn.Close()
}

// See statsd data types here: http://statsd.readthedocs.org/en/latest/types.html

// Incr - Increment a counter metric. Often used to note a particular event
func (c *StatsdClient) Incr(stat string, count int64) error {
	if 0 != count {
		return c.send(stat, "%d|c", count)
	}
	return nil
}

// Decr - Decrement a counter metric. Often used to note a particular event
func (c *StatsdClient) Decr(stat string, count int64) error {
	if 0 != count {
		return c.send(stat, "%d|c", -count)
	}
	return nil
}

// Timing - Track a duration event
func (c *StatsdClient) Timing(stat string, delta int64) error {
	return c.send(stat, "%d|ms", delta)
}

// Gauge - Gauges are a constant data type. They are not subject to averaging,
// and they don’t change unless you change them. That is, once you set a gauge value,
// it will be a flat line on the graph until you change it again
func (c *StatsdClient) Gauge(stat string, value int64) error {
	return c.send(stat, "%d|g", value)
}

// Absolute - Send absolute-valued metric (not averaged/aggregated)
func (c *StatsdClient) Absolute(stat string, value int64) error {
	return c.send(stat, "%d|a", value)
}

// Total - Send a metric that is continously increasing, e.g. read operations since boot
func (c *StatsdClient) Total(stat string, value int64) error {
	return c.send(stat, "%d|t", value)
}

// write a UDP packet with the statsd event
func (c *StatsdClient) send(stat string, format string, value int64) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}
	format = fmt.Sprintf("%s%s:%s", c.prefix, stat, format)
	//fmt.Printf("SENDING %s%s:%s\n", c.prefix, stat, format)
	_, err := fmt.Fprintf(c.conn, format, value)
	return err
}

// SendEvent - Sends stats from an event object
func (c *StatsdClient) SendEvent(e event.Event) error {
	if c.conn == nil {
		return fmt.Errorf("cannot send stats, not connected to StatsD server")
	}
	for _, stat := range e.Stats() {
		//fmt.Printf("SENDING EVENT %s%s\n", c.prefix, stat)
		_, err := fmt.Fprintf(c.conn, "%s%s", c.prefix, stat)
		if nil != err {
			return err
		}
	}
	return nil
}
