package mock

import "testing"

func TestHzToTonePeriod(t *testing.T) {

	hz := 440.0
	period := FreqHzToTonePeriod(hz)
	expectedPeriod := uint16(145)
	if period != expectedPeriod {
		t.Fatalf("Expected %d as period, but got %d", expectedPeriod, period)
	}

	hz2 := TonePeriodToFreqHz(period)
	if round(hz2) != round(hz) {
		t.Fatalf("Expected %d as freq, but got %d", int(hz), int(hz2))
	}

}
