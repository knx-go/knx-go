package gac

import (
	"strings"
	"testing"
)

func TestParseDPTsCanonicalizesValues(t *testing.T) {
	raw := " 1.001,DPST-9-1,dpst_5-1 "
	dpts, err := parseDPTs(raw)
	if err != nil {
		t.Fatalf("parseDPTs() error = %v", err)
	}
	if len(dpts) != 3 {
		t.Fatalf("parseDPTs() len = %d, want 3", len(dpts))
	}
	expected := []string{"1.001", "9.001", "5.001"}
	for i, dp := range dpts {
		if string(dp) != expected[i] {
			t.Fatalf("dpts[%d].Type = %s, want %s", i, dp, expected[i])
		}
	}
}

func TestCharsetReader(t *testing.T) {
	data := strings.NewReader("test")
	r, err := charsetReader("utf-8", data)
	if err != nil {
		t.Fatalf("charsetReader() error = %v", err)
	}
	if r != data {
		t.Fatalf("charsetReader() returned new reader")
	}
	if _, err := charsetReader("latin1", data); err == nil {
		t.Fatal("charsetReader() expected error for unsupported charset")
	}
}
