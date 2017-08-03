package statsd

import (
	"math"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// -------------------------------------------------------------------

type KVint64 struct {
	Key   string
	Value int64
}
type KVfloat64 struct {
	Key   string
	Value float64
}

type KVint64Sorter []KVint64
type KVfloat64Sorter []KVfloat64

func (a KVint64Sorter) Len() int      { return len(a) }
func (a KVint64Sorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a KVint64Sorter) Less(i, j int) bool {
	if a[i].Key == a[j].Key {
		return a[i].Value < a[j].Value
	}
	return a[i].Key < a[j].Key
}

func (a KVfloat64Sorter) Len() int      { return len(a) }
func (a KVfloat64Sorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a KVfloat64Sorter) Less(i, j int) bool {
	if a[i].Key == a[j].Key {
		return a[i].Value < a[j].Value
	}
	return a[i].Key < a[j].Key
}

// Normalise the number of decimal places for easier comparisons in the tests
func (a KVfloat64Sorter) Normalise(precision int) {
	for k, v := range a {
		v.Value = toFixed(v.Value, precision)
		a[k] = v
	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// -------------------------------------------------------------------

func TestBufferedInt64(t *testing.T) {
	hostname, err := os.Hostname()
	if nil != err {
		t.Fatal("Cannot read host name:", err)
	}

	regexTotal := regexp.MustCompile(`^(.*)\:([+\-]?\d+)\|(\w).*$`)
	prefix := "myproject."

	tt := []struct {
		name     string
		function string
		suffix   string
		input    []KVint64
		expected []KVint64
	}{
		{
			name:     "total",
			function: "total",
			suffix:   "t",
			input: []KVint64{
				{"a:b:c", 5},
				{"d:e:f", 2},
				{"x:b:c", 5},
				{"g.h.i", 1},
				{"zz.%HOST%", 1}, // also test %HOST% replacement
			},
			expected: []KVint64{
				{"a:b:c", 5},
				{"d:e:f", 2},
				{"g.h.i", 1},
				{"x:b:c", 5},
				{"zz." + hostname, 1}, // also test %HOST% replacement
			},
		},
		{
			name:     "gauge",
			function: "gauge",
			suffix:   "g",
			input: []KVint64{
				{"a:b:c", 5},
				{"d:e:f", 2},
				{"a:b:c", 2}, // this should override the previous one
				{"g.h.i", 1},
				{"zz.%HOST%", 1}, // also test %HOST% replacement
			},
			expected: []KVint64{
				{"a:b:c", 2},
				{"d:e:f", 2},
				{"g.h.i", 1},
				{"zz." + hostname, 1}, // also test %HOST% replacement
			},
		},
		{
			name:     "gaugedelta",
			function: "gaugedelta",
			suffix:   "g",
			input: []KVint64{
				{"a:b:c", +5},
				{"d:e:f", -2},
				{"a:b:c", -2},
				{"g.h.i", +1},
				{"zz.%HOST%", 1}, // also test %HOST% replacement
			},
			expected: []KVint64{
				{"a:b:c", 3},
				{"d:e:f", -2},
				{"g.h.i", +1},
				{"zz." + hostname, 1}, // also test %HOST% replacement
			},
		},
		{
			name:     "increment",
			function: "increment",
			suffix:   "c",
			input: []KVint64{
				{"a:b:c", 5},
				{"d:e:f", 2},
				{"a:b:c", -2},
				{"g.h.i", 1},
				{"zz.%HOST%", 1}, // also test %HOST% replacement
			},
			expected: []KVint64{
				{"a:b:c", 3},
				{"d:e:f", 2},
				{"g.h.i", 1},
				{"zz." + hostname, 1}, // also test %HOST% replacement
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			ln, udpAddr := newLocalListenerUDP(t)
			defer ln.Close()

			t.Log("Starting new UDP listener at", udpAddr.String())
			time.Sleep(50 * time.Millisecond)

			client := NewStatsdClient(udpAddr.String(), prefix)
			buffered := NewStatsdBuffer(time.Millisecond*20, client)

			ch := make(chan string)

			go doListenUDP(t, ln, ch, len(tc.expected))
			time.Sleep(50 * time.Millisecond)

			err = buffered.CreateSocket()
			if nil != err {
				t.Fatal(err)
			}
			defer buffered.Close()

			for _, entry := range tc.input {
				switch tc.function { // send metric
				case "total":
					err = buffered.Total(entry.Key, entry.Value)
				case "gauge":
					err = buffered.Gauge(entry.Key, entry.Value)
				case "gaugedelta":
					err = buffered.GaugeDelta(entry.Key, entry.Value)
				case "increment":
					if entry.Value < 0 {
						err = buffered.Decr(entry.Key, int64(math.Abs(float64(entry.Value))))
					} else {
						err = buffered.Incr(entry.Key, entry.Value)
					}
				}
				if nil != err {
					t.Error(err)
				}
			}

			received := 0
			var actual []KVint64
			for received < len(tc.expected) {
				batch := <-ch
				for _, x := range strings.Split(batch, "\n") {
					x = strings.TrimSpace(x)
					if "" == x {
						continue
					}
					if !strings.HasPrefix(x, prefix) {
						t.Errorf("Metric without expected prefix: expected '%s', actual '%s'", prefix, x)
						return
					}
					received++
					vv := regexTotal.FindStringSubmatch(x)
					//t.Log(vv, x)
					if len(vv) < 4 {
						t.Error("Expecting more tokens", len(vv))
						continue
					}
					if vv[3] != tc.suffix {
						t.Errorf("Metric without expected suffix: expected '%s', actual '%s'", tc.suffix, vv[3])
					}
					v, err := strconv.ParseInt(vv[2], 10, 64)
					if err != nil {
						t.Error(err)
					}
					actual = append(actual, KVint64{Key: vv[1][len(prefix):], Value: v})
				}
			}

			sort.Sort(KVint64Sorter(actual))

			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", tc.expected, tc.expected, actual, actual)
			}

			time.Sleep(500 * time.Millisecond)
		})
	}
}

func TestBufferedFloat64(t *testing.T) {
	hostname, err := os.Hostname()
	if nil != err {
		t.Fatal("Cannot read host name:", err)
	}

	regexTotal := regexp.MustCompile(`^(.*)\:([+\-]?\d+(?:\.\d+)?)\|(\w).*$`)
	prefix := "myproject."

	tt := []struct {
		name     string
		function string
		suffix   string
		input    KVfloat64Sorter
		expected KVfloat64Sorter
	}{
		{
			name:     "fgauge",
			function: "fgauge",
			suffix:   "g",
			input: KVfloat64Sorter{
				{"a:b:c", 5.2},
				{"d:e:f", 2.3},
				{"a:b:c", 2.2}, // this should override the previous one
				{"g.h.i", 1.2},
				{"zz.%HOST%", 1.1}, // also test %HOST% replacement
			},
			expected: KVfloat64Sorter{
				{"a:b:c", 2.2},
				{"d:e:f", 2.3},
				{"g.h.i", 1.2},
				{"zz." + hostname, 1.1}, // also test %HOST% replacement
			},
		},
		{
			name:     "fgaugedelta",
			function: "fgaugedelta",
			suffix:   "g",
			input: KVfloat64Sorter{
				{"a:b:c", +5.1},
				{"d:e:f", -2.2},
				{"a:b:c", -2.1},
				{"g.h.i", +1.3},
				{"zz.%HOST%", 1.4}, // also test %HOST% replacement
			},
			expected: KVfloat64Sorter{
				{"a:b:c", 3.0},
				{"d:e:f", -2.2},
				{"g.h.i", +1.3},
				{"zz." + hostname, 1.4}, // also test %HOST% replacement
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			ln, udpAddr := newLocalListenerUDP(t)
			defer ln.Close()

			t.Log("Starting new UDP listener at", udpAddr.String())
			time.Sleep(50 * time.Millisecond)

			client := NewStatsdClient(udpAddr.String(), prefix)
			buffered := NewStatsdBuffer(time.Millisecond*20, client)

			ch := make(chan string)

			go doListenUDP(t, ln, ch, len(tc.expected))
			time.Sleep(50 * time.Millisecond)

			err = buffered.CreateSocket()
			if nil != err {
				t.Fatal(err)
			}
			defer buffered.Close()

			for _, entry := range tc.input {
				switch tc.function { // send metric
				case "fgauge":
					err = buffered.FGauge(entry.Key, entry.Value)
				case "fgaugedelta":
					err = buffered.FGaugeDelta(entry.Key, entry.Value)
				}
				if nil != err {
					t.Error(err)
				}
			}

			received := 0
			var actual KVfloat64Sorter
			for received < len(tc.expected) {
				batch := <-ch
				for _, x := range strings.Split(batch, "\n") {
					x = strings.TrimSpace(x)
					if "" == x {
						continue
					}
					if !strings.HasPrefix(x, prefix) {
						t.Errorf("Metric without expected prefix: expected '%s', actual '%s'", prefix, x)
						return
					}
					received++
					vv := regexTotal.FindStringSubmatch(x)
					//t.Log(vv, x)
					if len(vv) < 4 {
						t.Error("Expecting more tokens", len(vv))
						continue
					}
					if vv[3] != tc.suffix {
						t.Errorf("Metric without expected suffix: expected '%s', actual '%s'", tc.suffix, vv[3])
					}
					v, err := strconv.ParseFloat(vv[2], 64)
					if err != nil {
						t.Error(err)
					}
					actual = append(actual, KVfloat64{Key: vv[1][len(prefix):], Value: v})
				}
			}

			actual.Normalise(2) // keep 2 decimal digits
			sort.Sort(actual)

			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("did not receive all metrics: Expected: \n%T %v, \nActual: \n%T %v ", tc.expected, tc.expected, actual, actual)
			}

			time.Sleep(500 * time.Millisecond)
		})
	}
}

func TestBufferedAbsolute(t *testing.T) {
	hostname, err := os.Hostname()
	if nil != err {
		t.Fatal("Cannot read host name:", err)
	}

	regexAbsolute := regexp.MustCompile(`^(.*)\:([+\-]?\d+)\|(\w).*$`)
	prefix := "myproject."

	tt := []struct {
		name     string
		suffix   string
		input    KVint64Sorter
		expected KVint64Sorter
	}{
		{
			name:   "absolute",
			suffix: "a",
			input: KVint64Sorter{
				{"a:b:c", 5},
				{"d:e:f", 2},
				{"a:b:c", 8},
				{"g.h.i", 1},
				{"zz.%HOST%", 1}, // also test %HOST% replacement
			},
			expected: KVint64Sorter{
				{"a:b:c", 5},
				{"a:b:c", 8},
				{"d:e:f", 2},
				{"g.h.i", 1},
				{"zz." + hostname, 1}, // also test %HOST% replacement
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			ln, udpAddr := newLocalListenerUDP(t)
			defer ln.Close()

			t.Log("Starting new UDP listener at", udpAddr.String())
			time.Sleep(50 * time.Millisecond)

			client := NewStatsdClient(udpAddr.String(), prefix)
			buffered := NewStatsdBuffer(time.Millisecond*20, client)

			ch := make(chan string)

			go doListenUDP(t, ln, ch, len(tc.expected))
			time.Sleep(50 * time.Millisecond)

			err = buffered.CreateSocket()
			if nil != err {
				t.Fatal(err)
			}
			defer buffered.Close()

			for _, entry := range tc.input {
				err = buffered.Absolute(entry.Key, entry.Value)
				if nil != err {
					t.Error(err)
				}
			}

			received := 0
			var actual KVint64Sorter
			for received < len(tc.expected) {
				batch := <-ch
				for _, x := range strings.Split(batch, "\n") {
					x = strings.TrimSpace(x)
					if "" == x {
						continue
					}
					if !strings.HasPrefix(x, prefix) {
						t.Errorf("Metric without expected prefix: expected '%s', actual '%s'", prefix, x)
						return
					}
					received++
					vv := regexAbsolute.FindStringSubmatch(x)
					//t.Log(vv, x)
					if len(vv) < 4 {
						t.Error("Expecting more tokens", len(vv))
						continue
					}
					if vv[3] != tc.suffix {
						t.Errorf("Metric without expected suffix: expected '%s', actual '%s'", tc.suffix, vv[3])
					}
					v, err := strconv.ParseInt(vv[2], 10, 64)
					if err != nil {
						t.Error(err)
					}
					actual = append(actual, KVint64{Key: vv[1][len(prefix):], Value: v})
				}
			}

			sort.Sort(actual)

			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("did not receive all metrics: \nExpected: \n%T %v, \nActual: \n%T %v ", tc.expected, tc.expected, actual, actual)
			}

			time.Sleep(500 * time.Millisecond)
		})
	}
}

