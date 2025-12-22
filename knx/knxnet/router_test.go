package knxnet

import (
	"testing"
	"time"
)

func TestDeviceStateString(t *testing.T) {
	tests := []struct {
		state DeviceState
		want  string
	}{
		{DeviceStateOk, "Ok"},
		{DeviceStateKNXError, "KNX error"},
		{DeviceStateIPError, "IP error"},
		{DeviceStateReserved, "Reserved"},
		{DeviceState(0xaa), "Unknown device status 0xaa"},
	}

	for _, tc := range tests {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("state %v: expected %q, got %q", uint8(tc.state), tc.want, got)
		}
	}
}

func TestRoutingLostUnpack(t *testing.T) {
	var msg RoutingLost
	data := []byte{0x04, byte(DeviceStateKNXError), 0x12, 0x34}

	n, err := msg.Unpack(data)
	if err != nil {
		t.Fatalf("Unpack() error = %v", err)
	}

	if n != uint(len(data)) {
		t.Fatalf("Unpack() bytes = %d, want %d", n, len(data))
	}

	if msg.Status != DeviceStateKNXError {
		t.Errorf("Status = %v, want %v", msg.Status, DeviceStateKNXError)
	}

	if msg.Count != 0x1234 {
		t.Errorf("Count = %#x, want 0x1234", msg.Count)
	}
}

func TestRoutingLostUnpackShort(t *testing.T) {
	var msg RoutingLost
	if _, err := msg.Unpack([]byte{0x04, byte(DeviceStateOk), 0x12}); err == nil {
		t.Fatal("expected error for short buffer")
	}
}

func TestRoutingBusyUnpack(t *testing.T) {
	var msg RoutingBusy
	data := []byte{0x06, byte(DeviceStateIPError), 0x01, 0xf4, 0x12, 0x34}

	n, err := msg.Unpack(data)
	if err != nil {
		t.Fatalf("Unpack() error = %v", err)
	}

	if n != uint(len(data)) {
		t.Fatalf("Unpack() bytes = %d, want %d", n, len(data))
	}

	if msg.Status != DeviceStateIPError {
		t.Errorf("Status = %v, want %v", msg.Status, DeviceStateIPError)
	}

	wantWait := 500 * time.Millisecond
	if msg.WaitTime != wantWait {
		t.Errorf("WaitTime = %v, want %v", msg.WaitTime, wantWait)
	}

	if msg.Control != 0x1234 {
		t.Errorf("Control = %#x, want 0x1234", msg.Control)
	}
}
