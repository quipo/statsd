package statsd

import "sync"

var std Statsd
var mu sync.Mutex

func init() {
	Unconfigure()
}

// Configure creates a global StatsD client.
// TODO: Use a buffered client instead!
func Configure(host string, prefix string) error {
	mu.Lock()
	defer mu.Unlock()

	client := NewStatsdClient(host, prefix)
	err := client.CreateSocket()
	if err != nil {
		return err
	}

	std = client
	return nil
}

// Unconfigure resets the global StatsD client to its default state, which is
// to silently drop all events.
func Unconfigure() {
	std = &NoopClient{}
}

// These functions write to the global StatsD client if one has been configured,
// otherwise do nothing.

func Incr(stat string, count int64) error {
	return std.Incr(stat, count)
}

func Decr(stat string, count int64) error {
	return std.Decr(stat, count)
}

func Timing(stat string, delta int64) error {
	return std.Timing(stat, delta)
}

func Absolute(stat string, value int64) error {
	return std.Absolute(stat, value)
}

func Total(stat string, value int64) error {
	return std.Total(stat, value)
}
