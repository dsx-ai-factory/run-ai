/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	workloadsComposer "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	assetsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	userCredentialPlaceholderPrefix     = workloadsComposer.UserCredentialsPrefix + ":"
	redactedPersonalCredentialReference = "<redacted>"
)

var userCredentialPlaceholderPattern = regexp.MustCompile(`^USERCREDENTIAL:([^,]+)(?:,(generic-secret|docker-registry))?$`)

type personalCredentialRepresentation string

const (
	personalCredentialRepresentationGeneric personalCredentialRepresentation = "generic-secret"
	personalCredentialRepresentationDocker  personalCredentialRepresentation = "docker-registry"
)

func processAIApplicationCredentials(ctx context.Context, applicationName string, spec *api.AiApplicationSpec) (bool, *api.Error) {
	usesPersonalChartAuth := isPersonalAIApplicationAuth(spec.Auth)
	if usesPersonalChartAuth && spec.Auth.AssetId == "" {
		return false, api.NewValidationError("spec.auth.assetId is required when spec.auth.userCredential is true.")
	}

	values, validationErr := personalCredentialValues(spec)
	if validationErr != nil {
		return false, validationErr
	}

	placeholders, validationErr := collectPersonalCredentialPlaceholders(values)
	if validationErr != nil {
		return false, validationErr
	}
	usesPersonalCredentials := usesPersonalChartAuth || len(placeholders) > 0
	if !usesPersonalCredentials {
		return false, nil
	}

	credentials, clientErr := assetsClient.Get().ListCredentialsAssets(ctx, true)
	if clientErr != nil {
		return false, api.NewError(int(clientErr.Code), "failed to validate personal credentials", nil).WithDetails(clientErr)
	}

	if usesPersonalChartAuth {
		if validationErr := validatePersonalChartAuthOwnership(spec.Auth.AssetId, credentials.Entries); validationErr != nil {
			return false, validationErr
		}
	}
	if len(placeholders) == 0 {
		return true, nil
	}

	resolvedValues, validationErr := resolvePersonalCredentialPlaceholders(applicationName, values, credentials.Entries)
	if validationErr != nil {
		return false, validationErr
	}
	spec.Values = resolvedValues
	spec.ValuesBase64 = ""
	return true, nil
}

