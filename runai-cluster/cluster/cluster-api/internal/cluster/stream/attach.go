/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"net/http"

	corev1 "k8s.io/api/core/v1"
)

type AttachOptions struct {
	Options
}

func (o *AttachOptions) ToK8sPodAttachOptions() (*corev1.PodAttachOptions, error) {
	attachOptions := corev1.PodAttachOptions{
		Stdin:     o.Stdin,
		Stdout:    o.Stdout,
		Stderr:    o.Stderr,
		TTY:       o.TTY,
		Container: o.Container,
	}

	return &attachOptions, nil
}

type AttachOptionsBuilder struct {
	*OptionsBuilder
}

func NewAttachOptionsBuilder() *AttachOptionsBuilder {
	return &AttachOptionsBuilder{NewOptionsBuilder()}
}

func (aob *AttachOptionsBuilder) Build() (AttachOptions, error) {
	return AttachOptions{aob.OptionsBuilder.Build()}, nil
}

func AttachOptionsFromRequest(r *http.Request) (*AttachOptions, error) {
	builder := NewAttachOptionsBuilder()

	err := fromGetRequest(r, builder.OptionsBuilder)
	if err != nil {
		return nil, err
	}

	options, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return &options, nil
}
