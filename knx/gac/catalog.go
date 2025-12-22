package gac

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knx-go/knx-go/knx/cemi"
)

// Catalog represents a collection of group addresses parsed from or exported to
// a KNX group address exchange document.
type Catalog struct {
	byName       map[string]*Group
	byAddress    map[string]*Group
	ordered      []*Group
	AddressStyle cemi.GroupAddrFormat
}

// Groups returns all known group addresses as value copies.
func (c *Catalog) Groups() []Group {
	groups := make([]Group, 0, len(c.ordered))
	for _, g := range c.ordered {
		groups = append(groups, cloneGroup(g))
	}
	return groups
}

// FormatAddress renders the provided address using the catalog's addressing
// style.
func (c *Catalog) FormatAddress(addr cemi.GroupAddr) string {
	as := c.AddressStyle
	if as == cemi.GroupAddrFormatUnknown {
		as = cemi.GroupAddrFormatThreeLevels
	}
	return addr.Format(as)
}

// Lookup returns the group definition for the provided name. Lookups are case
// insensitive.
func (c *Catalog) Lookup(name string) (*Group, bool) {
	if c == nil {
		return nil, false
	}
	g, ok := c.byName[strings.ToLower(strings.TrimSpace(name))]
	return g, ok
}

// LookupByAddress returns the group definition that matches the provided group
// address.
func (c *Catalog) LookupByAddress(addr cemi.GroupAddr) (*Group, bool) {
	if c == nil {
		return nil, false
	}
	g, ok := c.byAddress[addr.String()]
	return g, ok
}

func (c *Catalog) walkRange(rng exchangeRange, path []string) error {
	currentPath := append(append([]string(nil), path...), strings.TrimSpace(rng.Name))

	for _, addr := range rng.Addresses {
		if err := c.addGroup(addr, currentPath); err != nil {
			return err
		}
	}

	for _, child := range rng.Ranges {
		if err := c.walkRange(child, currentPath); err != nil {
			return err
		}
	}

	return nil
}

func (c *Catalog) addGroup(addr exchangeGroup, path []string) error {
	name := strings.TrimSpace(addr.Name)
	if name == "" {
		return fmt.Errorf("gac: group address without name in path %q", strings.Join(path, "/"))
	}

	normalized := strings.ToLower(name)
	if _, exists := c.byName[normalized]; exists {
		return fmt.Errorf("gac: duplicated group name %q", name)
	}

	rawValue, err := strconv.ParseUint(strings.TrimSpace(addr.Address), 10, 16)
	if err != nil {
		return fmt.Errorf("gac: invalid address for group %q: %w", name, err)
	}
	if rawValue == 0 {
		return fmt.Errorf("gac: group %q has invalid address 0", name)
	}

	datapoints, err := parseDPTs(addr.DPTs)
	if err != nil {
		return fmt.Errorf("gac: group %q has invalid datapoint definition: %w", name, err)
	}

	group := &Group{
		Name:    name,
		Address: cemi.GroupAddr(rawValue),
		DPTs:    datapoints,
		Path:    append([]string(nil), path...),
		Style:   c.AddressStyle,
	}

	if _, exists := c.byAddress[group.Address.String()]; exists {
		return fmt.Errorf("gac: duplicated group address %s", group.Address)
	}

	c.byName[normalized] = group
	c.byAddress[group.Address.String()] = group
	c.ordered = append(c.ordered, group)

	return nil
}
