/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package snapshotkey is the one implementation of the NVIDIA Snapshot key.
//
// The restore webhook computes the key when a pod is created; the capture
// plugin recomputes it on the Ready replica to check it still matches. If the
// two ever disagreed by a single field that check would fail permanently and
// snapshots would switch off for the workload, with nothing logged as an
// error. Hence one package, imported by both.
//
// # What the key covers
//
// The key is a hash of the replica's pod as admitted, so each version of the
// replica gets one snapshot.
//
// Hashed: the target container, run-to-completion init containers, their
// volumes, and the rest of the pod spec - security context, hostname, runtime
// class.
//
// Not hashed: metadata, sidecars, the service-account token volume, probes,
// lifecycle hooks, scheduling, and Snapshot's own additions.
//
// The list is an exclude-list, not an allow-list: a field this package does
// not name stays in the hash. A new Kubernetes field therefore enters the key
// automatically, which is the safe default - a field nobody considered changes
// the key rather than being silently ignored.
package snapshotkey

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Version is folded into the hash. Bump it whenever the shape
// changes, so snapshots taken under an older definition are never reused
// against a newer one.
const Version = "v1"

// nvidiaGPUResourceName matches dromos' karta accessor, whose HasGPU decides
// eligibility. The two must agree: a pod that is eligible but unkeyable opts
// into a feature that then never runs.
const nvidiaGPUResourceName = "nvidia.com/gpu"

// keyLength is how much of the digest the key keeps. 12 hex characters leave
// room for the name prefix within a 63-character label value.
const keyLength = 12

// ErrNotKeyable means the pod cannot be keyed: it has no controller owner to
// scope the key to, or no single target container. Callers treat it as "cold
// start", never as an error worth failing admission over.
var ErrNotKeyable = errors.New("pod is not keyable")

// OptedIn reports whether the pod has OptInLabel set to "true"
func OptedIn(pod *corev1.Pod) bool {
	return pod.Labels[constants.SnapshotOptInLabel] == "true"
}

// TargetContainer returns the one container requesting nvidia.com/gpu
// ErrNotKeyable if there is not exactly one.
func TargetContainer(pod *corev1.Pod) (string, error) {
	var found []string
	for _, container := range pod.Spec.Containers {
		if containerHasGPU(container.Resources) {
			found = append(found, container.Name)
		}
	}

	if len(found) != 1 {
		return "", fmt.Errorf("%w: want exactly one container requesting %s, got %d",
			ErrNotKeyable, nvidiaGPUResourceName, len(found))
	}
	return found[0], nil
}

// Scope returns the identity the key is scoped to: the UID of the pod's
// controller owner.
// ErrNotKeyable for a pod with no controller owner, whose key would be unique
// per replica and so never reusable.
func Scope(pod *corev1.Pod) (types.UID, error) {
	owner := metav1.GetControllerOf(pod)
	if owner == nil {
		return "", fmt.Errorf("%w: pod has no controller owner to scope the key to", ErrNotKeyable)
	}
	return owner.UID, nil
}

// Shape returns the exact bytes ComputeKey hashes: the pod reduced to the
// fields that identify it, as the package comment lists them.
func Shape(pod *corev1.Pod) ([]byte, error) {
	target, err := TargetContainer(pod)
	if err != nil {
		return nil, err
	}

	specForHashing := pod.Spec.DeepCopy()

	targetContainer := containerNamed(specForHashing.Containers, target)
	if targetContainer == nil {
		return nil, fmt.Errorf("%w: container %q is not in the pod", ErrNotKeyable, target)
	}

	// Only the target container. Sidecars are excluded, and so are any other
	// regular containers, such as Knative's queue-proxy.
	stripContainer(targetContainer)
	specForHashing.Containers = []corev1.Container{*targetContainer}

	// Init containers that run to completion are part of the pod's shape;
	// those with restartPolicy Always are sidecars and are not.
	var initContainers []corev1.Container
	for _, container := range specForHashing.InitContainers {
		if isSidecar(container) {
			continue
		}
		stripContainer(&container)
		initContainers = append(initContainers, container)
	}
	specForHashing.InitContainers = initContainers

	// "Their volumes" (design 8.3): a volume no hashed container mounts belongs
	// to a sidecar we dropped. Hashing it would let an unrelated sidecar
	// upgrade - a new istio or queue-proxy volume - change every opted-in
	// pod's key and invalidate every stored checkpoint.
	specForHashing.Volumes = mountedVolumes(specForHashing)

	stripDebugContainers(specForHashing)
	stripScheduling(specForHashing)
	stripSnapshotAdditions(specForHashing)
	sortSpec(specForHashing)

	return json.Marshal(specForHashing)
}

// ComputeKey returns the snapshot key for a pod as admitted.
func ComputeKey(pod *corev1.Pod) (string, error) {
	scope, err := Scope(pod)
	if err != nil {
		return "", err
	}

	// The target container's name is already inside the shape bytes, so it is
	// not mixed in separately.
	shape, err := Shape(pod)
	if err != nil {
		return "", err
	}

	// Length-prefixed so no two different inputs can produce the same stream,
	// which plain concatenation would allow.
	digest := sha256.New()
	for _, part := range [][]byte{[]byte(Version), []byte(scope), shape} {
		fmt.Fprintf(digest, "%d:", len(part))
		digest.Write(part)
	}

	return hex.EncodeToString(digest.Sum(nil))[:keyLength], nil
}

// KeyOf returns the key the webhook stamped, and whether it was present.
//
// Absence is meaningful: an opted-in pod with no key means the webhook did not
// run, not that there was no snapshot.
func KeyOf(pod *corev1.Pod) (string, bool) {
	key, ok := pod.Labels[constants.SnapshotKeyLabel]
	return key, ok
}

// SetKey stamps KeyLabel. The webhook is the only writer: the stored key is
// the record of the pod's shape at admission, and a key written later would
// make the drift check tautological.
func SetKey(pod *corev1.Pod, key string) {
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	pod.Labels[constants.SnapshotKeyLabel] = key
}

// Matches reports whether the pod still hashes to the key it carries - the
// drift check, for a replica changed by a webhook that ran after ours.
//
// A false result disables snapshots for the workload, so call it only where
// that is the intended response.
func Matches(pod *corev1.Pod) (bool, error) {
	stamped, ok := KeyOf(pod)
	if !ok {
		return false, fmt.Errorf("%w: pod carries no %s label", ErrNotKeyable, constants.SnapshotKeyLabel)
	}

	computed, err := ComputeKey(pod)
	if err != nil {
		return false, err
	}
	return computed == stamped, nil
}

// SnapshotName returns the SnapshotJob and PodSnapshot name for a key.
//
// Derived from the key alone, so the webhook needs no workload identity to
// look an object up. The workload's name travels as a label on the objects
// instead. At most 63 characters, since it is used as a label value.
func SnapshotName(key string) string {
	return "runai-snap-" + key
}
