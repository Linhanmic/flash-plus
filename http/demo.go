/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"time"

	"github.com/Linhanmic/flash-plus/event"
	m "github.com/getgauge/flash/gauge_messages"
)

func send(e chan event.Event, ev event.Event, d time.Duration) {
	e <- ev
	time.Sleep(d)
}

func specStart(fileName, heading string, tags ...string) event.Event {
	return event.Event{Type: event.Spec, Status: event.Progress, Name: heading, FileName: fileName, Tags: tags}
}

func specEnd(fileName, heading string, status event.Status) event.Event {
	return event.Event{Type: event.Spec, Status: status, Name: heading, FileName: fileName}
}

func scenarioStart(specFile, name string) event.Event {
	return event.Event{Type: event.Scenario, Status: event.Progress, Name: name, SpecFileName: specFile}
}

func scenarioEnd(specFile, name string, status event.Status) event.Event {
	return event.Event{Type: event.Scenario, Status: status, Name: name, SpecFileName: specFile}
}

func conceptStart(specFile, scenario, name string) event.Event {
	return event.Event{Type: event.Concept, Status: event.Progress, Name: name, ConceptName: name, SpecFileName: specFile, ScenarioName: scenario}
}

func conceptEnd(specFile, scenario, name string, status event.Status) event.Event {
	return event.Event{Type: event.Concept, Status: status, Name: name, ConceptName: name, SpecFileName: specFile, ScenarioName: scenario}
}

func stepStart(specFile, scenario, name string) event.Event {
	return event.Event{Type: event.Step, Status: event.Progress, Name: name, SpecFileName: specFile, ScenarioName: scenario}
}

func stepEnd(specFile, scenario, name string, status event.Status, errMsg, stack string) event.Event {
	return event.Event{Type: event.Step, Status: status, Name: name, SpecFileName: specFile, ScenarioName: scenario, ErrorMessage: errMsg, StackTrace: stack}
}

func dynamicLabel(value string) []*m.Parameter {
	return []*m.Parameter{{Name: "label", Value: value, ParameterType: m.Parameter_Dynamic}}
}

func execInfo(specFile, scenario, actual string, params []*m.Parameter, failed bool, errMsg, stack string) *m.ExecutionInfo {
	return &m.ExecutionInfo{
		CurrentSpec:     &m.SpecInfo{FileName: specFile},
		CurrentScenario: &m.ScenarioInfo{Name: scenario},
		CurrentStep: &m.StepInfo{
			Step:         &m.ExecuteStepRequest{ActualStepText: actual, Parameters: params},
			IsFailed:     failed,
			ErrorMessage: errMsg,
			StackTrace:   stack,
		},
	}
}

