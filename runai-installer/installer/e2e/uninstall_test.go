//go:build e2e

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package e2e_test

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/e2e/utils"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/uninstall"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/apiclient"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func namespaceExists(name string) (bool, error) {
	cfg, err := utils.GetK8sConfig()
	if err != nil {
		return false, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return false, err
	}
	_, err = clientset.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func newClientset() (*kubernetes.Clientset, error) {
	cfg, err := utils.GetK8sConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func newDynamicClient() (dynamic.Interface, error) {
	cfg, err := utils.GetK8sConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}

// listMutatingWebhooks returns names of MutatingWebhookConfigurations matching
// the given label selector (e.g. "app=runai").
func listMutatingWebhooks(labelSelector string) ([]string, error) {
	clientset, err := newClientset()
	if err != nil {
		return nil, err
	}
	list, err := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().List(
		context.Background(), metav1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	return names, nil
}

// listValidatingWebhooks returns names of ValidatingWebhookConfigurations matching
// the given label selector (e.g. "app=runai").
func listValidatingWebhooks(labelSelector string) ([]string, error) {
	clientset, err := newClientset()
	if err != nil {
		return nil, err
	}
	list, err := clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(
		context.Background(), metav1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, item := range list.Items {
		names = append(names, item.Name)
	}
	return names, nil
}

// listResourceInterfacesByLabelKeyPrefix returns names of ResourceInterface objects
// that have at least one label key starting with the given prefix.
// Kubernetes label selectors do not support key-prefix matching, so filtering
// is done client-side after a full list.
func listResourceInterfacesByLabelKeyPrefix(labelKeyPrefix string) ([]string, error) {
	dynamicClient, err := newDynamicClient()
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{Group: "optimization.nvidia.com", Version: "v1", Resource: "resourceinterfaces"}
	list, err := dynamicClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// The optimization.nvidia.com CRD is not installed on this cluster — no instances can exist.
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, item := range list.Items {
		for key := range item.GetLabels() {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, item.GetName())
				break
			}
		}
	}
	return names, nil
}

// listClusterRolesByLabelKeyPrefix returns names of ClusterRoles that have at
// least one label key starting with the given prefix.
func listClusterRolesByLabelKeyPrefix(labelKeyPrefix string) ([]string, error) {
	clientset, err := newClientset()
	if err != nil {
		return nil, err
	}
	list, err := clientset.RbacV1().ClusterRoles().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, item := range list.Items {
		for key := range item.GetLabels() {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, item.Name)
				break
			}
		}
	}
	return names, nil
}

// listPriorityClassesByLabelKeyPrefix returns names of PriorityClasses that have
// at least one label key starting with the given prefix.
func listPriorityClassesByLabelKeyPrefix(labelKeyPrefix string) ([]string, error) {
	clientset, err := newClientset()
	if err != nil {
		return nil, err
	}
	list, err := clientset.SchedulingV1().PriorityClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, item := range list.Items {
		for key := range item.GetLabels() {
			if strings.HasPrefix(key, labelKeyPrefix) {
				names = append(names, item.Name)
				break
			}
		}
	}
	return names, nil
}

// assertNoRunAIResources asserts that no RunAI resources exist in the cluster.
// context is included in failure messages to indicate when the check is performed
// (e.g. "before uninstall on a clean env", "after uninstall").
func assertNoRunAIResources(context string) {
	_, cpErr := getHelmChartVersion(step.CPNamespace, step.CPReleaseName)
	Expect(cpErr).To(HaveOccurred(), "CP release should not exist %s", context)
	Expect(cpErr.Error()).To(ContainSubstring("release: not found"), "CP release lookup failed for an unexpected reason %s", context)

	_, clusterErr := getHelmChartVersion(step.ClusterNamespace, step.ClusterReleaseName)
	Expect(clusterErr).To(HaveOccurred(), "cluster release should not exist %s", context)
	Expect(clusterErr.Error()).To(ContainSubstring("release: not found"), "cluster release lookup failed for an unexpected reason %s", context)

	cpNsExists, err := namespaceExists(step.CPNamespace)
	Expect(err).ToNot(HaveOccurred())
	Expect(cpNsExists).To(BeFalse(), "CP namespace should not exist %s", context)

	clusterNsExists, err := namespaceExists(step.ClusterNamespace)
	Expect(err).ToNot(HaveOccurred())
	Expect(clusterNsExists).To(BeFalse(), "cluster namespace should not exist %s", context)

	reservationNsExists, err := namespaceExists(step.ReservationNamespace)
	Expect(err).ToNot(HaveOccurred())
	Expect(reservationNsExists).To(BeFalse(), "runai-reservation namespace should not exist %s", context)

	scaleAdjustNsExists, err := namespaceExists(step.ScaleAdjustNamespace)
	Expect(err).ToNot(HaveOccurred())
	Expect(scaleAdjustNsExists).To(BeFalse(), "runai-scale-adjust namespace should not exist %s", context)

	mutatingWebhooks, err := listMutatingWebhooks("app=runai")
	Expect(err).ToNot(HaveOccurred())
	Expect(mutatingWebhooks).To(BeEmpty(), "no RunAI MutatingWebhookConfigurations should exist %s", context)

	validatingWebhooks, err := listValidatingWebhooks("app=runai")
	Expect(err).ToNot(HaveOccurred())
	Expect(validatingWebhooks).To(BeEmpty(), "no RunAI ValidatingWebhookConfigurations should exist %s", context)

	resourceInterfaces, err := listResourceInterfacesByLabelKeyPrefix("run.ai/")
	Expect(err).ToNot(HaveOccurred())
	Expect(resourceInterfaces).To(BeEmpty(), "no RunAI ResourceInterfaces should exist %s", context)

	clusterRoles, err := listClusterRolesByLabelKeyPrefix("run.ai/")
	Expect(err).ToNot(HaveOccurred())
	Expect(clusterRoles).To(BeEmpty(), "no RunAI ClusterRoles should exist %s", context)

	priorityClasses, err := listPriorityClassesByLabelKeyPrefix("run.ai/")
	Expect(err).ToNot(HaveOccurred())
	Expect(priorityClasses).To(BeEmpty(), "no RunAI PriorityClasses should exist %s", context)
}

var _ = Describe("Uninstall E2E", func() {
	BeforeEach(func() {
		apiclient.ResetClients()
		Expect(utils.Uninstall(testEnv.TempDir)).To(Succeed())
	})

	It("uninstall should remove all RunAI resources from the cluster", func() {
		utils.InstallVersion(installVersion(), testEnv, nil)

		_, cpErr := getHelmChartVersion(step.CPNamespace, step.CPReleaseName)
		Expect(cpErr).ToNot(HaveOccurred(), "CP release should exist after install")

		_, clusterErr := getHelmChartVersion(step.ClusterNamespace, step.ClusterReleaseName)
		Expect(clusterErr).ToNot(HaveOccurred(), "cluster release should exist after install")

		cfg := &config.UninstallConfig{
			DevMode:          true,
			LogsDir:          testEnv.TempDir,
			StepTimeout:      10 * time.Minute,
			ExecutionTimeout: 30 * time.Minute,
		}

		deps, err := uninstall.ProductionDeps()
		Expect(err).ToNot(HaveOccurred())

		err = orchestrator.RunUninstall(cfg, deps)
		Expect(err).ToNot(HaveOccurred())

		assertNoRunAIResources("after uninstall")
	})
})
