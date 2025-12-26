package dpt

import "testing"

func TestEncodeDPTFromString_OK(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dptName string
		in      string
		want    []byte
	}{
		{
			name:    "bool DPT 1.001 true",
			dptName: "1.001",
			in:      "true",
			want:    DPT_1001(true).Pack(),
		},
		{
			name:    "uint8 DPT 17.001 34",
			dptName: "17.001",
			in:      "34",
			want:    DPT_17001(34).Pack(),
		},
		{
			name:    "float DPT 9.001 21.5",
			dptName: "9.001",
			in:      "21.5",
			want:    DPT_9001(21.5).Pack(),
		},
		{
			name:    "string DPT 16.001 ciao",
			dptName: "16.001",
			in:      "test",
			want:    DPT_16001("test").Pack(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := EncodeDPTFromStringN(tc.dptName, tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != string(tc.want) {
				t.Fatalf("bytes mismatch:\n got:  %v\n want: %v", got, tc.want)
			}
		})
	}
}

func TestEncodeDPTFromString_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dptName string
		in      string
	}{
		{
			name:    "unsupported DPT",
			dptName: "999.999",
			in:      "1",
		},
		{
			name:    "invalid bool",
			dptName: "1.001",
			in:      "notabool",
		},
		{
			name:    "overflow uint8 for DPT 17.001",
			dptName: "17.001",
			in:      "999",
		},
		{
			name:    "complex struct DPT not supported by generic parser",
			dptName: "242.600",
			in:      "whatever",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := EncodeDPTFromStringN(tc.dptName, tc.in)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
