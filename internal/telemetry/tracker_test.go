package telemetry

import "testing"

func TestGlobalTrackerInit(t *testing.T) {
	if GlobalTracker == nil {
		t.Error("GlobalTracker should be initialized")
	}
	if GlobalTracker.RingBuffer == nil {
		t.Error("GlobalTracker.RingBuffer should be initialized")
	}
}
