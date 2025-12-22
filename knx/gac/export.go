package gac

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/knx-go/knx-go/knx/dpt"
)

// Export encodes the catalog as a KNX group address exchange document.
func (c *Catalog) Export(w io.Writer) error {
	if c == nil {
		return fmt.Errorf("gac: catalog is nil")
	}

	root := newRangeNode("")
	for _, group := range c.ordered {
		if group == nil {
			continue
		}

		node := root
		for _, segment := range group.Path {
			node = node.child(segment)
		}
		node.Groups = append(node.Groups, group)
	}

	root.finalize()

	doc := exchangeDocument{XMLNS: namespace}
	for _, child := range root.children {
		if !child.hasRange {
			continue
		}
		doc.Ranges = append(doc.Ranges, child.toExchangeRange())
	}

	if _, err := io.WriteString(w, xml.Header); err != nil {
		return fmt.Errorf("gac: failed to write XML header: %w", err)
	}

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		return fmt.Errorf("gac: failed to encode catalog: %w", err)
	}
	if err := encoder.Flush(); err != nil {
		return fmt.Errorf("gac: failed to flush encoder: %w", err)
	}

	return nil
}

func formatDPTsForExport(dpts []dpt.DataPointType) string {
	if len(dpts) == 0 {
		return ""
	}

	rendered := make([]string, 0, len(dpts))
	for _, dp := range dpts {
		value := dp.DPST()
		if value != "" {
			rendered = append(rendered, value)
		}
	}

	if len(rendered) == 0 {
		return ""
	}

	return strings.Join(rendered, ", ")
}
