package mock

import (
	"reflect"
	"testing"

	"github.com/quipo/statsd/event"
)

func TestNoopBehavior(t *testing.T) {
	mockClient := MockStatsdClient{}
	err := mockClient.CreateSocket()
	if err != nil {
		t.Fail()
	}
	err = mockClient.CreateTCPSocket()
	if err != nil {
		t.Fail()
	}
	err = mockClient.Close()
	if err != nil {
		t.Fail()
	}
	err = mockClient.Incr("incr", 1)
	if err != nil {
		t.Fail()
	}
	err = mockClient.Decr("decr", 2)
	if err != nil {
		t.Fail()
	}
	err = mockClient.Timing("timing", 3)
	if err != nil {
		t.Fail()
	}
	err = mockClient.PrecisionTiming("precisionTiming", 4)
	if err != nil {
		t.Fail()
	}
	err = mockClient.Gauge("gauge", 5)
	if err != nil {
		t.Fail()
	}
	err = mockClient.GaugeDelta("gaugeDelta", 6)
	if err != nil {
		t.Fail()
	}
	err = mockClient.Absolute("absolute", 7)
	if err != nil {
		t.Fail()
	}
	err = mockClient.Total("total", 8)
	if err != nil {
		t.Fail()
	}

	err = mockClient.FGauge("fgauge", 10.0)
	if err != nil {
		t.Fail()
	}
	err = mockClient.FGaugeDelta("fgaugeDelta", 11.0)
	if err != nil {
		t.Fail()
	}
	err = mockClient.FAbsolute("fabsolute", 12.0)
	if err != nil {
		t.Fail()
	}

	err = mockClient.SendEvents(make(map[string]event.Event))
	if err != nil {
		t.Fail()
	}
}

