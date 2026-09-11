/*----------------------------------------------------------------
 *  Flash Plus - Gauge execution progress reporter
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package event

import (
	"fmt"
	"regexp"
	"strings"

	m "github.com/getgauge/flash/gauge_messages"
)

// Param is a JSON-serializable Gauge step/concept parameter.
type Param struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

var placeholderRE = regexp.MustCompile(`<[^<>]+>`)

func paramsFromRequest(req *m.ExecuteStepRequest) []Param {
	if req == nil {
		return nil
	}
	out := make([]Param, 0, len(req.GetParameters()))
	for _, p := range req.GetParameters() {
		if p == nil {
			continue
		}
		out = append(out, Param{Name: p.GetName(), Value: displayParamValue(p)})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func displayParamValue(p *m.Parameter) string {
	switch p.GetParameterType() {
	case m.Parameter_Table, m.Parameter_Special_Table:
		return "<table>"
	default:
		v := p.GetValue()
		if len(v) > 120 {
			v = v[:117] + "..."
		}
		return v
	}
}

func quoteParam(value string) string {
	if value == "<table>" {
		return value
	}
	if len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return value
	}
	return fmt.Sprintf("\"%s\"", value)
}

// ResolveStepText replaces Gauge placeholders with runtime parameter values.
// ActualStepText keeps <name> as written in the spec; the resolved values live in Parameters.
func ResolveStepText(actual, parsed string, params []Param) string {
	text := actual
	if text == "" {
		text = parsed
	}
	if text == "" || len(params) == 0 {
		return text
	}
	used := make([]bool, len(params))
	for i, p := range params {
		if p.Name == "" {
			continue
		}
		placeholder := "<" + p.Name + ">"
		if !strings.Contains(text, placeholder) {
			continue
		}
		text = strings.Replace(text, placeholder, quoteParam(p.Value), 1)
		used[i] = true
	}
	for i, p := range params {
		if used[i] {
			continue
		}
		if strings.Contains(text, "{}") {
			text = strings.Replace(text, "{}", quoteParam(p.Value), 1)
			used[i] = true
			continue
		}
		loc := placeholderRE.FindStringIndex(text)
		if loc == nil {
			break
		}
		text = text[:loc[0]] + quoteParam(p.Value) + text[loc[1]:]
		used[i] = true
	}
	return text
}

func stepDisplayName(req *m.ExecuteStepRequest, unknown string) (string, []Param) {
	params := paramsFromRequest(req)
	if req == nil {
		return unknown, params
	}
	name := ResolveStepText(req.GetActualStepText(), req.GetParsedStepText(), params)
	if name == "" {
		return unknown, params
	}
	return name, params
}
