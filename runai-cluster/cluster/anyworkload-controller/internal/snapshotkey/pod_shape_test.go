/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshotkey

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

// update rewrites the keyShape files: go test ./internal/snapshotkey -update
var update = flag.Bool("update", false, "rewrite the golden files")

// The keyShape files hold the shape bytes and the resulting key, one per
// supported workload type. A diff names the field that moved, which a changed
// key on its own does not. A key that changes without Version changing orphans
// every stored snapshot, so these are meant to fail loudly.
type keyShape struct {
	Key   string          `json:"key"`
	Shape json.RawMessage `json:"shape"`
}

func gpuResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
	}
}

func ownedBy(kind, name, uid string) []metav1.OwnerReference {
	return []metav1.OwnerReference{{
		APIVersion: "apps/v1",
		Kind:       kind,
		Name:       name,
		UID:        types.UID(uid),
		Controller: ptr.To(true),
	}}
}

func goldenPods() map[string]*corev1.Pod {
	return map[string]*corev1.Pod{
		// Knative revision: the inference path, with the queue-proxy sidecar
		// and the service-account token volume both excluded.
		"knative-revision": {
			ObjectMeta: metav1.ObjectMeta{
				Name:            "llama-00001-deployment-7d4f8b6c9d-abcde",
				Namespace:       "runai-team-a",
				Labels:          map[string]string{constants.SnapshotOptInLabel: "true", "serving.knative.dev/revision": "llama-00001"},
				OwnerReferences: ownedBy("ReplicaSet", "llama-00001-deployment-7d4f8b6c9d", ownerUID),
			},
			Spec: corev1.PodSpec{
				Hostname:         "llama",
				RuntimeClassName: ptr.To("nvidia"),
				Containers: []corev1.Container{
					{
						Name:    "user-container",
						Image:   "registry.example.com/llama@sha256:abc",
						Command: []string{"vllm", "serve"},
						Env: []corev1.EnvVar{
							{Name: "MODEL", Value: "llama-3"},
							{Name: "PORT", Value: "8000"},
						},
						Resources:      gpuResources(),
						ReadinessProbe: &corev1.Probe{},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "models", MountPath: "/models"},
							{Name: "kube-api-access-x7k2p", MountPath: "/var/run/secrets/kubernetes.io/serviceaccount"},
						},
					},
					{Name: "queue-proxy", Image: "knative/queue"},
				},
				Volumes: []corev1.Volume{
					{Name: "models"},
					{Name: "kube-api-access-x7k2p"},
				},
				NodeName:          "gpu-node-7",
				SchedulerName:     "runai-scheduler",
				PriorityClassName: "low",
			},
		},
		// Deployment with a run-to-completion init container, which is part of
		// the pod's shape and so is hashed.
		"deployment-with-init": {
			ObjectMeta: metav1.ObjectMeta{
				Name:            "trainer-6b9c4f7d8e-xyzab",
				Namespace:       "runai-team-b",
				Labels:          map[string]string{constants.SnapshotOptInLabel: "true"},
				OwnerReferences: ownedBy("ReplicaSet", "trainer-6b9c4f7d8e", "11111111-2222-3333-4444-555555555555"),
			},
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					{Name: "fetch-model", Image: "busybox:1.37.0", Command: []string{"sh", "-c", "fetch"}},
					{Name: "logs", Image: "fluentbit", RestartPolicy: ptr.To(corev1.ContainerRestartPolicyAlways)},
				},
				Containers: []corev1.Container{{
					Name:      "trainer",
					Image:     "registry.example.com/trainer@sha256:def",
					Resources: gpuResources(),
				}},
				SecurityContext: &corev1.PodSecurityContext{RunAsUser: ptr.To(int64(1000))},
			},
		},
		// StatefulSet: a different owner kind, and a pod-level security
		// context that is hashed while the seccomp profile is not.
		"statefulset": {
			ObjectMeta: metav1.ObjectMeta{
				Name:            "ranker-0",
				Namespace:       "runai-team-c",
				Labels:          map[string]string{constants.SnapshotOptInLabel: "true"},
				OwnerReferences: ownedBy("StatefulSet", "ranker", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name:      "ranker",
					Image:     "registry.example.com/ranker@sha256:123",
					Resources: gpuResources(),
					EnvFrom: []corev1.EnvFromSource{
						{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "creds"}}},
						{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "settings"}}},
					},
				}},
				SecurityContext: &corev1.PodSecurityContext{
					FSGroup:        ptr.To(int64(2000)),
					SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeLocalhost},
				},
			},
		},
	}
}

func TestGolden(t *testing.T) {
	for name, pod := range goldenPods() {
		t.Run(name, func(t *testing.T) {
			shape, err := Shape(pod)
			if err != nil {
				t.Fatalf("Shape: %v", err)
			}
			key, err := ComputeKey(pod)
			if err != nil {
				t.Fatalf("Compute: %v", err)
			}

			// Indented so a diff is readable line by line.
			var indented bytes.Buffer
			if err := json.Indent(&indented, shape, "", "  "); err != nil {
				t.Fatalf("indent: %v", err)
			}
			got, err := json.MarshalIndent(keyShape{Key: key, Shape: indented.Bytes()}, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			got = append(got, '\n')

			path := filepath.Join("testdata", name+".json")
			if *update {
				if err := os.WriteFile(path, got, 0o600); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				return
			}

			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden (run with -update to create): %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("shape or key changed for %s.\n"+
					"If this is intended, bump Version and rerun with -update.\n"+
					"got:\n%s\nwant:\n%s", name, got, want)
			}
		})
	}
}
