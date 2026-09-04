/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/util"
)

type Options struct {
	Pod          string
	Container    string
	TailLines    *int64
	Timestamps   bool
	Follow       bool
	Previous     bool
	LimitBytes   int64
	SinceSeconds time.Duration //nolint:staticcheck
	SinceTime    string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) ToK8sLogsOptions() (*corev1.PodLogOptions, error) {
	logOptions := corev1.PodLogOptions{
		Container:  o.Container,
		TailLines:  o.TailLines,
		Timestamps: o.Timestamps,
		Follow:     o.Follow,
		Previous:   o.Previous,
	}

	if len(o.SinceTime) > 0 {
		t, err := util.ParseRFC3339(o.SinceTime, metav1.Now)
		if err != nil {
			return nil, err
		}

		logOptions.SinceTime = &t
	}

	if o.LimitBytes != 0 {
		logOptions.LimitBytes = &o.LimitBytes
	}

	if o.SinceSeconds != 0 {
		// round up to the nearest second
		sec := int64(o.SinceSeconds.Round(time.Second).Seconds())
		logOptions.SinceSeconds = &sec
	}

	return &logOptions, nil
}

func (o *Options) Validate() error {
	if len(o.SinceTime) > 0 && o.SinceSeconds != 0 {
		return errors.New("at most one of `sinceTime` or `sinceSeconds` may be specified")
	}

	if o.LimitBytes < 0 {
		return errors.New("--limit-bytes must be greater than 0")
	}

	if o.SinceSeconds < 0 {
		return errors.New("sinceSeconds must be greater than 0")
	}

	if o.TailLines != nil && *o.TailLines < -1 {
		return errors.New("--tail must be greater than or equal to -1")
	}

	return nil
}

type OptionsBuilder struct {
	options Options
}

func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{}
}

func (b *OptionsBuilder) WithPod(pod string) *OptionsBuilder {
	b.options.Pod = pod
	return b
}

func (b *OptionsBuilder) WithContainer(container string) *OptionsBuilder {
	b.options.Container = container
	return b
}

func (b *OptionsBuilder) WithTailLines(tailLines int64) *OptionsBuilder {
	b.options.TailLines = &tailLines
	return b
}

func (b *OptionsBuilder) WithTimestamps(timestamps bool) *OptionsBuilder {
	b.options.Timestamps = timestamps
	return b
}

func (b *OptionsBuilder) WithFollow(follow bool) *OptionsBuilder {
	b.options.Follow = follow
	return b
}

func (b *OptionsBuilder) WithPrevious(previous bool) *OptionsBuilder {
	b.options.Previous = previous
	return b
}

func (b *OptionsBuilder) WithLimitBytes(limitBytes int64) *OptionsBuilder {
	b.options.LimitBytes = limitBytes
	return b
}

func (b *OptionsBuilder) WithSinceSeconds(sinceSeconds int64) *OptionsBuilder {
	b.options.SinceSeconds = time.Duration(sinceSeconds) * time.Second
	return b
}

func (b *OptionsBuilder) WithSinceTime(sinceTime string) *OptionsBuilder {
	b.options.SinceTime = sinceTime
	return b
}

func (b *OptionsBuilder) Build() (Options, error) {
	if err := b.options.Validate(); err != nil {
		return Options{}, err
	}

	return b.options, nil
}

func (b *OptionsBuilder) FromOptions(options Options) *OptionsBuilder {
	b.options = options
	return b
}

func FromGetRequest(r *http.Request) (Options, error) {
	builder := NewOptionsBuilder()
	intBase := 10

	pod := r.URL.Query().Get("pod")
	if pod != "" {
		builder.WithPod(pod)
	}

	container := r.URL.Query().Get("container")
	if container != "" {
		builder.WithContainer(container)
	}

	tailStr := r.URL.Query().Get("tailLines")
	if tailStr != "" {
		tailLines, err := parseInt64(r.URL.Query().Get("tailLines"), intBase)
		if err != nil {
			return Options{}, fmt.Errorf("invalid tailLines value: %v", err)
		}

		builder.WithTailLines(tailLines)
	}

	timestampStr := r.URL.Query().Get("timestamps")
	if timestampStr != "" {
		timestamps, err := strconv.ParseBool(timestampStr)
		if err != nil {
			return Options{}, fmt.Errorf("invalid timestamps value: %v", err)
		}

		builder.WithTimestamps(timestamps)
	}

	followStr := r.URL.Query().Get("follow")
	if followStr != "" {
		follow, err := strconv.ParseBool(followStr)
		if err != nil {
			return Options{}, fmt.Errorf("invalid follow value: %v", err)
		}

		builder.WithFollow(follow)
	}

	previousStr := r.URL.Query().Get("previous")
	if previousStr != "" {
		previous, err := strconv.ParseBool(previousStr)
		if err != nil {
			return Options{}, fmt.Errorf("invalid previous value: %v", err)
		}

		builder.WithPrevious(previous)
	}

	limitBytesStr := r.URL.Query().Get("limitBytes")
	if limitBytesStr != "" {
		limitBytes, err := parseInt64(limitBytesStr, intBase)
		if err != nil {
			return Options{}, fmt.Errorf("invalid limitBytes value: %v", err)
		}

		builder.WithLimitBytes(limitBytes)
	}

	sinceSecondsStr := r.URL.Query().Get("sinceSeconds")
	if sinceSecondsStr != "" {
		sinceSeconds, err := parseInt64(sinceSecondsStr, intBase)
		if err != nil {
			return Options{}, fmt.Errorf("invalid sinceSeconds value: %v", err)
		}

		builder.WithSinceSeconds(sinceSeconds)
	}

	sinceTime := r.URL.Query().Get("sinceTime")
	if sinceTime != "" {
		builder.WithSinceTime(sinceTime)
	}

	return builder.Build()
}

func parseInt64(value string, base int) (int64, error) {
	if value == "" {
		return 0, nil
	}

	return strconv.ParseInt(value, base, 64)
}
