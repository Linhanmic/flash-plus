package http

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Linhanmic/flash-plus/event"
)

func TestReportOutputDirUsesEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GAUGE_PROJECT_ROOT", root)
	t.Setenv("gauge_reports_dir", "reports")
	t.Setenv("overwrite_reports", "true")
	t.Setenv("FLASH_REPORTS_DIR", "")
	dir, err := ReportOutputDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "reports", "flash-plus")
	if dir != want {
		t.Fatalf("got %s want %s", dir, want)
	}
}

func TestReportOutputDirOverrideAndTimestamp(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GAUGE_PROJECT_ROOT", root)
	t.Setenv("FLASH_REPORTS_DIR", "out")
	t.Setenv("overwrite_reports", "false")
	dir, err := ReportOutputDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dir, filepath.Join(root, "out", "flash-plus")+string(os.PathSeparator)) {
		t.Fatalf("timestamped dir = %s", dir)
	}
}

func TestSessionEndIsSticky(t *testing.T) {
	s := &Session{}
	s.SetReason(ReasonEnd)
	s.SetReason(ReasonStop)
	if s.Reason() != ReasonEnd {
		t.Fatalf("got %s", s.Reason())
	}
}

func TestSaveStandaloneReportOnPause(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GAUGE_PROJECT_ROOT", root)
	t.Setenv("gauge_reports_dir", "reports")
	t.Setenv("overwrite_reports", "true")

	static := fstest.MapFS{
		"index.html":    {Data: []byte(`<link href="/assets/app.css"><script type="module" src="/assets/app.js"></script>`)},
		"assets/app.js": {Data: []byte(`console.log(window.__FLASH_SNAPSHOT__.reason)`)} ,
		"assets/app.css": {Data: []byte(`body{color:#111}`)},
	}
	hub := NewWebSocketHub(ServerInfo{Project: "demo", Timestamp: "now"})
	hub.Broadcast(event.Event{Type: event.Spec, Name: "Env", FileName: "env.spec", Status: event.Pass})

	saver := NewSaver(static)
	dir, err := saver.Save(hub, ReasonPause)
	if err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(html)
	for _, want := range []string{
		"window.__FLASH_SNAPSHOT__",
		`"reason":"pause"`,
		"env.spec",
		"console.log",
		"body{color:#111}",
		`"type":"end"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("saved HTML missing %s\n%s", want, text)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "snapshot.json")); err != nil {
		t.Fatal(err)
	}
	if second, err := saver.Save(hub, ReasonStop); err != nil || second != dir {
		t.Fatalf("second save: %s %v", second, err)
	}
}

func TestSaveUsesEndStatus(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GAUGE_PROJECT_ROOT", root)
	t.Setenv("gauge_reports_dir", "reports")
	t.Setenv("overwrite_reports", "true")
	static := fstest.MapFS{
		"index.html":    {Data: []byte(`<script type="module" src="/assets/app.js"></script>`)},
		"assets/app.js": {Data: []byte(`void 0`)},
	}
	hub := NewWebSocketHub(ServerInfo{Project: "demo"})
	hub.Broadcast(event.Event{Type: event.End, Status: event.Fail, ProjectName: "demo"})
	saver := NewSaver(static)
	dir, err := saver.Save(hub, ReasonEnd)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(dir, "snapshot.json"))
	if !strings.Contains(string(body), `"reason":"end"`) || !strings.Contains(string(body), `"status":"fail"`) {
		t.Fatalf("snapshot: %s", body)
	}
}
