package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/cemi"
	"github.com/knx-go/knx-go/knx/gac"
	"github.com/knx-go/knx-go/pgknx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen for KNX group events",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applyListenConfig(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listen()
		},
	}

	cmd.Flags().StringVarP(&groupFile, "group-file", "f", "", "path to a KNX group address export (XML)")
	cmd.Flags().StringVarP(&postgresqlDSN, "postgresql-dsn", "d", "", "PostgreSQL Data Source Name used to persist events")

	root.AddCommand(cmd)
}

func applyListenConfig(cmd *cobra.Command) {
	if value := strings.TrimSpace(viper.GetString("listen.group_file")); value != "" && !flagChanged(cmd, "group-file") {
		groupFile = value
	}
	if value := strings.TrimSpace(viper.GetString("listen.postrgesql-dsn")); value != "" && !flagChanged(cmd, "postgresql-dsn") {
		postgresqlDSN = value
	}
}

func listen() error {
	catalog, err := loadCatalog(groupFile)
	if err != nil {
		return err
	}
	if catalog != nil {
		style := catalog.AddressStyle.String()
		if style == "" {
			style = cemi.GroupAddrFormatThreeLevels.String()
		}
		fmt.Printf("Loaded %d group addresses from %s (%s addressing)\n", len(catalog.Groups()), groupFile, style)
	}

	var store *pgknx.Store
	if len(postgresqlDSN) > 0 {
		store, err = pgknx.New(postgresqlDSN, viper.GetBool("debug"))
		if err != nil {
			return err
		}
		defer store.Close()
		fmt.Println("Database recording enabled")
	}

	for {
		client, err := knx.NewGroupTunnel(fmt.Sprintf("%s:%s", server, port), knx.DefaultTunnelConfig)
		if err != nil {
			fmt.Printf("Error while creating: %v\n", err)
			time.Sleep(time.Second)
			continue
		}
		defer client.Close()

		for event := range client.Inbound() {
			timestamp := time.Now().Format(time.RFC3339Nano)
			destination := event.Destination.String()
			if catalog != nil {
				group, ok := catalog.LookupByAddress(event.Destination)
				if ok || group != nil {
					destination = fmt.Sprintf("%s (%s)", catalog.FormatAddress(event.Destination), group.Name)
					description, _ := describeDatapoints(group, event.Data)
					fmt.Printf("[%s] %s %s -> %s %s\n", timestamp, event.Command, event.Source, destination, description)
				}
			} else {
				fmt.Printf("[%s] %s %s -> %s %v\n", timestamp, event.Command, event.Source, destination, event.Data)
			}

			if store != nil {
				if err := store.CreateEvent(context.Background(), eventFromDetails(event, catalog)); err != nil {
					fmt.Fprintf(os.Stderr, "failed to record event: %v\n", err)
				}
			}
		}

		fmt.Println("tunnel channel closed")
		return errors.New("tunnel channel closed")
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
