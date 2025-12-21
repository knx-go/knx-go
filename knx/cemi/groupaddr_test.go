package cemi

import (
	"testing"
)

func Test_GroupAddresses(t *testing.T) {
	type Addr struct {
		Src     string
		Valid   bool
		Printed string
	}

	addrs := []Addr{
		{"1/2/3", true, "1/2/3"},
		{"31/7/255", true, "31/7/255"},
		{"31/2040", true, "3/7/248"},
		{"65535", true, "31/7/255"},
		{"82/8/260", false, ""},
		{"84/230", false, ""},
		{"31/2060", false, ""},
		{"0/0/0", false, ""},
		{"0/0", false, ""},
		{"0", false, ""},
		{"123/foobar", false, ""},
		{"1000/2000/3000", false, ""},
	}

	for _, a := range addrs {
		ga, err := NewGroupAddrString(a.Src)
		if a.Valid {
			if err != nil {
				t.Errorf("%#v has error %s.", a.Src, err)
			} else if ga.String() != a.Printed {
				t.Errorf("%#v wrongly parsed: %#v is differenct form expected (%#v)", a.Src, ga.String(), a.Printed)
			}
		} else if err == nil {
			t.Errorf("%#v invalid parsed.", a.Src)
		}
	}
}
