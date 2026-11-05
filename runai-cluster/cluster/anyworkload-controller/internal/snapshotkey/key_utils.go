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

// Snapshot's own additions to a restore pod, which podcontract.Build writes.
// They must not reach the hash: Matches runs on a Ready replica, which on the
// restore path is a pod the webhook already rewrote. Leave them in and every
// restored replica fails the drift check, which turns snapshots off for the
// workload with nothing reported as an error.
const (
	snapshotVolumePrefix    = "snapshot-"
	snapshotEnvPrefix       = "SNAPSHOT_"
	snapshotMountPathPrefix = "/var/run/snapshot"
)

// serviceAccountTokenPrefix is the generated name of the projected token
// volume. It is excluded because it is remounted per pod and would otherwise
// change the key on every replica.
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
		if isExcludedVolumeName(mount.Name) || strings.HasPrefix(mount.MountPath, snapshotMountPathPrefix) {
			continue
		}
		kept = append(kept, mount)
	}
	return kept
}

func isExcludedVolumeName(name string) bool {
	return strings.HasPrefix(name, serviceAccountTokenPrefix) ||
		strings.HasPrefix(name, snapshotVolumePrefix)
}

func withoutSnapshotEnv(env []corev1.EnvVar) []corev1.EnvVar {
	var kept []corev1.EnvVar
	for _, variable := range env {
		if strings.HasPrefix(variable.Name, snapshotEnvPrefix) {
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

func sortContainer(container *corev1.Container) {
	sort.Slice(container.Env, func(i, j int) bool { return container.Env[i].Name < container.Env[j].Name })
	sort.Slice(container.VolumeMounts, func(i, j int) bool {
		if container.VolumeMounts[i].MountPath != container.VolumeMounts[j].MountPath {
			return container.VolumeMounts[i].MountPath < container.VolumeMounts[j].MountPath
		}
		return container.VolumeMounts[i].Name < container.VolumeMounts[j].Name
	})
	sort.Slice(container.EnvFrom, func(i, j int) bool {
		return envFromKey(container.EnvFrom[i]) < envFromKey(container.EnvFrom[j])
	})
}

func envFromKey(source corev1.EnvFromSource) string {
	switch {
	case source.ConfigMapRef != nil:
		return "c/" + source.ConfigMapRef.Name
	case source.SecretRef != nil:
		return "s/" + source.SecretRef.Name
	default:
		return source.Prefix
	}
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
