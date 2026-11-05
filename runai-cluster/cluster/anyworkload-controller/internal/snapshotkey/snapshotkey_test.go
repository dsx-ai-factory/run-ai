/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotkey

import (
	"errors"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
)

const ownerUID = "0c3f6b9a-1d2e-4f50-8a7b-6c5d4e3f2a1b"

// gpuPod is a minimal keyable pod: one controller owner, one GPU container.
func gpuPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "llama-00001-deployment-7d4f8b6c9d-abcde",
			Namespace: "runai-team-a",
			Labels:    map[string]string{constants.SnapshotOptInLabel: "true"},
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "apps/v1",
				Kind:       "ReplicaSet",
				Name:       "llama-00001-deployment-7d4f8b6c9d",
				UID:        ownerUID,
				Controller: ptr.To(true),
			}},
		},
		Spec: corev1.PodSpec{
			Hostname: "llama",
			Containers: []corev1.Container{{
				Name:    "vllm",
				Image:   "registry.example.com/llama@sha256:abc",
				Command: []string{"vllm", "serve"},
				Env: []corev1.EnvVar{
					{Name: "MODEL", Value: "llama-3"},
					{Name: "PORT", Value: "8000"},
				},
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"nvidia.com/gpu": resource.MustParse("1"),
					},
				},
			}},
		},
	}
}

func keyOrFail(t *testing.T, pod *corev1.Pod) string {
	t.Helper()
	key, err := ComputeKey(pod)
	if err != nil {
		t.Fatalf("ComputeKey: %v", err)
	}
	return key
}

func TestOptedInRequiresTheValueTrue(t *testing.T) {
	// The webhook's objectSelector matches on the value, so a presence-only
	// check here would capture workloads the webhook never restores into.
	cases := map[string]bool{"true": true, "false": false, "": false, "TRUE": false}
	for value, want := range cases {
		pod := gpuPod()
		pod.Labels[constants.SnapshotOptInLabel] = value
		if got := OptedIn(pod); got != want {
			t.Errorf("OptedIn(%q) = %v, want %v", value, got, want)
		}
	}

	pod := gpuPod()
	pod.Labels = nil
	if OptedIn(pod) {
		t.Error("OptedIn(no labels) = true, want false")
	}
}

func TestReplicasOfOneVersionShareAKey(t *testing.T) {
	first, second := gpuPod(), gpuPod()
	second.Name = "llama-00001-deployment-7d4f8b6c9d-zyxwv"
	second.Spec.NodeName = "gpu-node-7"

	if keyOrFail(t, first) != keyOrFail(t, second) {
		t.Error("replicas of one version produced different keys")
	}
}

func TestAnEditChangesTheKey(t *testing.T) {
	base := keyOrFail(t, gpuPod())

	edits := map[string]func(*corev1.Pod){
		"image":     func(p *corev1.Pod) { p.Spec.Containers[0].Image = "registry.example.com/llama@sha256:def" },
		"command":   func(p *corev1.Pod) { p.Spec.Containers[0].Command = []string{"vllm", "serve", "--tp", "2"} },
		"env value": func(p *corev1.Pod) { p.Spec.Containers[0].Env[0].Value = "llama-4" },
		"new env": func(p *corev1.Pod) {
			p.Spec.Containers[0].Env = append(p.Spec.Containers[0].Env, corev1.EnvVar{Name: "A", Value: "b"})
		},
		"hostname":      func(p *corev1.Pod) { p.Spec.Hostname = "other" },
		"runtime class": func(p *corev1.Pod) { p.Spec.RuntimeClassName = ptr.To("nvidia") },
		"gpu count":     func(p *corev1.Pod) { p.Spec.Containers[0].Resources.Limits["nvidia.com/gpu"] = resource.MustParse("2") },
		"different owner": func(p *corev1.Pod) {
			p.OwnerReferences[0].UID = "99999999-1d2e-4f50-8a7b-6c5d4e3f2a1b"
		},
	}

	for name, edit := range edits {
		pod := gpuPod()
		edit(pod)
		if keyOrFail(t, pod) == base {
			t.Errorf("%s did not change the key", name)
		}
	}
}

