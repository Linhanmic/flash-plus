/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package event

import (
	"os"
	"path/filepath"

	m "github.com/getgauge/flash/gauge_messages"
)

type Status string
type EventType string

const (
	Pass     Status    = "pass"
	Fail     Status    = "fail"
	Progress Status    = "progress"
	Skip     Status    = "skip"

	Suite    EventType = "suite"
	Spec     EventType = "spec"
	Scenario EventType = "scenario"
	Concept  EventType = "concept"
	Step     EventType = "step"
	End      EventType = "end"
)

// Event is a JSON-serializable Gauge execution event.
type Event struct {
	Type             EventType `json:"type"`
	Status           Status    `json:"status,omitempty"`
	Name             string    `json:"name,omitempty"`
	FileName         string    `json:"fileName,omitempty"`
	SpecFileName     string    `json:"specFileName,omitempty"`
	ScenarioName     string    `json:"scenarioName,omitempty"`
	ConceptName      string    `json:"conceptName,omitempty"`
	ErrorMessage     string    `json:"errorMessage,omitempty"`
	StackTrace       string    `json:"stackTrace,omitempty"`
	ProjectName      string    `json:"projectName,omitempty"`
	Tags             []string  `json:"tags,omitempty"`
	SpecsFailed      int32     `json:"specsFailed,omitempty"`
	SpecsSkipped     int32     `json:"specsSkipped,omitempty"`
	ExecutionTime    int64     `json:"executionTime,omitempty"`
	SuccessRate      float32   `json:"successRate,omitempty"`
	Environment      string    `json:"environment,omitempty"`
	Parameters       []Param   `json:"parameters,omitempty"`
}

func ProjectNameFromEnv() string {
	if name := os.Getenv("GAUGE_PROJECT_ROOT"); name != "" {
		return filepath.Base(name)
	}
	return ""
}

func statusOf(hasStarted, failed bool) Status {
	if hasStarted {
		return Progress
	}
	if failed {
		return Fail
	}
	return Pass
}

func NewSuiteEvent(i *m.ExecutionInfo) Event {
	name := ProjectNameFromEnv()
	if i != nil && i.GetProjectName() != "" {
		name = i.GetProjectName()
	}
	if name == "" {
		name = "Gauge"
	}
	return Event{
		Type:        Suite,
		Status:      Progress,
		Name:        name,
		ProjectName: name,
	}
}

func NewSpecEvent(i *m.ExecutionInfo, hasStarted bool) Event {
	if i == nil || i.GetCurrentSpec() == nil {
		return Event{
			Type:   Spec,
			Status: statusOf(hasStarted, false),
			Name:   "(unknown spec)",
		}
	}
	spec := i.GetCurrentSpec()
	return Event{
		Type:     Spec,
		Status:   statusOf(hasStarted, spec.GetIsFailed()),
		Name:     spec.GetName(),
		FileName: spec.GetFileName(),
		Tags:     spec.GetTags(),
	}
}

func NewScenarioEvent(i *m.ExecutionInfo, hasStarted bool) Event {
	if i == nil || i.GetCurrentScenario() == nil {
		return Event{
			Type:   Scenario,
			Status: statusOf(hasStarted, false),
			Name:   "(unknown scenario)",
		}
	}
	scn := i.GetCurrentScenario()
	fileName := ""
	if spec := i.GetCurrentSpec(); spec != nil {
		fileName = spec.GetFileName()
	}
	return Event{
		Type:         Scenario,
		Status:       statusOf(hasStarted, scn.GetIsFailed()),
		Name:         scn.GetName(),
		SpecFileName: fileName,
		Tags:         scn.GetTags(),
	}
}

func NewStepEvent(i *m.ExecutionInfo, hasStarted bool) Event {
	if i == nil || i.GetCurrentStep() == nil {
		return Event{
			Type:   Step,
			Status: statusOf(hasStarted, false),
			Name:   "(unknown step)",
		}
	}
	step := i.GetCurrentStep()
	name, params := stepDisplayName(step.GetStep(), "(unknown step)")
	specFile := ""
	if spec := i.GetCurrentSpec(); spec != nil {
		specFile = spec.GetFileName()
	}
	scenarioName := ""
	if scn := i.GetCurrentScenario(); scn != nil {
		scenarioName = scn.GetName()
	}
	ev := Event{
		Type:         Step,
		Status:       statusOf(hasStarted, step.GetIsFailed()),
		Name:         name,
		ScenarioName: scenarioName,
		SpecFileName: specFile,
		Parameters:   params,
	}
	if !hasStarted && step.GetIsFailed() {
		ev.ErrorMessage = step.GetErrorMessage()
		ev.StackTrace = step.GetStackTrace()
	}
	return ev
}

func NewConceptEvent(i *m.ExecutionInfo, hasStarted bool) Event {
	name := "(unknown concept)"
	var params []Param
	failed := false
	specFile := ""
	scenarioName := ""
	if i != nil {
		if spec := i.GetCurrentSpec(); spec != nil {
			specFile = spec.GetFileName()
		}
		if scn := i.GetCurrentScenario(); scn != nil {
			scenarioName = scn.GetName()
		}
		if step := i.GetCurrentStep(); step != nil {
			failed = step.GetIsFailed()
			name, params = stepDisplayName(step.GetStep(), "(unknown concept)")
		}
	}
	ev := Event{
		Type:         Concept,
		Status:       statusOf(hasStarted, failed),
		Name:         name,
		ConceptName:  name,
		SpecFileName: specFile,
		ScenarioName: scenarioName,
		Parameters:   params,
	}
	if !hasStarted && failed && i != nil && i.GetCurrentStep() != nil {
		ev.ErrorMessage = i.GetCurrentStep().GetErrorMessage()
		ev.StackTrace = i.GetCurrentStep().GetStackTrace()
	}
	return ev
}

func NewEndEvent(isFailed bool) Event {
	s := Pass
	if isFailed {
		s = Fail
	}
	return Event{Type: End, Status: s}
}

func NewEndEventFromSuite(r *m.ProtoSuiteResult) Event {
	if r == nil {
		return NewEndEvent(false)
	}
	ev := NewEndEvent(r.GetFailed())
	ev.Name = r.GetProjectName()
	ev.ProjectName = r.GetProjectName()
	ev.SpecsFailed = r.GetSpecsFailedCount()
	ev.SpecsSkipped = r.GetSpecsSkippedCount()
	ev.ExecutionTime = r.GetExecutionTime()
	ev.SuccessRate = r.GetSuccessRate()
	ev.Environment = r.GetEnvironment()
	return ev
}
