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

func Start(ctx context.Context, e chan event.Event, staticFS fs.FS, session *Session) error {
	if session == nil {
		session = &Session{}
	}
	hub := NewWebSocketHub(ServerInfo{
		Project:   event.ProjectNameFromEnv(),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})
	saver := NewSaver(staticFS)
	saved := make(chan struct{})

	go func() {
		defer close(saved)
		handle := func(ev event.Event) {
			hub.Broadcast(ev)
			if ev.Type == event.End {
				session.SetReason(ReasonEnd)
				if _, err := saver.Save(hub, ReasonEnd); err != nil {
					log.Printf("[Flash Plus] failed to save report: %v\n", err)
				}
			}
		}
		saveIfNeeded := func() {
			if saver.Saved() {
				return
			}
			if _, err := saver.Save(hub, session.Reason()); err != nil {
				log.Printf("[Flash Plus] failed to save report: %v\n", err)
			}
		}
		for {
			select {
			case <-ctx.Done():
				deadline := time.After(400 * time.Millisecond)
				for {
					select {
					case ev, ok := <-e:
						if !ok {
							saveIfNeeded()
							return
						}
						handle(ev)
					case <-deadline:
						saveIfNeeded()
						return
					}
				}
			case ev, ok := <-e:
				if !ok {
					saveIfNeeded()
					return
				}
				handle(ev)
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
		select {
		case <-saved:
		case <-time.After(2 * time.Second):
		}
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
