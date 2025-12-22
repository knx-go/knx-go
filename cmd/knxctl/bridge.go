package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/cemi"
	"github.com/knx-go/knx-go/knx/knxnet"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "Bridge traffic between two KNX endpoints",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			applyBridgeConfig(cmd)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBridgeLoop(bridgeOther)
		},
	}

	cmd.Flags().StringVarP(&bridgeOther, "other", "o", "", "KNX endpoint to bridge to (host:port)")

	root.AddCommand(cmd)
}

func applyBridgeConfig(cmd *cobra.Command) {
	if value := strings.TrimSpace(viper.GetString("bridge.other")); value != "" && !flagChanged(cmd, "other") {
		bridgeOther = value
	}
}

type bridgeRelay interface {
	relay(data cemi.LData) error

	Inbound() <-chan cemi.Message
	Close()
}

type tunnelRelay struct {
	*knx.Tunnel
}

func (relay tunnelRelay) relay(data cemi.LData) error {
	return relay.Send(&cemi.LDataReq{LData: data})
}

type routerRelay struct {
	*knx.Router
}

func (relay routerRelay) relay(data cemi.LData) error {
	return relay.Send(&cemi.LDataInd{LData: data})
}

type bridge struct {
	tunnel *knx.Tunnel
	other  bridgeRelay
	logger *log.Logger
}

func runBridgeLoop(otherAddr string) error {
	trimmed := strings.TrimSpace(otherAddr)
	if trimmed == "" {
		return errors.New("other KNX endpoint must be specified")
	}

	gatewayAddr := fmt.Sprintf("%s:%s", server, port)
	logger := log.New(os.Stdout, "", log.LstdFlags)

	for {
		br, err := newBridge(gatewayAddr, trimmed, logger)
		if err != nil {
			logger.Printf("Error while creating bridge: %v", err)
			time.Sleep(time.Second)
			continue
		}

		logger.Printf("Bridge connected: %s <-> %s", gatewayAddr, trimmed)

		if err := br.serve(); err != nil {
			logger.Printf("Bridge terminated with error: %v", err)
		}

		br.close()

		time.Sleep(time.Second)
	}
}

func newBridge(gatewayAddr, otherAddr string, logger *log.Logger) (*bridge, error) {
	tunnel, err := knx.NewTunnel(gatewayAddr, knxnet.TunnelLayerData, knx.DefaultTunnelConfig)
	if err != nil {
		return nil, err
	}

	var other bridgeRelay

	addr, err := net.ResolveUDPAddr("udp4", otherAddr)
	if err != nil {
		tunnel.Close()
		return nil, err
	}

	if addr.IP.IsMulticast() {
		router, err := knx.NewRouter(otherAddr, knx.DefaultRouterConfig)
		if err != nil {
			tunnel.Close()
			return nil, err
		}

		other = routerRelay{router}
	} else {
		otherTunnel, err := knx.NewTunnel(otherAddr, knxnet.TunnelLayerData, knx.DefaultTunnelConfig)
		if err != nil {
			tunnel.Close()
			return nil, err
		}

		other = tunnelRelay{otherTunnel}
	}

	return &bridge{tunnel: tunnel, other: other, logger: logger}, nil
}

func (br *bridge) serve() error {
	for {
		select {
		case msg, open := <-br.tunnel.Inbound():
			if !open {
				return errors.New("tunnel channel closed")
			}

			if ind, ok := msg.(*cemi.LDataInd); ok {
				if br.logger != nil {
					br.logger.Printf("tunnel -> other: %+v", ind)
				}
				if err := br.other.relay(ind.LData); err != nil {
					return err
				}
			}
		case msg, open := <-br.other.Inbound():
			if !open {
				return errors.New("other channel closed")
			}

			if ind, ok := msg.(*cemi.LDataInd); ok {
				if br.logger != nil {
					br.logger.Printf("other -> tunnel: %+v", ind)
				}
				if err := br.tunnel.Send(&cemi.LDataReq{LData: ind.LData}); err != nil {
					return err
				}
			}
		}
	}
}

func (br *bridge) close() {
	if br == nil {
		return
	}

	if br.tunnel != nil {
		br.tunnel.Close()
	}

	if br.other != nil {
		br.other.Close()
	}
}
