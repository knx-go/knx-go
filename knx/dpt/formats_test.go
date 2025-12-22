package dpt

import (
	"errors"
	"math"
	"testing"
)

func TestUnpackB1(t *testing.T) {
	var value bool
	if err := unpackB1([]byte{0x01}, &value); err != nil {
		t.Fatalf("unpackB1() error = %v", err)
	}
	if !value {
		t.Fatalf("unpackB1() = %v, want true", value)
	}
	if err := unpackB1([]byte{}, &value); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("unpackB1() error = %v, want %v", err, ErrInvalidLength)
	}
}

func TestUnpackB2ReservedBits(t *testing.T) {
	var b0, b1 bool
	if err := unpackB2(0x03, &b0, &b1); err != nil {
		t.Fatalf("unpackB2() unexpected error = %v", err)
	}
	if !b0 || !b1 {
		t.Fatalf("unpackB2() values = %v, %v, want true, true", b0, b1)
	}
	if err := unpackB2(0x10, &b0, &b1); !errors.Is(err, ErrBadReservedBits) {
		t.Fatalf("unpackB2() error = %v, want %v", err, ErrBadReservedBits)
	}
}

func TestPackUnpackF16(t *testing.T) {
	samples := []float32{-123.45, 0, 327.68}
	for _, sample := range samples {
		data := packF16(sample)
		var decoded float32
		if err := unpackF16(data, &decoded); err != nil {
			t.Fatalf("unpackF16(%v) error = %v", sample, err)
		}
		if math.Abs(float64(decoded-sample)) > 0.5 {
			t.Fatalf("round trip for %v = %v", sample, decoded)
		}
	}
	var decoded float32
	if err := unpackF16([]byte{0x00}, &decoded); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("unpackF16 short error = %v, want %v", err, ErrInvalidLength)
	}
}

func TestPackUnpackU16(t *testing.T) {
	value := uint16(0xCAFE)
	data := packU16(value)
	var decoded uint16
	if err := unpackU16(data, &decoded); err != nil {
		t.Fatalf("unpackU16() error = %v", err)
	}
	if decoded != value {
		t.Fatalf("unpackU16() = 0x%X, want 0x%X", decoded, value)
	}
	if err := unpackU16([]byte{0x01, 0x02}, &decoded); !errors.Is(err, ErrInvalidLength) {
		t.Fatalf("unpackU16 short error = %v, want %v", err, ErrInvalidLength)
	}
}
