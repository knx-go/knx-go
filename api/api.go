package api

import (
	"context"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/gac"
)

type api struct {
	catalog *gac.Catalog
	ctx     context.Context
	store   Store
	tunnel  knx.GroupTunnel
}
