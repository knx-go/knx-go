package api

import (
	"context"

	"github.com/knx-go/knx-go/pgknx"
)

type Store interface {
	CreateEvent(context.Context, pgknx.Event) error
	GroupLastEvent(context.Context, string) (pgknx.Event, error)
	GroupEventsHistory(context.Context, string, pgknx.HistoryOptions) ([]pgknx.Event, error)
}
