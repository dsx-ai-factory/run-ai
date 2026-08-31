/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	workloadsv2 "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"gopkg.in/yaml.v3"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type yamlWorkloadSubmission struct {
	Metadata     map[string]any `json:"metadata"     jsonschema:"Complete Workloads V2 creation metadata. Supports name, useGivenNameAsPrefix, projectId or clusterId, priority, category, preemptibility, and configuration."`
	ManifestYAML string         `json:"manifestYAML" jsonschema:"Full Kubernetes workload manifest as YAML text. Pass the content itself, not a file path."`
}

const maxEncodedYAMLManifestLength = 2 * 1024 * 1024

type yamlManifestIdentity struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name         string `yaml:"name"`
		GenerateName string `yaml:"generateName"`
		Namespace    string `yaml:"namespace"`
	} `yaml:"metadata"`
}

func submitYAMLWorkload(ctx context.Context, client *runaiclient.Client, in submitWorkloadInput) (submitWorkloadOutput, error) {
	metadata, manifest, err := decodeYAMLSubmission(in.YAML)
	if err != nil {
		return submitWorkloadOutput{}, err
	}
	out := newYAMLSubmissionOutput(metadata, manifest, in.DryRun)
	project, err := resolveYAMLSubmissionProject(ctx, client, out.ProjectID, out.ClusterID, manifest.Metadata.Namespace)
	if err != nil {
		return submitWorkloadOutput{}, err
	}
	out.ProjectID = project.Id
	out.ClusterID = project.ClusterId
	if !in.DryRun {
		if err := confirmSubmission(ctx, in.Confirm, out, submitSourceYAML); err != nil {
			return out, err
		}
	}
	created, err := createYAMLWorkload(ctx, client, metadata, in.YAML.ManifestYAML, in.DryRun)
	if err != nil {
		return submitWorkloadOutput{}, err
	}
	return applyYAMLSubmissionResult(out, created, in.DryRun), nil
}

func newYAMLSubmissionOutput(metadata workloadsv2.WorkloadV2MetadataCreateParams, manifest yamlManifestIdentity, dryRun bool) submitWorkloadOutput {
	name := metadata.GetName()
	nameIsPrefix := name != "" && metadata.GetUseGivenNameAsPrefix()
	if name == "" {
		name = manifest.Metadata.Name
	}
	if name == "" {
		name = manifest.Metadata.GenerateName
		nameIsPrefix = name != ""
	}
	return submitWorkloadOutput{
		DryRun:       dryRun,
		Name:         name,
		WorkloadType: manifest.Kind,
		ProjectID:    metadata.GetProjectId(),
		ClusterID:    metadata.GetClusterId(),
		NameIsPrefix: nameIsPrefix,
	}
}

func createYAMLWorkload(ctx context.Context, client *runaiclient.Client, metadata workloadsv2.WorkloadV2MetadataCreateParams, manifest string, dryRun bool) (*workloadsv2.WorkloadV2, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(manifest))
	req := workloadsv2.NewWorkloadV2CreateRequest(metadata)
	req.SetManifestBase64Encoded(encoded)
	created, resp, err := client.WorkloadsV2.WorkloadsV2API.CreateWorkloadV2(ctx).
		WorkloadV2CreateRequest(*req).
		DryRun(dryRun).
		Execute()
	if err != nil {
		// Some Workloads V2 versions return desiredPhase:"" for a successful
		// dry-run. The generated enum rejects that value after it has already
		// decoded the rest of the workload. Accept only this known, non-mutating
		// 2xx compatibility case; all other decode failures remain errors.
		if dryRun && resp != nil && resp.StatusCode < 300 && created != nil &&
			strings.Contains(err.Error(), "is not a valid DesiredPhase") {
			return created, nil
		}
		mapped := runaimcp.MapHTTPError(resp, err)
		if dryRun {
			return nil, fmt.Errorf("dry-run YAML workload: %w", mapped)
		}
		if resp == nil || resp.StatusCode < 300 {
			return nil, fmt.Errorf("create YAML workload: request outcome is unknown; inspect workloads before retrying: %w", mapped)
		}
		return nil, fmt.Errorf("create YAML workload: %w", mapped)
	}
	if created == nil {
		if dryRun {
			return nil, fmt.Errorf("dry-run YAML workload: nil response: %w", runaimcp.ErrUpstream)
		}
		return nil, fmt.Errorf("create YAML workload: nil response; creation outcome is unknown, inspect workloads before retrying: %w", runaimcp.ErrUpstream)
	}
	return created, nil
}

func applyYAMLSubmissionResult(out submitWorkloadOutput, created *workloadsv2.WorkloadV2, dryRun bool) submitWorkloadOutput {
	if !dryRun {
		out.ID = created.Metadata.Id
	}
	if created.Metadata.Name != "" {
		out.Name = created.Metadata.Name
	}
	if created.Metadata.ProjectId != "" {
		out.ProjectID = created.Metadata.ProjectId
	}
	if created.Metadata.ClusterId != "" {
		out.ClusterID = created.Metadata.ClusterId
	}
	if created.Metadata.Gvk.Kind != "" {
		out.WorkloadType = created.Metadata.Gvk.Kind
	}
	return out
}