func TestReorderingDoesNotChangeTheKey(t *testing.T) {
	base := keyOrFail(t, gpuPod())

	reorders := map[string]func(*corev1.Pod){
		"volumes and mounts": func(p *corev1.Pod) {
			p.Spec.Volumes = []corev1.Volume{{Name: "b"}, {Name: "a"}}
			p.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
				{Name: "b", MountPath: "/b"}, {Name: "a", MountPath: "/a"},
			}
		},
	}

	ordered := gpuPod()
	ordered.Spec.Volumes = []corev1.Volume{{Name: "a"}, {Name: "b"}}
	ordered.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
		{Name: "a", MountPath: "/a"}, {Name: "b", MountPath: "/b"},
	}
	withVolumes := keyOrFail(t, ordered)

	for name, reorder := range reorders {
		pod := gpuPod()
		reorder(pod)

		want := base
		if name == "volumes and mounts" {
			want = withVolumes
		}
		if keyOrFail(t, pod) != want {
			t.Errorf("reordering %s changed the key", name)
		}
	}
}

// Env order is meaning, not presentation: Kubernetes expands $(VAR) only
// against earlier entries, so [A=x, B=$(A)] and [B=$(A), A=x] are different
// effective environments. Sorting them would let a replica restore from a
// checkpoint built with different configuration.
func TestReorderingEnvChangesTheKey(t *testing.T) {
	base := keyOrFail(t, gpuPod())

	reversed := gpuPod()
	reversed.Spec.Containers[0].Env = []corev1.EnvVar{
		{Name: "PORT", Value: "8000"},
		{Name: "MODEL", Value: "llama-3"},
	}

	if keyOrFail(t, reversed) == base {
		t.Error("reordering env did not change the key")
	}
}

// The same for envFrom, where the last source wins a duplicated key.
func TestReorderingEnvFromChangesTheKey(t *testing.T) {
	withSources := func(names ...string) *corev1.Pod {
		pod := gpuPod()
		for _, name := range names {
			pod.Spec.Containers[0].EnvFrom = append(pod.Spec.Containers[0].EnvFrom,
				corev1.EnvFromSource{ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: name},
				}})
		}
		return pod
	}

	if keyOrFail(t, withSources("cm1", "cm2")) == keyOrFail(t, withSources("cm2", "cm1")) {
		t.Error("reordering envFrom did not change the key")
	}
}

func TestExcludedFieldsDoNotChangeTheKey(t *testing.T) {
	base := keyOrFail(t, gpuPod())

	excluded := map[string]func(*corev1.Pod){
		"pod name":     func(p *corev1.Pod) { p.Name = "llama-00001-deployment-7d4f8b6c9d-other" },
		"labels":       func(p *corev1.Pod) { p.Labels["unrelated"] = "x" },
		"annotations":  func(p *corev1.Pod) { p.Annotations = map[string]string{"a": "b"} },
		"node name":    func(p *corev1.Pod) { p.Spec.NodeName = "gpu-node-7" },
		"nodeSelector": func(p *corev1.Pod) { p.Spec.NodeSelector = map[string]string{"nvidia.com/gpu.present": "true"} },
		"affinity":     func(p *corev1.Pod) { p.Spec.Affinity = &corev1.Affinity{} },
		"tolerations":  func(p *corev1.Pod) { p.Spec.Tolerations = []corev1.Toleration{{Key: "nvidia.com/gpu"}} },
		"scheduler":    func(p *corev1.Pod) { p.Spec.SchedulerName = "runai-scheduler" },
		"priority":     func(p *corev1.Pod) { p.Spec.PriorityClassName = "low" },
		"probes": func(p *corev1.Pod) {
			p.Spec.Containers[0].LivenessProbe = &corev1.Probe{}
			p.Spec.Containers[0].ReadinessProbe = &corev1.Probe{}
			p.Spec.Containers[0].StartupProbe = &corev1.Probe{}
		},
		"lifecycle": func(p *corev1.Pod) { p.Spec.Containers[0].Lifecycle = &corev1.Lifecycle{} },
		"sidecar": func(p *corev1.Pod) {
			p.Spec.InitContainers = append(p.Spec.InitContainers, corev1.Container{
				Name:          "queue-proxy",
				Image:         "knative/queue",
				RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways),
			})
		},
		"extra regular container": func(p *corev1.Pod) {
			p.Spec.Containers = append(p.Spec.Containers, corev1.Container{Name: "sidecar", Image: "busybox"})
		},
		"service-account token volume": func(p *corev1.Pod) {
			p.Spec.Volumes = append(p.Spec.Volumes, corev1.Volume{Name: "kube-api-access-x7k2p"})
			p.Spec.Containers[0].VolumeMounts = append(p.Spec.Containers[0].VolumeMounts,
				corev1.VolumeMount{Name: "kube-api-access-x7k2p", MountPath: "/var/run/secrets/kubernetes.io/serviceaccount"})
		},
	}

	for name, mutate := range excluded {
		pod := gpuPod()
		mutate(pod)
		if keyOrFail(t, pod) != base {
			t.Errorf("%s changed the key but is excluded", name)
		}
	}
}

