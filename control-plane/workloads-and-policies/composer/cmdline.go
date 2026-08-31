/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"

	"golang.org/x/exp/utf8string"
)

type StateEnum int

const (
	StateStart StateEnum = iota
	StateArg
	StateQuotes
)

// ParseCommandLine based on: https://newbedev.com/parse-a-command-line-string-into-flags-and-arguments-in-golang
// this function splits a string based on unix shell rules
// if no values found, return nil cause that's how k8s expects it
func ParseCommandLine(origin string) ([]string, error) {
	command := utf8string.NewString(origin)
	var args []string

	state := StateStart

	current := ""
	quote := ""
	escapeNext := false

	for i := 0; i < command.RuneCount(); i++ {
		c := command.At(i)

		if c == '\\' {
			escapeNext = true
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if state == StateQuotes {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = StateStart
			}
			continue
		}

		if c == '"' || c == '\'' {
			state = StateQuotes
			quote = string(c)
			continue
		}

		if state == StateArg {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = StateStart
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = StateArg
			current += string(c)
		}
	}

	if state == StateQuotes {
		return []string{}, fmt.Errorf("unclosed quote in command line: %s", command)
	}

	if current != "" {
		args = append(args, current)
	}

	if len(args) == 0 {
		return nil, nil
	}

	return args, nil
}
