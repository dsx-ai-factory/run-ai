/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"fmt"
	"strings"

	assets "github.com/run-ai/runai/runai-clients/assets-service-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type credentialKeyValue struct {
	Key   string `json:"key"   jsonschema:"Key stored in the generic secret."`
	Value string `json:"value" jsonschema:"Secret value. It is sent upstream but is never returned by this tool."`
}

type credentialSpec struct {
	ExistingSecretName string               `json:"existingSecretName,omitempty" jsonschema:"Shared only: existing Kubernetes Secret to import instead of values. Mutually exclusive with all value fields."`
	AccessKeyID        string               `json:"accessKeyId,omitempty"         jsonschema:"Access-key credential value. Required with secretAccessKey when credentialType=accessKey and no existingSecretName is used."`
	SecretAccessKey    string               `json:"secretAccessKey,omitempty"     jsonschema:"Secret access-key value. Never returned by this tool."`
	User               string               `json:"user,omitempty"                jsonschema:"User name for password or dockerRegistry credentials."`
	Password           string               `json:"password,omitempty"            jsonschema:"Password for password or dockerRegistry credentials. Never returned by this tool."`
	URL                string               `json:"url,omitempty"                 jsonschema:"Registry URL. Required for dockerRegistry when no existingSecretName is used."`
	KeyValuePairs      []credentialKeyValue `json:"keyValuePairs,omitempty"       jsonschema:"Key/value entries for genericSecret. Values are never returned by this tool."`
	NGCAPIKey          string               `json:"ngcApiKey,omitempty"           jsonschema:"NGC API key for ngcApiKey credentials. Never returned by this tool."`
}

func (in manageCredentialsInput) Validate() error {
	action := common.WriteAction(in.Action)
	if !action.Valid() {
		return fmt.Errorf("action must be one of: create, update, delete")
	}
	adapter, err := credentialAdapterByType(in.CredentialType)
	if err != nil {
		return err
	}
	if in.Private && action != common.ActionCreate {
		return fmt.Errorf("private is create-only; update and delete derive it from credentialId")
	}
	if in.Private && !adapter.userCredentialAllowed() {
		return fmt.Errorf("user credentials support credentialType dockerRegistry, genericSecret, or ngcApiKey")
	}
	if in.CredentialID != "" {
		if err := common.ValidateUUID("credentialId", in.CredentialID); err != nil {
			return err
		}
	}
	if in.IgnoreMissing && action != common.ActionDelete {
		return fmt.Errorf("ignoreMissing is only valid with action=delete")
	}

	switch action {
	case common.ActionCreate:
		return in.validateCreate()
	case common.ActionUpdate:
		return in.validateUpdate()
	case common.ActionDelete:
		return in.validateDelete()
	default:
		return fmt.Errorf("unsupported action %q", in.Action)
	}
}

func (in manageCredentialsInput) validateCreate() error {
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("name is required for create")
	}
	if in.CredentialID != "" {
		return fmt.Errorf("credentialId must not be set for create")
	}
	if in.Spec == nil {
		return fmt.Errorf("spec is required for create")
	}
	if err := in.validateCreateScope(); err != nil {
		return err
	}
	return validateCredentialSpec(in.Private, in.CredentialType, in.Spec)
}

func (in manageCredentialsInput) validateUpdate() error {
	if in.CredentialID == "" {
		return fmt.Errorf("credentialId is required for update")
	}
	if in.Name != "" {
		return fmt.Errorf("name must not be set for update")
	}
	if in.Description == nil {
		return fmt.Errorf("description is required for update; secret values cannot be updated")
	}
	if in.Spec != nil {
		return fmt.Errorf("spec is create-only; upstream updates support description metadata only")
	}
	return in.validateNoScopeFilters()
}

func (in manageCredentialsInput) validateDelete() error {
	if in.CredentialID == "" {
		return fmt.Errorf("credentialId is required for delete")
	}
	if in.Name != "" {
		return fmt.Errorf("name must not be set for delete")
	}
	if in.Description != nil || in.Spec != nil {
		return fmt.Errorf("description and spec must not be set for delete")
	}
	return in.validateNoScopeFilters()
}

func (in manageCredentialsInput) validateCreateScope() error {
	if in.Private {
		if in.Scope != "" && in.Scope != string(assets.SCOPE_TENANT) {
			return fmt.Errorf("user credentials always use scope=tenant")
		}
		if in.ClusterID != "" || in.DepartmentID != "" || in.ProjectID != nil {
			return fmt.Errorf("user credentials must not set clusterId, departmentId, or projectId")
		}
		return nil
	}

	if in.Scope == string(assets.SCOPE_TENANT) &&
		(in.CredentialType == kindGenericSecret || in.CredentialType == kindNGCAPIKey) {
		return fmt.Errorf("shared %s credentials do not support scope=tenant", in.CredentialType)
	}

	switch assets.Scope(in.Scope) {
	case assets.SCOPE_TENANT:
		if in.ClusterID != "" || in.DepartmentID != "" || in.ProjectID != nil {
			return fmt.Errorf("tenant-scoped credentials must not set clusterId, departmentId, or projectId")
		}
	case assets.SCOPE_CLUSTER:
		if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
			return err
		}
		if in.DepartmentID != "" || in.ProjectID != nil {
			return fmt.Errorf("cluster-scoped credentials must not set departmentId or projectId")
		}
	case assets.SCOPE_DEPARTMENT:
		if strings.TrimSpace(in.DepartmentID) == "" {
			return fmt.Errorf("departmentId is required for scope=department")
		}
		if in.ClusterID != "" {
			return fmt.Errorf("department-scoped credentials must not set clusterId; the API derives it from departmentId")
		}
		if in.ProjectID != nil {
			return fmt.Errorf("department-scoped credentials must not set projectId")
		}
	case assets.SCOPE_PROJECT:
		if in.ClusterID != "" {
			if err := common.ValidateUUID("clusterId", in.ClusterID); err != nil {
				return err
			}
		}
		if in.ProjectID == nil {
			return fmt.Errorf("projectId is required for scope=project")
		}
		if in.DepartmentID != "" {
			return fmt.Errorf("project-scoped credentials must not set departmentId")
		}
	default:
		return fmt.Errorf("scope must be one of: tenant, cluster, department, project")
	}
	return nil
}

