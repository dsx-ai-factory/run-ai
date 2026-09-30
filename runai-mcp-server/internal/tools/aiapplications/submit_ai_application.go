/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aiapplications

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type aiApplicationChartRef struct {
	Repo         string `json:"repo,omitempty"         jsonschema:"HTTP, HTTPS, or OCI Helm repository reference, up to 250 characters. Use with chartName and chartVersion, and do not combine with chartURL or chartId."`
	ChartName    string `json:"chartName,omitempty"    jsonschema:"Chart name within repo, up to 250 characters. Required when repo is used."`
	ChartVersion string `json:"chartVersion,omitempty" jsonschema:"Semantic chart version, up to 50 characters. Required when repo is used."`
	ChartURL     string `json:"chartURL,omitempty"     jsonschema:"Direct HTTPS URL to a packaged Helm chart (.tgz), up to 250 characters. Do not combine with repository fields or chartId."`
	ChartID      string `json:"chartId,omitempty"      jsonschema:"Identifier of a chart archive previously uploaded to Run:ai. Do not combine with repository fields or chartURL."`
}

type aiApplicationAuth struct {
	AssetID        string `json:"assetId"                  jsonschema:"Credential asset ID used to authenticate to a private Helm repository or chart URL."`
	UserCredential bool   `json:"userCredential,omitempty" jsonschema:"Set to true when assetId identifies a personal credential owned by the caller. Omitted or false means a shared credential."`
}

type aiApplicationSpec struct {
	ChartRef     aiApplicationChartRef `json:"chartRef"              jsonschema:"Helm chart origin. Provide exactly one source: repo with chartName and chartVersion, chartURL, or chartId."`
	Values       map[string]any        `json:"values,omitempty"       jsonschema:"Helm values. Personal credentials use whole-string USERCREDENTIAL:<name>[,generic-secret|,docker-registry]."`
	ValuesBase64 *string               `json:"valuesBase64,omitempty" jsonschema:"Base64 Helm values. Personal credentials use the same USERCREDENTIAL grammar as values."`
	Auth         *aiApplicationAuth    `json:"auth,omitempty"         jsonschema:"Authentication for a private Helm chart source."`
}

type submitAIApplicationInput struct {
	ClusterID string            `json:"clusterId"         jsonschema:"UUID of the cluster where the AI application is scheduled."`
	ProjectID string            `json:"projectId"         jsonschema:"ID of the project that owns the AI application."`
	Name      string            `json:"name"              jsonschema:"Unique AI application name, up to 63 characters. Use lowercase letters, digits, and hyphens; start with a letter and end with a letter or digit."`
	Spec      aiApplicationSpec `json:"spec"              jsonschema:"Complete AI application Helm chart specification."`
	Confirm   bool              `json:"confirm,omitempty" jsonschema:"Required to submit the AI application when the client cannot show an interactive confirmation prompt."`
}

func (in submitAIApplicationInput) Validate() error {
	if err := validateAIApplicationChartRef(in.Spec.ChartRef); err != nil {
		return err
	}
	if in.Spec.Values != nil && in.Spec.ValuesBase64 != nil {
		return fmt.Errorf("spec.values and spec.valuesBase64 are mutually exclusive")
	}
	if in.Spec.ValuesBase64 != nil {
		if _, err := base64.StdEncoding.DecodeString(*in.Spec.ValuesBase64); err != nil {
			return fmt.Errorf("spec.valuesBase64 must be valid base64")
		}
	}
	if in.Spec.Auth != nil {
		if in.Spec.Auth.AssetID == "" {
			return fmt.Errorf("spec.auth.assetId is required")
		}
	}
	return nil
}

func validateAIApplicationChartRef(ref aiApplicationChartRef) error {
	repoSet := ref.Repo != "" || ref.ChartName != "" || ref.ChartVersion != ""
	if repoSet && (ref.Repo == "" || ref.ChartName == "" || ref.ChartVersion == "") {
		return fmt.Errorf("spec.chartRef.repo, chartName, and chartVersion must all be set for a repository chart")
	}
	sources := 0
	if repoSet {
		sources++
	}
	if ref.ChartURL != "" {
		sources++
	}
	if ref.ChartID != "" {
		sources++
	}
	if sources != 1 {
		return fmt.Errorf("spec.chartRef must provide exactly one source: repo with chartName and chartVersion, chartURL, or chartId")
	}
	return nil
}

type submitAIApplicationOutput struct {
	ID            string `json:"id,omitempty"          jsonschema:"Created AI application ID."`
	Name          string `json:"name"                  jsonschema:"AI application name."`
	ClusterID     string `json:"clusterId"             jsonschema:"Cluster where the AI application is scheduled."`
	ProjectID     string `json:"projectId"             jsonschema:"Project that owns the AI application."`
	DesiredPhase  string `json:"desiredPhase,omitempty" jsonschema:"Desired lifecycle phase returned by the API."`
	Phase         string `json:"phase,omitempty"       jsonschema:"Current Kubernetes or Helm release phase returned by the API."`
	ActionStatus  string `json:"actionStatus,omitempty" jsonschema:"Current AI application action status, such as creating, deployed, or failed."`
	ActionMessage string `json:"actionMessage,omitempty" jsonschema:"Message associated with the current AI application action status."`
}

