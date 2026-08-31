/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package terminal

import (
	"errors"
	"fmt"

	"github.com/pterm/pterm"
)

// ErrNoDefault is returned when running non-interactively (Yes()==true) and no
// default value was supplied for a prompt that requires one.
var ErrNoDefault = errors.New("non-interactive prompt requires a default value")

// Confirm asks a yes/no question.
//
// When Yes()==true the prompt is skipped and the function returns true,
// regardless of defaultYes. defaultYes is only used in interactive mode to
// pick the value when the user hits Enter.
func (t *Terminal) Confirm(question string, defaultYes bool) (bool, error) {
	if t.yes {
		return true, nil
	}
	if t.confirmFn != nil {
		return t.confirmFn(question, defaultYes)
	}
	return pterm.DefaultInteractiveConfirm.
		WithDefaultValue(defaultYes).
		Show(question)
}

// Select presents a list of options and returns the chosen one.
//
// When Yes()==true the prompt is skipped and defaultOption is returned.
// defaultOption must be one of the options. If defaultOption is empty in
// non-interactive mode, ErrNoDefault is returned.
func (t *Terminal) Select(question string, options []string, defaultOption string) (string, error) {
	if t.yes {
		if defaultOption == "" {
			return "", fmt.Errorf("%w: %s", ErrNoDefault, question)
		}
		return defaultOption, nil
	}
	prompt := pterm.DefaultInteractiveSelect.WithOptions(options)
	if defaultOption != "" {
		prompt = prompt.WithDefaultOption(defaultOption)
	}
	return prompt.Show(question)
}

// Input prompts the user for free-form text.
//
// When Yes()==true the prompt is skipped and defaultValue is returned. If
// defaultValue is empty in non-interactive mode, ErrNoDefault is returned.
func (t *Terminal) Input(question string, defaultValue string) (string, error) {
	if t.yes {
		if defaultValue == "" {
			return "", fmt.Errorf("%w: %s", ErrNoDefault, question)
		}
		return defaultValue, nil
	}
	p := pterm.DefaultInteractiveTextInput
	prompt := &p
	if defaultValue != "" {
		prompt = prompt.WithDefaultValue(defaultValue)
	}
	return prompt.Show(question)
}
