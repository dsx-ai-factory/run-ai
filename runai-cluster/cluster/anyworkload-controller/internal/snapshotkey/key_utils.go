/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotkey

import (
	"reflect"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// Snapshot's own additions to a restore pod, named exactly as podcontract
// writes them (api/podcontract: restore_pod.go, protocol.go). They must not
// reach the hash: Matches runs on a Ready replica, which on the restore path is
// a pod the webhook already rewrote. Leave them in and every restored replica
// fails the drift check, which turns snapshots off for the workload with
// nothing reported as an error.
//
// Matched exactly rather than by prefix, so a workload's own "snapshot-" volume
// or "SNAPSHOT_" variable still counts towards its shape.
var (
	snapshotVolumeNames = map[string]bool{
		"snapshot-control": true, // podcontract.SnapshotControlVolumeName
	}

	snapshotMountPaths = map[string]bool{
		"/snapshot-control":  true, // podcontract.SnapshotControlMountPath
		"/tmp/snapshot-cuda": true, // podcontract.CuInterposeMountPath
	}

	// The DYN_ spellings are podcontract's legacy names, written alongside the
	// current ones. They do not share a prefix with them, so a prefix match
	// would silently leave them in the hash.
	snapshotEnvNames = map[string]bool{
		"SNAPSHOT_CONTROL_DIR":                true, // podcontract.SnapshotControlDirEnv
		"DYN_SNAPSHOT_CONTROL_DIR":            true, // podcontract.LegacySnapshotControlDirEnv
		"SNAPSHOT_RESTORE_STANDBY":            true, // podcontract.RestoreStandbyModeEnv
		"DYN_SNAPSHOT_RESTORE_STANDBY":        true, // podcontract.LegacyRestoreStandbyModeEnv
		"SNAPSHOT_JOB_UID":                    true, // podcontract.SnapshotJobUIDEnv
		"SNAPSHOT_NON_TARGET_ARTIFACT_SUBDIR": true, // podcontract.NonTargetArtifactSubdirEnv
	}
)

const serviceAccountTokenPrefix = "kube-api-access-"

func containerNamed(containers []corev1.Container, name string) *corev1.Container {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}

func isSidecar(container corev1.Container) bool {
	return container.RestartPolicy != nil && *container.RestartPolicy == corev1.ContainerRestartPolicyAlways
}

// stripContainer blanks the per-container fields that are not part of the
// pod's shape (key calculation), plus Snapshot's own additions to the target container.
func stripContainer(container *corev1.Container) {
	container.LivenessProbe = nil
	container.ReadinessProbe = nil
	container.StartupProbe = nil
	container.Lifecycle = nil

	container.Env = withoutSnapshotEnv(container.Env)
	container.VolumeMounts = keptMounts(container.VolumeMounts)

	// podcontract sets a localhost seccomp profile on the restored container.
	if container.SecurityContext != nil {
		container.SecurityContext.SeccompProfile = nil
		if reflect.DeepEqual(*container.SecurityContext, corev1.SecurityContext{}) {
			// An emptied struct still serialises as {}, which a cold pod's
			// absent one does not.
			container.SecurityContext = nil
		}
	}
}

// stripScheduling removes where the pod runs. Scheduling is not part of the
// pod's shape, and the restore webhook adds node affinity of its own.
func stripScheduling(spec *corev1.PodSpec) {
	spec.NodeName = ""
	spec.NodeSelector = nil
	spec.Affinity = nil
	spec.Tolerations = nil
	spec.SchedulerName = ""
	spec.PriorityClassName = ""
	spec.Priority = nil
	spec.TopologySpreadConstraints = nil
	// Gates that hold a pod back from scheduling; webhooks add and remove them.
	spec.SchedulingGates = nil
}

// stripDebugContainers removes ephemeral containers. kubectl debug injects one
// into a running pod, so leaving them in would let debugging a replica change
// its key, which the capture plugin reads as drift and responds to by turning
// snapshots off for the whole workload.
func stripDebugContainers(spec *corev1.PodSpec) {
	spec.EphemeralContainers = nil
}

// stripSnapshotAdditions removes the pod-level fields podcontract adds.
func stripSnapshotAdditions(spec *corev1.PodSpec) {
	if spec.SecurityContext != nil {
		spec.SecurityContext.SeccompProfile = nil
		if reflect.DeepEqual(*spec.SecurityContext, corev1.PodSecurityContext{}) {
			spec.SecurityContext = nil
		}
	}
}

// mountedVolumes keeps only the volumes some hashed container mounts.
//
// This subsumes the excluded-name filter: keptMounts has already dropped the
// service-account token and Snapshot's own mounts, so those volumes are
// unreachable here and fall out. It also keeps the two lists symmetric - a
// volume is hashed exactly when a mount for it is.
func mountedVolumes(spec *corev1.PodSpec) []corev1.Volume {
	mounted := map[string]bool{}
	for _, containers := range [][]corev1.Container{spec.InitContainers, spec.Containers} {
		for _, container := range containers {
			for _, mount := range container.VolumeMounts {
				mounted[mount.Name] = true
			}
			// A raw block volume is consumed through VolumeDevices and has no
			// VolumeMount, so it would otherwise be dropped and a change to the
			// backing claim would not change the key.
			for _, device := range container.VolumeDevices {
				mounted[device.Name] = true
			}
		}
	}

	var kept []corev1.Volume
	for _, volume := range spec.Volumes {
		if !mounted[volume.Name] {
			continue
		}
		kept = append(kept, volume)
	}
	return kept
}

func keptMounts(mounts []corev1.VolumeMount) []corev1.VolumeMount {
	var kept []corev1.VolumeMount
	for _, mount := range mounts {
		if isExcludedVolumeName(mount.Name) || snapshotMountPaths[mount.MountPath] {
			continue
		}
		kept = append(kept, mount)
	}
	return kept
}

func isExcludedVolumeName(name string) bool {
	// The token volume keeps its generated suffix, so it is the one name still
	// matched by prefix.
	return strings.HasPrefix(name, serviceAccountTokenPrefix) || snapshotVolumeNames[name]
}

// expandsStrippedEnv reports whether a retained variable expands a
// Snapshot-owned one, and names the first it finds.
//
// Removing a variable loses its position, and Kubernetes expands $(VAR) only
// against entries declared earlier. So [SNAPSHOT_CONTROL_DIR, X=$(SNAPSHOT_CONTROL_DIR)]
// and [X=$(SNAPSHOT_CONTROL_DIR), SNAPSHOT_CONTROL_DIR] strip to the same bytes
// while resolving X differently - one expanded, one literal. Keying them the
// same would let a replica restore from a checkpoint built with a different
// environment, so such a pod is refused instead.
func expandsStrippedEnv(container corev1.Container) (string, bool) {
	// Snapshot's own variables are gone by now - Shape strips them before
	// calling this - so every variable here is one the hash keeps.
	for _, variable := range container.Env {
		for name := range snapshotEnvNames {
			if strings.Contains(variable.Value, "$("+name+")") {
				return name, true
			}
		}
	}
	return "", false
}

func withoutSnapshotEnv(env []corev1.EnvVar) []corev1.EnvVar {
	var kept []corev1.EnvVar
	for _, variable := range env {
		if snapshotEnvNames[variable.Name] {
			continue
		}
		kept = append(kept, variable)
	}
	return kept
}

// sortSpec orders what Kubernetes does not, so that map iteration or a
// controller reordering a slice cannot change the key.
func sortSpec(spec *corev1.PodSpec) {
	sort.Slice(spec.Volumes, func(i, j int) bool { return spec.Volumes[i].Name < spec.Volumes[j].Name })

	for i := range spec.Containers {
		sortContainer(&spec.Containers[i])
	}
	for i := range spec.InitContainers {
		sortContainer(&spec.InitContainers[i])
	}
}

// Env and EnvFrom are deliberately left in order. Kubernetes expands $(VAR) in
// env only against entries that appear earlier, and applies envFrom sources in
// order with the last duplicate key winning - so two orderings of the same
// entries are two different effective environments. Sorting them would give
// both the same key, and a replica could restore from a checkpoint built with
// different configuration.
//
// This narrows design 8.3's "env, mounts, and volumes sorted": mounts and
// volumes carry no meaning in their order, and env does.
func sortContainer(container *corev1.Container) {
	sort.Slice(container.VolumeMounts, func(i, j int) bool {
		if container.VolumeMounts[i].MountPath != container.VolumeMounts[j].MountPath {
			return container.VolumeMounts[i].MountPath < container.VolumeMounts[j].MountPath
		}
		return container.VolumeMounts[i].Name < container.VolumeMounts[j].Name
	})
}

func containerHasGPU(resources corev1.ResourceRequirements) bool {
	for name := range resources.Requests {
		if strings.Contains(string(name), nvidiaGPUResourceName) {
			return true
		}
	}
	for name := range resources.Limits {
		if strings.Contains(string(name), nvidiaGPUResourceName) {
			return true
		}
	}
	return false
}
