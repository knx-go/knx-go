package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/dpt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read KNX group values",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applyReadConfig(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return readValue()
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "KNX group address to target")
	cmd.Flags().StringVar(&groupName, "group-name", "", "KNX group name to resolve via the catalog")
	cmd.Flags().StringVarP(&groupFile, "group-file", "f", "", "path to a KNX group address export (XML)")
	cmd.Flags().BoolVar(&waitForResponse, "wait-response", true, "wait for a response to the sent command")
	cmd.Flags().DurationVar(&waitTimeout, "timeout", 5*time.Second, "maximum time to wait for a response when wait-response is enabled")

	root.AddCommand(cmd)
}

func applyReadConfig(cmd *cobra.Command) {
	if value := strings.TrimSpace(viper.GetString("read.group")); value != "" && !flagChanged(cmd, "group") {
		group = value
	}
	if value := strings.TrimSpace(viper.GetString("read.group_name")); value != "" && !flagChanged(cmd, "group-name") {
		groupName = value
	}
	if value := strings.TrimSpace(viper.GetString("read.group_file")); value != "" && !flagChanged(cmd, "group-file") {
		groupFile = value
	}
	if !flagChanged(cmd, "wait-response") {
		if viper.IsSet("read.wait_response") {
			waitForResponse = viper.GetBool("read.wait_response")
		} else {
			waitForResponse = true
		}
	}
	if viper.IsSet("read.timeout") && !flagChanged(cmd, "timeout") {
		if duration, ok := normalizeDuration(viper.Get("read.timeout")); ok {
			waitTimeout = duration
		}
	}
}

func readValue() error {
	catalog, err := loadCatalog(groupFile)
	if err != nil {
		return err
	}

	client, err := knx.NewGroupTunnel(fmt.Sprintf("%s:%s", server, port), knx.DefaultTunnelConfig)
	if err != nil {
		fmt.Printf("Error while creating: %v\n", err)
		return err
	}
	defer client.Close()

	destination, groupMeta, err := resolveGroupTarget(catalog)
	if err != nil {
		return err
	}

	dptType, err := resolveDatapointType(groupMeta)
	if err != nil {
		return err
	}

	event := knx.GroupEvent{
		Command:     knx.GroupRead,
		Destination: destination,
	}

	drainInbound(client.Inbound())

	if err := client.Send(event); err != nil {
		fmt.Printf("Error while sending: %v\n", err)
		return err
	}

	if !waitForResponse {
		return nil
	}

	decoder, _ := dpt.Produce(dptType)
	timer := time.NewTimer(waitTimeout)
	defer timer.Stop()

	for {
		select {
		case msg, ok := <-client.Inbound():
			if !ok {
				return errors.New("connection closed before receiving a response")
			}

			if msg.Destination != destination {
				continue
			}

			if msg.Command == knx.GroupRead {
				continue
			}

			if decoder == nil {
				fmt.Printf("%+v: data=% X\n", msg, msg.Data)
				return nil
			}

			if err := decoder.Unpack(msg.Data); err != nil {
				fmt.Printf("%+v: decode error: %v\n", msg, err)
				return nil
			}

			fmt.Printf("%v\n", decoder)
			return nil
		case <-timer.C:
			return errors.New("timeout waiting for response")
		}
	}
}

func drainInbound(inbound <-chan knx.GroupEvent) {
	for {
		select {
		case <-inbound:
			continue
		default:
			return
		}
	}
}
