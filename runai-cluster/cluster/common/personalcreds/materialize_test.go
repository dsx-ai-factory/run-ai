/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildDestination(t *testing.T) {
	key := []byte("905c18f0839affcddca117dcac633f79")
	encrypted := []byte{
		168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
		224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
		201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0,
	}
	source := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "source",
			Namespace:   "runai",
			Labels:      map[string]string{AssetLabel: DockerRegistryAssetKind},
			Annotations: map[string]string{"source": "annotation"},
		},
		Data: map[string][]byte{"token": encrypted},
	}

	dest, err := BuildDestination("destination", "project", source, key)
	if err != nil {
		t.Fatalf("BuildDestination() error = %v", err)
	}
	if got := string(dest.Data["token"]); got != "private-value-1" {
		t.Fatalf("decrypted token = %q", got)
	}
	if dest.Type != corev1.SecretTypeDockerConfigJson {
		t.Fatalf("type = %q", dest.Type)
	}
	if dest.Labels[CopyLabel] != "true" {
		t.Fatalf("copy label missing")
	}
	if _, found := dest.Labels[AssetLabel]; found {
		t.Fatalf("destination retained asset classification")
	}
	if _, found := dest.Labels[UserCredentialsLabel]; found {
		t.Fatalf("destination retained personal credential classification")
	}
	if source.Labels[CopyLabel] != "" {
		t.Fatalf("source labels were mutated")
	}
}
