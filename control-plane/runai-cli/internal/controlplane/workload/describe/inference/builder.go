/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// Builder builds inference Fields
type Builder struct {
	fields Fields
}

// NewBuilder creates a new inference Fields builder
func NewBuilder() *Builder {
	return &Builder{
		fields: Fields{},
	}
}

func (b *Builder) LatestCreatedRevisionName(lcr string) *Builder {
	b.fields.LatestCreatedRevisionName = &lcr
	return b
}

func (b *Builder) LatestReadyRevisionName(lrr string) *Builder {
	b.fields.LatestReadyRevisionName = &lrr
	return b
}

func (b *Builder) MaxReplicas(val *int32) *Builder {
	b.fields.MaxReplicas = val
	return b
}

func (b *Builder) MinReplicas(val *int32) *Builder {
	b.fields.MinReplicas = val
	return b
}

func (b *Builder) Autoscaling(autoscaling openapi.AutoScaling) *Builder {
	if autoscaling.InitialReplicas.IsSet() {
		b.fields.InitialReplicas = autoscaling.InitialReplicas.Get()
	}
	if autoscaling.ActivationReplicas.IsSet() {
		b.fields.ActivationReplicas = autoscaling.ActivationReplicas.Get()
	}
	if autoscaling.Metric.IsSet() {
		b.fields.Metric = autoscaling.Metric.Get()
	}
	if autoscaling.MetricThreshold.IsSet() {
		b.fields.MetricThreshold = autoscaling.MetricThreshold.Get()
	}
	if autoscaling.MetricThresholdPercentage.IsSet() {
		b.fields.MetricThresholdPercentage = autoscaling.MetricThresholdPercentage.Get()
	}
	if autoscaling.ScaleDownDelaySeconds.IsSet() {
		b.fields.ScaleDownDelaySeconds = autoscaling.ScaleDownDelaySeconds.Get()
	}
	if autoscaling.ScaleToZeroRetentionSeconds.IsSet() {
		b.fields.ScaleToZeroRetentionSeconds = autoscaling.ScaleToZeroRetentionSeconds.Get()
	}
	if autoscaling.ConcurrencyHardLimit.IsSet() {
		b.fields.ConcurrencyHardLimit = autoscaling.ConcurrencyHardLimit.Get()
	}
	if autoscaling.MinReplicas.IsSet() {
		b.fields.MinReplicas = autoscaling.MinReplicas.Get()
	}
	if autoscaling.MaxReplicas.IsSet() {
		b.fields.MaxReplicas = autoscaling.MaxReplicas.Get()
	}
	return b
}

func (b *Builder) ServingConfiguration(servingConfiguration openapi.ServingConfiguration) *Builder {
	if servingConfiguration.InitializationTimeoutSeconds.IsSet() {
		b.fields.InitializationTimeoutSeconds = servingConfiguration.InitializationTimeoutSeconds.Get()
	}
	if servingConfiguration.RequestTimeoutSeconds.IsSet() {
		b.fields.RequestTimeoutSeconds = servingConfiguration.RequestTimeoutSeconds.Get()
	}
	return b
}

// Build returns the built Fields
func (b *Builder) Build() *Fields {
	return &b.fields
}

// BuildFromWorkload creates inference Fields from workload data
func BuildFromWorkload(
	additionalFields map[string]interface{},
	requestedPods *openapi.RequestedPods,
	autoScaling *openapi.AutoScaling,
	servingConfiguration *openapi.ServingConfiguration,
) *Fields {
	builder := NewBuilder()

	// Set revision names from additional fields
	if value, ok := getStringFromAdditionalFields(additionalFields, "latestCreatedRevisionName"); ok {
		builder.LatestCreatedRevisionName(value)
	}
	if value, ok := getStringFromAdditionalFields(additionalFields, "latestReadyRevisionName"); ok {
		builder.LatestReadyRevisionName(value)
	}

	// Set replica counts
	if requestedPods != nil {
		if requestedPods.Max.IsSet() {
			builder.MaxReplicas(requestedPods.Max.Get())
		}
		if requestedPods.Min.IsSet() {
			builder.MinReplicas(requestedPods.Min.Get())
		}
	}

	// Set autoscaling and serving configuration
	if autoScaling != nil {
		builder.Autoscaling(*autoScaling)
	}
	if servingConfiguration != nil {
		builder.ServingConfiguration(*servingConfiguration)
	}

	return builder.Build()
}

func getStringFromAdditionalFields(fields map[string]interface{}, key string) (string, bool) {
	if fields == nil {
		return "", false
	}
	value, ok := fields[key]
	if !ok {
		return "", false
	}
	strValue, ok := value.(string)
	if !ok || strValue == "" {
		return "", false
	}
	return strValue, true
}

// SetFields sets the inference fields on the target pointer
func SetFields(target **Fields, fields *Fields) {
	*target = fields
}

// BuildAndSetFields builds inference fields from workload data and sets them on the target
func BuildAndSetFields(
	target **Fields,
	additionalFields map[string]interface{},
	requestedPods *openapi.RequestedPods,
	autoScaling *openapi.AutoScaling,
	servingConfiguration *openapi.ServingConfiguration,
) {
	fields := BuildFromWorkload(additionalFields, requestedPods, autoScaling, servingConfiguration)
	SetFields(target, fields)
}
