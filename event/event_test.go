package event

import (
	"os"
	"testing"

	m "github.com/getgauge/flash/gauge_messages"
)

func TestNewSpecEventNilSafe(t *testing.T) {
	ev := NewSpecEvent(nil, true)
	if ev.Type != Spec || ev.Status != Progress {
		t.Fatalf("unexpected event: %+v", ev)
	}

	ev = NewSpecEvent(&m.ExecutionInfo{}, false)
	if ev.Name != "(unknown spec)" {
		t.Fatalf("expected placeholder name, got %q", ev.Name)
	}
}

func TestNewSpecEventPassFail(t *testing.T) {
	info := &m.ExecutionInfo{
		CurrentSpec: &m.SpecInfo{Name: "Login", FileName: "specs/login.spec", IsFailed: true, Tags: []string{"auth"}},
	}
	started := NewSpecEvent(info, true)
	if started.Status != Progress || started.FileName != "specs/login.spec" {
		t.Fatalf("start: %+v", started)
	}
	ended := NewSpecEvent(info, false)
	if ended.Status != Fail {
		t.Fatalf("end status = %s", ended.Status)
	}
}

func TestNewStepEventCapturesError(t *testing.T) {
	info := &m.ExecutionInfo{
		CurrentSpec:     &m.SpecInfo{Name: "Login", FileName: "specs/login.spec"},
		CurrentScenario: &m.ScenarioInfo{Name: "fails"},
		CurrentStep: &m.StepInfo{
			Step:         &m.ExecuteStepRequest{ActualStepText: "Click button"},
			IsFailed:     true,
			ErrorMessage: "element not found",
			StackTrace:   "at foo:1",
		},
	}
	started := NewStepEvent(info, true)
	if started.Status != Progress || started.ErrorMessage != "" {
		t.Fatalf("start should not include error: %+v", started)
	}
	ended := NewStepEvent(info, false)
	if ended.Status != Fail || ended.ErrorMessage != "element not found" || ended.StackTrace != "at foo:1" {
		t.Fatalf("end: %+v", ended)
	}
}

func TestNewStepEventNilStepRequest(t *testing.T) {
	info := &m.ExecutionInfo{CurrentStep: &m.StepInfo{}}
	ev := NewStepEvent(info, true)
	if ev.Name != "(unknown step)" {
		t.Fatalf("got %q", ev.Name)
	}
}

func TestNewConceptEvent(t *testing.T) {
	info := &m.ExecutionInfo{
		CurrentSpec:     &m.SpecInfo{Name: "Login", FileName: "specs/login.spec"},
		CurrentScenario: &m.ScenarioInfo{Name: "ok"},
		CurrentStep: &m.StepInfo{
			Step:         &m.ExecuteStepRequest{ActualStepText: "Login as \"user\""},
			IsFailed:     true,
			ErrorMessage: "inner step failed",
			StackTrace:   "at concept:1",
		},
	}
	started := NewConceptEvent(info, true)
	if started.Type != Concept || started.Status != Progress || started.Name != "Login as \"user\"" || started.ErrorMessage != "" {
		t.Fatalf("start: %+v", started)
	}
	if started.SpecFileName != "specs/login.spec" || started.ScenarioName != "ok" || started.ConceptName != "Login as \"user\"" {
		t.Fatalf("start refs: %+v", started)
	}
	ended := NewConceptEvent(info, false)
	if ended.Status != Fail || ended.ErrorMessage != "inner step failed" || ended.StackTrace != "at concept:1" {
		t.Fatalf("end: %+v", ended)
	}
	unknown := NewConceptEvent(nil, true)
	if unknown.Name != "(unknown concept)" || unknown.Status != Progress {
		t.Fatalf("nil: %+v", unknown)
	}
}

func TestNewEndEventFromSuite(t *testing.T) {
	ev := NewEndEventFromSuite(nil)
	if ev.Type != End || ev.Status != Pass {
		t.Fatalf("%+v", ev)
	}
	ev = NewEndEventFromSuite(&m.ProtoSuiteResult{
		Failed:            true,
		ProjectName:       "demo",
		SpecsFailedCount:  2,
		SpecsSkippedCount: 1,
		ExecutionTime:     1234,
		Environment:       "ci",
	})
	if ev.Status != Fail || ev.ProjectName != "demo" || ev.SpecsSkipped != 1 || ev.ExecutionTime != 1234 {
		t.Fatalf("%+v", ev)
	}
}

func TestProjectNameFromEnv(t *testing.T) {
	t.Setenv("GAUGE_PROJECT_ROOT", "/tmp/my-project")
	if got := ProjectNameFromEnv(); got != "my-project" {
		t.Fatalf("got %q", got)
	}
	os.Unsetenv("GAUGE_PROJECT_ROOT")
	if ProjectNameFromEnv() != "" {
		t.Fatal("expected empty")
	}
}

func TestNewSuiteEvent(t *testing.T) {
	t.Setenv("GAUGE_PROJECT_ROOT", "/work/uHIL")
	ev := NewSuiteEvent(&m.ExecutionInfo{ProjectName: "from-gauge"})
	if ev.Type != Suite || ev.ProjectName != "from-gauge" {
		t.Fatalf("%+v", ev)
	}
	ev = NewSuiteEvent(nil)
	if ev.ProjectName != "uHIL" {
		t.Fatalf("fallback from env: %+v", ev)
	}
}
