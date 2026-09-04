/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCredentialID      = "6d9a5d97-e66c-4106-9ea9-157481e599d2"
	testOtherCredentialID = "88c87cd3-29e6-4de7-9092-b566c2f7fc0f"
	testThirdCredentialID = "4c8e05d4-df7c-4af7-85ac-5e40aec247ce"
	testClusterID         = "d1405b3f-761a-4d4b-acb6-a00327ca853c"
)

func credentialResponse(id, name, kind string, private bool, scope, description string) map[string]any {
	meta := map[string]any{
		"id": id, "name": name, "kind": kind, "private": private, "scope": scope,
		"createdBy": "user", "updatedBy": "user",
		"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
	}
	if description != "" {
		meta["description"] = description
	}
	return map[string]any{"meta": meta, "spec": map[string]any{}}
}

func TestManageCredentialsInputValidate(t *testing.T) {
	description := "updated"
	baseSecret := &credentialSpec{KeyValuePairs: []credentialKeyValue{{Key: "TOKEN", Value: "secret"}}}
	cases := []struct {
		name    string
		in      manageCredentialsInput
		wantErr string
	}{
		{
			name: "shared_create",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "accessKey", Name: "s3",
				Scope: "cluster", ClusterID: testClusterID,
				Spec: &credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"},
			},
		},
		{
			name: "shared_create_from_existing_secret",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "accessKey", Name: "existing-s3",
				Scope: "cluster", ClusterID: testClusterID,
				Spec: &credentialSpec{ExistingSecretName: "s3-credentials"},
			},
		},
		{
			name: "department_create_without_cluster",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "accessKey", Name: "department",
				Scope: "department", DepartmentID: "department-a",
				Spec: &credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"},
			},
		},
		{
			name: "project_create_without_cluster",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "accessKey", Name: "project",
				Scope: "project", ProjectID: int32Ptr(1),
				Spec: &credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"},
			},
		},
		{
			name: "personal_create",
			in: manageCredentialsInput{
				Action: "create", Private: true, CredentialType: "genericSecret",
				Name: "mine", Spec: baseSecret,
			},
		},
		{
			name: "shared_tenant_generic_secret_rejected",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "genericSecret",
				Name: "shared-secret", Scope: "tenant", Spec: baseSecret,
			},
			wantErr: "do not support scope=tenant",
		},
		{
			name: "shared_tenant_ngc_api_key_rejected",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "ngcApiKey",
				Name: "shared-ngc", Scope: "tenant", Spec: &credentialSpec{NGCAPIKey: "secret"},
			},
			wantErr: "do not support scope=tenant",
		},
		{
			name: "update_description_only",
			in: manageCredentialsInput{
				Action: "update", CredentialType: "password",
				CredentialID: testCredentialID, Description: &description,
			},
		},
		{
			name: "personal_access_key_rejected",
			in: manageCredentialsInput{
				Action: "create", Private: true, CredentialType: "accessKey",
				Name: "bad", Spec: &credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"},
			},
			wantErr: "user credentials support",
		},
		{
			name: "shared_scope_required",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "genericSecret",
				Name: "bad", Spec: baseSecret,
			},
			wantErr: "scope must be one of",
		},
		{
			name: "existing_and_values_rejected",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "password",
				Name: "bad", Scope: "tenant",
				Spec: &credentialSpec{ExistingSecretName: "existing", User: "u", Password: "p"},
			},
			wantErr: "mutually exclusive",
		},
		{
			name: "department_cluster_rejected",
			in: manageCredentialsInput{
				Action: "create", CredentialType: "accessKey", Name: "department",
				Scope: "department", ClusterID: testClusterID, DepartmentID: "department-a",
				Spec: &credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"},
			},
			wantErr: "must not set clusterId",
		},
		{
			name: "generic_secret_defers_value_and_duplicate_policy_to_api",
			in: manageCredentialsInput{
				Action: "create", Private: true, CredentialType: "genericSecret", Name: "generic",
				Spec: &credentialSpec{KeyValuePairs: []credentialKeyValue{
					{Key: "same", Value: ""},
					{Key: "same", Value: "later"},
				}},
			},
		},
		{
			name: "foreign_spec_field_rejected",
			in: manageCredentialsInput{
				Action: "create", Private: true, CredentialType: "ngcApiKey",
				Name: "bad", Spec: &credentialSpec{NGCAPIKey: "key", Password: "not-valid-here"},
			},
			wantErr: "not valid for credentialType",
		},
		{
			name: "update_secret_values_rejected",
			in: manageCredentialsInput{
				Action: "update", CredentialType: "genericSecret",
				CredentialID: testCredentialID, Description: &description, Spec: baseSecret,
			},
			wantErr: "spec is create-only",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestManageCredentialsMissingMandatoryFields(t *testing.T) {
	description := "new"
	validAccessKey := &credentialSpec{AccessKeyID: "key", SecretAccessKey: "secret"}
	cases := []struct {
		name    string
		in      manageCredentialsInput
		wantErr string
	}{
		{"action", manageCredentialsInput{}, "action must be one of"},
		{"credential_type", manageCredentialsInput{
			Action: "create",
		}, "credentialType must be one of"},
		{"create_name", manageCredentialsInput{
			Action: "create", CredentialType: kindAccessKey,
			Scope: "tenant", Spec: validAccessKey,
		}, "name is required for create"},
		{"create_spec", manageCredentialsInput{
			Action: "create", CredentialType: kindAccessKey,
			Name: "missing-spec", Scope: "tenant",
		}, "spec is required for create"},
		{"access_key_values", manageCredentialsInput{
			Action: "create", CredentialType: kindAccessKey,
			Name: "access", Scope: "tenant", Spec: &credentialSpec{},
		}, "spec.accessKeyId and spec.secretAccessKey are required"},
		{"password_values", manageCredentialsInput{
			Action: "create", CredentialType: kindPassword,
			Name: "password", Scope: "tenant", Spec: &credentialSpec{User: "user"},
		}, "spec.password is required"},
		{"docker_registry_values", manageCredentialsInput{
			Action: "create", CredentialType: kindDockerRegistry,
			Name: "registry", Scope: "tenant",
			Spec: &credentialSpec{User: "user", Password: "secret"},
		}, "spec.url is required"},
		{"generic_secret_values", manageCredentialsInput{
			Action: "create", Private: true, CredentialType: kindGenericSecret,
			Name: "generic", Spec: &credentialSpec{},
		}, "spec.keyValuePairs must contain at least one entry"},
		{"ngc_api_key_value", manageCredentialsInput{
			Action: "create", Private: true, CredentialType: kindNGCAPIKey,
			Name: "ngc", Spec: &credentialSpec{},
		}, "spec.ngcApiKey is required"},
		{"cluster_id", manageCredentialsInput{
			Action: "create", CredentialType: kindAccessKey,
			Name: "cluster", Scope: "cluster", Spec: validAccessKey,
		}, "clusterId is required"},
		{"department_id", manageCredentialsInput{
			Action: "create", CredentialType: kindAccessKey,
			Name: "department", Scope: "department", Spec: validAccessKey,
		}, "departmentId is required"},
		{"project_id", manageCredentialsInput{
			Action: "create", CredentialType: kindAccessKey,
			Name: "project", Scope: "project", Spec: validAccessKey,
		}, "projectId is required"},
		{"update_id", manageCredentialsInput{
			Action: "update", CredentialType: kindAccessKey,
			Description: &description,
		}, "credentialId is required for update"},
		{"update_description", manageCredentialsInput{
			Action: "update", CredentialType: kindAccessKey,
			CredentialID: testCredentialID,
		}, "description is required for update"},
		{"delete_id", manageCredentialsInput{
			Action: "delete", CredentialType: kindAccessKey,
		}, "credentialId is required for delete"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func int32Ptr(v int32) *int32 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}