// RunDemo publishes a realistic execution timeline for local UI debugging.
func RunDemo(e chan event.Event) {
	send(e, event.Event{Type: event.Suite, Status: event.Progress, Name: "flash-plus-demo", ProjectName: "flash-plus-demo"}, 200*time.Millisecond)

	envSpec := "specs/env.spec"
	envHeading := "Demo environment"
	envScenario := "Check environment label"
	envConcept := "aaa <label>"
	envStep := "Verify demo environment label is <label>"
	send(e, specStart(envSpec, envHeading), 200*time.Millisecond)
	send(e, scenarioStart(envSpec, envScenario), 180*time.Millisecond)
	send(e, event.NewConceptEvent(execInfo(envSpec, envScenario, envConcept, dynamicLabel("staging"), false, "", ""), true), 180*time.Millisecond)
	send(e, event.NewStepEvent(execInfo(envSpec, envScenario, envStep, dynamicLabel("staging"), false, "", ""), true), 180*time.Millisecond)
	send(e, event.NewStepEvent(execInfo(envSpec, envScenario, envStep, dynamicLabel("staging"), false, "", ""), false), 180*time.Millisecond)
	send(e, event.NewStepEvent(execInfo(envSpec, envScenario, envStep, dynamicLabel("staging"), false, "", ""), true), 180*time.Millisecond)
	send(e, event.NewStepEvent(execInfo(envSpec, envScenario, envStep, dynamicLabel("staging"), false, "", ""), false), 180*time.Millisecond)
	send(e, event.NewConceptEvent(execInfo(envSpec, envScenario, envConcept, dynamicLabel("staging"), false, "", ""), false), 180*time.Millisecond)
	send(e, scenarioEnd(envSpec, envScenario, event.Pass), 180*time.Millisecond)
	send(e, specEnd(envSpec, envHeading, event.Pass), 200*time.Millisecond)

	loginSpec := "specs/login.spec"
	loginHeading := "User login"
	okScenario := "Successful login with valid credentials"
	failScenario := "Login fails with a wrong password"
	loginConcept := "Login as \"demo_user\""

	send(e, specStart(loginSpec, loginHeading, "auth"), 250*time.Millisecond)
	send(e, scenarioStart(loginSpec, okScenario), 200*time.Millisecond)
	send(e, conceptStart(loginSpec, okScenario, loginConcept), 180*time.Millisecond)
	send(e, stepStart(loginSpec, okScenario, "Enter username \"alice\" and password \"secret\""), 180*time.Millisecond)
	send(e, stepEnd(loginSpec, okScenario, "Enter username \"alice\" and password \"secret\"", event.Pass, "", ""), 180*time.Millisecond)
	send(e, stepStart(loginSpec, okScenario, "Click \"Sign in\""), 180*time.Millisecond)
	send(e, stepEnd(loginSpec, okScenario, "Click \"Sign in\"", event.Pass, "", ""), 180*time.Millisecond)
	send(e, conceptEnd(loginSpec, okScenario, loginConcept, event.Pass), 180*time.Millisecond)
	send(e, stepStart(loginSpec, okScenario, "The dashboard should show \"Welcome alice\""), 160*time.Millisecond)
	send(e, stepEnd(loginSpec, okScenario, "The dashboard should show \"Welcome alice\"", event.Pass, "", ""), 160*time.Millisecond)
	send(e, scenarioEnd(loginSpec, okScenario, event.Pass), 200*time.Millisecond)

	send(e, scenarioStart(loginSpec, failScenario), 200*time.Millisecond)
	send(e, stepStart(loginSpec, failScenario, "Open the login page"), 160*time.Millisecond)
	send(e, stepEnd(loginSpec, failScenario, "Open the login page", event.Pass, "", ""), 160*time.Millisecond)
	send(e, conceptStart(loginSpec, failScenario, loginConcept), 180*time.Millisecond)
	send(e, stepStart(loginSpec, failScenario, "Enter username \"alice\" and password \"wrong\""), 200*time.Millisecond)
	send(e, stepEnd(loginSpec, failScenario, "Enter username \"alice\" and password \"wrong\"", event.Fail, "Expected error banner \"Invalid password\", got \"Welcome alice\"", "at steps/login.js:42\n    throw new Error(\"Expected error banner\")"), 250*time.Millisecond)
	send(e, conceptEnd(loginSpec, failScenario, loginConcept, event.Fail), 180*time.Millisecond)
	send(e, scenarioEnd(loginSpec, failScenario, event.Fail), 200*time.Millisecond)
	send(e, specEnd(loginSpec, loginHeading, event.Fail), 250*time.Millisecond)

	searchSpec := "specs/search.spec"
	searchHeading := "Search results"
	searchScenario := "Find an existing item"
	send(e, specStart(searchSpec, searchHeading), 200*time.Millisecond)
	send(e, scenarioStart(searchSpec, searchScenario), 180*time.Millisecond)
	send(e, stepStart(searchSpec, searchScenario, "Search for \"flash-plus\""), 160*time.Millisecond)
	send(e, stepEnd(searchSpec, searchScenario, "Search for \"flash-plus\"", event.Pass, "", ""), 160*time.Millisecond)
	send(e, stepStart(searchSpec, searchScenario, "The result list should contain \"Flash Plus\""), 160*time.Millisecond)
	send(e, stepEnd(searchSpec, searchScenario, "The result list should contain \"Flash Plus\"", event.Pass, "", ""), 160*time.Millisecond)
	send(e, scenarioEnd(searchSpec, searchScenario, event.Pass), 180*time.Millisecond)
	send(e, specEnd(searchSpec, searchHeading, event.Pass), 200*time.Millisecond)

	send(e, event.Event{
		Type:          event.End,
		Status:        event.Fail,
		ProjectName:   "flash-plus-demo",
		SpecsFailed:   1,
		SpecsSkipped:  0,
		ExecutionTime: 4200,
		SuccessRate:   50,
		Environment:   "default",
	}, 0)
}
