/*----------------------------------------------------------------
 *  Flash - 前后端分离版本
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package event

import m "github.com/getgauge/flash/gauge_messages"

type Status string
type EventType string

const (
	Pass     Status    = "pass"
	Fail     Status    = "fail"
	Progress Status    = "progress"
	Spec     EventType = "spec"
	Scenario EventType = "scenario"
	Step     EventType = "step"
	End      EventType = "end"
)

// Event 接口定义方法
type Event interface {
	GetType() EventType
	GetStatus() Status
}

// BaseEvent 公共字段
type BaseEvent struct {
	Name   string    `json:"name"`
	Status Status    `json:"status"`
	Type   EventType `json:"type"`
}

func (e BaseEvent) GetType() EventType { return e.Type }
func (e BaseEvent) GetStatus() Status  { return e.Status }

type SpecEvent struct {
	BaseEvent
	FileName string `json:"fileName"`
}

type ScenarioEvent struct {
	BaseEvent
	SpecFileName string `json:"specFileName"`
}

type StepEvent struct {
	BaseEvent
	ScenarioName string `json:"scenarioName"`
	SpecFileName string `json:"specFileName"`
}

type EndEvent struct {
	BaseEvent
}

func NewSpecEvent(i *m.ExecutionInfo, hasStarted bool) SpecEvent {
	s := Pass
	if i.CurrentSpec.GetIsFailed() {
		s = Fail
	}
	if hasStarted {
		s = Progress
	}
	return SpecEvent{
		BaseEvent: BaseEvent{
			Name:   i.CurrentSpec.Name,
			Status: s,
			Type:   Spec,
		},
		FileName: i.CurrentSpec.FileName,
	}
}

func NewScenarioEvent(i *m.ExecutionInfo, hasStarted bool) ScenarioEvent {
	s := Pass
	if i.CurrentScenario.GetIsFailed() {
		s = Fail
	}
	if hasStarted {
		s = Progress
	}
	return ScenarioEvent{
		BaseEvent: BaseEvent{
			Name:   i.CurrentScenario.Name,
			Status: s,
			Type:   Scenario,
		},
		SpecFileName: i.CurrentSpec.FileName,
	}
}

func NewStepEvent(i *m.ExecutionInfo, hasStarted bool) StepEvent {
	s := Pass
	if i.CurrentStep.GetIsFailed() {
		s = Fail
	}
	if hasStarted {
		s = Progress
	}
	return StepEvent{
		BaseEvent: BaseEvent{
			Name:   i.CurrentStep.Step.ActualStepText,
			Status: s,
			Type:   Step,
		},
		ScenarioName: i.CurrentScenario.Name,
		SpecFileName: i.CurrentSpec.FileName,
	}
}

func NewEndEvent(isFailed bool) EndEvent {
	s := Pass
	if isFailed {
		s = Fail
	}
	return EndEvent{
		BaseEvent: BaseEvent{
			Status: s,
			Type:   End,
		},
	}
}