func TestMockStatsdClient_RecordCreateSocketEventsTo(t *testing.T) {
	var createSocketEvents []UnvaluedEvent
	mockClient := (&MockStatsdClient{}).RecordCreateSocketEventsTo(&createSocketEvents)
	err := mockClient.CreateSocket()
	if err != nil {
		t.Logf("Got non-nil err from mock CreateSocket")
		t.Fail()
	}
	expectedEvents := []UnvaluedEvent{UnvaluedEvent{}}
	if !reflect.DeepEqual(createSocketEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, createSocketEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordCreateTCPSocketEventsTo(t *testing.T) {
	var createTCPSocketEvents []UnvaluedEvent
	mockClient := (&MockStatsdClient{}).RecordCreateTCPSocketEventsTo(&createTCPSocketEvents)
	err := mockClient.CreateTCPSocket()
	if err != nil {
		t.Logf("Got non-nil err from mock CreateTCPSocket")
		t.Fail()
	}
	expectedEvents := []UnvaluedEvent{UnvaluedEvent{}}
	if !reflect.DeepEqual(createTCPSocketEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, createTCPSocketEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordCloseEventsTo(t *testing.T) {
	var closeEvents []UnvaluedEvent
	mockClient := (&MockStatsdClient{}).RecordCloseEventsTo(&closeEvents)
	err := mockClient.Close()
	if err != nil {
		t.Logf("Got non-nil err from mock Close")
		t.Fail()
	}
	expectedEvents := []UnvaluedEvent{UnvaluedEvent{}}
	if !reflect.DeepEqual(closeEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, closeEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordIncrEventsTo(t *testing.T) {
	var incrEvents []Int64Event
	mockClient := (&MockStatsdClient{}).RecordIncrEventsTo(&incrEvents)
	err := mockClient.Incr("incr", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock Incr")
		t.Fail()
	}
	expectedEvents := []Int64Event{Int64Event{MetricName: "incr", EventValue: 1}}
	if !reflect.DeepEqual(incrEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, incrEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_Decr(t *testing.T) {
	var decrEvents []Int64Event
	mockClient := (&MockStatsdClient{}).RecordDecrEventsTo(&decrEvents)
	err := mockClient.Decr("decr", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock Decr")
		t.Fail()
	}
	expectedEvents := []Int64Event{Int64Event{MetricName: "decr", EventValue: 1}}
	if !reflect.DeepEqual(decrEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, decrEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_Timing(t *testing.T) {
	var timingEvents []Int64Event
	mockClient := (&MockStatsdClient{}).RecordIncrEventsTo(&timingEvents)
	err := mockClient.Incr("timing", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock Timing")
		t.Fail()
	}
	expectedEvents := []Int64Event{Int64Event{MetricName: "timing", EventValue: 1}}
	if !reflect.DeepEqual(timingEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, timingEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordPrecisionTimingEventsTo(t *testing.T) {
	var durationEvents []DurationEvent
	mockClient := (&MockStatsdClient{}).RecordPrecisionTimingEventsTo(&durationEvents)
	err := mockClient.PrecisionTiming("precisionTiming", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock PrecisionTiming")
		t.Fail()
	}
	expectedEvents := []DurationEvent{DurationEvent{MetricName: "precisionTiming", EventValue: 1}}
	if !reflect.DeepEqual(durationEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, durationEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordGaugeEventsTo(t *testing.T) {
	var gaugeEvents []Int64Event
	mockClient := (&MockStatsdClient{}).RecordGaugeEventsTo(&gaugeEvents)
	err := mockClient.Gauge("gauge", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock Gauge")
		t.Fail()
	}
	expectedEvents := []Int64Event{Int64Event{MetricName: "gauge", EventValue: 1}}
	if !reflect.DeepEqual(gaugeEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, gaugeEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordFGaugeDeltaEventsTo(t *testing.T) {
	var gaugeDeltaEvents []Int64Event
	mockClient := (&MockStatsdClient{}).RecordGaugeDeltaEventsTo(&gaugeDeltaEvents)
	err := mockClient.GaugeDelta("gaugeDelta", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock GaugeDelta")
		t.Fail()
	}
	expectedEvents := []Int64Event{Int64Event{MetricName: "gaugeDelta", EventValue: 1}}
	if !reflect.DeepEqual(gaugeDeltaEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, gaugeDeltaEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordAbsoluteEventsTo(t *testing.T) {
	var absoluteEvents []Int64Event
	mockClient := (&MockStatsdClient{}).RecordAbsoluteEventsTo(&absoluteEvents)
	err := mockClient.Absolute("absolute", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock Absolute")
		t.Fail()
	}
	expectedEvents := []Int64Event{Int64Event{MetricName: "absolute", EventValue: 1}}
	if !reflect.DeepEqual(absoluteEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, absoluteEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordTotalEventsTo(t *testing.T) {
	var totalEvents []Int64Event
	mockClient := (&MockStatsdClient{}).RecordTotalEventsTo(&totalEvents)
	err := mockClient.Total("total", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock Total")
		t.Fail()
	}
	expectedEvents := []Int64Event{Int64Event{MetricName: "total", EventValue: 1}}
	if !reflect.DeepEqual(totalEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, totalEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordFGaugeEventsTo(t *testing.T) {
	var fgaugeEvents []Float64Event
	mockClient := (&MockStatsdClient{}).RecordFGaugeEventsTo(&fgaugeEvents)
	err := mockClient.FGauge("fgauge", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock FGauge")
		t.Fail()
	}
	expectedEvents := []Float64Event{Float64Event{MetricName: "fgauge", EventValue: 1}}
	if !reflect.DeepEqual(fgaugeEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, fgaugeEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordFGaugeDeltaEventsTo2(t *testing.T) {
	var fgaugeDeltaEvents []Float64Event
	mockClient := (&MockStatsdClient{}).RecordFGaugeDeltaEventsTo(&fgaugeDeltaEvents)
	err := mockClient.FGaugeDelta("fgaugeDelta", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock FGaugeDelta")
		t.Fail()
	}
	expectedEvents := []Float64Event{Float64Event{MetricName: "fgaugeDelta", EventValue: 1}}
	if !reflect.DeepEqual(fgaugeDeltaEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, fgaugeDeltaEvents)
		t.Fail()
	}
}

func TestMockStatsdClient_RecordFAbsoluteEventsTo(t *testing.T) {
	var fabsoluteEvents []Float64Event
	mockClient := (&MockStatsdClient{}).RecordFAbsoluteEventsTo(&fabsoluteEvents)
	err := mockClient.FAbsolute("fabsolute", 1)
	if err != nil {
		t.Logf("Got non-nil err from mock FAbsolute")
		t.Fail()
	}
	expectedEvents := []Float64Event{Float64Event{MetricName: "fabsolute", EventValue: 1}}
	if !reflect.DeepEqual(fabsoluteEvents, expectedEvents) {
		t.Logf("Expected %s, saw %s", expectedEvents, fabsoluteEvents)
		t.Fail()
	}
}

//func TestRecordingBuilders(t *testing.T) {
//	var createTcpSocketEvents []UnvaluedEvent
//	var closeEvents []UnvaluedEvent
//	var incrEvents []Int64Event
//	var decrEvents []Int64Event
//	var timingEvents []Int64Event
//	var precidionTimingEvents []DurationEvent
//	var gaugeEvents []Int64Event
//	var gaugeDeltaEvents []Int64Event
//	var absoluteEvents []Int64Event
//	var totalEvents []Int64Event
//	var fgaugeEvents []Float64Event
//	var fgaugeDeltaEvents []Float64Event
//	var fabsoluteEvents []Float64Event
//
//	mockClient := &MockStatsdClient{}.
//		RecordCreateSocketEventsTo(&createSocketEvents).
//		RecordCreateTCPSocketEventsTo(&createTcpSocketEvents).
//		RecordCloseEventsTo(&closeEvents).
//		RecordIncrEventsTo(&incrEvents).
//		RecordDecrEventsTo(&decrEvents).
//		RecordTimingEventsTo(&timingEvents).
//		RecordPrecisionTimingEventsTo(&precidionTimingEvents).
//		RecordGaugeEventsTo(&gaugeEvents).
//		RecordGaugeDeltaEventsTo(&gaugeDeltaEvents).
//		RecordAbsoluteEventsTo(&absoluteEvents).
//		RecordTotalEventsTo(&totalEvents).
//		RecordFGaugeEventsTo(&fgaugeEvents).
//		RecordFGaugeDeltaEventsTo(&fgaugeDeltaEvents).
//		RecordFAbsoluteEventsTo(&fabsoluteEvents)
//
//	err := mockClient.CreateSocket()
//	if err != nil {
//		t.Fail()
//	}
//	err = mockClient.CreateTCPSocket()
//	if err != nil {
//		t.Fail()
//	}
//	err = mockClient.Close()
//	if err != nil {
//		t.Fail()
//	}
//	err = mockClient.Incr("incr", 1)
//	if err != nil {
//		t.Fail()
//	}
//	err = mockClient.Decr("decr", 1)
//	if err != nil {
//		t.Fail()
//	}
//	err = mockClient.Timing("timing", 1)
//	if err != nil {
//		t.Fail()
//	}
//}
