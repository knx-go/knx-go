package api

import (
	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/gac"
)

type Options struct {
	// Catalog optionally provides group metadata for name lookups
	Catalog *gac.Catalog
	// ListenAddress defines the HTTP listen address. Defaults to ":8080"
	ListenAddress string
	// PgKNX Store to leverage
	Store Store
	// knx.GroupTunnel to communicate with the KNX bus
	Tunnel knx.GroupTunnel
}
