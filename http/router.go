/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func NewRouter(hub *WebSocketHub, staticFS fs.FS) *mux.Router {
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	r.HandleFunc("/api/events", hub.GetEvents).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/events/stream", hub.HandleWebSocket).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/info", hub.GetInfo).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/health", healthCheck).Methods("GET", "OPTIONS")
	// Alias for original Flash websocket path.
	r.HandleFunc("/progress", hub.HandleWebSocket).Methods("GET", "OPTIONS")

	if staticFS != nil {
		fileServer := http.FileServer(http.FS(staticFS))
		r.PathPrefix("/").Handler(spaHandler(staticFS, fileServer))
	}

	return r
}

func spaHandler(staticFS fs.FS, fileServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(staticFS, path); err != nil {
			r.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
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
