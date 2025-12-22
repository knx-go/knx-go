package gac

import (
	"github.com/knx-go/knx-go/knx/cemi"
	"github.com/knx-go/knx-go/knx/dpt"
)

// Group describes a single group address that was declared in the exchange
// file.
type Group struct {
	Name    string
	Address cemi.GroupAddr
	DPTs    []dpt.DataPointType
	Path    []string
	Style   cemi.GroupAddrFormat
}

// AddressString returns the textual representation of the group address using
// the catalog's address style.
func (g *Group) AddressString() string {
	if g == nil {
		return ""
	}
	return g.Address.Format(g.Style)
}

func cloneGroup(g *Group) Group {
	cloned := Group{
		Name:    g.Name,
		Address: g.Address,
		Path:    append([]string(nil), g.Path...),
		Style:   g.Style,
	}
	if len(g.DPTs) > 0 {
		cloned.DPTs = append([]dpt.DataPointType(nil), g.DPTs...)
	}
	return cloned
}
