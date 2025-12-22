package dpt

import "testing"

func TestNormalizeDPTFormats(t *testing.T) {
	cases := map[string]string{
		"dpst-1-1": "1.001",
		"DPT_9-1":  "9.001",
		"1.23":     "1.023",
	}
	for input, want := range cases {
		got, err := NormaliseDPTID(input)
		if err != nil {
			t.Fatalf("normalizeDPT(%q) error = %v", input, err)
		}
		if got != want {
			t.Fatalf("normalizeDPT(%q) = %q, want %q", input, got, want)
		}
	}
	if _, err := NormaliseDPTID("invalid"); err == nil {
		t.Fatal("normalizeDPT() expected error for invalid input")
	}
}
