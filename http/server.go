/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Linhanmic/flash-plus/event"
)

const flashServerPort = "FLASH_SERVER_PORT"

func Start(ctx context.Context, e chan event.Event, staticFS fs.FS) error {
	hub := NewWebSocketHub(ServerInfo{
		Project:   event.ProjectNameFromEnv(),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-e:
				if !ok {
					return
				}
				hub.Broadcast(ev)
			}
		}
	}()

	router := NewRouter(hub, staticFS)
	port := GetPort()
	addr := fmt.Sprintf("0.0.0.0:%d", port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	fmt.Printf("[Flash Plus] Starting progress reporting at http://127.0.0.1:%d\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func GetPort() int {
	port := os.Getenv(flashServerPort)
	if port != "" {
		p, err := strconv.Atoi(port)
		if err == nil && p > 0 {
			return p
		}
		fmt.Printf("[Flash Plus] Cannot use %s='%s' value. Error: %s\n", flashServerPort, port, err)
	}
	l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 0})
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
