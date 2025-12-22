package dpt

import "testing"

func TestListSupportedTypesContainsKnownDatapoints(t *testing.T) {
	t.Parallel()

	types := ListSupportedTypes()
	if len(types) == 0 {
		t.Fatal("expected supported types list to be non-empty")
	}

	want := map[string]bool{"1.001": false, "5.001": false, "28.001": false}
	for _, name := range types {
		if _, ok := want[name]; ok {
			want[name] = true
		}
	}

	for name, seen := range want {
		if !seen {
			t.Fatalf("expected type %s to be registered", name)
		}
	}
}

func TestProduceReturnsFreshInstances(t *testing.T) {
	t.Parallel()

	value, ok := Produce("1.001")
	if !ok {
		t.Fatalf("expected to produce datapoint 1.001")
	}
	if _, ok := value.(*DPT_1001); !ok {
		t.Fatalf("expected *DPT_1001, got %T", value)
	}

	another, ok := Produce("1.001")
	if !ok {
		t.Fatalf("expected to produce datapoint 1.001 on subsequent call")
	}
	if value == another {
		t.Fatal("expected Produce to return distinct instances")
	}

	if _, ok := Produce("0.000"); ok {
		t.Fatal("expected unknown datapoint to return ok=false")
	}
}