func (in manageCredentialsInput) validateNoScopeFilters() error {
	if in.Scope != "" || in.ClusterID != "" || in.DepartmentID != "" || in.ProjectID != nil {
		return fmt.Errorf("scope, clusterId, departmentId, and projectId are create-only")
	}
	return nil
}

func validateCredentialSpec(private bool, kind string, spec *credentialSpec) error {
	hasExisting := strings.TrimSpace(spec.ExistingSecretName) != ""
	if private && hasExisting {
		return fmt.Errorf("spec.existingSecretName is not supported for user credentials")
	}
	if hasExisting && hasCredentialValues(spec) {
		return fmt.Errorf("spec.existingSecretName is mutually exclusive with credential value fields")
	}

	switch kind {
	case kindAccessKey:
		if !hasExisting {
			if err := requireCredentialFields(kind,
				requiredCredentialField{"accessKeyId", spec.AccessKeyID},
				requiredCredentialField{"secretAccessKey", spec.SecretAccessKey},
			); err != nil {
				return err
			}
		}
	case kindPassword:
		if !hasExisting {
			if err := requireCredentialFields(kind,
				requiredCredentialField{"user", spec.User},
				requiredCredentialField{"password", spec.Password},
			); err != nil {
				return err
			}
		}
	case kindDockerRegistry:
		if !hasExisting {
			if err := requireCredentialFields(kind,
				requiredCredentialField{"user", spec.User},
				requiredCredentialField{"password", spec.Password},
				requiredCredentialField{"url", spec.URL},
			); err != nil {
				return err
			}
		}
	case kindGenericSecret:
		if err := validateGenericSecretPairs(hasExisting, spec.KeyValuePairs); err != nil {
			return err
		}
	case kindNGCAPIKey:
		if spec.NGCAPIKey == "" {
			return fmt.Errorf("spec.ngcApiKey is required for ngcApiKey")
		}
	}
	return rejectFieldsForOtherKinds(kind, spec)
}

type requiredCredentialField struct {
	name  string
	value string
}

func requireCredentialFields(kind string, fields ...requiredCredentialField) error {
	missing := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.value == "" {
			missing = append(missing, "spec."+field.name)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	if len(missing) == 1 {
		return fmt.Errorf("%s is required for %s", missing[0], kind)
	}
	return fmt.Errorf("%s and %s are required for %s",
		strings.Join(missing[:len(missing)-1], ", "), missing[len(missing)-1], kind)
}

func hasCredentialValues(spec *credentialSpec) bool {
	return spec.AccessKeyID != "" ||
		spec.SecretAccessKey != "" ||
		spec.User != "" ||
		spec.Password != "" ||
		spec.URL != "" ||
		len(spec.KeyValuePairs) > 0 ||
		spec.NGCAPIKey != ""
}

func validateGenericSecretPairs(hasExisting bool, pairs []credentialKeyValue) error {
	if !hasExisting && len(pairs) == 0 {
		return fmt.Errorf("spec.keyValuePairs must contain at least one entry for genericSecret")
	}
	for i, pair := range pairs {
		if strings.TrimSpace(pair.Key) == "" {
			return fmt.Errorf("spec.keyValuePairs[%d].key is required", i)
		}
	}
	return nil
}

func rejectFieldsForOtherKinds(kind string, spec *credentialSpec) error {
	allowed := map[string]bool{"existingSecretName": kind != kindNGCAPIKey}
	switch kind {
	case kindAccessKey:
		allowed["accessKeyId"], allowed["secretAccessKey"] = true, true
	case kindPassword:
		allowed["user"], allowed["password"] = true, true
	case kindDockerRegistry:
		allowed["user"], allowed["password"], allowed["url"] = true, true, true
	case kindGenericSecret:
		allowed["keyValuePairs"] = true
	case kindNGCAPIKey:
		allowed["ngcApiKey"] = true
	}

	for field, present := range credentialSpecFields(spec) {
		if present && !allowed[field] {
			return fmt.Errorf("spec.%s is not valid for credentialType=%s", field, kind)
		}
	}
	return nil
}

func credentialSpecFields(spec *credentialSpec) map[string]bool {
	return map[string]bool{
		"existingSecretName": spec.ExistingSecretName != "",
		"accessKeyId":        spec.AccessKeyID != "",
		"secretAccessKey":    spec.SecretAccessKey != "",
		"user":               spec.User != "",
		"password":           spec.Password != "",
		"url":                spec.URL != "",
		"keyValuePairs":      len(spec.KeyValuePairs) > 0,
		"ngcApiKey":          spec.NGCAPIKey != "",
	}
}
