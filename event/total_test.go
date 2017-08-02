package event

import (
	"reflect"
	"testing"
)

func TestTotalUpdate(t *testing.T) {
	e1 := &Total{Name: "test", Value: int64(15)}
	e2 := &Total{Name: "test", Value: int64(-10)}
	e3 := &Total{Name: "test", Value: int64(8)}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:13|t"} // only the last value is flushed
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}
