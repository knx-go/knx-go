package dpt

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Datapoint represents a datapoint type. The format is the identifier (e.g. "1.001")
type DataPointType string

// Produce returns a new datapoint value instance for the datapoint type.
func (d DataPointType) Produce() (DatapointValue, bool) {
	return Produce(string(d))
}

func (d DataPointType) DPST() string {
	dpt, k := d.Produce()
	if !k {
		return ""
	}
	t := reflect.TypeOf(dpt).Elem()
	name := t.Name()

	tn, _ := strconv.Atoi(name[4 : len(name)-3])
	sn, _ := strconv.Atoi(name[len(name)-3:])

	return fmt.Sprintf("DPST-%d-%d", tn, sn)
}

func NormaliseDPTID(raw string) (string, error) {
	// Normalize
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	if normalized == "" {
		return "", fmt.Errorf("invalid datapoint identifier %q, stirng is empty", raw)
	}

	// Split
	var parts []string
	if strings.HasPrefix(normalized, "DPST-") || strings.HasPrefix(normalized, "DPT-") {
		parts = strings.Split(normalized, "-")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid datapoint identifier %q, DPST- format requires 2 dashes", raw)
		}
		parts = parts[1:]
	}
	if strings.Contains(normalized, ".") {
		parts = strings.Split(normalized, ".")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid datapoint identifier %q, dot format requires 1 dot", raw)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid datapoint identifier %q, neither dot nor dash format", raw)
	}

	// Segments
	main, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid datapoint identifier %q, first part not parsable %v", raw, err)
	}

	sub, err := strconv.Atoi(strings.Join(parts[1:], ""))
	if err != nil {
		return "", fmt.Errorf("invalid datapoint identifier %q, second part not parsable: %v", raw, err)
	}

	// Validate
	if main < 0 || sub < 0 {
		return "", fmt.Errorf("invalid datapoint identifier %d.%d", main, sub)
	}

	// Print
	if sub < 1000 {
		return fmt.Sprintf("%d.%03d", main, sub), nil
	}
	return fmt.Sprintf("%d.%d", main, sub), nil
}
