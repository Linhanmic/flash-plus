package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Linhanmic/flash-plus/event"
	"github.com/gorilla/websocket"
)

func TestHealthAndEmptyEvents(t *testing.T) {
	hub := NewWebSocketHub(ServerInfo{Project: "demo", Timestamp: "now"})
	srv := httptest.NewServer(NewRouter(hub, nil))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("health status %d", res.StatusCode)
	}

	res, err = http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if strings.TrimSpace(string(body)) != "[]" {
		t.Fatalf("expected empty array, got %s", body)
	}

	res, err = http.Get(srv.URL + "/api/info")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var info ServerInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}
	if info.Project != "demo" {
		t.Fatalf("project = %q", info.Project)
	}
}

func TestBroadcastReachesWebsocketAndRest(t *testing.T) {
	hub := NewWebSocketHub(ServerInfo{Project: "demo"})
	srv := httptest.NewServer(NewRouter(hub, nil))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/events/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	hub.Broadcast(event.Event{Type: event.Spec, Status: event.Progress, Name: "Login", FileName: "a.spec"})
	hub.Broadcast(event.Event{Type: event.End, Status: event.Fail, SpecsFailed: 1})

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var first event.Event
	if err := conn.ReadJSON(&first); err != nil {
		t.Fatal(err)
	}
	if first.Type != event.Spec || first.Name != "Login" {
		t.Fatalf("first %+v", first)
	}
	var end event.Event
	if err := conn.ReadJSON(&end); err != nil {
		t.Fatal(err)
	}
	if end.Type != event.End || end.Status != event.Fail {
		t.Fatalf("end %+v", end)
	}

	res, err := http.Get(srv.URL + "/api/events")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var events []event.Event
	if err := json.NewDecoder(res.Body).Decode(&events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("want 2 events, got %d", len(events))
	}

	select {
	case <-hub.Done():
	case <-time.After(time.Second):
		t.Fatal("done not signaled")
	}
}

func TestWebsocketReplaysHistory(t *testing.T) {
	hub := NewWebSocketHub(ServerInfo{})
	hub.Broadcast(event.Event{Type: event.Suite, ProjectName: "hist", Status: event.Progress})
	hub.Broadcast(event.Event{Type: event.Spec, Name: "S", FileName: "s.spec", Status: event.Pass})

	srv := httptest.NewServer(NewRouter(hub, nil))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/events/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev event.Event
	if err := conn.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != event.Suite || ev.ProjectName != "hist" {
		t.Fatalf("replay %+v", ev)
	}
	if err := conn.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Name != "S" {
		t.Fatalf("replay spec %+v", ev)
	}
}

func TestGetPortFromEnv(t *testing.T) {
	t.Setenv("FLASH_SERVER_PORT", "18080")
	if p := GetPort(); p != 18080 {
		t.Fatalf("got %d", p)
	}
}

func TestCorsPreflight(t *testing.T) {
	hub := NewWebSocketHub(ServerInfo{})
	srv := httptest.NewServer(NewRouter(hub, nil))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodOptions, srv.URL+"/api/health", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
	if res.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS header")
	}
}
