package main

import "testing"

func TestEncodeDatapointValueBooleanNumeric(t *testing.T) {
	payload, err := encodeDatapointValue("1.001", "0")
	if err != nil {
		t.Fatalf("encodeDatapointValue returned error: %v", err)
	}

	if len(payload) != 1 || payload[0] != 0x00 {
		t.Fatalf("expected payload to contain a single zero byte, got %v", payload)
	}

	payload, err = encodeDatapointValue("1.001", "1")
	if err != nil {
		t.Fatalf("encodeDatapointValue returned error: %v", err)
	}

	if len(payload) != 1 || payload[0] != 0x01 {
		t.Fatalf("expected payload to contain a single 0x01 byte, got %v", payload)
	}
}
