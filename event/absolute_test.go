package event

import (
	"reflect"
	"testing"
)

func TestAbsoluteUpdate(t *testing.T) {
	e1 := &Absolute{Name: "test", Values: []int64{15}}
	e2 := &Absolute{Name: "test", Values: []int64{-10}}
	e3 := &Absolute{Name: "test", Values: []int64{8}}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:15|a", "test:-10|a", "test:8|a"} // only the last value is flushed
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}
