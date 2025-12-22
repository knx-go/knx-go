package gac

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/knx-go/knx-go/knx/dpt"
)

// Import parses the provided reader as a KNX group address exchange document
// and returns the resulting catalog.
func Import(r io.Reader) (*Catalog, error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charsetReader

	var doc exchangeDocument
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("gac: decode error: %w", err)
	}

	catalog := &Catalog{
		byName:    make(map[string]*Group),
		byAddress: make(map[string]*Group),
	}

	for _, rng := range doc.Ranges {
		if err := catalog.walkRange(rng, nil); err != nil {
			return nil, err
		}
	}

	return catalog, nil
}

func parseDPTs(raw string) ([]dpt.DataPointType, error) {
	fields := strings.Split(raw, ",")
	result := make([]dpt.DataPointType, 0, len(fields))
	for _, field := range fields {
		canonical, err := dpt.NormaliseDPTID(field)
		if err != nil {
			return nil, err
		}

		if _, ok := dpt.Produce(canonical); !ok {
			return nil, fmt.Errorf("gac: unknown datapoint type %q", field)
		}
		result = append(result, dpt.DataPointType(canonical))
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

// charsetReader is a minimal wrapper to allow parsing UTF-8 documents without
// pulling in x/net/html/charset when not needed.
func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	// ETS typically writes UTF-8 which is handled transparently by the XML
	// decoder. Any other charset is unsupported at the moment.
	charset = strings.TrimSpace(strings.ToLower(charset))
	if charset != "" && charset != "utf-8" && charset != "utf8" {
		return nil, fmt.Errorf("gac: unsupported charset %q", charset)
	}
	return input, nil
}
