package statsd

// Statsd is an interface to a StatsD client (buffered/unbuffered)
type Statsd interface {
	Close() error
	Incr(stat string, count int64) error
	Decr(stat string, count int64) error
	Timing(stat string, delta int64) error
	Gauge(stat string, value int64) error
	Absolute(stat string, value int64) error
	Total(stat string, value int64) error
}
