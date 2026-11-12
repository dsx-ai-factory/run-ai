/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotrestore

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ai-dynamo/snapshot/api/podcontract"
	snapshotv1alpha1 "github.com/ai-dynamo/snapshot/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/internal/snapshotkey"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Reasons a pod cold-starts, never reported as failures.
var (
	ErrUnplaceable       = errors.New("the workload's placement excludes every node the agent runs on")
	errNoReader          = errors.New("no API reader to look up the snapshot")
	errSnapshotNotUsable = errors.New("snapshot is not ready, or is marked failed")
	errNoAgentNodes      = errors.New("no nodes are known to run Snapshot's agent")
	errOwnStartupProbe   = errors.New("the target container defines its own startup probe")
)

// Mutator stamps the snapshot key on every opted-in pod, and turns the pod into
// a restore pod when a ready snapshot exists for that key.
type Mutator struct {
	// Uncached: the grant is get on podsnapshots, and an informer needs list
	// and watch, so a cached read would stall admission.
	apiReader client.Reader

	// Where Snapshot's agent runs; only those nodes can take a restore pod.
	agentNodeSelector map[string]string
}

// NewMutator returns the restore webhook handler.
func NewMutator(apiReader client.Reader, agentNodeSelector map[string]string) *Mutator {
	return &Mutator{apiReader: apiReader, agentNodeSelector: agentNodeSelector}
}

// Handle stamps the key and restores from a ready snapshot, or admits the pod
// unchanged. Every failure admits: not restoring costs a cold start, rejecting
// stops the workload running. Opt-in is not rechecked - the objectSelector
// already matched.
func (m *Mutator) Handle(ctx context.Context, request admission.Request) admission.Response {
	if request.Operation != admissionv1.Create {
		// Only CREATE is registered; anything else means we have drifted apart.
		return admission.Allowed("unexpected operation, admitted unchanged")
	}

	pod := &corev1.Pod{}
	if err := json.Unmarshal(request.Object.Raw, pod); err != nil {
		return admission.Allowed("pod could not be read, admitted unchanged")
	}

	// Not keyable is ordinary, not a failure: the replica cold-starts.
	key, err := snapshotkey.ComputeKey(pod)
	if err != nil {
		return admission.Allowed("pod is not keyable, admitted unchanged")
	}

	snapshotkey.SetSnapshotKey(pod, key)

	restored, err := m.restore(ctx, pod, key)
	if err != nil {
		// Still worth stamping: the key is what gets this snapshot captured.
		restored = pod
	}

	mutated, marshalErr := json.Marshal(restored)
	if marshalErr != nil {
		return admission.Allowed("pod could not be written, admitted unchanged")
	}
	return admission.PatchResponseFromRaw(request.Object.Raw, mutated)
}

// restore rewrites the pod to start from its snapshot. The error says why it
// cannot, and is never a failure to report.
func (m *Mutator) restore(ctx context.Context, pod *corev1.Pod, key string) (*corev1.Pod, error) {
	if m.apiReader == nil {
		return nil, errNoReader
	}

	if len(m.agentNodeSelector) == 0 {
		return nil, errNoAgentNodes
	}

	snapshot := &snapshotv1alpha1.PodSnapshot{}
	name := types.NamespacedName{Name: snapshotkey.SnapshotName(key), Namespace: pod.Namespace}
	if err := m.apiReader.Get(ctx, name, snapshot); err != nil {
		// Not found is ordinary: this key has no snapshot yet.
		return nil, err
	}

	// Failed means Snapshot marked the checkpoint itself at fault.
	if !snapshotv1alpha1.IsPodSnapshotSucceeded(snapshot) || snapshotv1alpha1.IsPodSnapshotFailed(snapshot) {
		return nil, errSnapshotNotUsable
	}

	target, err := snapshotkey.TargetContainer(pod)
	if err != nil {
		return nil, err
	}

	// podcontract replaces the target's startup probe with the restore gate, so
	// a workload that defines one would lose it. Same rule as its own seccomp
	// profile: cold-start instead.
	if hasOwnStartupProbe(pod, target) {
		return nil, errOwnStartupProbe
	}

	restored, err := podcontract.Build(pod,
		podcontract.Request{SnapshotName: snapshot.Name, SourceContainer: target},
		podcontract.Options{
			SeccompProfile: podcontract.DefaultSeccompLocalhostProfile,
			// Withhold readiness until the restore completes.
			EnableStartupGate: true,
		})
	if err != nil {
		return nil, err
	}

	if err := RequireNodeSelector(&restored.Spec, m.agentNodeSelector); err != nil {
		// Admitting an unplaceable pod would strand the replica in Pending;
		// cold-starting runs it where the workload asked.
		return nil, err
	}

	return restored, nil
}

// hasOwnStartupProbe reports whether the target brings a probe the restore
// gate would replace.
func hasOwnStartupProbe(pod *corev1.Pod, target string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == target {
			return container.StartupProbe != nil
		}
	}
	return false
}