func personalCredentialValues(spec *api.AiApplicationSpec) (map[string]interface{}, *api.Error) {
	if spec.Values != nil || spec.ValuesBase64 == "" {
		return spec.Values, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(spec.ValuesBase64)
	if err != nil {
		return nil, api.NewValidationError("spec.valuesBase64 must be valid base64")
	}
	if !bytes.Contains(decoded, []byte(userCredentialPlaceholderPrefix)) {
		return nil, nil
	}

	values := make(map[string]interface{})
	if err := yaml.Unmarshal(decoded, &values); err != nil {
		return nil, api.NewValidationError("spec.valuesBase64 containing personal credentials must be a valid YAML object.")
	}
	return values, nil
}

// collectPersonalCredentialPlaceholders recursively walks parsed Helm values,
// collecting whole-value USERCREDENTIAL references and rejecting malformed or embedded references.
func collectPersonalCredentialPlaceholders(value interface{}) ([]string, *api.Error) {
	var placeholders []string
	var visit func(interface{}) *api.Error
	visit = func(current interface{}) *api.Error {
		switch typed := current.(type) {
		case map[string]interface{}:
			for _, child := range typed {
				if err := visit(child); err != nil {
					return err
				}
			}
		case []interface{}:
			for _, child := range typed {
				if err := visit(child); err != nil {
					return err
				}
			}
		case string:
			if !strings.Contains(typed, userCredentialPlaceholderPrefix) {
				return nil
			}
			if !userCredentialPlaceholderPattern.MatchString(typed) {
				return api.NewValidationError("personal credential placeholders must occupy the complete value and use the supported USERCREDENTIAL grammar")
			}
			placeholders = append(placeholders, typed)
		}
		return nil
	}
	if err := visit(value); err != nil {
		return nil, err
	}
	return placeholders, nil
}

func resolvePersonalCredentialPlaceholders(applicationName string, values map[string]interface{}, credentials []assetsApi.CredentialsListResponseEntry) (map[string]interface{}, *api.Error) {
	credentialsByName := make(map[string][]assetsApi.CredentialsListResponseEntry, len(credentials))
	for _, credential := range credentials {
		name := credential.Meta.Name
		credentialsByName[name] = append(credentialsByName[name], credential)
	}
	resolved, err := resolvePersonalCredentialValue(applicationName, values, credentialsByName)
	if err != nil {
		return nil, err
	}
	if resolved == nil {
		return nil, nil
	}
	return resolved.(map[string]interface{}), nil
}

func resolvePersonalCredentialValue(applicationName string, value interface{}, credentialsByName map[string][]assetsApi.CredentialsListResponseEntry) (interface{}, *api.Error) {
	switch typed := value.(type) {
	case map[string]interface{}:
		resolved := make(map[string]interface{}, len(typed))
		for key, child := range typed {
			next, err := resolvePersonalCredentialValue(applicationName, child, credentialsByName)
			if err != nil {
				return nil, err
			}
			resolved[key] = next
		}
		return resolved, nil
	case []interface{}:
		resolved := make([]interface{}, len(typed))
		for index, child := range typed {
			next, err := resolvePersonalCredentialValue(applicationName, child, credentialsByName)
			if err != nil {
				return nil, err
			}
			resolved[index] = next
		}
		return resolved, nil
	case string:
		if !strings.Contains(typed, userCredentialPlaceholderPrefix) {
			return typed, nil
		}
		matches := userCredentialPlaceholderPattern.FindStringSubmatch(typed)
		if matches == nil {
			return nil, api.NewValidationError("personal credential placeholders must occupy the complete value and use the supported USERCREDENTIAL grammar")
		}
		sourceName, err := resolvePersonalCredentialSource(matches[1], personalCredentialRepresentation(matches[2]), credentialsByName)
		if err != nil {
			return nil, err
		}
		if err := validatePersonalCredentialDestinationName(applicationName, sourceName); err != nil {
			return nil, err
		}
		return userCredentialPlaceholderPrefix + sourceName, nil
	default:
		return value, nil
	}
}

func resolvePersonalCredentialSource(name string, representation personalCredentialRepresentation, credentialsByName map[string][]assetsApi.CredentialsListResponseEntry) (string, *api.Error) {
	matches := credentialsByName[name]
	switch len(matches) {
	case 0:
		return "", api.NewValidationError(fmt.Sprintf("personal credential %q was not found", name))
	case 1:
	default:
		return "", api.NewValidationError(fmt.Sprintf("personal credential name %q is ambiguous", name))
	}

	return canonicalPersonalCredentialSource(matches[0], representation)
}

// canonicalPersonalCredentialSource returns the encrypted runai-namespace Secret name
// generated for a personal credential, selecting the requested NGC representation when needed.
func canonicalPersonalCredentialSource(credential assetsApi.CredentialsListResponseEntry, representation personalCredentialRepresentation) (string, *api.Error) {
	var ngcPrefix string
	switch credential.Meta.Kind {
	case assetsApi.ASSETKIND_GENERIC_SECRET, assetsApi.ASSETKIND_DOCKER_REGISTRY:
		if representation != "" {
			return "", api.NewValidationError("credential representation is supported only for NGC credentials")
		}
	case assetsApi.ASSETKIND_NGC_API_KEY:
		switch representation {
		case personalCredentialRepresentationGeneric:
			ngcPrefix = workloadsComposer.NgcApiKeyGenericSecretPrefix
		case personalCredentialRepresentationDocker:
			ngcPrefix = workloadsComposer.NgcApiKeyDockerRegistryPrefix
		default:
			return "", api.NewValidationError("NGC credential placeholders must specify generic-secret or docker-registry")
		}
	default:
		return "", api.NewValidationError(fmt.Sprintf(
			"personal credential kind %q is not supported for AI Application values",
			credential.Meta.Kind,
		))
	}

	assetID := credential.Meta.Id
	reference, buildErr := workloadsComposer.BuildUserCredentialName(
		credential.Meta.Name,
		&assetID,
		nil,
		workloadsApi.AssetKind(credential.Meta.Kind),
		ngcPrefix,
	)
	if buildErr != nil {
		return "", api.NewValidationError(fmt.Sprintf("personal credential %q has an invalid asset ID", credential.Meta.Name))
	}
	return strings.TrimPrefix(reference, userCredentialPlaceholderPrefix), nil
}

func validatePersonalChartAuthOwnership(assetID string, credentials []assetsApi.CredentialsListResponseEntry) *api.Error {
	matches := make([]assetsApi.CredentialsListResponseEntry, 0, 1)
	for _, credential := range credentials {
		if credential.Meta.Id == assetID {
			matches = append(matches, credential)
		}
	}
	if len(matches) != 1 {
		return api.NewValidationError("personal authentication credential is invalid")
	}
	return nil
}

func validatePersonalCredentialDestinationName(applicationName, sourceName string) *api.Error {
	destinationName := fmt.Sprintf("pc-%s-%s", applicationName, sourceName)
	if problems := validation.IsDNS1123Subdomain(destinationName); len(problems) != 0 {
		return api.NewValidationError(fmt.Sprintf("personal credential destination name is invalid: %s", strings.Join(problems, "; ")))
	}
	return nil
}

func validateApplicationPersonalCredentialsCapability(ctx context.Context, clusterID string) *api.Error {
	if capErr := services_utils.ValidateClusterSupportsCapability(ctx, clusterID, checker.MinVersionOfApplicationPersonalCredentials); capErr != nil {
		return api.NewError(int(capErr.GetCode()), capErr.GetMessage(), nil)
	}
	return nil
}

func isPersonalAIApplicationAuth(auth api.AiApplicationAuth) bool {
	return auth.UserCredential != nil && *auth.UserCredential
}

func sanitizeAIApplicationResponse(ctx context.Context, application *api.AiApplication) {
	if application == nil {
		return
	}
	currentUser := runaictx.GetString(ctx, runaictx.CfSubjectID)
	if currentUser != nil && *currentUser == application.CreatedBy {
		return
	}

	if isPersonalAIApplicationAuth(application.Spec.Auth) {
		application.Spec.Auth.AssetId = ""
	}
	if application.Spec.Values != nil {
		application.Spec.Values = convutils.MapMe(redactPersonalCredentialValue(application.Spec.Values))
	}
	if application.Spec.ValuesBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(application.Spec.ValuesBase64)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to decode AI Application valuesBase64 while sanitizing response: %s", err)
		} else if strings.Contains(string(decoded), userCredentialPlaceholderPrefix) {
			application.Spec.ValuesBase64 = ""
		}
	}
}

func redactPersonalCredentialValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		copy := make(map[string]interface{}, len(typed))
		for key, child := range typed {
			copy[key] = redactPersonalCredentialValue(child)
		}
		return copy
	case []interface{}:
		copy := make([]interface{}, len(typed))
		for index, child := range typed {
			copy[index] = redactPersonalCredentialValue(child)
		}
		return copy
	case string:
		if strings.HasPrefix(typed, userCredentialPlaceholderPrefix) {
			return redactedPersonalCredentialReference
		}
	}
	return value
}
