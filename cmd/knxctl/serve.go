package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/knx-go/knx-go/api"
	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/pgknx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Expose KNX group operations over HTTP",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applyServeConfig(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&serveListenAddr, "listen", "l", ":8080", "HTTP listen address")
	cmd.Flags().StringVarP(&groupFile, "group-file", "f", "", "path to a KNX group address export (XML)")
	cmd.Flags().StringVarP(&postgresqlDSN, "postgresql-dsn", "d", "", "PostgreSQL Data Source Name used to persist events")

	root.AddCommand(cmd)
}

func applyServeConfig(cmd *cobra.Command) {
	if value := strings.TrimSpace(viper.GetString("serve.listen")); value != "" && !flagChanged(cmd, "listen") {
		serveListenAddr = value
	}
	if value := strings.TrimSpace(viper.GetString("serve.group_file")); value != "" && !flagChanged(cmd, "group-file") {
		groupFile = value
	}
	if value := strings.TrimSpace(viper.GetString("serve.postrgesql-dsn")); value != "" && !flagChanged(cmd, "postgresql-dsn") {
		postgresqlDSN = value
	}
}

func runServe(ctx context.Context) error {
	catalog, err := loadCatalog(groupFile)
	if err != nil {
		return err
	}
	if catalog != nil {
		fmt.Printf("Loaded %d group addresses from %s\n", len(catalog.Groups()), strings.TrimSpace(groupFile))
	}

	knxgt, err := knx.NewGroupTunnel(fmt.Sprintf("%s:%s", server, port), knx.DefaultTunnelConfig)
	if err != nil {
		fmt.Printf("Error while creating: %v\n", err)
		return err
	}
	defer knxgt.Close()

	parent := ctx
	if parent == nil {
		parent = context.Background()
	}

	runCtx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := pgknx.New(postgresqlDSN, viper.GetBool("debug"))
	if err != nil {
		return err
	}

	opts := api.Options{
		Store:   store,
		Catalog: catalog,
		Tunnel:  knxgt,
	}
	if err := api.Run(runCtx, opts); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}

	return nil
}
