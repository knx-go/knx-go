package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/knx-go/knx-go/knx/gac"
)

func loadCatalog(path string) (*gac.Catalog, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil, nil
	}

	file, err := os.Open(trimmed)
	if err != nil {
		return nil, fmt.Errorf("failed to open group file %q: %w", trimmed, err)
	}
	defer file.Close()

	catalog, err := gac.Import(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group file %q: %w", trimmed, err)
	}

	return catalog, nil
}