// Stripping a Snapshot-owned variable loses its position, and $(VAR) expands
// only against entries declared earlier. Two orderings would strip to the same
// bytes while resolving the reference differently - expanded in one, literal in
// the other - so such a pod is refused rather than keyed into a false match.
func TestEnvExpandingAStrippedVariableIsNotKeyable(t *testing.T) {
	ordered := func(first, second corev1.EnvVar) *corev1.Pod {
		pod := gpuPod()
		pod.Spec.Containers[0].Env = []corev1.EnvVar{first, second}
		return pod
	}
	control := corev1.EnvVar{Name: "SNAPSHOT_CONTROL_DIR", Value: "/snapshot-control"}
	reference := corev1.EnvVar{Name: "MY_PATH", Value: "$(SNAPSHOT_CONTROL_DIR)/x"}

	// Both orderings are refused: expanded in one, literal in the other.
	for name, pod := range map[string]*corev1.Pod{
		"reference after the variable":  ordered(control, reference),
		"reference before the variable": ordered(reference, control),
	} {
		if _, err := ComputeKey(pod); !errors.Is(err, ErrNotKeyable) {
			t.Errorf("%s: err = %v, want ErrNotKeyable", name, err)
		}
	}
}

// Command, args and subPathExpr are expanded against the finished environment,
// so stripping a Snapshot-owned variable hides whether they resolved at all:
// two pods would run different commands, or mount different paths, under one
// key.
func TestExpandedFieldsReferencingAStrippedVariableAreNotKeyable(t *testing.T) {
	cases := map[string]func(*corev1.Pod){
		"command": func(p *corev1.Pod) {
			p.Spec.Containers[0].Command = []string{"$(SNAPSHOT_CONTROL_DIR)/run"}
		},
		"args": func(p *corev1.Pod) {
			p.Spec.Containers[0].Args = []string{"--dir", "$(SNAPSHOT_CONTROL_DIR)"}
		},
		"subPathExpr": func(p *corev1.Pod) {
			p.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
				{Name: "models", MountPath: "/models", SubPathExpr: "$(SNAPSHOT_CONTROL_DIR)"},
			}
			p.Spec.Volumes = []corev1.Volume{{Name: "models"}}
		},
		"init container command": func(p *corev1.Pod) {
			p.Spec.InitContainers = []corev1.Container{
				{Name: "fetch", Image: "busybox", Command: []string{"$(SNAPSHOT_JOB_UID)"}},
			}
		},
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			pod := gpuPod()
			mutate(pod)
			if _, err := ComputeKey(pod); !errors.Is(err, ErrNotKeyable) {
				t.Errorf("err = %v, want ErrNotKeyable", err)
			}
		})
	}
}

// "$$" escapes to a literal "$", so "$$(NAME)" is text Kubernetes never
// expands. Refusing it would cost a cold start for a pod that is perfectly
// keyable.
func TestEscapedReferenceIsStillKeyable(t *testing.T) {
	keyable := map[string]string{
		"escaped":            "$$(SNAPSHOT_CONTROL_DIR)/x",
		"escaped in command": "literal $$(SNAPSHOT_JOB_UID) here",
		"unterminated":       "$(SNAPSHOT_CONTROL_DIR/x",
		"not a reference":    "$SNAPSHOT_CONTROL_DIR",
		"other variable":     "$(HOME)/x",
	}

	for name, value := range keyable {
		t.Run(name, func(t *testing.T) {
			pod := gpuPod()
			pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env,
				corev1.EnvVar{Name: "MY_PATH", Value: value})

			if _, err := ComputeKey(pod); err != nil {
				t.Errorf("ComputeKey = %v, want keyable", err)
			}
		})
	}
}

// A doubled escape restores the reference: "$$" is one literal "$", and the
// "$(NAME)" that follows is expanded.
func TestDoubleEscapedReferenceIsNotKeyable(t *testing.T) {
	pod := gpuPod()
	pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env,
		corev1.EnvVar{Name: "MY_PATH", Value: "$$$(SNAPSHOT_CONTROL_DIR)"})

	if _, err := ComputeKey(pod); !errors.Is(err, ErrNotKeyable) {
		t.Errorf("err = %v, want ErrNotKeyable", err)
	}
}

