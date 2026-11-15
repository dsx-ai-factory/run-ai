/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"encoding/json"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	environmentservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment"
	trainingclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/training/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	assetsapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"k8s.io/utils/ptr"
)

var _ = Describe("Environment asset submission", func() {
	BeforeEach(func() {
		previousSpec, previousTxn := Spec, txn
		previousURL, previousToken, previousCluster, previousProject := cpUrl, token, clusterId, projectId
		DeferCleanup(func() {
			Spec, txn = previousSpec, previousTxn
			cpUrl, token, clusterId, projectId = previousURL, previousToken, previousCluster, previousProject
			environmentservice.SetClient(nil)
		})
		Spec = spec.NewTrainingSpec()
		txn = transaction.NewTxn()
		cpUrl, token, clusterId, projectId = "http://localhost", "", "cluster-1", "123"
	})

	It("resolves the asset name and keeps explicit variables in the submitted request", func() {
		command := &cobra.Command{Use: "submit"}
		project := "project"
		Spec.SetFlags(command, &project)
		Expect(command.ParseFlags([]string{
			"--environment", "my-environment", "--environment-variable", "A=x=y", "-e", "B=", "--dry-run",
		})).To(Succeed())

		lister := cpmocks.NewMockEnvironmentDescribeLister(gomock.NewController(GinkgoT()))
		lister.EXPECT().List(gomock.Any(), ptr.To[int32](123), ptr.To(clusterId), ptr.To("my-environment"), ptr.To("Training")).
			Return(&assetsapi.EnvironmentListResponse{Entries: []assetsapi.EnvironmentAsset{
				{Meta: assetsapi.AssetMeta{Id: "environment-id", Name: "my-environment"}},
			}}, nil)
		environmentservice.SetClient(lister)
		Expect(resolveEnvironmentAsset()).To(Succeed())

		client, err := trainingclient.NewTrainingClient(cpUrl, token)
		Expect(err).NotTo(HaveOccurred())
		body, err := client.Submit(txn.GetCtx(), "example", projectId, clusterId, Spec)
		Expect(err).NotTo(HaveOccurred())
		var request struct {
			Assets struct {
				Environment string `json:"environment"`
			} `json:"assets"`
			Spec struct {
				EnvironmentVariables []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"environmentVariables"`
			} `json:"spec"`
		}
		Expect(json.Unmarshal([]byte(body), &request)).To(Succeed())
		Expect(request.Assets.Environment).To(Equal("environment-id"))
		Expect(request.Spec.EnvironmentVariables).To(HaveLen(2))
		Expect(request.Spec.EnvironmentVariables[0].Name).To(Equal("A"))
		Expect(request.Spec.EnvironmentVariables[0].Value).To(Equal("x=y"))
		Expect(request.Spec.EnvironmentVariables[1].Name).To(Equal("B"))
		Expect(request.Spec.EnvironmentVariables[1].Value).To(BeEmpty())
	})

})
