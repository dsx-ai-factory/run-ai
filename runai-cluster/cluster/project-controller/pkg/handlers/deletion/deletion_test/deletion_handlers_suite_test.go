/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion_test

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Test Globals
var scheme *runtime.Scheme

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv2.AddToScheme(scheme)).Should(Succeed())

	config.SetForTest(test.RunaiConfigForTests())
})

// gvkRestoringClient wraps a client.Client and backfills the TypeMeta
// (apiVersion/kind) on objects returned by List. The controller-runtime fake
// client strips TypeMeta from listed objects, whereas a real API server returns
// it; some deletion handlers read the kind off listed objects to build their
// status messages, so the test client mirrors real-client behavior here.
type gvkRestoringClient struct {
	client.Client
}

func (c gvkRestoringClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if err := c.Client.List(ctx, list, opts...); err != nil {
		return err
	}
	return apimeta.EachListItem(list, func(obj runtime.Object) error {
		gvks, _, err := c.Scheme().ObjectKinds(obj)
		if err != nil || len(gvks) == 0 {
			return nil
		}
		obj.GetObjectKind().SetGroupVersionKind(gvks[0])
		return nil
	})
}

func TestDeletionHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunTest(t, "Project Resource Deletion Handlers Tests", "../../../../../bin/test/results/project_resource_deletion_handlers_test_results.xml")
}

type TestDeletionHandler struct {
	HandlerCalled bool
}

func (handler *TestDeletionHandler) OnDelete(project *kaiv1alpha1.Project) ([]kaiv1alpha1.ProjectCondition, error) {
	handler.HandlerCalled = true
	return []kaiv1alpha1.ProjectCondition{}, nil
}
