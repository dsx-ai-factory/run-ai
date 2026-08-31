/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package scale_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/scale"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nim"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type standardInferencePaginator struct{}

func (p *standardInferencePaginator) List(_ context.Context, _, _ int32, filters []string, _ bool) (*openapi.GetWorkloads200Response, error) {
	for _, value := range filters {
		if value == "type=="+api.BackendNimServiceWorkloadType {
			return &openapi.GetWorkloads200Response{}, nil
		}
	}

	return &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
		{Id: "inference-id", Name: "llama-repro", Type: api.Inference.QueryParam()},
	}}, nil
}

var _ = Describe("NIM scale Jira regression", func() {
	BeforeEach(func() {
		workload.SetPaginator(&standardInferencePaginator{})
		DeferCleanup(func() { workload.SetPaginator(nil) })
	})

	It("rejects a standard inference workload before calling the NIM API", func() {
		instance := config.GetInstance()
		Expect(instance.Initializing(config.Options{
			Debug:      true,
			ConfigPath: GinkgoT().TempDir(),
			ConfigFile: "config.json",
		})).To(Succeed())
		originalArgs := os.Args
		os.Args = []string{"runai", "inference", "nim", "scale", "llama-repro", "-p", "thea", "--replicas", "0"}
		DeferCleanup(func() { os.Args = originalArgs })
		updateCalled := false
		searchWorkload := func(_ *transaction.Txn, domain, token, name, project, cluster string, fb *filter.Builder) (*list.Output, error) {
			return nim.GetByNameAndProj(transaction.NewTxn(), domain, token, name, project, cluster, fb)
		}
		updateWorkload := func(_ *transaction.Txn, _, _, _, _ string, _ *spec.GeneralSpec) (string, error) {
			updateCalled = true
			return "", nil
		}
		scaleCommand := scale.NewCommand(
			api.NimService,
			api.Inference,
			spec.NewNimSpec(),
			cmd.NewEmptyCobraCommand,
			scale.NewScaleText(api.Inference, api.NimService),
			instance,
			helper.Cmd{},
			searchWorkload,
			updateWorkload,
			factories.GetCompletionFnMock(),
		)
		Expect(scaleCommand.RegisterFlags()).To(Succeed())
		command := scaleCommand.GetCommand()
		command.SetArgs([]string{"llama-repro", "-p", "thea", "--replicas", "0"})

		err := command.Execute()

		Expect(err).To(MatchError("workload not found"))
		Expect(updateCalled).To(BeFalse())
	})
})
