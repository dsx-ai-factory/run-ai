/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"log"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/integration-tests/common"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// projectDeletionPollInterval and projectDeletionTimeout bound the wait for a
	// test project to be fully removed during Cleanup (see Cleanup for why).
	projectDeletionPollInterval = 500 * time.Millisecond
	projectDeletionTimeout      = 30 * time.Second
)

type TestConfig struct {
	KubeClient        client.Client
	Ctx               context.Context
	TestNsName        string
	TestNamespace     *corev1.Namespace
	TestProject       *kaiv1alpha1.Project
	TestDepartment    *kaiv1alpha1.Department
	DefaultNodePool   client.Object
	DefaultDepartment *kaiv1alpha1.Department

	KubeConfig *rest.Config
	Scheme     *runtime.Scheme
}

func GetTestConfig(asserter gomega.Gomega) *TestConfig {
	err := initConnectivity()
	if err != nil {
		asserter.Expect(err).NotTo(gomega.HaveOccurred(), "failed to get connectivity objects")
	}

	testConfig := &TestConfig{
		Ctx:        context.Background(),
		KubeClient: kubeClient,
		KubeConfig: kubeConfig,
		Scheme:     scheme,
	}

	testConfig.createBaseRunaiObjects(asserter)
	return testConfig
}

func (c *TestConfig) Cleanup() {
	err := c.KubeClient.Delete(c.Ctx, c.TestDepartment)
	if err != nil {
		fmt.Printf("failed to delete department: %s", err)
	}
	err = c.KubeClient.Delete(c.Ctx, c.TestProject)
	if err != nil {
		fmt.Printf("failed to delete project: %s", err)
	}
	err = c.KubeClient.Delete(c.Ctx, c.TestNamespace)
	if err != nil {
		fmt.Printf("failed to delete namespace: %s", err)
	}
	c.waitForProjectDeleted()
	err = c.KubeClient.Delete(c.Ctx, c.DefaultNodePool)
	if err != nil {
		fmt.Printf("failed to delete default node pool: %s", err)
	}
	err = c.KubeClient.Delete(c.Ctx, c.DefaultDepartment)
	if err != nil {
		fmt.Printf("failed to delete default department: %s", err)
	}
}

// waitForProjectDeleted polls until the test project no longer exists, bounded by
// projectDeletionTimeout. Best-effort: it logs on timeout rather than failing, to
// preserve Cleanup's non-fatal behavior.
func (c *TestConfig) waitForProjectDeleted() {
	key := client.ObjectKeyFromObject(c.TestProject)
	err := wait.PollUntilContextTimeout(c.Ctx, projectDeletionPollInterval, projectDeletionTimeout, true,
		func(ctx context.Context) (bool, error) {
			getErr := c.KubeClient.Get(ctx, key, &kaiv1alpha1.Project{})
			return errors.IsNotFound(getErr), nil
		})
	if err != nil {
		fmt.Printf("timed out waiting for project %s to be deleted before removing default node pool: %s", key.Name, err)
	}
}

// ensureHealthyDefaultNodePool guarantees a non-terminating default node pool exists
// before a suite runs.
func (c *TestConfig) ensureHealthyDefaultNodePool(asserter gomega.Gomega, defaultNodePool client.Object) {
	existing := newEmptyNodePoolObject()
	getErr := kubeClient.Get(c.Ctx, client.ObjectKey{Name: v1alpha1.DefaultNodePoolName}, existing)
	if getErr == nil && !existing.GetDeletionTimestamp().IsZero() {
		log.Printf("default node pool is stuck terminating (finalizers %v); clearing finalizers to recover it",
			existing.GetFinalizers())
		existing.SetFinalizers(nil)
		if uErr := kubeClient.Update(c.Ctx, existing); uErr != nil && !errors.IsNotFound(uErr) {
			fmt.Printf("failed clearing finalizers on terminating default node pool: %s", uErr)
		}
		if wErr := wait.PollUntilContextTimeout(c.Ctx, projectDeletionPollInterval, projectDeletionTimeout, true,
			func(ctx context.Context) (bool, error) {
				return errors.IsNotFound(kubeClient.Get(ctx, client.ObjectKey{Name: v1alpha1.DefaultNodePoolName}, newEmptyNodePoolObject())), nil
			}); wErr != nil {
			fmt.Printf("timed out waiting for stuck default node pool to be deleted: %s", wErr)
		}
	}

	err := kubeClient.Create(c.Ctx, defaultNodePool)
	if err != nil && !errors.IsAlreadyExists(err) {
		asserter.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create default node pool")
	}
}

func (c *TestConfig) createBaseRunaiObjects(asserter gomega.Gomega) {
	projName := common.GetRandomName("project")
	depName := common.GetRandomName("department")

	projId := common.GetRandomName("project-id")
	depId := common.GetRandomName("department-id")

	testProj1 := BuildProjectObject(projName, depName, projId, depId)
	testNs1 := BuildNamespaceObject(fmt.Sprintf("runai-%s", projName), &projName)
	testDep1 := buildDepartmentObject(depName, depId)
	defaultNodePool := buildDefaultNodePoolObject()
	defaultDepartment := buildDepartmentObject("default", "10")

	c.ensureHealthyDefaultNodePool(asserter, defaultNodePool)
	err := kubeClient.Create(c.Ctx, defaultDepartment)
	if err != nil && !errors.IsAlreadyExists(err) {
		asserter.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create default department")
	}
	err = kubeClient.Create(c.Ctx, testDep1)
	if err != nil && !errors.IsAlreadyExists(err) {
		asserter.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create department")
	}
	err = kubeClient.Create(c.Ctx, testProj1)
	if err != nil && !errors.IsAlreadyExists(err) {
		asserter.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create project")
	}
	err = kubeClient.Create(c.Ctx, testNs1)
	if err != nil && !errors.IsAlreadyExists(err) {
		asserter.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create namespace")
	}
	log.Printf("created project: %s, namespace: %s", projName, testNs1.Name)
	c.TestDepartment = testDep1
	c.TestProject = testProj1
	c.TestNamespace = testNs1
	c.TestNsName = testNs1.Name
	c.DefaultNodePool = defaultNodePool
	c.DefaultDepartment = defaultDepartment
}
