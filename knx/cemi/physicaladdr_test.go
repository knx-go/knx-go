package cemi

import (
	"testing"
)

func Test_PhysicalAddresses(t *testing.T) {
	type Addr struct {
		Src     string
		Valid   bool
		Printed string
	}

	addrs := []Addr{
		{"1.2.3", true, "1.2.3"},
		{"1.3.255", true, "1.3.255"},
		{"1.3.0", true, "1.3.0"},
		{"75.235", true, "4.11.235"},
		{"65535", true, "15.15.255"},
		{"15.15.255", true, "15.15.255"},
		{"15.15.0", true, "15.15.0"},
		{"13057", true, "3.3.1"},
		{"16.17.255", false, ""},
		{"1..0", false, ""},
		{"15.15.", false, ""},
		{" . .15", false, ""},
		{"18.15.240", false, ""},
		{"1.3.450", false, ""},
		{"1.450", false, ""},
		{"-2", false, ""},
		{".400", false, ""},
		{"-11.0.0", false, ""},
		{"0.0.0", false, ""},
		{"0.0", false, ""},
		{"0", false, ""},
	}

	for _, a := range addrs {
		ia, err := NewPhysicalAddrString(a.Src)
		if a.Valid {
			if err != nil {
				t.Errorf("%#v has error %s.", a.Src, err)
			} else if ia.String() != a.Printed {
				t.Errorf("%#v wrongly parsed: %#v is differenct form expected (%#v)", a.Src, ia.String(), a.Printed)
			}
		} else if err == nil {
			t.Errorf("%#v invalid parsed.", a.Src)
		}
	}
}
