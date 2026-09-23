//go:build e2e

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const lwsControllerDeployment = "lws-controller-manager"

var runaiConfigGVR = schema.GroupVersionResource{Group: "run.ai", Version: "v1", Resource: "runaiconfigs"}

// LWSControllerInstalled reports whether the LWS controller Deployment exists in
// the given namespace, regardless of whether Helm or an OLM operator put it there.
func LWSControllerInstalled(namespace string) error {
	cfg, err := GetK8sConfig()
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	_, err = clientset.AppsV1().Deployments(namespace).Get(context.Background(), lwsControllerDeployment, metav1.GetOptions{})
	return err
}

// RunaiConfigLWSNamespace returns spec.lws.namespace of the runai RunaiConfig.
func RunaiConfigLWSNamespace() (string, error) {
	cfg, err := GetK8sConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get k8s config: %w", err)
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create dynamic client: %w", err)
	}
	runaiConfig, err := dyn.Resource(runaiConfigGVR).Namespace("runai").Get(context.Background(), "runai", metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get runaiconfig: %w", err)
	}
	namespace, _, err := unstructured.NestedString(runaiConfig.Object, "spec", "lws", "namespace")
	return namespace, err
}
