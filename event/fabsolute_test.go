package event

import (
	"reflect"
	"testing"
)

func TestFAbsoluteUpdate(t *testing.T) {
	e1 := &FAbsolute{Name: "test", Values: []float64{15.3}}
	e2 := &FAbsolute{Name: "test", Values: []float64{-10.1}}
	e3 := &FAbsolute{Name: "test", Values: []float64{8.3}}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:15.3|a", "test:-10.1|a", "test:8.3|a"} // only the last value is flushed
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}
