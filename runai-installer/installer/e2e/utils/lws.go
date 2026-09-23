//go:build e2e

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	runaiNamespace          = "runai"
	runaiPublicConfigMap    = "runai-public"
	lwsControllerDeployment = "lws-controller-manager"
)

var runaiConfigGVR = schema.GroupVersionResource{Group: "run.ai", Version: "v1", Resource: "runaiconfigs"}

// publicStatus mirrors the dependency entries of the runai-public ConfigMap the
// Run:ai operator publishes (dependencies.optional.<component>).
type publicStatus struct {
	Available *bool  `yaml:"available"`
	Reason    string `yaml:"reason"`
}

type runaiPublic struct {
	Dependencies struct {
		Optional map[string]publicStatus `yaml:"optional"`
	} `yaml:"dependencies"`
}

// LWSControllerAvailable fails unless the LWS controller Deployment in the given
// namespace has at least one available replica, regardless of whether Helm or
// an OLM operator put it there.
func LWSControllerAvailable(namespace string) error {
	cfg, err := GetK8sConfig()
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), lwsControllerDeployment, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if deployment.Status.AvailableReplicas == 0 {
		return fmt.Errorf("deployment %s/%s has no available replicas", namespace, lwsControllerDeployment)
	}
	return nil
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
	runaiConfig, err := dyn.Resource(runaiConfigGVR).Namespace(runaiNamespace).Get(context.Background(), runaiNamespace, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get runaiconfig: %w", err)
	}
	namespace, _, err := unstructured.NestedString(runaiConfig.Object, "spec", "lws", "namespace")
	return namespace, err
}

// RunaiPublicLWSAvailable fails unless the Run:ai operator reports LWS as an
// available optional dependency in the runai-public ConfigMap, which is what
// makes the platform offer distributed inference.
func RunaiPublicLWSAvailable() error {
	cfg, err := GetK8sConfig()
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %w", err)
	}
	configMap, err := clientset.CoreV1().ConfigMaps(runaiNamespace).Get(context.Background(), runaiPublicConfigMap, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get %s configmap: %w", runaiPublicConfigMap, err)
	}
	var public runaiPublic
	if err := yaml.Unmarshal([]byte(configMap.Data[runaiPublicConfigMap]), &public); err != nil {
		return fmt.Errorf("failed to parse %s configmap: %w", runaiPublicConfigMap, err)
	}
	lws := public.Dependencies.Optional["lws"]
	if lws.Available == nil || !*lws.Available {
		return fmt.Errorf("runai-public reports LWS unavailable: %s", lws.Reason)
	}
	return nil
}