func confirmSubmission(ctx context.Context, confirmed bool, out submitWorkloadOutput, source string) error {
	name := out.Name
	if name == "" {
		name = out.WorkloadType
	}
	details := map[string]any{
		"source":       source,
		"workloadType": out.WorkloadType,
		"projectId":    out.ProjectID,
		"clusterId":    out.ClusterID,
	}
	if out.NameIsPrefix {
		details["nameIsPrefix"] = true
	}
	return common.ConfirmWrite(ctx, confirmed, "create", "workload", name, "", details)
}

func decodeYAMLSubmission(in *yamlWorkloadSubmission) (workloadsv2.WorkloadV2MetadataCreateParams, yamlManifestIdentity, error) {
	if in == nil {
		return workloadsv2.WorkloadV2MetadataCreateParams{}, yamlManifestIdentity{}, fmt.Errorf("yaml is required: %w", runaimcp.ErrValidation)
	}
	if base64.StdEncoding.EncodedLen(len(in.ManifestYAML)) > maxEncodedYAMLManifestLength {
		return workloadsv2.WorkloadV2MetadataCreateParams{}, yamlManifestIdentity{}, fmt.Errorf(
			"yaml.manifestYAML exceeds the API maximum encoded size of %d bytes: %w",
			maxEncodedYAMLManifestLength, runaimcp.ErrValidation)
	}
	metadata, err := decodeYAMLMetadata(in.Metadata)
	if err != nil {
		return metadata, yamlManifestIdentity{}, err
	}
	manifest, err := decodeYAMLManifest(in.ManifestYAML)
	return metadata, manifest, err
}

func decodeYAMLMetadata(raw map[string]any) (workloadsv2.WorkloadV2MetadataCreateParams, error) {
	var metadata workloadsv2.WorkloadV2MetadataCreateParams
	if err := common.DecodeTypedRequest(raw, &metadata); err != nil {
		return metadata, fmt.Errorf("decode yaml.metadata: %v: %w", err, runaimcp.ErrValidation)
	}
	return metadata, nil
}

func decodeYAMLManifest(raw string) (yamlManifestIdentity, error) {
	var manifest yamlManifestIdentity
	yamlDecoder := yaml.NewDecoder(strings.NewReader(raw))
	if err := yamlDecoder.Decode(&manifest); err != nil {
		return manifest, fmt.Errorf("decode yaml.manifestYAML: %v: %w", err, runaimcp.ErrValidation)
	}
	var extraDocument any
	if err := yamlDecoder.Decode(&extraDocument); err != io.EOF {
		if err != nil {
			return manifest, fmt.Errorf("decode yaml.manifestYAML: %v: %w", err, runaimcp.ErrValidation)
		}
		return manifest, fmt.Errorf("yaml.manifestYAML must contain exactly one YAML document: %w", runaimcp.ErrValidation)
	}
	return manifest, nil
}

func validateYAMLSubmissionInput(in submitWorkloadInput) error {
	if in.YAML == nil {
		return fmt.Errorf("yaml is required when source=yaml")
	}
	if in.Native != nil {
		return fmt.Errorf("native must not be set when source=yaml")
	}
	if strings.TrimSpace(in.YAML.ManifestYAML) == "" {
		return fmt.Errorf("yaml.manifestYAML is required")
	}

	metadata, manifest, err := decodeYAMLSubmission(in.YAML)
	if err != nil {
		return err
	}
	if strings.TrimSpace(manifest.APIVersion) == "" || strings.TrimSpace(manifest.Kind) == "" {
		return fmt.Errorf("yaml.manifestYAML must contain apiVersion and kind: %w", runaimcp.ErrValidation)
	}
	return validateYAMLSubmissionScope(metadata, manifest.Metadata.Namespace)
}

func validateYAMLSubmissionScope(metadata workloadsv2.WorkloadV2MetadataCreateParams, namespace string) error {
	projectID, clusterID := metadata.GetProjectId(), metadata.GetClusterId()
	hasProject := metadata.HasProjectId()
	hasCluster := metadata.HasClusterId()
	switch {
	case !hasProject && !hasCluster:
		return fmt.Errorf("yaml.metadata must contain projectId or clusterId: %w", runaimcp.ErrValidation)
	case hasProject && hasCluster:
		return fmt.Errorf("yaml.metadata.projectId and yaml.metadata.clusterId are mutually exclusive: %w", runaimcp.ErrValidation)
	}
	if hasProject {
		return common.ValidateNumericString("yaml.metadata.projectId", projectID)
	}
	if strings.TrimSpace(namespace) == "" {
		return fmt.Errorf("yaml.manifestYAML metadata.namespace is required when yaml.metadata uses clusterId: %w", runaimcp.ErrValidation)
	}
	return common.ValidateUUID("yaml.metadata.clusterId", clusterID)
}
