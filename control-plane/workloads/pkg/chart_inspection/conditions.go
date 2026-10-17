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
)

var conditionTypedValue = regexp.MustCompile(`^("(?:[^"\\]|\\.)*"|'[^']*'|true\b|false\b|null\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)`)

type conditionParser struct {
	text       string
	pos, count int
	references []string
}

// Validate the restricted visibility grammar and collect referenced values paths.
// This parses comparisons, boolean literals, &&, || and parentheses; it does not
// evaluate expressions or run a general-purpose template/expression language.
func parseCondition(value string) ([]string, error) {
	if len(value) > 4096 {
		return nil, fmt.Errorf("condition is too long")
	}
	p := &conditionParser{text: value}
	if err := p.or(0); err != nil {
		return nil, err
	}
	p.whitespace()
	if p.pos != len(value) {
		return nil, fmt.Errorf("unexpected condition text")
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
func (p *conditionParser) or(depth int) error {
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
func (p *conditionParser) and(depth int) error {
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

func (p *conditionParser) primary(depth int) error {
	p.count++
	if p.count > 128 || depth > 16 {
		return fmt.Errorf("condition is too complex")
	}
	if p.take("(") {
		if err := p.or(depth + 1); err != nil {
			return err
		}
		if !p.take(")") {
			return fmt.Errorf("missing closing condition parenthesis")
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
		return fmt.Errorf("invalid condition path: %w", err)
	}
	if !p.take("!=") && !p.take("=") {
		return fmt.Errorf("condition requires = or !=")
	}
	p.references = append(p.references, valuePath)
	p.whitespace()
	if p.pos < len(p.text) && p.text[p.pos] == '=' {
		return fmt.Errorf("condition requires a single equality operator")
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
				return fmt.Errorf("invalid condition literal: %w", err)
			}
			if number, ok := value.(float64); ok && (math.IsNaN(number) || math.IsInf(number, 0)) {
				return fmt.Errorf("condition number must be finite")
			}
		}
		p.pos += len(match)
		return nil
	}
	if p.pos < len(p.text) && (p.text[p.pos] == '"' || p.text[p.pos] == '\'') {
		return fmt.Errorf("unterminated condition string")
	}
	for p.pos < len(p.text) && !strings.ContainsRune("&|()", rune(p.text[p.pos])) {
		p.pos++
	}
	return nil
}
