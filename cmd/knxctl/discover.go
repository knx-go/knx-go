package main

import (
	"fmt"
	"time"

	"github.com/knx-go/knx-go/knx"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover KNX server information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return discover()
		},
	}

	root.AddCommand(cmd)
}

func discover() error {
	client, err := knx.DescribeTunnel(fmt.Sprintf("%s:%s", server, port), time.Millisecond*750)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", client)

	return nil
}
