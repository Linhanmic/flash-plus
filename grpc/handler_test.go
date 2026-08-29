package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/Linhanmic/flash-plus/event"
	gm "github.com/getgauge/flash/gauge_messages"
)

func TestHandlerEmitsLifecycleEvents(t *testing.T) {
	ch := make(chan event.Event, 16)
	h := NewHandler(nil, ch)
	ctx := context.Background()
	info := &gm.ExecutionInfo{
		ProjectName: "demo",
		CurrentSpec: &gm.SpecInfo{Name: "Login", FileName: "login.spec"},
		CurrentScenario: &gm.ScenarioInfo{Name: "ok"},
		CurrentStep: &gm.StepInfo{
			Step:     &gm.ExecuteStepRequest{ActualStepText: "open page"},
			IsFailed: false,
		},
	}

	if _, err := h.NotifyExecutionStarting(ctx, &gm.ExecutionStartingRequest{CurrentExecutionInfo: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifySpecExecutionStarting(ctx, &gm.SpecExecutionStartingRequest{CurrentExecutionInfo: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifyScenarioExecutionStarting(ctx, &gm.ScenarioExecutionStartingRequest{CurrentExecutionInfo: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifyStepExecutionStarting(ctx, &gm.StepExecutionStartingRequest{CurrentExecutionInfo: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifyStepExecutionEnding(ctx, &gm.StepExecutionEndingRequest{CurrentExecutionInfo: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifyScenarioExecutionEnding(ctx, &gm.ScenarioExecutionEndingRequest{CurrentExecutionInfo: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifySpecExecutionEnding(ctx, &gm.SpecExecutionEndingRequest{CurrentExecutionInfo: info}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifySuiteResult(ctx, &gm.SuiteExecutionResult{SuiteResult: &gm.ProtoSuiteResult{Failed: false, ProjectName: "demo"}}); err != nil {
		t.Fatal(err)
	}

	got := drain(ch, 8)
	types := make([]event.EventType, len(got))
	for i, e := range got {
		types[i] = e.Type
	}
	want := []event.EventType{event.Suite, event.Spec, event.Scenario, event.Step, event.Step, event.Scenario, event.Spec, event.End}
	if len(types) != len(want) {
		t.Fatalf("got %v", types)
	}
	for i := range want {
		if types[i] != want[i] {
			t.Fatalf("index %d: got %s want %s", i, types[i], want[i])
		}
	}
	if got[3].Status != event.Progress || got[4].Status != event.Pass {
		t.Fatalf("step statuses %s %s", got[3].Status, got[4].Status)
	}
}

func TestHandlerNilRequests(t *testing.T) {
	ch := make(chan event.Event, 8)
	h := NewHandler(nil, ch)
	ctx := context.Background()
	if _, err := h.NotifySpecExecutionStarting(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := h.NotifySuiteResult(ctx, nil); err != nil {
		t.Fatal(err)
	}
	got := drain(ch, 1)
	if got[0].Type != event.End {
		t.Fatalf("got %+v", got[0])
	}
}

func TestKillCancels(t *testing.T) {
	ch := make(chan event.Event, 1)
	h := NewHandler(nil, ch)
	ctx, cancel := context.WithCancel(context.Background())
	h.SetCancel(cancel)
	if _, err := h.Kill(ctx, &gm.KillProcessRequest{}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("kill did not cancel")
	}
}

func drain(ch <-chan event.Event, n int) []event.Event {
	out := make([]event.Event, 0, n)
	for i := 0; i < n; i++ {
		select {
		case e := <-ch:
			out = append(out, e)
		case <-time.After(time.Second):
			return out
		}
	}
	return out
}
