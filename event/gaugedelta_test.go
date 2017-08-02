package event

import (
	"reflect"
	"testing"
)

func TestGaugeDeltaUpdate(t *testing.T) {
	e1 := &GaugeDelta{Name: "test", Value: int64(15)}
	e2 := &GaugeDelta{Name: "test", Value: int64(-10)}
	e3 := &GaugeDelta{Name: "test", Value: int64(15)}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:+20|g"}
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}

func TestGaugeDeltaUpdateNegative(t *testing.T) {
	e1 := &GaugeDelta{Name: "test", Value: int64(-15)}
	e2 := &GaugeDelta{Name: "test", Value: int64(10)}
	e3 := &GaugeDelta{Name: "test", Value: int64(-15)}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:-20|g"}
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}
