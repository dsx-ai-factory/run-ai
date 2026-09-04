/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/client"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	workloadmanager "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

// SubmitOptions contains all the options for submitting a workload
type SubmitOptions struct {
	Name           string
	File           string
	Project        string
	Priority       string
	Category       string
	NamePrefix     string
	Preemptibility workloadmanager.NullablePreemptibility
	Mnnvl          workloadmanager.NullableMnnvlMode
}

// Submit handles the workload submission workflow
func Submit(
	txn *transaction.Txn,
	instance *config.Instance,
	options *SubmitOptions,
) (string, string, error) {
	var projectId string
	manifestData, err := ReadAndParseManifest(options.File)
	if err != nil {
		return "", "", err
	}

	projectName, err := resolveProject(options.Project, manifestData.NamespaceFromYAML, instance.GetConfigValue(config.ClusterProjectName))
	if err != nil {
		return "", "", err
	}

	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	token := instance.GetToken()
	clusterId := instance.GetConfigValue(config.ClusterUUID)

	// empty project id means we use the namespace in the yaml.
	if projectName != "" {
		projectId, err = project.GetProjectId(txn, cpUrl, token, clusterId, projectName)
		if err != nil {
			return "", "", fmt.Errorf(errorFailedToGetProjectID, err)
		}
	}
	workloadName, useGivenNameAsPrefix := determineNameAndPrefix(
		options.Name,
		manifestData.NameFromYAML,
		options.NamePrefix,
		manifestData.KindFromYAML,
	)

	manifestJSON := manifestData.ParsedManifest

	workloadClient, err := client.NewWorkloadV2(cpUrl, token)
	if err != nil {
		return "", "", fmt.Errorf(errorFailedToCreateClient, err)
	}

	submitResponse, err := workloadClient.Submit(
		txn.GetCtx(),
		workloadName,
		useGivenNameAsPrefix,
		projectId,
		options.Priority,
		options.Category,
		manifestJSON,
		clusterId,
		options.Preemptibility,
		options.Mnnvl,
	)
	if err != nil {
		return "", "", fmt.Errorf(errorFailedToSubmitWorkload, err)
	}

	return submitResponse.Name, manifestData.KindFromYAML, nil
}

// determineNameAndPrefix implements the name handling logic
// Priority: --name-prefix or [NAME] argument > YAML metadata.name > default "workload kind"
func determineNameAndPrefix(nameArg string, nameFromYAML string, namePrefixFlag string, KindFromYAML string) (string, bool) {
	if namePrefixFlag != "" {
		return namePrefixFlag, true
	}

	if nameArg != "" {
		return nameArg, false
	}

	if nameFromYAML != "" {
		return "", false
	}

	return KindFromYAML, true
}
