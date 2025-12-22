package gac

import (
	"bytes"
	"strings"
	"testing"

	"github.com/knx-go/knx-go/knx/cemi"
)

const sampleExchange = `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<GroupAddress-Export xmlns="http://knx.org/xml/ga-export/01">
  <GroupRange Name="Root" RangeStart="1" RangeEnd="512">
    <GroupAddress Name="light" Address="1" DPTs="DPST-1-1" />
    <GroupRange Name="Nested" RangeStart="2" RangeEnd="16">
      <GroupAddress Name="temperature" Address="15" DPTs="DPST-9-1" />
    </GroupRange>
  </GroupRange>
</GroupAddress-Export>`

const invalidDPTExchange = `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<GroupAddress-Export xmlns="http://knx.org/xml/ga-export/01">
  <GroupRange Name="Root" RangeStart="1" RangeEnd="512">
    <GroupAddress Name="invalid" Address="1" DPTs="DPST-99-999" />
  </GroupRange>
</GroupAddress-Export>`

const twoLevelExchange = `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<GroupAddress-Export xmlns="http://knx.org/xml/ga-export/01">
  <GroupRange Name="Root" RangeStart="1" RangeEnd="512">
    <GroupAddress Name="sensor" Address="257" DPTs="DPST-9-1" />
  </GroupRange>
</GroupAddress-Export>`

const freeExchange = `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
<GroupAddress-Export xmlns="http://knx.org/xml/ga-export/01">
  <GroupRange Name="Root" RangeStart="1" RangeEnd="512">
    <GroupAddress Name="raw" Address="1234" DPTs="DPST-1-1" />
  </GroupRange>
</GroupAddress-Export>`

func TestImportCatalog(t *testing.T) {
	catalog, err := Import(strings.NewReader(sampleExchange))
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	groups := catalog.Groups()
	if len(groups) != 2 {
		t.Fatalf("Groups() len = %d, want 2", len(groups))
	}

	light, ok := catalog.Lookup("light")
	if !ok {
		t.Fatalf("Lookup(light) returned false")
	}
	if light.Address != cemi.GroupAddr(1) {
		t.Errorf("light address = %s, want 0/0/1", light.Address)
	}
	if len(light.Path) != 1 || light.Path[0] != "Root" {
		t.Errorf("light path = %v, want [Root]", light.Path)
	}
	if len(light.DPTs) != 1 || light.DPTs[0] != "1.001" {
		t.Errorf("light DPTs = %v, want [1.001]", light.DPTs)
	}
	if light.DPTs[0].DPST() != "DPST-1-1" {
		t.Errorf("light original DPT = %s, want DPST-1-1", light.DPTs[0].DPST())
	}
	if light.AddressString() != "0/0/1" {
		t.Errorf("light AddressString() = %s, want 0/0/1", light.AddressString())
	}

	temp, ok := catalog.Lookup("TEMPERATURE")
	if !ok {
		t.Fatalf("Lookup(TEMPERATURE) returned false")
	}
	if temp.Address != cemi.GroupAddr(15) {
		t.Errorf("temperature address = %s, want 0/0/15", temp.Address)
	}
	if len(temp.Path) != 2 || temp.Path[0] != "Root" || temp.Path[1] != "Nested" {
		t.Errorf("temperature path = %v, want [Root Nested]", temp.Path)
	}
	if len(temp.DPTs) != 1 || temp.DPTs[0] != "9.001" {
		t.Errorf("temperature DPTs = %v, want [9.001]", temp.DPTs)
	}
	if temp.AddressString() != "0/0/15" {
		t.Errorf("temperature AddressString() = %s, want 0/0/15", temp.AddressString())
	}

	byAddr, ok := catalog.LookupByAddress(cemi.GroupAddr(15))
	if !ok {
		t.Fatalf("LookupByAddress(15) returned false")
	}
	if byAddr.Name != "temperature" {
		t.Errorf("LookupByAddress(15) name = %s, want temperature", byAddr.Name)
	}
}

func TestImportCatalogInvalidDPT(t *testing.T) {
	if _, err := Import(strings.NewReader(invalidDPTExchange)); err == nil {
		t.Fatal("Import() expected error for invalid DPT, got nil")
	}
}

