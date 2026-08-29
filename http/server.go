/*----------------------------------------------------------------
 *  Flash - 前后端分离版本
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/getgauge/flash-server/event"
)

func Start(e chan event.Event, staticFS fs.FS) {
	hub := NewWebSocketHub()

	// 从 channel 读取事件并广播
	go func() {
		for ev := range e {
			hub.Broadcast(ev)
		}
	}()

	router := NewRouter(hub)

	// 挂载前端静态文件
	if staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))
		router.PathPrefix("/").Handler(fileServer)
	}

	addr := ":8080"
	fmt.Printf("[Flash] API server listening on http://localhost%s\n", addr)
	fmt.Printf("[Flash] WebSocket endpoint: ws://localhost%s/api/events/stream\n", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
