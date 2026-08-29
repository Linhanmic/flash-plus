/*----------------------------------------------------------------
 *  Flash - 前后端分离版本
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package grpc

import (
	"fmt"
	"net"

	"github.com/getgauge/flash-server/event"
	gm "github.com/getgauge/flash/gauge_messages"
	"google.golang.org/grpc"
)

func Start(e chan event.Event) {
	address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		panic("failed to resolve TCP address")
	}
	l, err := net.ListenTCP("tcp", address)
	if err != nil {
		panic("failed to start gRPC server")
	}
	server := grpc.NewServer(grpc.MaxRecvMsgSize(1024 * 1024 * 1024))
	h := NewHandler(server, e)
	gm.RegisterReporterServer(server, h)
	fmt.Printf("[Flash] gRPC server listening on port:%d\n", l.Addr().(*net.TCPAddr).Port)
	server.Serve(l)
}
