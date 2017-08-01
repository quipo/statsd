package statsd

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/quipo/statsd/event"
)

// MockNetConn is a mock for net.Conn
type MockNetConn struct {
	buf bytes.Buffer
}

func (mock *MockNetConn) Read(b []byte) (n int, err error) {
	return mock.buf.Read(b)
}
func (mock *MockNetConn) Write(b []byte) (n int, err error) {
	return mock.buf.Write(append(b, '\n'))
}
func (mock MockNetConn) Close() error {
	mock.buf.Truncate(0)
	return nil
}
func (mock MockNetConn) LocalAddr() net.Addr {
	return nil
}
func (mock MockNetConn) RemoteAddr() net.Addr {
	return nil
}
func (mock MockNetConn) SetDeadline(t time.Time) error {
	return nil
}
func (mock MockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (mock MockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

/*
// TODO: use this function instead mocking net.Conn
// usage: client, server := GetTestConnection("tcp", t)
// usage: client, server := GetTestConnection("udp", t)
func GetTestConnection(connType string, t *testing.T) (client, server net.Conn) {
	ln, err := net.Listen(connType, "127.0.0.1")
	if nil != err {
		t.Error("TCP errpr:", err)
	}
	go func() {
		defer ln.Close()
		server, err = ln.Accept()
		if nil != err {
			t.Error("TCP Accept errpr:", err)
		}
	}()

	client, err = net.Dial(connType, ln.Addr().String())
	if nil != err {
		t.Error("TCP Dial error:", err)
	}
	return client, server
}
*/

func TestTotal(t *testing.T) {
	ln, udpAddr := newLocalListenerUDP(t)
	defer ln.Close()
	t.Log("Starting new UDP listener at", udpAddr.String())
	time.Sleep(50 * time.Millisecond)

	prefix := "myproject."

	client := NewStatsdClient(udpAddr.String(), prefix)

	ch := make(chan string)

	s := map[string]int64{
		"a:b:c": 5,
		"d:e:f": 2,
		"x:b:c": 5,
		"g.h.i": 1,
	}

	expected := make(map[string]int64)
	for k, v := range s {
		expected[k] = v
	}

	// also test %HOST% replacement
	s["zz.%HOST%"] = 1
	hostname, err := os.Hostname()
	expected["zz."+hostname] = 1
	if nil != err {
		t.Error("Cannot read host name:", err.Error())
	}

	go doListenUDP(t, ln, ch, len(s))

	err = client.CreateSocket()
	if nil != err {
		t.Fatal("Create socket error:", err)
	}
	defer client.Close()

	for k, v := range s {
		err = client.Total(k, v)
		if nil != err {
			t.Error(err)
		}
	}

	actual := make(map[string]int64)

	re := regexp.MustCompile(`^(.*)\:(\d+)\|(\w).*$`)

	for i := len(s); i > 0; i-- {
		x, open := <-ch
		if !open {
			t.Logf("CLOSED CHANNEL")
			break
		}
		x = strings.TrimSpace(x)
		//t.Logf(x)
		if !strings.HasPrefix(x, prefix) {
			t.Errorf("Metric without expected prefix: expected '%s', actual '%s'", prefix, x)
			break
		}
		vv := re.FindStringSubmatch(x)
		if vv[3] != "t" {
			t.Errorf("Metric without expected suffix: expected 't', actual '%s'", vv[3])
		}
		v, err := strconv.ParseInt(vv[2], 10, 64)
		if err != nil {
			t.Error("Cannot parse int:", err)
		}
		actual[vv[1][len(prefix):]] = v
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}

func newLocalListenerUDP(t *testing.T) (*net.UDPConn, *net.UDPAddr) {
	addr := fmt.Sprintf(":%d", getFreePort())
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Error("UDP error:", err)
		return nil, nil
	}
	ln, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Error("UDP Listen error:", err)
		return ln, udpAddr
	}
	t.Logf("Started new local UDP listener @ %s\n", udpAddr)
	return ln, udpAddr
}

func doListenUDP(t *testing.T, conn *net.UDPConn, ch chan string, n int) {
	var wg sync.WaitGroup
	wg.Add(n)

	for n > 0 {
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c *net.UDPConn, ch chan string, wg *sync.WaitGroup) {
			t.Logf("Reading from UDP socket @ %s\n", conn.LocalAddr().String())
			buffer := make([]byte, 1024)
			size, err := c.Read(buffer)
			// size, address, err := sock.ReadFrom(buffer) <- This starts printing empty and nil values below immediatly
			if err != nil {
				t.Logf("Error reading from UDP socket. Buffer: %s, Size: %d, Error: %s\n", string(buffer), size, err)
				//t.Fatal(err)
			}
			t.Logf("Read buffer: \n------------------\n%s\n------------------\n* Size: %d\n", string(buffer), size)
			ch <- string(buffer)
			wg.Done()
		}(conn, ch, &wg)
		n--
	}
	wg.Wait()
	t.Logf("Finished listening on UDP socket @ %s\n", conn.LocalAddr().String())
}