// A sidecar is not hashed, so what it expands cannot affect the key and must
// not make the pod unkeyable.
func TestSidecarExpandingAStrippedVariableIsStillKeyable(t *testing.T) {
	pod := gpuPod()
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name:    "queue-proxy",
		Image:   "knative/queue",
		Command: []string{"$(SNAPSHOT_CONTROL_DIR)/run"},
	})

	if keyOrFail(t, pod) != keyOrFail(t, gpuPod()) {
		t.Error("a sidecar's reference changed the key")
	}
}

// Snapshot's own variables are still stripped when nothing expands them, so a
// restored pod keeps matching the cold pod it came from.
func TestStrippedVariableAloneIsStillKeyable(t *testing.T) {
	pod := gpuPod()
	pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env,
		corev1.EnvVar{Name: "SNAPSHOT_CONTROL_DIR", Value: "/snapshot-control"})

	if keyOrFail(t, pod) != keyOrFail(t, gpuPod()) {
		t.Error("a Snapshot-owned variable changed the key")
	}
}

// A raw block volume is consumed through VolumeDevices and has no VolumeMount,
// so it must still count as "their volume" - otherwise a change to the backing
// claim would not change the key and a replica could restore against different
// storage.
func TestVolumeDeviceVolumeIsHashed(t *testing.T) {
	withClaim := func(claim string) *corev1.Pod {
		pod := gpuPod()
		pod.Spec.Containers[0].VolumeDevices = []corev1.VolumeDevice{{Name: "block", DevicePath: "/dev/xvda"}}
		pod.Spec.Volumes = []corev1.Volume{{
			Name: "block",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: claim},
			},
		}}
		return pod
	}

	if keyOrFail(t, withClaim("data-a")) == keyOrFail(t, withClaim("data-b")) {
		t.Error("changing the claim behind a VolumeDevice did not change the key")
	}
}

// Snapshot's additions are matched by exact name, so a workload's own
// similarly-named fields still count towards its shape.
func TestUserOwnedSnapshotLikeFieldsAreHashed(t *testing.T) {
	base := keyOrFail(t, gpuPod())

	cases := map[string]func(*corev1.Pod){
		"env named SNAPSHOT_SOMETHING": func(p *corev1.Pod) {
			p.Spec.Containers[0].Env = append(p.Spec.Containers[0].Env,
				corev1.EnvVar{Name: "SNAPSHOT_TUNING", Value: "on"})
		},
		"volume named snapshot-cache": func(p *corev1.Pod) {
			p.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{Name: "snapshot-cache", MountPath: "/cache"}}
			p.Spec.Volumes = []corev1.Volume{{Name: "snapshot-cache"}}
		},
	}

	for name, mutate := range cases {
		pod := gpuPod()
		mutate(pod)
		if keyOrFail(t, pod) == base {
			t.Errorf("%s did not change the key", name)
		}
	}
}

// Design 8.3 hashes "their volumes": the ones the hashed containers mount. A
// volume belonging only to a dropped sidecar is not one of them - otherwise an
// istio or queue-proxy upgrade would rekey every opted-in pod in the cluster.
func TestSidecarOnlyVolumeIsNotHashed(t *testing.T) {
	base := keyOrFail(t, gpuPod())

	pod := gpuPod()
	pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{
		Name:         "queue-proxy",
		Image:        "knative/queue",
		VolumeMounts: []corev1.VolumeMount{{Name: "proxy-certs", MountPath: "/certs"}},
	})
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{Name: "proxy-certs"})

	if got := keyOrFail(t, pod); got != base {
		t.Errorf("a sidecar-only volume changed the key: %s -> %s", base, got)
	}
}

// A volume the target container mounts is part of the shape, so changing it
// must change the key.
func TestTargetVolumeIsHashed(t *testing.T) {
	base := keyOrFail(t, gpuPod())

	pod := gpuPod()
	pod.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{Name: "models", MountPath: "/models"}}
	pod.Spec.Volumes = []corev1.Volume{{Name: "models"}}

	if keyOrFail(t, pod) == base {
		t.Error("a volume the target mounts did not change the key")
	}
}

// A run-to-completion init container is part of the pod's shape, unlike a
// sidecar, so it must change the key.
func TestRunToCompletionInitContainerChangesTheKey(t *testing.T) {
	pod := gpuPod()
	pod.Spec.InitContainers = []corev1.Container{{Name: "fetch-model", Image: "busybox"}}

	if keyOrFail(t, pod) == keyOrFail(t, gpuPod()) {
		t.Error("a run-to-completion init container did not change the key")
	}
}