func TestBufferedFAbsolute(t *testing.T) {
	hostname, err := os.Hostname()
	if nil != err {
		t.Fatal("Cannot read host name:", err)
	}

	regexAbsolute := regexp.MustCompile(`^(.*)\:([+\-]?\d+(?:\.\d+)?)\|(\w).*$`)
	prefix := "myproject."

	tt := []struct {
		name     string
		suffix   string
		input    KVfloat64Sorter
		expected KVfloat64Sorter
	}{
		{
			name:   "fabsolute",
			suffix: "a",
			input: KVfloat64Sorter{
				{"a:b:c", 5.2},
				{"d:e:f", 2.1},
				{"x:b:c", 5.1},
				{"g.h.i", 1.1},
				{"zz.%HOST%", 1.5}, // also test %HOST% replacement
			},
			expected: KVfloat64Sorter{
				{"a:b:c", 5.2},
				{"d:e:f", 2.1},
				{"g.h.i", 1.1},
				{"x:b:c", 5.1},
				{"zz." + hostname, 1.5}, // also test %HOST% replacement
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			ln, udpAddr := newLocalListenerUDP(t)
			defer ln.Close()

			t.Log("Starting new UDP listener at", udpAddr.String())
			time.Sleep(50 * time.Millisecond)

			client := NewStatsdClient(udpAddr.String(), prefix)
			buffered := NewStatsdBuffer(time.Millisecond*20, client)

			ch := make(chan string)

			go doListenUDP(t, ln, ch, len(tc.expected))
			time.Sleep(50 * time.Millisecond)

			err = buffered.CreateSocket()
			if nil != err {
				t.Fatal(err)
			}
			defer buffered.Close()

			for _, entry := range tc.input {
				err = buffered.FAbsolute(entry.Key, entry.Value)
				if nil != err {
					t.Error(err)
				}
			}

			received := 0
			var actual KVfloat64Sorter
			for received < len(tc.expected) {
				batch := <-ch
				for _, x := range strings.Split(batch, "\n") {
					x = strings.TrimSpace(x)
					if "" == x {
						continue
					}
					if !strings.HasPrefix(x, prefix) {
						t.Errorf("Metric without expected prefix: expected '%s', actual '%s'", prefix, x)
						return
					}
					received++
					vv := regexAbsolute.FindStringSubmatch(x)
					//t.Log(vv, x)
					if len(vv) < 4 {
						t.Error("Expecting more tokens", len(vv))
						continue
					}
					if vv[3] != tc.suffix {
						t.Errorf("Metric without expected suffix: expected '%s', actual '%s'", tc.suffix, vv[3])
					}
					v, err := strconv.ParseFloat(vv[2], 64)
					if err != nil {
						t.Error(err)
					}
					actual = append(actual, KVfloat64{Key: vv[1][len(prefix):], Value: toFixed(v, 2)})
				}
			}

			sort.Sort(actual)

			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("did not receive all metrics: \nExpected: \n%T %v, \nActual: \n%T %v ", tc.expected, tc.expected, actual, actual)
			}

			time.Sleep(500 * time.Millisecond)
		})
	}
}
