package gac

import "strconv"

type rangeNode struct {
	Name       string
	Groups     []*Group
	children   []*rangeNode
	index      map[string]*rangeNode
	hasRange   bool
	rangeStart uint16
	rangeEnd   uint16
}

func newRangeNode(name string) *rangeNode {
	return &rangeNode{Name: name}
}

func (n *rangeNode) child(name string) *rangeNode {
	if n.index == nil {
		n.index = make(map[string]*rangeNode)
	}
	if child, ok := n.index[name]; ok {
		return child
	}
	child := newRangeNode(name)
	n.children = append(n.children, child)
	n.index[name] = child
	return child
}

func (n *rangeNode) finalize() (uint16, uint16, bool) {
	var minAddr, maxAddr uint16
	var has bool

	for _, group := range n.Groups {
		addr := uint16(group.Address)
		if !has || addr < minAddr {
			minAddr = addr
		}
		if !has || addr > maxAddr {
			maxAddr = addr
		}
		has = true
	}

	for _, child := range n.children {
		childMin, childMax, childHas := child.finalize()
		if !childHas {
			continue
		}
		if !has || childMin < minAddr {
			minAddr = childMin
		}
		if !has || childMax > maxAddr {
			maxAddr = childMax
		}
		has = true
	}

	if has {
		n.hasRange = true
		n.rangeStart = minAddr
		n.rangeEnd = maxAddr
	}

	return minAddr, maxAddr, has
}

func (n *rangeNode) toExchangeRange() exchangeRange {
	rng := exchangeRange{
		Name:       n.Name,
		RangeStart: n.rangeStart,
		RangeEnd:   n.rangeEnd,
	}

	if len(n.Groups) > 0 {
		rng.Addresses = make([]exchangeGroup, 0, len(n.Groups))
		for _, group := range n.Groups {
			rng.Addresses = append(rng.Addresses, exchangeGroup{
				Name:    group.Name,
				Address: strconv.FormatUint(uint64(group.Address), 10),
				DPTs:    formatDPTsForExport(group.DPTs),
			})
		}
	}

	for _, child := range n.children {
		if !child.hasRange {
			continue
		}
		rng.Ranges = append(rng.Ranges, child.toExchangeRange())
	}

	return rng
}
