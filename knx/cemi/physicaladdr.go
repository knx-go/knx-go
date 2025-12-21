package cemi

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type PhysicalAddressFormat uint8

const (
	PhysicalAddressFormatUnknown PhysicalAddressFormat = iota
	PhysicalAddressFormatFree
	PhysicalAddressFormatTwoLevels
	PhysicalAddressFormatThreeLevels
)

// PhysicalAddr is an individual address for a KNX device. Its format is defined in
// 03_03_02 Data Link Layer General v01.02.02 AS.pdf
// 1.4.2 Individual Address
// Individual address zero (0.0.0) is not allowed.
type PhysicalAddr uint16

// NewPhysicalAddrString parses the given string to an individual address.
// Supported formats are
// %d.%d.%d ([0..15], [0..15], [0..255]),
// %d.%d ([0..255], [0..255]) and
// %d ([0..65535]). Validity is checked.
func NewPhysicalAddrString(addr string) (PhysicalAddr, error) {
	var nums []int

	for s := range strings.SplitSeq(addr, ".") {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		nums = append(nums, i)
	}

	switch len(nums) {
	case 3:
		if nums[0] < 0 || nums[0] > 15 ||
			nums[1] < 0 || nums[1] > 15 ||
			nums[2] < 0 || nums[2] > 255 {
			return 0, fmt.Errorf("invalid area, line or device address in %s", addr)
		}
		if nums[0] == 0 && nums[1] == 0 && nums[2] == 0 {
			return 0, errors.New("invalid individual address 0.0.0")
		}
		return PhysicalAddr(uint8(nums[0])&0xF)<<12 | PhysicalAddr(uint8(nums[1])&0xF)<<8 | PhysicalAddr(uint8(nums[2])), nil
	case 2:
		if nums[0] < 0 || nums[0] > 255 || nums[1] < 0 || nums[1] > 255 {
			return 0, fmt.Errorf("invalid subnetwork or device address in %s", addr)
		}
		if nums[0] == 0 && nums[1] == 0 {
			return 0, errors.New("invalid individual address 0.0")
		}
		return PhysicalAddr(uint8(nums[0]))<<8 | PhysicalAddr(uint8(nums[1])), nil
	case 1:
		if nums[0] <= 0 || nums[0] > 65535 {
			return 0, fmt.Errorf("invalid raw individual address in %s", addr)
		}
		return PhysicalAddr(nums[0]), nil
	}

	return 0, errors.New("string cannot be parsed to an individual address")
}

// String generates a string representation "a.b.c" where
// a = Area Address = 4 bits
// b = Line Address = 4 bits
// c = Device Address = 8 bits
func (addr PhysicalAddr) String() string {
	return addr.Format(PhysicalAddressFormatThreeLevels)
}

// Format individual address
func (addr PhysicalAddr) Format(f PhysicalAddressFormat) string {
	switch f {
	case PhysicalAddressFormatFree:
		return strconv.FormatUint(uint64(addr), 10)
	case PhysicalAddressFormatTwoLevels:
		main := uint8(addr>>12) & 0x1F
		sub := uint16(addr) & 0x7FF
		return fmt.Sprintf("%d.%d", main, sub)
	case PhysicalAddressFormatThreeLevels:
		fallthrough
	default:
		main := uint8(addr>>12) & 0xF
		middle := uint8(addr>>8) & 0xF
		sub := uint8(addr)
		return fmt.Sprintf("%d.%d.%d", main, middle, sub)
	}
}
