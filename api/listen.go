package api

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/gac"
	"github.com/knx-go/knx-go/pgknx"
)

func (a *api) listen(ctx context.Context) {
	in := a.tunnel.Inbound()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-in:
			if !ok {
				return
			}

			if err := a.store.CreateEvent(ctx, eventFromDetails(event, a.catalog)); err != nil {
				fmt.Fprintf(os.Stderr, "failed to record event: %v\n", err)
			}
		}
	}
}

func eventFromDetails(event knx.GroupEvent, catalog *gac.Catalog) pgknx.Event {
	recorded := pgknx.Event{
		Timestamp:   time.Now().UTC(),
		Command:     event.Command.String(),
		Source:      event.Source.String(),
		Data:        event.Data,
		Destination: event.Destination.String(),
	}
	if catalog == nil {
		return recorded
	}

	recorded.Destination = catalog.FormatAddress(event.Destination)
	group, ok := catalog.LookupByAddress(event.Destination)
	if !ok || group == nil {
		return recorded
	}
	recorded.Group = group.Name
	if len(group.DPTs) > 0 {
		recorded.DPT = string(group.DPTs[0])
	}
	if description, ok := describeDatapoints(group, event.Data); ok {
		recorded.Decoded = description
	}
	return recorded
}

func describeDatapoints(group *gac.Group, data []byte) (string, bool) {
	for _, dp := range group.DPTs {
		value, ok := dp.Produce()
		if !ok {
			continue
		}

		if err := value.Unpack(data); err != nil {
			continue
		}

		return fmt.Sprint(value), true
	}
	return "", false
}