// Matches runs on a Ready replica, which on the restore path is a pod the
// webhook already rewrote. If Snapshot's own additions reached the hash, every
// restored replica would fail the drift check and snapshots would switch off
// for the workload with nothing logged as an error.
func TestRestoredPodHashesLikeTheColdPod(t *testing.T) {
	cold := gpuPod()
	key := keyOrFail(t, cold)

	restored := gpuPod()
	SetKey(restored, key)
	restored.Spec.SecurityContext = &corev1.PodSecurityContext{
		SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeLocalhost},
	}
	restored.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
		SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeLocalhost},
	}
	// Exactly what podcontract.Build writes, legacy spellings included: the
	// DYN_ names share no prefix with the current ones, so a prefix match would
	// leave them in the hash and every restored replica would fail this check.
	for _, name := range []string{
		"SNAPSHOT_RESTORE_STANDBY", "DYN_SNAPSHOT_RESTORE_STANDBY",
		"SNAPSHOT_CONTROL_DIR", "DYN_SNAPSHOT_CONTROL_DIR",
	} {
		restored.Spec.Containers[0].Env = append(restored.Spec.Containers[0].Env,
			corev1.EnvVar{Name: name, Value: "/snapshot-control"})
	}
	restored.Spec.Volumes = append(restored.Spec.Volumes, corev1.Volume{Name: "snapshot-control"})
	restored.Spec.Containers[0].VolumeMounts = append(restored.Spec.Containers[0].VolumeMounts,
		corev1.VolumeMount{Name: "snapshot-control", MountPath: "/snapshot-control"})

	matches, err := Matches(restored)
	if err != nil {
		t.Fatalf("Matches: %v", err)
	}
	if !matches {
		t.Error("a restored pod did not match the key of the cold pod it came from")
	}
}

func TestNotKeyable(t *testing.T) {
	cases := map[string]func(*corev1.Pod){
		"no owner references": func(p *corev1.Pod) { p.OwnerReferences = nil },
		"owner is not the controller": func(p *corev1.Pod) {
			p.OwnerReferences[0].Controller = ptr.To(false)
		},
		"no gpu container": func(p *corev1.Pod) {
			p.Spec.Containers[0].Resources = corev1.ResourceRequirements{}
		},
		"two gpu containers": func(p *corev1.Pod) {
			second := p.Spec.Containers[0].DeepCopy()
			second.Name = "vllm-2"
			p.Spec.Containers = append(p.Spec.Containers, *second)
		},
	}

	for name, mutate := range cases {
		pod := gpuPod()
		mutate(pod)
		if _, err := ComputeKey(pod); !errors.Is(err, ErrNotKeyable) {
			t.Errorf("%s: err = %v, want ErrNotKeyable", name, err)
		}
	}
}

func TestKeyOfAndSetKey(t *testing.T) {
	pod := gpuPod()
	if _, ok := KeyOf(pod); ok {
		t.Error("KeyOf found a key on an unstamped pod")
	}

	// An opted-in pod with no key means the webhook did not run.
	pod.Labels = nil
	SetKey(pod, "3f9a1c2b7d4e")
	key, ok := KeyOf(pod)
	if !ok || key != "3f9a1c2b7d4e" {
		t.Errorf("KeyOf after SetKey = %q, %v; want the stamped key", key, ok)
	}
}

func TestMatchesIsFalseAfterAnEdit(t *testing.T) {
	pod := gpuPod()
	SetKey(pod, keyOrFail(t, pod))

	pod.Spec.Containers[0].Image = "registry.example.com/llama@sha256:def"

	matches, err := Matches(pod)
	if err != nil {
		t.Fatalf("Matches: %v", err)
	}
	if matches {
		t.Error("Matches = true after the image changed")
	}
}

func TestMatchesOnAnUnstampedPod(t *testing.T) {
	if _, err := Matches(gpuPod()); !errors.Is(err, ErrNotKeyable) {
		t.Errorf("err = %v, want ErrNotKeyable", err)
	}
}

func TestSnapshotNameIsAValidLabelValue(t *testing.T) {
	key := keyOrFail(t, gpuPod())
	name := SnapshotName(key)

	if len(name) > validation.DNS1123LabelMaxLength {
		t.Errorf("name %q is %d characters, want at most %d",
			name, len(name), validation.DNS1123LabelMaxLength)
	}
	if errs := validation.IsValidLabelValue(name); len(errs) > 0 {
		t.Errorf("name %q is not a valid label value: %v", name, errs)
	}
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		t.Errorf("name %q is not a valid object name: %v", name, errs)
	}
	if !strings.Contains(name, key) {
		t.Errorf("name %q does not contain the key %q", name, key)
	}
}
