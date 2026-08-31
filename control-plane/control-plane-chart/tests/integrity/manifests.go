/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package integrity

import (
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Manifests struct {
	Deployments  []appsv1.Deployment
	StatefulSets []appsv1.StatefulSet
	DaemonSets   []appsv1.DaemonSet
	Jobs         []batchv1.Job
	CronJobs     []batchv1.CronJob
	Secrets      []corev1.Secret
	ConfigMaps   []corev1.ConfigMap
}

func parseManifests(t *testing.T, rendered string) *Manifests {
	t.Helper()
	out := &Manifests{}
	for i, doc := range splitYAMLDocs(rendered) {
		var meta struct {
			Kind     string            `json:"kind"`
			Metadata metav1.ObjectMeta `json:"metadata"`
		}
		if err := yaml.Unmarshal([]byte(doc), &meta); err != nil {
			t.Fatalf("doc %d: parse kind: %v", i, err)
		}
		switch meta.Kind {
		case "Deployment":
			var d appsv1.Deployment
			mustUnmarshal(t, i, doc, &d)
			out.Deployments = append(out.Deployments, d)
		case "StatefulSet":
			var s appsv1.StatefulSet
			mustUnmarshal(t, i, doc, &s)
			out.StatefulSets = append(out.StatefulSets, s)
		case "DaemonSet":
			var d appsv1.DaemonSet
			mustUnmarshal(t, i, doc, &d)
			out.DaemonSets = append(out.DaemonSets, d)
		case "Job":
			var j batchv1.Job
			mustUnmarshal(t, i, doc, &j)
			out.Jobs = append(out.Jobs, j)
		case "CronJob":
			var c batchv1.CronJob
			mustUnmarshal(t, i, doc, &c)
			out.CronJobs = append(out.CronJobs, c)
		case "Secret":
			var s corev1.Secret
			mustUnmarshal(t, i, doc, &s)
			out.Secrets = append(out.Secrets, s)
		case "ConfigMap":
			var cm corev1.ConfigMap
			mustUnmarshal(t, i, doc, &cm)
			out.ConfigMaps = append(out.ConfigMaps, cm)
		}
	}
	return out
}

func mustUnmarshal(t *testing.T, idx int, doc string, target any) {
	t.Helper()
	if err := yaml.Unmarshal([]byte(doc), target); err != nil {
		t.Fatalf("doc %d: unmarshal: %v", idx, err)
	}
}

func splitYAMLDocs(input string) []string {
	var docs []string
	var cur strings.Builder
	flush := func() {
		s := cur.String()
		if strings.TrimSpace(s) != "" {
			docs = append(docs, s)
		}
		cur.Reset()
	}
	for _, line := range strings.Split(input, "\n") {
		if strings.HasPrefix(line, "---") {
			flush()
			continue
		}
		cur.WriteString(line)
		cur.WriteByte('\n')
	}
	flush()
	return docs
}