func registerSubmitAIApplication(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[submitAIApplicationInput, submitAIApplicationOutput]{
		Name: "submit_ai_application",
		Description: "Call when the user asks to submit a new AI application for scheduling from a Helm chart. " +
			"Provide clusterId, projectId, name, and exactly one chart source: an HTTP/HTTPS/OCI repository with " +
			"chartName and chartVersion, a direct HTTPS chartURL, or a previously uploaded chartId. " +
			"The spec can include inline or base64-encoded Helm values and shared or personal credentials for a private chart source. " +
			"Personal runtime references must be exact whole strings using USERCREDENTIAL:<name>; NGC references require ,generic-secret or ,docker-registry. Submission requires confirmation. " +
			"Returns the created application ID and its desired, action, and release phases. " +
			"Do not use this tool to inspect existing applications; use list_ai_applications. " +
			"Do not use it to submit an individual workload; use submit_workload.",
		ReadOnly:    false,
		Destructive: false,
		Handler: func(ctx context.Context, in submitAIApplicationInput) (submitAIApplicationOutput, error) {
			return submitAIApplicationRun(ctx, client, in)
		},
	})
}

func submitAIApplicationRun(ctx context.Context, client *runaiclient.Client, in submitAIApplicationInput) (submitAIApplicationOutput, error) {
	out := submitAIApplicationOutput{
		Name: in.Name, ClusterID: in.ClusterID, ProjectID: in.ProjectID,
	}

	if err := common.ConfirmWrite(ctx, in.Confirm, "create", "AI application", in.Name, "", aiApplicationConfirmationDetails(in)); err != nil {
		return out, err
	}

	created, httpResp, err := client.Workloads.AIApplicationsAPI.CreateApplication(ctx).
		AIApplicationCreationRequest(toAIApplicationCreationRequest(in)).Execute()
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNoContent {
			return out, nil
		}
		mapped := runaimcp.MapHTTPError(httpResp, err)
		outcomeUnknown := httpResp == nil || httpResp.StatusCode < http.StatusMultipleChoices ||
			httpResp.StatusCode >= http.StatusInternalServerError
		if outcomeUnknown {
			return submitAIApplicationOutput{}, fmt.Errorf("create AI application %q: request outcome is unknown; use list_ai_applications before retrying: %w", in.Name, mapped)
		}
		return submitAIApplicationOutput{}, fmt.Errorf("create AI application %q: %w", in.Name, mapped)
	}
	if created == nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNoContent {
			return out, nil
		}
		return submitAIApplicationOutput{}, fmt.Errorf("create AI application %q: nil response; creation outcome is unknown, use list_ai_applications before retrying: %w", in.Name, runaimcp.ErrUpstream)
	}
	out.ID = created.GetId()
	if created.GetName() != "" {
		out.Name = created.GetName()
	}
	if created.GetProjectId() != "" {
		out.ProjectID = created.GetProjectId()
	}
	out.DesiredPhase = created.GetDesiredPhase()
	out.Phase = created.GetPhase()
	out.ActionStatus = created.GetActionStatus()
	out.ActionMessage = created.GetActionMessage()
	return out, nil
}

func toAIApplicationCreationRequest(in submitAIApplicationInput) workloads.AIApplicationCreationRequest {
	ref := workloads.AIApplicationSpecChartRef{}
	if in.Spec.ChartRef.Repo != "" {
		ref.SetRepo(in.Spec.ChartRef.Repo)
	}
	if in.Spec.ChartRef.ChartName != "" {
		ref.SetChartName(in.Spec.ChartRef.ChartName)
	}
	if in.Spec.ChartRef.ChartVersion != "" {
		ref.SetChartVersion(in.Spec.ChartRef.ChartVersion)
	}
	if in.Spec.ChartRef.ChartURL != "" {
		ref.SetChartURL(in.Spec.ChartRef.ChartURL)
	}
	if in.Spec.ChartRef.ChartID != "" {
		ref.SetChartId(in.Spec.ChartRef.ChartID)
	}
	spec := workloads.AIApplicationSpec{ChartRef: ref}
	if in.Spec.Values != nil {
		spec.SetValues(in.Spec.Values)
	}
	if in.Spec.ValuesBase64 != nil {
		spec.SetValuesBase64(*in.Spec.ValuesBase64)
	}
	if in.Spec.Auth != nil {
		auth := workloads.AIApplicationAuth{}
		auth.SetAssetId(in.Spec.Auth.AssetID)
		if in.Spec.Auth.UserCredential {
			auth.SetUserCredential(true)
		}
		spec.SetAuth(auth)
	}
	return workloads.AIApplicationCreationRequest{ClusterId: in.ClusterID, ProjectId: in.ProjectID, Name: in.Name, Spec: spec}
}

func aiApplicationConfirmationDetails(in submitAIApplicationInput) map[string]any {
	details := map[string]any{
		"clusterId": in.ClusterID,
		"projectId": in.ProjectID,
	}
	if in.Spec.ChartRef.Repo != "" {
		details["chartSource"] = "repository"
		details["chartName"] = in.Spec.ChartRef.ChartName
		details["chartVersion"] = in.Spec.ChartRef.ChartVersion
	} else if in.Spec.ChartRef.ChartURL != "" {
		details["chartSource"] = "url"
	} else {
		details["chartSource"] = "uploadedChart"
		details["chartId"] = in.Spec.ChartRef.ChartID
	}
	return details
}
