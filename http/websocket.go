/*----------------------------------------------------------------
 *  Flash - 前后端分离版本
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/getgauge/flash-server/event"
	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	mu          sync.RWMutex
	connections []*websocket.Conn
	events      []event.Event
	upgrader    websocket.Upgrader
	done        chan struct{}
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		done: make(chan struct{}),
	}
}

// HandleWebSocket 处理 WebSocket 连接
func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.connections = append(h.connections, conn)
	// 回放历史事件
	for _, e := range h.events {
		conn.WriteJSON(e)
	}
	h.mu.Unlock()
}

// Broadcast 广播事件给所有连接
func (h *WebSocketHub) Broadcast(e event.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.events = append(h.events, e)

	// 如果是结束事件，通知所有客户端
	if e.GetType() == event.End {
		for _, conn := range h.connections {
			conn.WriteJSON(e)
			conn.Close()
		}
		h.connections = nil
		close(h.done)
		return
	}

	for _, conn := range h.connections {
		conn.WriteJSON(e)
	}
}

// GetEvents 获取历史事件（REST API）
func (h *WebSocketHub) GetEvents(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.events)
}

// Done 返回结束信号 channel
func (h *WebSocketHub) Done() <-chan struct{} {
	return h.done
}
