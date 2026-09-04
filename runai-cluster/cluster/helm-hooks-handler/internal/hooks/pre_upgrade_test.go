/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hooks_test

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"log"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	schemeInit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"

	pre_hooks "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/pre-hooks"

	testu "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("PreUpgrade Hook", func() {
	var (
		resources *clients.Clients
		server    *ghttp.Server
	)
	BeforeEach(func() {
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/_version"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, Version{
					Version: "2.22.5",
				}),
			),
		)
	})

	It("succeed where both runai and runai-public configmap are not found", func() {

		resources = &clients.Clients{
			LocalClusterClient:        fake.NewClientBuilder().WithScheme(schemeInit.Scheme()).Build(),
			LocalClusterDynamicClient: testu.NewTestDynamicClient(false, false),
		}
		resources.RemoteClusterClient = resources.LocalClusterClient
		resources.RemoteClusterDynamicClient = resources.LocalClusterDynamicClient
		Expect(resources.LocalClusterClient.Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: runv1.DefaultNamespace,
			},
		})).To(Succeed())
		hook := pre_hooks.NewPreUpgradeHook(&hooks.ClusterHookParams{
			ControlPlaneDetails: clients.ControlPlaneDetails{
				URL: server.URL(),
			},
			ClusterDetails: hooks.ClusterDetails{
				Version: "2.21.8",
			},
		}, &common.CustomCA{})
		Expect(hook.Call(context.Background(), resources)).To(Succeed())
	})

	It("should succeed for newer versions where both runai and runai-public configmap exists", func() {
		log.Println("Starting test")
		resources = &clients.Clients{
			LocalClusterClient:        fake.NewClientBuilder().WithScheme(schemeInit.Scheme()).Build(),
			LocalClusterDynamicClient: testu.NewTestDynamicClient(true, true),
		}
		resources.RemoteClusterClient = resources.LocalClusterClient
		resources.RemoteClusterDynamicClient = resources.LocalClusterDynamicClient
		Expect(resources.LocalClusterClient.Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: runv1.DefaultNamespace,
			},
		})).To(Succeed())
		hook := pre_hooks.NewPreUpgradeHook(&hooks.ClusterHookParams{
			ControlPlaneDetails: clients.ControlPlaneDetails{
				URL: server.URL(),
			},
			ClusterDetails: hooks.ClusterDetails{
				Version: "2.21.8",
			},
		}, &common.CustomCA{})
		Expect(hook.Call(context.Background(), resources)).To(Succeed())
	})
})
