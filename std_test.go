package statsd

import (
	"testing"
)

func TestConfigure(t *testing.T) {

	// assert that global is a noop client
	_, ok := std.(*NoopClient)
	if !ok { 
		t.Errorf("expected std to be a NoopClient, got a %#v", std)
	}

	// assert that after calling Configure, the global is a StatsdClient
	Configure("localhost:8888", "prefix")
	_, ok = std.(*StatsdClient)
	if !ok {
		t.Errorf("expected std to be a StatsdClient, got a %#v", std)
	}

	// assert that after calling Unconfigure, the global is back to being a
	// NoopClient
	Unconfigure()
	_, ok = std.(*NoopClient)
	if !ok {
		t.Errorf("expected std to be a NoopClient, got a %#v", std)
	}
}
