/*----------------------------------------------------------------
 *  Flash - 前后端分离版本
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(hub *WebSocketHub) *mux.Router {
	r := mux.NewRouter()

	// CORS 中间件
	r.Use(corsMiddleware)

	// API 路由
	r.HandleFunc("/api/events", hub.GetEvents).Methods("GET")
	r.HandleFunc("/api/events/stream", hub.HandleWebSocket).Methods("GET")
	r.HandleFunc("/api/health", healthCheck).Methods("GET")

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
