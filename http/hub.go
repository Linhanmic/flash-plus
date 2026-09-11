/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Linhanmic/flash-plus/event"
	"github.com/gorilla/websocket"
)

type clientConn struct {
	conn      *websocket.Conn
	send      chan event.Event
	closeOnce sync.Once
}

func (c *clientConn) closeSend() {
	c.closeOnce.Do(func() { close(c.send) })
}

type WebSocketHub struct {
	mu          sync.RWMutex
	connections map[*clientConn]struct{}
	events      []event.Event
	upgrader    websocket.Upgrader
	doneOnce    sync.Once
	done        chan struct{}
	info        ServerInfo
}

type ServerInfo struct {
	Project   string `json:"project"`
	Timestamp string `json:"timestamp"`
	Finished  bool   `json:"finished"`
	Status    string `json:"status,omitempty"`
}

func NewWebSocketHub(info ServerInfo) *WebSocketHub {
	if info.Timestamp == "" {
		info.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}
	if info.Project == "" {
		info.Project = event.ProjectNameFromEnv()
	}
	if info.Project == "" {
		info.Project = "Gauge"
	}
	return &WebSocketHub{
		connections: make(map[*clientConn]struct{}),
		events:      make([]event.Event, 0),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		done: make(chan struct{}),
		info: info,
	}
}

func (h *WebSocketHub) Info() ServerInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.info
}

func (h *WebSocketHub) Events() []event.Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]event.Event, len(h.events))
	copy(out, h.events)
	return out
}

func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &clientConn{
		conn: conn,
		send: make(chan event.Event, 256),
	}

	h.mu.Lock()
	history := append([]event.Event(nil), h.events...)
	finished := h.info.Finished
	h.connections[client] = struct{}{}
	h.mu.Unlock()

	go h.writePump(client)
	for _, e := range history {
		select {
		case client.send <- e:
		default:
			h.remove(client)
			return
		}
	}
	if finished {
		client.closeSend()
	}
	h.readPump(client)
}

func (h *WebSocketHub) writePump(c *clientConn) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case e, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteJSON(e); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHub) readPump(c *clientConn) {
	defer h.remove(c)
	c.conn.SetReadLimit(1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *WebSocketHub) remove(c *clientConn) {
	h.mu.Lock()
	delete(h.connections, c)
	h.mu.Unlock()
	c.closeSend()
	_ = c.conn.Close()
}

func (h *WebSocketHub) Broadcast(e event.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if e.Type == event.Suite && e.ProjectName != "" {
		h.info.Project = e.ProjectName
	}

	h.events = append(h.events, e)

	if e.Type == event.End {
		h.info.Finished = true
		h.info.Status = string(e.Status)
		for c := range h.connections {
			select {
			case c.send <- e:
			default:
			}
			delete(h.connections, c)
			c.closeSend()
		}
		h.doneOnce.Do(func() { close(h.done) })
		return
	}

	stale := make([]*clientConn, 0)
	for c := range h.connections {
		select {
		case c.send <- e:
		default:
			stale = append(stale, c)
		}
	}
	for _, c := range stale {
		delete(h.connections, c)
		c.closeSend()
	}
}

func (h *WebSocketHub) GetEvents(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.events)
}

func (h *WebSocketHub) GetInfo(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.info)
}

func (h *WebSocketHub) Done() <-chan struct{} {
	return h.done
}
