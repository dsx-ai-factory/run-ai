/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"fmt"
	"net/http"
	"strconv"
)

type Options struct {
	Stdin     bool
	Stdout    bool
	Stderr    bool
	TTY       bool
	Pod       string
	Container string
	Command   []string

	TTYWidth  uint16
	TTYHeight uint16
}

type OptionsBuilder struct {
	options Options
}

func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{options: Options{}}
}

func (b *OptionsBuilder) WithPod(pod string) *OptionsBuilder {
	b.options.Pod = pod
	return b
}

func (b *OptionsBuilder) WithContainer(container string) *OptionsBuilder {
	b.options.Container = container
	return b
}

func (b *OptionsBuilder) WithStdin(stdBool bool) *OptionsBuilder {
	b.options.Stdin = stdBool
	return b
}

func (b *OptionsBuilder) WithStdout(stdBool bool) *OptionsBuilder {
	b.options.Stdout = stdBool
	return b
}

func (b *OptionsBuilder) WithStderr(stdBool bool) *OptionsBuilder {
	b.options.Stderr = stdBool
	return b
}

func (b *OptionsBuilder) WithTTY(ttyBool bool) *OptionsBuilder {
	b.options.TTY = ttyBool
	return b
}

func (b *OptionsBuilder) WithTTYWidth(size uint16) *OptionsBuilder {
	b.options.TTYWidth = size
	return b
}

func (b *OptionsBuilder) WithTTYHeight(size uint16) *OptionsBuilder {
	b.options.TTYHeight = size
	return b
}

func (b *OptionsBuilder) WithCommand(command []string) *OptionsBuilder {
	b.options.Command = command
	return b
}

func (b *OptionsBuilder) Build() Options {
	return b.options
}

func fromGetRequest(r *http.Request, builder *OptionsBuilder) error {
	pod := r.URL.Query().Get("pod")
	if pod != "" {
		builder.WithPod(pod)
	}

	container := r.URL.Query().Get("container")
	if container != "" {
		builder.WithContainer(container)
	}

	stdinStr := r.URL.Query().Get("stdin")
	if stdinStr != "" {
		stdinBool, err := strconv.ParseBool(stdinStr)
		if err != nil {
			return fmt.Errorf("invalid stdin value: %v", err)
		}

		builder.WithStdin(stdinBool)
	}

	stdoutStr := r.URL.Query().Get("stdout")
	if stdoutStr != "" {
		stdoutBool, err := strconv.ParseBool(stdoutStr)
		if err != nil {
			return fmt.Errorf("invalid stdout value: %v", err)
		}

		builder.WithStdout(stdoutBool)
	}

	stderrStr := r.URL.Query().Get("stderr")
	if stderrStr != "" {
		stderrBool, err := strconv.ParseBool(stderrStr)
		if err != nil {
			return fmt.Errorf("invalid stderr value: %v", err)
		}

		builder.WithStderr(stderrBool)
	}

	ttyStr := r.URL.Query().Get("tty")
	if ttyStr != "" {
		ttyBool, err := strconv.ParseBool(ttyStr)
		if err != nil {
			return fmt.Errorf("invalid tty value: %v", err)
		}

		builder.WithTTY(ttyBool)
	}

	ttyWidth := r.URL.Query().Get("width")
	if ttyWidth != "" {
		ttySize, err := ParseUInt16(ttyWidth)
		if err != nil {
			return fmt.Errorf("invalid tty width value: %v", err)
		}

		builder.WithTTYWidth(ttySize)
	}

	ttyHeight := r.URL.Query().Get("height")
	if ttyHeight != "" {
		ttySize, err := ParseUInt16(ttyHeight)
		if err != nil {
			return fmt.Errorf("invalid tty height value: %v", err)
		}

		builder.WithTTYHeight(ttySize)
	}

	return nil
}

func ParseUInt16(value string) (uint16, error) {
	uintValue, err := strconv.ParseInt(value, 10, 16)
	return uint16(uintValue), err
}
