/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
)

const (
	HeaderCommand = "Command"
	QueryCommand  = "command"
)

var (
	ErrQueryAndCommandBothEmpty         = errors.New("command header and command query parameters are both empty")
	ErrQueryAndCommandMutuallyExclusive = errors.New("command header and command query parameters are mutually exclusive")
)

type ExecOptions struct {
	Options
}

type ExecOptionsBuilder struct {
	*OptionsBuilder
}

func NewExecOptionsBuilder() *ExecOptionsBuilder {
	return &ExecOptionsBuilder{&OptionsBuilder{Options{}}}
}

func (bob *ExecOptionsBuilder) Build() (ExecOptions, error) {
	eo := ExecOptions{bob.OptionsBuilder.Build()}
	if err := eo.Validate(); err != nil {
		return ExecOptions{}, err
	}

	return eo, nil
}

func (o *ExecOptions) ToK8sPodExecOptions() (*corev1.PodExecOptions, error) {
	execOptions := corev1.PodExecOptions{
		Stdin:     o.Stdin,
		Stdout:    o.Stdout,
		Stderr:    o.Stderr,
		TTY:       o.TTY,
		Container: o.Container,
		Command:   o.Command,
	}

	return &execOptions, nil
}

func (o *ExecOptions) Validate() error {
	if len(o.Command) == 0 {
		return fmt.Errorf("exec command cannot be empty")
	}

	return nil
}

func ExecOptionsFromRequest(r *http.Request) (*ExecOptions, error) {
	builder := NewExecOptionsBuilder()

	err := fromGetRequest(r, builder.OptionsBuilder)
	if err != nil {
		return nil, err
	}

	command, err := getCommand(r)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve command: %w", err)
	}

	builder.WithCommand(command)

	options, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &options, nil
}

func getCommand(req *http.Request) ([]string, error) {
	query := req.URL.Query().Get(QueryCommand)
	header := req.Header.Get(HeaderCommand)

	if query == "" && header == "" {
		return nil, ErrQueryAndCommandBothEmpty
	}

	if query != "" && header != "" {
		return nil, ErrQueryAndCommandMutuallyExclusive
	}

	if query != "" {
		return decodeCommand(query)
	}

	return decodeCommand(header)
}

func decodeCommand(command string) ([]string, error) {
	enc, err := base64.StdEncoding.DecodeString(command)
	if err != nil {
		return nil, fmt.Errorf("failed to decode command (%s): %w", command, err)
	}

	var cmd []string

	err = json.Unmarshal(enc, &cmd)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
