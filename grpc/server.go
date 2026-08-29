/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/Linhanmic/flash-plus/event"
	gm "github.com/getgauge/flash/gauge_messages"
	"google.golang.org/grpc"
)

// Start listens on a random local port and prints the Gauge handshake line.
// Gauge extracts the port from stdout using the exact prefix "Listening on port:".
func Start(ctx context.Context, e chan event.Event, cancel context.CancelFunc) error {
	address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to resolve TCP address: %w", err)
	}
	l, err := net.ListenTCP("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	server := grpc.NewServer(grpc.MaxRecvMsgSize(1024 * 1024 * 1024))
	h := NewHandler(server, e)
	h.SetCancel(cancel)
	gm.RegisterReporterServer(server, h)

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	// Keep this format identical to getgauge/Flash — Gauge parses it.
	fmt.Printf("Listening on port:%d\n", l.Addr().(*net.TCPAddr).Port)
	return server.Serve(l)
}
