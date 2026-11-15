// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

var conditionTypedValue = regexp.MustCompile(`^("(?:[^"\\]|\\.)*"|'[^']*'|true\b|false\b|null\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)`)

const (
	maxConditionBytes = 4096
	maxConditionTerms = 128
	maxConditionDepth = 16
)

type conditionParser struct {
	text       string
	pos, count int
	references []string
}

// Validate the restricted visibility grammar and collect referenced values paths.
// This parses comparisons, boolean literals, &&, || and parentheses; it does not
// evaluate expressions or run a general-purpose template/expression language.
func parseCondition(value string) ([]string, *api.Error) {
	if len(value) > maxConditionBytes {
		return nil, api.NewValidationError(fmt.Sprintf("condition exceeds %d bytes", maxConditionBytes))
	}
	p := &conditionParser{text: value}
	if err := p.or(0); err != nil {
		return nil, err
	}
	p.whitespace()
	if p.pos != len(value) {
		return nil, api.NewValidationError("unexpected condition text")
	}
	return p.references, nil
}

func (p *conditionParser) whitespace() {
	for p.pos < len(p.text) && unicode.IsSpace(rune(p.text[p.pos])) {
		p.pos++
	}
}
func (p *conditionParser) take(value string) bool {
	p.whitespace()
	if strings.HasPrefix(p.text[p.pos:], value) {
		p.pos += len(value)
		return true
	}
	return false
}
func (p *conditionParser) or(depth int) *api.Error {
	if err := p.and(depth); err != nil {
		return err
	}
	for p.take("||") {
		if err := p.and(depth); err != nil {
			return err
		}
	}
	return nil
}
func (p *conditionParser) and(depth int) *api.Error {
	if err := p.primary(depth); err != nil {
		return err
	}
	for p.take("&&") {
		if err := p.primary(depth); err != nil {
			return err
		}
	}
	return nil
}

func (p *conditionParser) primary(depth int) *api.Error {
	p.count++
	if p.count > maxConditionTerms || depth > maxConditionDepth {
		return api.NewValidationError(fmt.Sprintf("condition exceeds %d terms or %d nesting levels", maxConditionTerms, maxConditionDepth))
	}
	if p.take("(") {
		if err := p.or(depth + 1); err != nil {
			return err
		}
		if !p.take(")") {
			return api.NewValidationError("missing closing condition parenthesis")
		}
		return nil
	}
	p.whitespace()
	start := p.pos
	quoted, escaped, brackets := false, false, 0
	for p.pos < len(p.text) {
		c := p.text[p.pos]
		if quoted {
			if c == '"' && !escaped {
				quoted = false
			}
			if c == '\\' && !escaped {
				escaped = true
			} else {
				escaped = false
			}
			p.pos++
			continue
		}
		if c == '"' {
			quoted = true
		}
		if c == '[' {
			brackets++
		}
		if c == ']' {
			brackets--
		}
		if brackets == 0 && (unicode.IsSpace(rune(c)) || strings.ContainsRune("=!&|()", rune(c))) {
			break
		}
		p.pos++
	}
	valuePath := p.text[start:p.pos]
	p.whitespace()
	if (valuePath == literalTrue || valuePath == literalFalse) && (p.pos == len(p.text) || !strings.ContainsRune("=!", rune(p.text[p.pos]))) {
		return nil
	}
	if _, err := parsePath(valuePath); err != nil {
		return api.NewValidationError(fmt.Sprintf("invalid condition path: %s", err.Message))
	}
	if !p.take("!=") && !p.take("=") {
		return api.NewValidationError("condition requires = or !=")
	}
	p.references = append(p.references, valuePath)
	p.whitespace()
	if p.pos < len(p.text) && p.text[p.pos] == '=' {
		return api.NewValidationError("condition requires a single equality operator")
	}
	match := conditionTypedValue.FindString(p.text[p.pos:])
	if match != "" && match[0] != '\'' && match[0] != '"' {
		end := p.pos
		for end < len(p.text) && !strings.ContainsRune("&|()", rune(p.text[end])) {
			end++
		}
		if strings.TrimSpace(p.text[p.pos:end]) != match {
			match = ""
		}
	}
	if match != "" {
		if match[0] != '\'' {
			var value any
			if err := json.Unmarshal([]byte(match), &value); err != nil {
				return api.NewValidationError(fmt.Sprintf("invalid condition literal: %v", err))
			}
			if number, ok := value.(float64); ok && (math.IsNaN(number) || math.IsInf(number, 0)) {
				return api.NewValidationError("condition number must be finite")
			}
		}
		p.pos += len(match)
		return nil
	}
	if p.pos < len(p.text) && (p.text[p.pos] == '"' || p.text[p.pos] == '\'') {
		return api.NewValidationError("unterminated condition string")
	}
	for p.pos < len(p.text) && !strings.ContainsRune("&|()", rune(p.text[p.pos])) {
		p.pos++
	}
	return nil
}

type declarationQuestion struct {
	key, parent string
	references  []string
}

// Resolve references after every field is known, allowing forward references.
// Parent visibility and show_if dependencies share one graph for cycle detection.
func (v *declarationValidator) checkConditions() *api.Error {
	known := map[string]bool{}
	for _, q := range v.questions {
		known[q.key] = true
	}
	graph := datastructs.NewDependenciesGraph[string]()
	for _, q := range v.questions {
		if q.parent != "" {
			graph.AddDependency(q.key, q.parent)
		}
		for _, reference := range q.references {
			segments, err := parsePath(reference)
			if err != nil {
				return err
			}
			key := canonicalPath(segments)
			if known[key] {
				graph.AddDependency(q.key, key)
			} else if _, exists := readPath(v.defaults, segments); v.defaults != nil && !exists {
				v.warnings = append(v.warnings, api.AiApplicationChartInspectionWarning{Code: "UNKNOWN_CONDITION_REFERENCE", Message: "Condition path is absent from chart defaults and declared fields", Path: reference})
			}
		}
	}
	if _, err := graph.FindOrder(); err != nil {
		return api.NewValidationError("question visibility conditions contain a cycle")
	}
	return nil
}
