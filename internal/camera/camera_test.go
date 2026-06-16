package camera

import (
	"testing"
	"time"

	"github.com/IDisposable/docker-wyze-bridge/internal/wyzeapi"
)

func TestCameraState(t *testing.T) {
	cam := NewCamera(wyzeapi.CameraInfo{Name: "test"}, "hd", true, false)

	if cam.GetState() != StateOffline {
		t.Errorf("initial state = %v, want Offline", cam.GetState())
	}

	cam.SetState(StateConnecting)
	if cam.GetState() != StateConnecting {
		t.Errorf("state = %v, want Connecting", cam.GetState())
	}

	cam.SetState(StateStreaming)
	if cam.GetState() != StateStreaming {
		t.Errorf("state = %v, want Streaming", cam.GetState())
	}
	if cam.ConnectedAt.IsZero() {
		t.Error("ConnectedAt should be set on streaming")
	}
	if cam.ErrorCount != 0 {
		t.Error("ErrorCount should reset on streaming")
	}
}

func TestBackoffDuration(t *testing.T) {
	cam := NewCamera(wyzeapi.CameraInfo{Name: "test"}, "hd", true, false)

	// First error: 5s * 1 = 5s
	d := cam.IncrementError()
	if d != 5*time.Second {
		t.Errorf("backoff after 1 error = %v, want 5s", d)
	}

	// Second error: 5s * 2 = 10s
	d = cam.IncrementError()
	if d != 10*time.Second {
		t.Errorf("backoff after 2 errors = %v, want 10s", d)
	}

	// Keep incrementing until we hit cap (need 60 errors total for 5min)
	for i := 2; i < 60; i++ {
		d = cam.IncrementError()
	}
	if d != 5*time.Minute {
		t.Errorf("max backoff = %v, want 5m", d)
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateOffline, "offline"},
		{StateDiscovering, "discovering"},
		{StateConnecting, "connecting"},
		{StateStreaming, "streaming"},
		{StateError, "error"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestStatusJSON(t *testing.T) {
	cam := NewCamera(wyzeapi.CameraInfo{
		Name:     "front_door",
		Nickname: "Front Door",
		Model:    "HL_CAM4",
		MAC:      "AABBCCDDEEFF",
		LanIP:    "192.168.1.10",
	}, "hd", true, false)

	status := cam.StatusJSON()
	if status["name"] != "front_door" {
		t.Errorf("name = %v", status["name"])
	}
	if status["model_name"] != "V4" {
		t.Errorf("model_name = %v", status["model_name"])
	}
	if status["quality"] != "hd" {
		t.Errorf("quality = %v", status["quality"])
	}
}
