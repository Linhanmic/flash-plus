/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package grpc

import (
	"context"

	"github.com/Linhanmic/flash-plus/event"
	gm "github.com/getgauge/flash/gauge_messages"
	"google.golang.org/grpc"
)

type handler struct {
	server *grpc.Server
	e      chan event.Event
	cancel context.CancelFunc
}

func NewHandler(s *grpc.Server, e chan event.Event) *handler {
	return &handler{server: s, e: e}
}

func (h *handler) SetCancel(cancel context.CancelFunc) {
	h.cancel = cancel
}

func (h *handler) SetOnKill(fn context.CancelFunc) {
	if fn != nil {
		h.cancel = fn
	}
}

func (h *handler) emit(ev event.Event) {
	h.e <- ev
}

func (h *handler) NotifyExecutionStarting(c context.Context, m *gm.ExecutionStartingRequest) (*gm.Empty, error) {
	var info *gm.ExecutionInfo
	if m != nil {
		info = m.CurrentExecutionInfo
	}
	h.emit(event.NewSuiteEvent(info))
	return &gm.Empty{}, nil
}

func (h *handler) NotifySpecExecutionStarting(c context.Context, m *gm.SpecExecutionStartingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewSpecEvent(m.CurrentExecutionInfo, true))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifyScenarioExecutionStarting(c context.Context, m *gm.ScenarioExecutionStartingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewScenarioEvent(m.CurrentExecutionInfo, true))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifyConceptExecutionStarting(c context.Context, m *gm.StepExecutionStartingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewConceptEvent(m.CurrentExecutionInfo, true))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifyConceptExecutionEnding(c context.Context, m *gm.StepExecutionEndingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewConceptEvent(m.CurrentExecutionInfo, false))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifyStepExecutionStarting(c context.Context, m *gm.StepExecutionStartingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewStepEvent(m.CurrentExecutionInfo, true))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifyStepExecutionEnding(c context.Context, m *gm.StepExecutionEndingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewStepEvent(m.CurrentExecutionInfo, false))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifyScenarioExecutionEnding(c context.Context, m *gm.ScenarioExecutionEndingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewScenarioEvent(m.CurrentExecutionInfo, false))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifySpecExecutionEnding(c context.Context, m *gm.SpecExecutionEndingRequest) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewSpecEvent(m.CurrentExecutionInfo, false))
	}
	return &gm.Empty{}, nil
}

func (h *handler) NotifyExecutionEnding(c context.Context, m *gm.ExecutionEndingRequest) (*gm.Empty, error) {
	return &gm.Empty{}, nil
}

func (h *handler) NotifySuiteResult(c context.Context, m *gm.SuiteExecutionResult) (*gm.Empty, error) {
	if m != nil {
		h.emit(event.NewEndEventFromSuite(m.SuiteResult))
	} else {
		h.emit(event.NewEndEvent(false))
	}
	return &gm.Empty{}, nil
}

func (h *handler) Kill(c context.Context, m *gm.KillProcessRequest) (*gm.Empty, error) {
	go h.stopServer()
	return &gm.Empty{}, nil
}

func (h *handler) stopServer() {
	if h.cancel != nil {
		h.cancel()
		return
	}
	if h.server != nil {
		h.server.GracefulStop()
	}
}
