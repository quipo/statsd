package event

import (
	"reflect"
	"testing"
)

func TestFGaugeDeltaUpdate(t *testing.T) {
	e1 := &FGaugeDelta{Name: "test", Value: float64(15.1)}
	e2 := &FGaugeDelta{Name: "test", Value: float64(-10.0)}
	e3 := &FGaugeDelta{Name: "test", Value: float64(15.1)}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:+20.2|g"}
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}

func TestFGaugeDeltaUpdateNegative(t *testing.T) {
	e1 := &FGaugeDelta{Name: "test", Value: float64(-15.1)}
	e2 := &FGaugeDelta{Name: "test", Value: float64(10.0)}
	e3 := &FGaugeDelta{Name: "test", Value: float64(-15.1)}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:-20.2|g"}
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}
