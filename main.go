/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Linhanmic/flash-plus/event"
	flashGrpc "github.com/Linhanmic/flash-plus/grpc"
	flashHttp "github.com/Linhanmic/flash-plus/http"
)

//go:embed dist
var distFS embed.FS

func main() {
	demo := flag.Bool("demo", false, "Run a local demo timeline without Gauge")
	flag.Parse()
	if os.Getenv("FLASH_PLUS_DEMO") == "1" {
		*demo = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session := &flashHttp.Session{}
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		sig := <-ch
		if sig == syscall.SIGINT {
			session.SetReason(flashHttp.ReasonPause)
		} else {
			session.SetReason(flashHttp.ReasonStop)
		}
		cancel()
	}()

	events := make(chan event.Event, 256)

	staticFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("failed to load embedded frontend: %v", err)
	}

	if *demo {
		log.Println("[Flash Plus] demo mode: publishing sample execution events")
		go flashHttp.RunDemo(events)
	} else {
		go func() {
			if err := flashGrpc.Start(ctx, events, func() {
				session.SetReason(flashHttp.ReasonStop)
				cancel()
			}); err != nil {
				log.Printf("[Flash Plus] gRPC server stopped: %v", err)
				session.SetReason(flashHttp.ReasonStop)
				cancel()
			}
		}()
	}

	if err := flashHttp.Start(ctx, events, staticFS, session); err != nil {
		log.Fatal(err)
	}
}
