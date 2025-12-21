package cemi

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type GroupAddrFormat uint8

const (
	GroupAddrFormatUnknown GroupAddrFormat = iota
	GroupAddrFormatFree
	GroupAddrFormatTwoLevels
	GroupAddrFormatThreeLevels
)

// String returns the canonical textual representation of the address style.
func (f GroupAddrFormat) String() string {
	switch f {
	case GroupAddrFormatFree:
		return "Free"
	case GroupAddrFormatTwoLevels:
		return "TwoLevel"
	case GroupAddrFormatThreeLevels:
		return "ThreeLevel"
	default:
		return ""
	}
}

func NewGroupAddrFormat(s string) (GroupAddrFormat, error) {
	switch s {
	case "":
		return GroupAddrFormatUnknown, nil
	case "Free":
		return GroupAddrFormatFree, nil
	case "TwoLevel":
		return GroupAddrFormatTwoLevels, nil
	case "ThreeLevel":
		return GroupAddrFormatThreeLevels, nil
	default:
		return GroupAddrFormatUnknown, fmt.Errorf("format %v not valid", s)
	}
}

// GroupAddr is an address for a KNX group object. Group address
// zero (0/0/0) is not allowed.
type GroupAddr uint16

// NewGroupAddr3 generates a group address from an "a/b/c"
// representation, where a is the Main Group [0..31], b is
// the Middle Group [0..7], c is the Sub Group [0..255].
func NewGroupAddr3(a, b, c uint8) GroupAddr {
	return GroupAddr(a&0x1F)<<11 | GroupAddr(b&0x7)<<8 | GroupAddr(c)
}

// NewGroupAddr2 generates a group address from and "a/b"
// representation, where a is the Main Group [0..31] and b is
// the Sub Group [0..2047].
func NewGroupAddr2(a uint8, b uint16) GroupAddr {
	return GroupAddr(a)<<8 | GroupAddr(b&0x7FF)
}

// NewGroupAddrString parses the given string to a group address.
// Supported formats are:
// %d/%d/%d ([0..31], [0..7], [0..255]),
// %d/%d ([0..31], [0..2047]) and
// %d ([0..65535]). Validity is checked.
func NewGroupAddrString(addr string) (GroupAddr, error) {
	var nums []int

	for s := range strings.SplitSeq(addr, "/") {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, fmt.Errorf("invalid parsing %s", err)
		}
		nums = append(nums, i)
	}

	switch len(nums) {
	case 3:
		if nums[0] < 0 || nums[0] > 31 ||
			nums[1] < 0 || nums[1] > 7 ||
			nums[2] < 0 || nums[2] > 255 {
			return 0, fmt.Errorf("invalid main, middle or sub group address in %s", addr)
		}
		if nums[0] == 0 && nums[1] == 0 && nums[2] == 0 {
			return 0, errors.New("invalid group address 0/0/0")
		}
		return NewGroupAddr3(uint8(nums[0]), uint8(nums[1]), uint8(nums[2])), nil
	case 2:
		if nums[0] < 0 || nums[0] > 31 ||
			nums[1] < 0 || nums[1] > 2047 {
			return 0, fmt.Errorf("invalid main or sub group address in %s", addr)
		}
		if nums[0] == 0 && nums[1] == 0 {
			return 0, errors.New("invalid group address 0/0")
		}
		return NewGroupAddr2(uint8(nums[0]), uint16(nums[1])), nil
	case 1:
		if nums[0] <= 0 || nums[0] > 65535 {
			return 0, fmt.Errorf("invalid raw group address in %s", addr)
		}
		return GroupAddr(nums[0]), nil
	}

	return 0, errors.New("string cannot be parsed to a group address")
}

// String generates a string representation with groups "a/b/c" where
// a = Main Group = 5 bits, b = Middle Group = 3 bits, c = Sub Group = 1 byte.
func (addr GroupAddr) String() string {
	return addr.Format(GroupAddrFormatThreeLevels)
}

// Format group address
func (addr GroupAddr) Format(f GroupAddrFormat) string {
	switch f {
	case GroupAddrFormatFree:
		return strconv.FormatUint(uint64(addr), 10)
	case GroupAddrFormatTwoLevels:
		main := uint8(addr>>11) & 0x1F
		sub := uint16(addr) & 0x7FF
		return fmt.Sprintf("%d/%d", main, sub)
	case GroupAddrFormatThreeLevels:
		fallthrough
	default:
		main := uint8(addr>>11) & 0x1F
		middle := uint8(addr>>8) & 0x7
		sub := uint8(addr)
		return fmt.Sprintf("%d/%d/%d", main, middle, sub)
	}
}