func TestImportCatalogTwoLevelStyle(t *testing.T) {
	catalog, err := Import(strings.NewReader(twoLevelExchange))
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	catalog.AddressStyle = cemi.GroupAddrFormatTwoLevels

	sensor, ok := catalog.Lookup("sensor")
	if !ok {
		t.Fatalf("Lookup(sensor) returned false")
	}
	if sensor.AddressString() != "0/1/1" {
		t.Fatalf("sensor AddressString() = %s, want 0/1/1", sensor.AddressString())
	}
	if formatted := catalog.FormatAddress(sensor.Address); formatted != "0/257" {
		t.Fatalf("FormatAddress() = %s, want 0/257", formatted)
	}
}

func TestImportCatalogFreeStyle(t *testing.T) {
	catalog, err := Import(strings.NewReader(freeExchange))
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	catalog.AddressStyle = cemi.GroupAddrFormatFree

	raw, ok := catalog.Lookup("raw")
	if !ok {
		t.Fatalf("Lookup(raw) returned false")
	}
	if raw.AddressString() != "0/4/210" {
		t.Fatalf("raw AddressString() = %s, want 0/4/210", raw.AddressString())
	}
	if formatted := catalog.FormatAddress(raw.Address); formatted != "1234" {
		t.Fatalf("FormatAddress() = %s, want 1234", formatted)
	}
}

func TestExportRoundTrip(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		style cemi.GroupAddrFormat
	}{
		{name: "ThreeLevel", input: sampleExchange, style: cemi.GroupAddrFormatThreeLevels},
		{name: "TwoLevel", input: twoLevelExchange, style: cemi.GroupAddrFormatTwoLevels},
		{name: "Free", input: freeExchange, style: cemi.GroupAddrFormatFree},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			catalog, err := Import(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Import() error = %v", err)
			}

			var buf bytes.Buffer
			if err := catalog.Export(&buf); err != nil {
				t.Fatalf("Export() error = %v", err)
			}

			roundTrip, err := Import(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("Import() round-trip error = %v", err)
			}

			assertCatalogsEqual(t, catalog, roundTrip)

			var second bytes.Buffer
			if err := roundTrip.Export(&second); err != nil {
				t.Fatalf("second Export() error = %v", err)
			}

			if buf.String() != second.String() {
				t.Fatalf("exported XML mismatch after second export\nfirst:\n%s\nsecond:\n%s", buf.String(), second.String())
			}
		})
	}
}

func assertCatalogsEqual(t *testing.T, expected, actual *Catalog) {
	t.Helper()

	expectedGroups := expected.Groups()
	actualGroups := actual.Groups()

	if len(actualGroups) != len(expectedGroups) {
		t.Fatalf("round-trip group count = %d, want %d", len(actualGroups), len(expectedGroups))
	}

	for _, group := range expectedGroups {
		rt, ok := actual.Lookup(group.Name)
		if !ok {
			t.Fatalf("round-trip missing group %q", group.Name)
		}
		if rt.Address != group.Address {
			t.Fatalf("round-trip address mismatch for %q: %s != %s", group.Name, rt.Address, group.Address)
		}
		if strings.Join(rt.Path, "/") != strings.Join(group.Path, "/") {
			t.Fatalf("round-trip path mismatch for %q: %v != %v", group.Name, rt.Path, group.Path)
		}
		if len(rt.DPTs) != len(group.DPTs) {
			t.Fatalf("round-trip DPT count mismatch for %q: %d != %d", group.Name, len(rt.DPTs), len(group.DPTs))
		}
		for i := range rt.DPTs {
			if rt.DPTs[i] != group.DPTs[i] {
				t.Fatalf("round-trip DPT mismatch for %q: %s != %s", group.Name, rt.DPTs[i], group.DPTs[i])
			}
			if rt.DPTs[i].DPST() != group.DPTs[i].DPST() {
				t.Fatalf("round-trip original DPT mismatch for %q: %s != %s", group.Name, rt.DPTs[i].DPST(), group.DPTs[i].DPST())
			}
		}
		if rt.Style != group.Style {
			t.Fatalf("round-trip style mismatch for %q: %v != %v", group.Name, rt.Style, group.Style)
		}
	}
}
