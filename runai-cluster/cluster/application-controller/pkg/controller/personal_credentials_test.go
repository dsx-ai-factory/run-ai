/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	personalcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPersonalCredentialLifecycle(t *testing.T) {
	const encryptionKey = "905c18f0839affcddca117dcac633f79"
	t.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, encryptionKey)

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := runv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	app := &runv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "project",
			UID:       types.UID("app-uid"),
			Annotations: map[string]string{
				personalcreds.Annotation: "- sourceSecret: source\n",
			},
		},
	}
	source := credentialSource(t, "source", "first", []byte(encryptionKey))
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, source).Build()
	controller := NewApplicationController(k8sClient)

	if err := controller.materializePersonalCredentials(context.Background(), app); err != nil {
		t.Fatalf("initial materialization failed: %v", err)
	}
	destKey := types.NamespacedName{
		Name: personalcreds.DestName(app.Name, source.Name), Namespace: app.Namespace,
	}
	var dest corev1.Secret
	if err := k8sClient.Get(context.Background(), destKey, &dest); err != nil {
		t.Fatal(err)
	}
	if got := string(dest.Data["token"]); got != "first" {
		t.Fatalf("initial value = %q", got)
	}
	if !metav1.IsControlledBy(&dest, app) {
		t.Fatal("destination is not application-owned")
	}
	if dest.Labels[personalCredentialsOwnerLabel] != app.Name {
		t.Fatal("destination is missing the application owner label")
	}

	if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(source), source); err != nil {
		t.Fatal(err)
	}
	source.Data["token"] = encryptForTest(t, []byte("second"), []byte(encryptionKey))
	if err := k8sClient.Update(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := controller.materializePersonalCredentials(context.Background(), app); err != nil {
		t.Fatalf("rotation failed: %v", err)
	}
	if err := k8sClient.Get(context.Background(), destKey, &dest); err != nil {
		t.Fatal(err)
	}
	if got := string(dest.Data["token"]); got != "second" {
		t.Fatalf("rotated value = %q", got)
	}

	if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(source), source); err != nil {
		t.Fatal(err)
	}
	source.Data["token"] = []byte("invalid-ciphertext")
	if err := k8sClient.Update(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := controller.materializePersonalCredentials(context.Background(), app); err == nil {
		t.Fatal("expected invalid ciphertext to fail")
	}
	if err := k8sClient.Get(context.Background(), destKey, &dest); err != nil {
		t.Fatal(err)
	}
	if got := string(dest.Data["token"]); got != "second" {
		t.Fatalf("working destination changed after retryable failure: %q", got)
	}

	delete(source.Labels, personalcreds.UserCredentialsLabel)
	if err := k8sClient.Update(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := controller.materializePersonalCredentials(context.Background(), app); err == nil {
		t.Fatal("expected removed personal credential label to fail")
	}
	if err := k8sClient.Get(context.Background(), destKey, &dest); !apierrors.IsNotFound(err) {
		t.Fatalf("revoked destination still exists after label removal: %v", err)
	}

	source.Labels[personalcreds.UserCredentialsLabel] = "true"
	source.Data["token"] = encryptForTest(t, []byte("third"), []byte(encryptionKey))
	if err := k8sClient.Update(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := controller.materializePersonalCredentials(context.Background(), app); err != nil {
		t.Fatalf("materialization after restoring source failed: %v", err)
	}
	if err := k8sClient.Get(context.Background(), destKey, &dest); err != nil {
		t.Fatal(err)
	}
	if got := string(dest.Data["token"]); got != "third" {
		t.Fatalf("restored value = %q", got)
	}

	if err := k8sClient.Delete(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := controller.materializePersonalCredentials(context.Background(), app); err == nil {
		t.Fatal("expected missing source to fail")
	}
	if err := k8sClient.Get(context.Background(), destKey, &dest); !apierrors.IsNotFound(err) {
		t.Fatalf("revoked destination still exists: %v", err)
	}
}

func TestPersonalCredentialRefusesForeignDestination(t *testing.T) {
	const encryptionKey = "905c18f0839affcddca117dcac633f79"
	t.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, encryptionKey)

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := runv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	app := &runv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app",
			Namespace: "project",
			UID:       types.UID("app-uid"),
			Annotations: map[string]string{
				personalcreds.Annotation: "- sourceSecret: source\n",
			},
		},
	}
	source := credentialSource(t, "source", "private-value", []byte(encryptionKey))
	foreign := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      personalcreds.DestName(app.Name, source.Name),
			Namespace: app.Namespace,
		},
		Data: map[string][]byte{"token": []byte("foreign-value")},
	}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, source, foreign).Build()
	controller := NewApplicationController(k8sClient)

	if err := controller.materializePersonalCredentials(context.Background(), app); err == nil {
		t.Fatal("expected a destination ownership collision")
	}
	var current corev1.Secret
	if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(foreign), &current); err != nil {
		t.Fatal(err)
	}
	if got := string(current.Data["token"]); got != "foreign-value" {
		t.Fatalf("foreign destination was changed to %q", got)
	}
}

func credentialSource(t *testing.T, name, value string, key []byte) *corev1.Secret {
	t.Helper()
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.NamespaceRunAi,
			Labels: map[string]string{
				personalcreds.UserCredentialsLabel: "true",
				personalcreds.AssetLabel:           personalcreds.GenericAssetKind,
			},
		},
		Data: map[string][]byte{"token": encryptForTest(t, []byte(value), key)},
	}
}

func encryptForTest(t *testing.T, plaintext, key []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	for i := range nonce {
		nonce[i] = byte(i + 1)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil)
}
