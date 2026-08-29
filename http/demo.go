/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package http

import (
	"time"

	"github.com/Linhanmic/flash-plus/event"
)

// RunDemo publishes a realistic execution timeline for local UI debugging.
func RunDemo(e chan event.Event) {
	send := func(ev event.Event, d time.Duration) {
		e <- ev
		time.Sleep(d)
	}

	send(event.Event{Type: event.Suite, Status: event.Progress, Name: "flash-plus-demo", ProjectName: "flash-plus-demo"}, 200*time.Millisecond)

	loginSpec := "specs/login.spec"
	send(event.Event{Type: event.Spec, Status: event.Progress, Name: "User login", FileName: loginSpec, Tags: []string{"auth"}}, 250*time.Millisecond)
	send(event.Event{Type: event.Scenario, Status: event.Progress, Name: "Successful login with valid credentials", SpecFileName: loginSpec}, 200*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Progress, Name: "Open the login page", ScenarioName: "Successful login with valid credentials", SpecFileName: loginSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Pass, Name: "Open the login page", ScenarioName: "Successful login with valid credentials", SpecFileName: loginSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Progress, Name: "Enter username \"alice\" and password \"secret\"", ScenarioName: "Successful login with valid credentials", SpecFileName: loginSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Pass, Name: "Enter username \"alice\" and password \"secret\"", ScenarioName: "Successful login with valid credentials", SpecFileName: loginSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Progress, Name: "Click \"Sign in\"", ScenarioName: "Successful login with valid credentials", SpecFileName: loginSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Pass, Name: "Click \"Sign in\"", ScenarioName: "Successful login with valid credentials", SpecFileName: loginSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Scenario, Status: event.Pass, Name: "Successful login with valid credentials", SpecFileName: loginSpec}, 200*time.Millisecond)

	send(event.Event{Type: event.Scenario, Status: event.Progress, Name: "Login fails with a wrong password", SpecFileName: loginSpec}, 200*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Progress, Name: "Open the login page", ScenarioName: "Login fails with a wrong password", SpecFileName: loginSpec}, 160*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Pass, Name: "Open the login page", ScenarioName: "Login fails with a wrong password", SpecFileName: loginSpec}, 160*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Progress, Name: "Enter username \"alice\" and password \"wrong\"", ScenarioName: "Login fails with a wrong password", SpecFileName: loginSpec}, 200*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Fail, Name: "Enter username \"alice\" and password \"wrong\"", ScenarioName: "Login fails with a wrong password", SpecFileName: loginSpec, ErrorMessage: "Expected error banner \"Invalid password\", got \"Welcome alice\"", StackTrace: "at steps/login.js:42\n    throw new Error(\"Expected error banner\")"}, 250*time.Millisecond)
	send(event.Event{Type: event.Scenario, Status: event.Fail, Name: "Login fails with a wrong password", SpecFileName: loginSpec}, 200*time.Millisecond)
	send(event.Event{Type: event.Spec, Status: event.Fail, Name: "User login", FileName: loginSpec}, 250*time.Millisecond)

	searchSpec := "specs/search.spec"
	send(event.Event{Type: event.Spec, Status: event.Progress, Name: "Search results", FileName: searchSpec}, 200*time.Millisecond)
	send(event.Event{Type: event.Scenario, Status: event.Progress, Name: "Find an existing item", SpecFileName: searchSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Progress, Name: "Search for \"flash-plus\"", ScenarioName: "Find an existing item", SpecFileName: searchSpec}, 160*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Pass, Name: "Search for \"flash-plus\"", ScenarioName: "Find an existing item", SpecFileName: searchSpec}, 160*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Progress, Name: "The result list should contain \"Flash Plus\"", ScenarioName: "Find an existing item", SpecFileName: searchSpec}, 160*time.Millisecond)
	send(event.Event{Type: event.Step, Status: event.Pass, Name: "The result list should contain \"Flash Plus\"", ScenarioName: "Find an existing item", SpecFileName: searchSpec}, 160*time.Millisecond)
	send(event.Event{Type: event.Scenario, Status: event.Pass, Name: "Find an existing item", SpecFileName: searchSpec}, 180*time.Millisecond)
	send(event.Event{Type: event.Spec, Status: event.Pass, Name: "Search results", FileName: searchSpec}, 200*time.Millisecond)

	send(event.Event{
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