func doListenTCP(t *testing.T, conn net.Listener, ch chan string, n int) {
	for n > 0 { // read n non-empty lines from TCP socket
		t.Logf("doListenTCP iteration")
		client, err := conn.Accept()

		if err != nil {
			t.Error(err)
			return
		}

		buf := make([]byte, 1024)
		c, err := client.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			t.Error(err)
			return
		}
		t.Logf("Read from TCP socket:\n----------\n%s\n----------\n", string(buf))
		for _, s := range bytes.Split(buf[:c], []byte{'\n'}) {
			if len(s) > 0 {
				n--
				ch <- string(s)
			}
		}

	}
}

func newLocalListenerTCP(t *testing.T) (string, net.Listener) {
	addr := fmt.Sprintf("127.0.0.1:%d", getFreePort())
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	return addr, ln
}

func TestTCP(t *testing.T) {
	addr, ln := newLocalListenerTCP(t)
	defer ln.Close()

	t.Log("Starting new TCP listener at", addr)
	time.Sleep(50 * time.Millisecond)

	prefix := "myproject."
	client := NewStatsdClient(addr, prefix)

	ch := make(chan string)

	s := map[string]int64{
		"a:b:c": 5,
		"d:e:f": 2,
		"x:b:c": 5,
		"g.h.i": 1,
	}

	expected := make(map[string]int64)
	for k, v := range s {
		expected[k] = v
	}

	// also test %HOST% replacement
	s["zz.%HOST%"] = 1
	hostname, err := os.Hostname()
	expected["zz."+hostname] = 1
	if nil != err {
		t.Error("Cannot read host name:", err.Error())
	}

	t.Logf("Sending stats to TCP Socket")
	err = client.CreateTCPSocket()
	if nil != err {
		t.Error(err)
	}
	defer client.Close()

	for k, v := range s {
		err = client.Total(k, v)
		if nil != err {
			t.Error(err)
		}
	}
	time.Sleep(60 * time.Millisecond)

	go doListenTCP(t, ln, ch, len(s))
	time.Sleep(50 * time.Millisecond)

	actual := make(map[string]int64)

	re := regexp.MustCompile(`^(.*)\:(\d+)\|(\w).*$`)

	for i := len(s); i > 0; i-- {
		//t.Logf("ITERATION %d\n", i)
		x, open := <-ch
		if !open {
			//t.Logf("CLOSED _____")
			break
		}
		x = strings.TrimSpace(x)
		if "" == x {
			//t.Logf("EMPTY STRING *****")
			break
		}
		//fmt.Println(x)
		if !strings.HasPrefix(x, prefix) {
			t.Errorf("Metric without expected prefix: expected '%s', actual '%s'", prefix, x)
			break
		}
		vv := re.FindStringSubmatch(x)
		if vv[3] != "t" {
			t.Errorf("Metric without expected suffix: expected 't', actual '%s'", vv[3])
		}
		v, err := strconv.ParseInt(vv[2], 10, 64)
		if err != nil {
			t.Error(err)
		}
		actual[vv[1][len(prefix):]] = v
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v \n", expected, expected, actual, actual)
	}
}

func TestSendEvents(t *testing.T) {
	c := NewStatsdClient("127.0.0.1:1201", "test")
	c.conn = &MockNetConn{} // mock connection

	// override with a small size
	UDPPayloadSize = 40

	e1 := &event.Increment{Name: "test1", Value: 123}
	e2 := &event.Increment{Name: "test2", Value: 432}
	e3 := &event.Increment{Name: "test3", Value: 111}
	e4 := &event.Gauge{Name: "test4", Value: 12435}

	events := map[string]event.Event{
		"test1": e1,
		"test2": e2,
		"test3": e3,
		"test4": e4,
	}

	err := c.SendEvents(events)
	if nil != err {
		t.Error(err)
	}

	b1 := make([]byte, UDPPayloadSize*3)
	n, err2 := c.conn.Read(b1)
	if nil != err2 {
		t.Error(err2)
	}
	nStats := len(strings.Split(strings.TrimSpace(string(b1[:n])), "\n"))
	if nStats != len(events) {
		t.Errorf("Was expecting %d events, got %d:  %s", len(events), nStats, string(b1))
	}
}

// getFreePort Ask the kernel for a free open port that is ready to use
func getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
