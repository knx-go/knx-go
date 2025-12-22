package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/cemi"
	"github.com/knx-go/knx-go/knx/dpt"
	"github.com/knx-go/knx-go/knx/gac"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send KNX group values",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applySendConfig(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return sendValue()
		},
	}

	cmd.Flags().StringVarP(&group, "group", "g", "", "KNX group address to target")
	cmd.Flags().StringVar(&groupName, "group-name", "", "KNX group name to resolve via the catalog")
	cmd.Flags().StringVar(&writeDPT, "dpt", "", "Datapoint type to use when encoding the value")
	cmd.Flags().StringVar(&valueRaw, "value", "", "Value to encode and send to the destination group")
	cmd.Flags().StringVarP(&groupFile, "group-file", "f", "", "path to a KNX group address export (XML)")
	cmd.Flags().BoolVar(&waitForResponse, "wait-response", false, "wait for a response to the sent command")
	cmd.Flags().DurationVar(&waitTimeout, "timeout", 5*time.Second, "maximum time to wait for a response when wait-response is enabled")

	root.AddCommand(cmd)
}

func applySendConfig(cmd *cobra.Command) {
	if value := strings.TrimSpace(viper.GetString("send.group")); value != "" && !flagChanged(cmd, "group") {
		group = value
	}
	if value := strings.TrimSpace(viper.GetString("send.group_name")); value != "" && !flagChanged(cmd, "group-name") {
		groupName = value
	}
	if value := strings.TrimSpace(viper.GetString("send.dpt")); value != "" && !flagChanged(cmd, "dpt") {
		writeDPT = value
	}
	if value := strings.TrimSpace(viper.GetString("send.value")); value != "" && !flagChanged(cmd, "value") {
		valueRaw = value
	}
	if value := strings.TrimSpace(viper.GetString("send.group_file")); value != "" && !flagChanged(cmd, "group-file") {
		groupFile = value
	}
	if viper.IsSet("send.wait_response") && !flagChanged(cmd, "wait-response") {
		waitForResponse = viper.GetBool("send.wait_response")
	}
	if viper.IsSet("send.timeout") && !flagChanged(cmd, "timeout") {
		if duration, ok := normalizeDuration(viper.Get("send.timeout")); ok {
			waitTimeout = duration
		}
	}
}

func sendValue() error {
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

	payload, err := encodeDatapointValue(dptType, valueRaw)
	if err != nil {
		return err
	}

	event := knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: destination,
		Data:        payload,
	}

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

			if decoder == nil {
				fmt.Printf("%+v: data=% X\n", msg, msg.Data)
				return nil
			}

			if err := decoder.Unpack(msg.Data); err != nil {
				fmt.Printf("%+v: decode error: %v\n", msg, err)
				return nil
			}

			fmt.Printf("%+v: %v\n", msg, decoder)
			return nil
		case <-timer.C:
			return errors.New("timeout waiting for response")
		}
	}
}

func resolveGroupTarget(catalog *gac.Catalog) (cemi.GroupAddr, *gac.Group, error) {
	var zero cemi.GroupAddr
	if trimmed := strings.TrimSpace(groupName); trimmed != "" {
		if catalog == nil {
			return zero, nil, fmt.Errorf("group name %q requested but no catalog was provided", trimmed)
		}

		group, ok := catalog.Lookup(trimmed)
		if !ok {
			return zero, nil, fmt.Errorf("unknown group name %q", trimmed)
		}

		return group.Address, group, nil
	}

	trimmed := strings.TrimSpace(group)
	if trimmed == "" {
		return zero, nil, errors.New("group address must be specified")
	}

	destination, err := cemi.NewGroupAddrString(trimmed)
	if err != nil {
		return zero, nil, err
	}

	if catalog != nil {
		if group, ok := catalog.LookupByAddress(destination); ok {
			return destination, group, nil
		}
	}

	return destination, nil, nil
}

func resolveDatapointType(group *gac.Group) (string, error) {
	if trimmed := strings.TrimSpace(writeDPT); trimmed != "" {
		canonical, err := canonicalizeDatapointType(trimmed)
		if err != nil {
			return "", err
		}
		if _, ok := dpt.Produce(canonical); !ok {
			return "", fmt.Errorf("unknown datapoint type %q", trimmed)
		}
		return canonical, nil
	}

	if group != nil {
		if len(group.DPTs) == 0 {
			return "", fmt.Errorf("group %q does not define any datapoint types", group.Name)
		}
		return string(group.DPTs[0]), nil
	}

	return "", errors.New("no datapoint type provided and none available from catalog")
}

func encodeDatapointValue(datapointType, literal string) ([]byte, error) {
	value, ok := dpt.Produce(datapointType)
	if !ok {
		return nil, fmt.Errorf("unknown datapoint type %q", datapointType)
	}

	trimmed := strings.TrimSpace(literal)
	if trimmed == "" {
		return nil, errors.New("value must be provided")
	}

	candidate := []byte(trimmed)
	if !json.Valid(candidate) {
		marshaled, err := json.Marshal(trimmed)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize value: %w", err)
		}
		candidate = marshaled
	}

	if err := json.Unmarshal(candidate, value); err != nil {
		if parsed, perr := strconv.ParseBool(trimmed); perr == nil {
			fallback := []byte(strconv.FormatBool(parsed))
			if err := json.Unmarshal(fallback, value); err == nil {
				return value.Pack(), nil
			}
		}
		return nil, fmt.Errorf("failed to decode value for datapoint %s: %w", datapointType, err)
	}

	return value.Pack(), nil
}

func canonicalizeDatapointType(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("datapoint type must not be empty")
	}

	normalized := strings.ToUpper(trimmed)
	normalized = strings.ReplaceAll(normalized, "_", "-")

	if strings.HasPrefix(normalized, "DPST-") || strings.HasPrefix(normalized, "DPT-") {
		segments := strings.Split(normalized, "-")
		if len(segments) < 3 {
			return "", fmt.Errorf("invalid datapoint type %q", raw)
		}

		mainPart := segments[1]
		subPart := strings.Join(segments[2:], "")

		main, err := strconv.Atoi(mainPart)
		if err != nil {
			return "", fmt.Errorf("invalid datapoint type %q", raw)
		}

		sub, err := strconv.Atoi(subPart)
		if err != nil {
			return "", fmt.Errorf("invalid datapoint type %q", raw)
		}

		return formatDatapoint(main, sub), nil
	}

	if strings.Contains(normalized, ".") {
		segments := strings.SplitN(normalized, ".", 2)
		if len(segments) != 2 {
			return "", fmt.Errorf("invalid datapoint type %q", raw)
		}

		main, err := strconv.Atoi(segments[0])
		if err != nil {
			return "", fmt.Errorf("invalid datapoint type %q", raw)
		}

		sub, err := strconv.Atoi(segments[1])
		if err != nil {
			return "", fmt.Errorf("invalid datapoint type %q", raw)
		}

		return formatDatapoint(main, sub), nil
	}

	return "", fmt.Errorf("invalid datapoint type %q", raw)
}

func formatDatapoint(main, sub int) string {
	return fmt.Sprintf("%d.%03d", main, sub)
}
