[![Check](https://github.com/knx-go/knx-go/actions/workflows/check.yaml/badge.svg?branch=master)](https://github.com/knx-go/knx-go/actions/workflows/check.yaml)
[![GoDoc](https://godoc.org/github.com/knx-go/knx-go?status.svg)](https://godoc.org/github.com/knx-go/knx-go)

# knx-go

This repository contains a collection of Go packages that provide the means to communicate with KNX
networks.

## Packages

 Package           | Description
-------------------|--------------------------------------------------------------------
 **cmd/knxctl**    | Tool to interact with KNX devices from CLI
 **knx**           | Abstractions to communicate with KNXnet/IP servers
 **knx/cemi**      | CEMI-encoded frames
 **knx/dpt**       | Datapoint types
 **knx/gac**       | GroupAddress Catalog
 **knx/knxnet**    | KNXnet/IP protocol services

## Installation

Simply run the following command.

	$ go get -u github.com/knx-go/knx-go/...

## Examples

### KNXnet/IP Group Client

If you simply want to send and receive group communication, the
[GroupTunnel](https://godoc.org/github.com/knx-go/knx-go/knx#GroupTunnel) or
[GroupRouter](https://godoc.org/github.com/knx-go/knx-go/knx#GroupRouter)
might be sufficient to you.

```go
package main

import (
	"log"
	"os"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/cemi"
	"github.com/knx-go/knx-go/knx/dpt"
	"github.com/knx-go/knx-go/knx/util"
)

func main() {
	// Setup logger for auxiliary logging. This enables us to see log messages from internal
	// routines.
	util.Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Connect to the gateway.
	client, err := knx.NewGroupTunnel("10.0.0.7:3671", knx.DefaultTunnelConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Close upon exiting. Even if the gateway closes the connection, we still have to clean up.
	defer client.Close()

	// Send 20.5°C to group 1/2/3.
	err = client.Send(knx.GroupEvent{
		Command:     knx.GroupWrite,
		Destination: cemi.NewGroupAddr3(1, 2, 3),
		Data:        dpt.DPT_9001(20.5).Pack(),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Receive messages from the gateway. The inbound channel is closed with the connection.
	for msg := range client.Inbound() {
		var temp dpt.DPT_9001

		err := temp.Unpack(msg.Data)
		if err != nil {
			continue
		}

		util.Logger.Printf("%+v: %v", msg, temp)
	}
}
```

In case you want to access a KNXnet/IP router instead of a gateway, simply replace

```go
client, err := knx.NewGroupTunnel("10.0.0.7:3671", knx.DefaultTunnelConfig)
```

with

```go
client, err := knx.NewGroupRouter("224.0.23.12:3671", knx.DefaultRouterConfig)
```

### KNXnet/IP CEMI Client

Use [Tunnel](https://godoc.org/github.com/knx-go/knx-go/knx#Tunnel) or
[Router](https://godoc.org/github.com/knx-go/knx-go/knx#Router) for finer control over the
communication with a gateway or router.

### KNX Bridge

The **knxctl bridge** tool (in package `cmd/knxctl`) has multiple use cases.

Expose a KNX network behind a gateway at `10.0.0.2:3671` on the multicast group `224.0.23.12:3671`.
This allows routers and router clients to access the network.

	$ knxctl bridge 10.0.0.2:3671 224.0.23.12:3671

Connect two KNX networks through gateways. In this example one gateway is at `10.0.0.2:3671`, the
other is at `10.0.0.3:3671`.

	$ knxctl bridge 10.0.0.2:3671 10.0.0.3:3671

### Discover all KNXnet/IP Servers

The following example shows how to discover all routers/gateways on a network.

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/kr/pretty"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/util"
)

func main() {
	// Setup logger for auxiliary logging. This enables us to see log messages from internal
	// routines.
	util.Logger = log.New(os.Stdout, "", log.LstdFlags)

	servers, err := knx.Discover("224.0.23.12:3671", time.Millisecond*750)
	if err != nil {
		log.Fatal(err)
	}

	util.Logger.Printf("%# v", pretty.Formatter(servers))
}
```

### Describe a Single KNXnet/IP Server

The following example shows how to get a description from a single server.

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/kr/pretty"

	"github.com/knx-go/knx-go/knx"
	"github.com/knx-go/knx-go/knx/util"
)

func main() {
	util.Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Describe KNXnet/IP server at given address and default port
	servers, err := knx.DescribeTunnel("192.168.1.254:3671", time.Millisecond*750)
	if err != nil {
		log.Fatal(err)
	}

	util.Logger.Printf("%# v", pretty.Formatter(servers))
}
```
