/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Linhanmic/flash-plus/event"
)

const (
	GaugeReportsDirEnv      = "gauge_reports_dir"
	OverwriteReportsEnv     = "overwrite_reports"
	FlashReportsDirEnv      = "FLASH_REPORTS_DIR"
	DefaultReportsDir       = "reports"
	ReportPluginDir         = "flash-plus"
	ReasonEnd               = "end"
	ReasonPause             = "pause"
	ReasonStop              = "stop"
	reportTimeFormat        = "2006-01-02_15.04.05"
)

type Snapshot struct {
	Info   ServerInfo     `json:"info"`
	Events []event.Event  `json:"events"`
	Reason string         `json:"reason,omitempty"`
}

type Session struct {
	mu     sync.Mutex
	reason string
}

func (s *Session) SetReason(reason string) {
	if s == nil || reason == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reason == ReasonEnd {
		return
	}
	s.reason = reason
}

func (s *Session) Reason() string {
	if s == nil {
		return ReasonStop
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.reason == "" {
		return ReasonStop
	}
	return s.reason
}

type Saver struct {
	mu       sync.Mutex
	saved    bool
	dir      string
	staticFS fs.FS
}

func NewSaver(staticFS fs.FS) *Saver {
	return &Saver{staticFS: staticFS}
}

func (s *Saver) Saved() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saved
}

func (s *Saver) Dir() string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dir
}

func (s *Saver) Save(hub *WebSocketHub, reason string) (string, error) {
	if s == nil || hub == nil {
		return "", fmt.Errorf("report saver is not configured")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.saved {
		return s.dir, nil
	}
	if reason == "" {
		reason = ReasonStop
	}

	info, events := hub.Snapshot()
	info.Finished = true
	info.Reason = reason
	if info.Status == "" {
		info.Status = string(statusFromEvents(events, reason))
	}
	if !hasEndEvent(events) {
		events = append(events, event.Event{
			Type:   event.End,
			Status: event.Status(info.Status),
			Name:   reason,
		})
	}
	snap := Snapshot{Info: info, Events: events, Reason: reason}

	dir, err := ReportOutputDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := writeSnapshotFiles(dir, s.staticFS, snap); err != nil {
		return "", err
	}
	s.saved = true
	s.dir = dir
	log.Printf("[Flash Plus] Report saved to %s (%s)\n", dir, reason)
	return dir, nil
}

func statusFromEvents(events []event.Event, reason string) event.Status {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Type == event.End && events[i].Status != "" {
			return events[i].Status
		}
	}
	for _, ev := range events {
		if ev.Status == event.Fail {
			return event.Fail
		}
	}
	if reason == ReasonEnd {
		return event.Pass
	}
	return event.Skip
}

func hasEndEvent(events []event.Event) bool {
	for _, ev := range events {
		if ev.Type == event.End {
			return true
		}
	}
	return false
}

func ShouldOverwriteReports() bool {
	return strings.EqualFold(os.Getenv(OverwriteReportsEnv), "true")
}

func reportsBaseDir() (string, error) {
	dir := os.Getenv(FlashReportsDirEnv)
	if dir == "" {
		dir = os.Getenv(GaugeReportsDirEnv)
	}
	if dir == "" {
		dir = DefaultReportsDir
	}
	if !filepath.IsAbs(dir) {
		root := os.Getenv("GAUGE_PROJECT_ROOT")
		if root != "" {
			dir = filepath.Join(root, dir)
		}
	}
	return filepath.Abs(dir)
}

func ReportOutputDir() (string, error) {
	base, err := reportsBaseDir()
	if err != nil {
		return "", err
	}
	if ShouldOverwriteReports() {
		return filepath.Join(base, ReportPluginDir), nil
	}
	return filepath.Join(base, ReportPluginDir, time.Now().Format(reportTimeFormat)), nil
}

var (
	scriptSrcRE = regexp.MustCompile(`src="(/assets/[^"]+)"`)
	styleHrefRE = regexp.MustCompile(`href="(/assets/[^"]+\.css)"`)
)

func writeSnapshotFiles(dir string, staticFS fs.FS, snap Snapshot) error {
	payload, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "snapshot.json"), payload, 0644); err != nil {
		return err
	}
	html, err := buildStandaloneHTML(staticFS, payload)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "index.html"), html, 0644)
}

func buildStandaloneHTML(staticFS fs.FS, payload []byte) ([]byte, error) {
	index, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		return fallbackHTML(payload), nil
	}
	cssPath := firstMatch(styleHrefRE, string(index))
	jsPath := firstMatch(scriptSrcRE, string(index))
	var css, js string
	if cssPath != "" {
		if b, err := fs.ReadFile(staticFS, strings.TrimPrefix(cssPath, "/")); err == nil {
			css = string(b)
		}
	}
	if jsPath != "" {
		if b, err := fs.ReadFile(staticFS, strings.TrimPrefix(jsPath, "/")); err == nil {
			js = strings.ReplaceAll(string(b), "</script>", "<\\/script>")
		}
	}
	if js == "" {
		return fallbackHTML(payload), nil
	}

	var buf bytes.Buffer
	buf.WriteString("<!DOCTYPE html>\n<html lang=\"zh-CN\">\n<head>\n")
	buf.WriteString("  <meta charset=\"UTF-8\">\n")
	buf.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	buf.WriteString("  <title>Flash Plus - Gauge Report</title>\n")
	buf.WriteString("  <script>window.__FLASH_SNAPSHOT__ = ")
	buf.Write(payload)
	buf.WriteString(";</script>\n")
	if css != "" {
		buf.WriteString("  <style>\n")
		buf.WriteString(css)
		buf.WriteString("\n  </style>\n")
	}
	buf.WriteString("</head>\n<body>\n  <div id=\"app\"></div>\n")
	buf.WriteString("  <script type=\"module\">\n")
	buf.WriteString(js)
	buf.WriteString("\n  </script>\n</body>\n</html>\n")
	return buf.Bytes(), nil
}

func firstMatch(re *regexp.Regexp, html string) string {
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func fallbackHTML(payload []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("<!DOCTYPE html><html><head><meta charset=\"UTF-8\"><title>Flash Plus Report</title>")
	buf.WriteString("<script>window.__FLASH_SNAPSHOT__ = ")
	buf.Write(payload)
	buf.WriteString(";</script></head><body>")
	buf.WriteString("<p>Flash Plus snapshot saved. Open this file after a full plugin build so the report UI can be inlined.</p>")
	buf.WriteString("<pre id=\"events\"></pre><script>")
	buf.WriteString("document.getElementById('events').textContent = JSON.stringify(window.__FLASH_SNAPSHOT__, null, 2);")
	buf.WriteString("</script></body></html>")
	return buf.Bytes()
}
