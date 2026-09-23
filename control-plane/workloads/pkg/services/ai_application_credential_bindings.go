/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	assetsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/helmvalues"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
)

func processAIApplicationCredentialBindings(ctx context.Context, request *api.AiApplicationCreationRequest, departmentID string) (bool, *api.Error) {
	spec := &request.Spec
	if len(spec.CredentialBindings) == 0 {
		return false, nil
	}
	variables := make([]string, 0, len(spec.CredentialBindings))
	for _, binding := range spec.CredentialBindings {
		variables = append(variables, binding.Variable)
		if _, err := uuid.Parse(binding.AssetId); err != nil {
			return false, api.NewValidationError("credential binding assetId must be a valid UUID")
		}
	}
	if err := helmvalues.ValidatePaths(variables); err != nil {
		return false, api.NewValidationError(err.Error())
	}
	if capErr := services_utils.ValidateClusterSupportsCapability(ctx, request.ClusterId, checker.MinVersionOfApplicationCredentialBindings); capErr != nil {
		return false, api.NewError(int(capErr.GetCode()), capErr.GetMessage(), nil)
	}
	if spec.ValuesBase64 != "" {
		if spec.Values != nil {
			return false, api.NewValidationError("spec.values and spec.valuesBase64 are mutually exclusive")
		}
		decoded, err := base64.StdEncoding.DecodeString(spec.ValuesBase64)
		if err != nil {
			return false, api.NewValidationError("spec.valuesBase64 must be valid base64")
		}
		if err := yaml.Unmarshal(decoded, &spec.Values); err != nil || spec.Values == nil {
			return false, api.NewValidationError("spec.valuesBase64 with credential bindings must be a YAML object")
		}
		spec.ValuesBase64 = ""
	}

	// Both lists are authorized by assets-service; the personal list is restricted to the caller.
	credentials := make(map[string]assetsApi.CredentialsListResponseEntry)
	for _, privateOnly := range []bool{false, true} {
		response, clientErr := assetsClient.Get().ListCredentialsAssets(ctx, privateOnly)
		if clientErr != nil {
			return false, api.NewError(int(clientErr.Code), "failed to validate selected credentials", nil).WithDetails(clientErr)
		}
		if response == nil {
			return false, api.NewError(503, "failed to validate selected credentials", nil)
		}
		for _, credential := range response.Entries {
			credentials[credential.Meta.Id] = credential
		}
	}
	usesPersonal := false
	for i := range spec.CredentialBindings {
		binding := &spec.CredentialBindings[i]
		credential, found := credentials[binding.AssetId]
		if !found || credential.Meta.DeletedAt != nil {
			return false, api.NewValidationError(fmt.Sprintf("Selected credential %q no longer exists or is not accessible. Please reselect.", binding.AssetId))
		}
		meta := credential.Meta
		if !credentialMatchesProject(meta, request.ProjectId, departmentID, request.ClusterId) {
			return false, api.NewValidationError(fmt.Sprintf("Selected credential %q is not available in the target project. Please reselect.", binding.AssetId))
		}
		personal := meta.GetPrivate()
		if personal && meta.CreatedBy != convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "") {
			return false, api.NewValidationError("Selected personal credential is not accessible. Please reselect.")
		}
		if !credentialKindCompatible(string(binding.CredentialType), meta.Kind) {
			return false, api.NewValidationError(fmt.Sprintf("Selected credential %q is incompatible with %q at %q. Please reselect.", binding.AssetId, binding.CredentialType, binding.Variable))
		}
		if personal {
			representation := personalCredentialRepresentation("")
			if meta.Kind == assetsApi.ASSETKIND_NGC_API_KEY {
				representation = personalCredentialRepresentationGeneric
				if string(binding.CredentialType) == "image-pull-secret" {
					representation = personalCredentialRepresentationDocker
				}
			}
			sourceName, err := canonicalPersonalCredentialSource(credential, representation)
			if err != nil {
				return false, err
			}
			if err := validatePersonalCredentialDestinationName(request.Name, sourceName); err != nil {
				return false, err
			}
		}
		// Never trust a caller-supplied personal/shared selector.
		binding.UserCredential = personal
		binding.ExistingSecretName = ""
		if !personal {
			switch meta.Kind {
			case assetsApi.ASSETKIND_GENERIC_SECRET:
				generic := credential.Spec.GetGenericSecret()
				binding.ExistingSecretName = generic.GetExistingSecretName()
			case assetsApi.ASSETKIND_DOCKER_REGISTRY:
				registry := credential.Spec.GetDockerRegistry()
				binding.ExistingSecretName = registry.GetExistingSecretName()
			}
		}
		usesPersonal = usesPersonal || personal
	}
	// These destinations are owned by the bindings. Drop overridden literals/placeholders before
	// persistence and legacy personal-credential processing; final names are written after chart loading.
	for _, variable := range variables {
		path, err := helmvalues.Parse(variable)
		if err != nil {
			return false, api.NewValidationError(err.Error())
		}
		paths := []helmvalues.Path{path}
		if flag := path.CreateFlag(); flag != "" {
			companion, err := helmvalues.Parse(flag)
			if err != nil {
				return false, api.NewValidationError(err.Error())
			}
			paths = append(paths, companion)
		}
		for _, owned := range paths {
			if _, err := owned.Get(spec.Values); err == nil {
				if err := owned.Set(spec.Values, nil); err != nil {
					return false, api.NewValidationError(err.Error())
				}
			}
		}
	}
	return usesPersonal, nil
}

func credentialKindCompatible(credentialType string, kind assetsApi.AssetKind) bool {
	switch credentialType {
	case "ngc-api-key":
		return kind == assetsApi.ASSETKIND_NGC_API_KEY
	case "image-pull-secret":
		return kind == assetsApi.ASSETKIND_DOCKER_REGISTRY || kind == assetsApi.ASSETKIND_NGC_API_KEY
	case "generic":
		return kind == assetsApi.ASSETKIND_GENERIC_SECRET || kind == assetsApi.ASSETKIND_NGC_API_KEY
	default:
		return false
	}
}

func credentialMatchesProject(meta assetsApi.AssetMeta, projectID, departmentID, clusterID string) bool {
	switch meta.Scope {
	case assetsApi.SCOPE_TENANT:
		return true
	case assetsApi.SCOPE_CLUSTER:
		return meta.GetClusterId() == clusterID
	case assetsApi.SCOPE_DEPARTMENT:
		return meta.GetClusterId() == clusterID && meta.GetDepartmentId() == departmentID
	case assetsApi.SCOPE_PROJECT:
		return meta.GetClusterId() == clusterID && fmt.Sprint(meta.GetProjectId()) == projectID
	default:
		return false
	}
}
