// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	apiconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction/info"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction/reporter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
	"github.com/spf13/cobra"
)

func newRecommendCommand() *cobra.Command {
	var image, tag, gpuType string
	command := &cobra.Command{
		Use: "recommend", Short: "Recommend GPU allocation for a NIM test/dev workload as JSON",
		Long:    "Recommend GPU allocation for an exact NIM image tag using the selected cluster. Does not submit a workload or reserve resources.",
		GroupID: consts.GroupInferenceBaseID, Args: cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error { return nil },
		// This JSON-only command skips the root upgrade flow, which can print
		// notices or restart the CLI after the response. Keep transaction reporting.
		PersistentPostRunE: reporter.Report,
		RunE: func(cmd *cobra.Command, _ []string) error {
			request := catalog.NewRightSizingRequest(instance.GetConfigValue(config.ClusterUUID), image, tag)
			if cmd.Flags().Changed("gpu-type") {
				request.SetGpuType(gpuType)
			}
			ctx := cmd.Context()
			if id := instance.GetTxn().GetCtx().Value(info.Id); id != nil {
				ctx = context.WithValue(ctx, info.Id, id)
			}
			return runRecommendation(ctx, instance.GetConfigValue(config.ControlPlaneUrl), instance.GetToken(), *request, cmd.OutOrStdout())
		},
	}
	command.Flags().StringVar(&image, "image", "", "Untagged nvcr.io/nim/organization/model image")
	command.Flags().StringVar(&tag, "tag", "", "Exact image tag")
	command.Flags().StringVar(&gpuType, "gpu-type", "", "Exact cluster GPU type (default: all reported types)")
	_ = command.MarkFlagRequired("image")
	_ = command.MarkFlagRequired("tag")
	return command
}

func runRecommendation(ctx context.Context, baseURL, token string, request catalog.RightSizingRequest, out io.Writer) error {
	configuration, err := apiconfig.CreateCatalogConfig(baseURL, token)
	if err != nil {
		return fmt.Errorf("configure sizing API: %w", err)
	}
	result, response, err := catalog.NewAPIClient(configuration).RightSizingAPI.RecommendInferenceResources(ctx).RightSizingRequest(request).Execute()
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return clierrors.HandleResponseError(response, err)
	}
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("write sizing recommendation: %w", err)
	}
	return nil
}
