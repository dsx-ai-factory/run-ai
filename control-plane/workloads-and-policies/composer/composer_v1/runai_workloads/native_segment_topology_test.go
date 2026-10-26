/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_workloads

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

func TestDistributedComposerPreservesSegmentAnnotationsForNativeFlow(t *testing.T) {
	const (
		topologyName     = "kai.scheduler/topology"
		segmentSize      = "kai.scheduler/segment-size"
		segmentPreferred = "kai.scheduler/segment-topology-preferred-placement"
	)

	framework := workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH
	numWorkers := int32(4)
	image := "ubuntu"
	spec := &workloadsApi.SupersetSpec{
		DistributedFramework: *workloadsApi.NewNullableDistributedFramework(&framework),
		NumWorkers:           *workloadsApi.NewNullableInt32(&numWorkers),
		Image:                *workloadsApi.NewNullableString(&image),
		Annotations: []workloadsApi.Annotation{
			segmentAnnotation(topologyName, "network-topology"),
			segmentAnnotation(segmentSize, "2"),
			segmentAnnotation(segmentPreferred, "topology.kubernetes.io/rack"),
		},
	}
	workload := &workloadsApi.SupersetWorkload{Spec: spec}
	target := &jobs.JobWorkload{}
	dinfComposer := NewDistributedWorkloadComposer("segmented", "user", nil, composer.FieldsComposerOptions{})

	if err := dinfComposer.Compose(context.Background(), workload, target); err != nil {
		t.Fatalf("unexpected native distributed composition error: %v", err)
	}

	annotations := target.Spec.Annotations.Deref()
	for key, expected := range map[string]string{
		topologyName:     "network-topology",
		segmentSize:      "2",
		segmentPreferred: "topology.kubernetes.io/rack",
	} {
		item, found := annotations[key]
		if !found || item.Value == nil || *item.Value != expected {
			t.Fatalf("expected annotation %q=%q, got %+v", key, expected, item)
		}
	}
}

func segmentAnnotation(name, value string) workloadsApi.Annotation {
	annotation := workloadsApi.NewAnnotation()
	annotation.SetName(name)
	annotation.SetValue(value)
	return *annotation
}
