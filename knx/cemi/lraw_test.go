package cemi

import "testing"

func TestLRawUnpackAllocatesAndCopies(t *testing.T) {
	var frame LRaw
	input := []byte{0x01, 0x02, 0x03}

	n, err := frame.Unpack(input)
	if err != nil {
		t.Fatalf("Unpack() error = %v", err)
	}
	if n != uint(len(input)) {
		t.Fatalf("Unpack() bytes = %d, want %d", n, len(input))
	}
	if len(frame) != len(input) {
		t.Fatalf("Unpack() len = %d, want %d", len(frame), len(input))
	}
	for i, b := range frame {
		if b != input[i] {
			t.Fatalf("frame[%d] = %d, want %d", i, b, input[i])
		}
	}
	if &frame[0] == &input[0] {
		t.Fatal("Unpack() reused caller buffer")
	}
}

func TestLRawUnpackReusesCapacity(t *testing.T) {
	frame := LRaw(make([]byte, 5))
	input := []byte{0x0A, 0x0B}

	n, err := frame.Unpack(input)
	if err != nil {
		t.Fatalf("Unpack() error = %v", err)
	}
	if n != uint(len(input)) {
		t.Fatalf("Unpack() bytes = %d, want %d", n, len(input))
	}
	if cap(frame) != 5 {
		t.Fatalf("cap(frame) = %d, want 5", cap(frame))
	}
	for i := range input {
		if frame[i] != input[i] {
			t.Fatalf("frame[%d] = %d, want %d", i, frame[i], input[i])
		}
	}
}

func TestLRawMessageCodes(t *testing.T) {
	if code := (LRawReq{}).MessageCode(); code != LRawReqCode {
		t.Fatalf("LRawReq MessageCode = %v, want %v", code, LRawReqCode)
	}
	if code := (LRawCon{}).MessageCode(); code != LRawConCode {
		t.Fatalf("LRawCon MessageCode = %v, want %v", code, LRawConCode)
	}
	if code := (LRawInd{}).MessageCode(); code != LRawConCode {
		t.Fatalf("LRawInd MessageCode = %v, want %v", code, LRawConCode)
	}
}
