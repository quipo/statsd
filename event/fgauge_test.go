package event

import (
	"reflect"
	"testing"
)

func TestFGaugeUpdate(t *testing.T) {
	e1 := &FGauge{Name: "test", Value: float64(15.1)}
	e2 := &FGauge{Name: "test", Value: float64(-10.1)}
	e3 := &FGauge{Name: "test", Value: float64(8.4)}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}
	err = e1.Update(e3)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:8.4|g"} // only the last value is flushed
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}

func TestFGaugeUpdateNegative(t *testing.T) {
	e1 := &FGauge{Name: "test", Value: float64(-10.1)}
	e2 := &FGauge{Name: "test", Value: float64(-3.4)}
	err := e1.Update(e2)
	if nil != err {
		t.Error(err)
	}

	expected := []string{"test:0|g", "test:-3.4|g"} // only the last value is flushed
	actual := e1.Stats()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("did not receive all metrics: Expected: %T %v, Actual: %T %v ", expected, expected, actual, actual)
	}
}
