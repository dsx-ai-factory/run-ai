/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package templates

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type getTemplateInput struct {
	TemplateID string `json:"templateId"          jsonschema:"Template UUID returned by list_templates."`
	ForUpdate  bool   `json:"forUpdate,omitempty" jsonschema:"When true, return a complete request body accepted by manage_template update instead of the read-only API response. Defaults to false."`
}

func (in getTemplateInput) Validate() error {
	return common.ValidateUUID("templateId", in.TemplateID)
}

type getTemplateOutput struct {
	TemplateID          string         `json:"templateId"              jsonschema:"Template UUID to pass to manage_template."`
	Name                string         `json:"name"                    jsonschema:"Template name."`
	WorkloadType        string         `json:"workloadType"            jsonschema:"Native workload type to pass to manage_template."`
	Template            map[string]any `json:"template,omitempty"      jsonschema:"Complete read-only template returned by the Get Template by ID API, with credential values such as awsSessionToken removed. Present unless forUpdate is true."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	UpdateRequest       map[string]any `json:"updateRequest,omitempty" jsonschema:"Complete request body accepted by manage_template update, with read-only response fields removed. Present when forUpdate is true."`
	ExpectedFingerprint string         `json:"expectedFingerprint,omitempty" jsonschema:"Best-effort stale-snapshot fingerprint to pass unchanged to manage_template with updateRequest. It detects changes visible before the update request but is not an atomic API precondition."` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func registerGetTemplate(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[getTemplateInput, getTemplateOutput]{
		Name: "get_template",
		Description: "Retrieve one complete Run:ai workload template by templateId. Returns the full non-secret template from the " +
			"Get Template by ID API, including its spec, assets, compliance and runnable information, and distributed " +
			"settings when present; credential values such as awsSessionToken are never returned. Set forUpdate to return only an update-ready request with all mutable fields and no " +
			"read-only response fields; pass that updateRequest and expectedFingerprint to manage_template for a full replacement with a best-effort stale-snapshot check. Use " +
			"list_templates first when the template ID is unknown.",
		ReadOnly: true,
		Handler: func(ctx context.Context, in getTemplateInput) (getTemplateOutput, error) {
			return getTemplateRun(ctx, client, in)
		},
	})
}

func getTemplateRun(ctx context.Context, client *runaiclient.Client, in getTemplateInput) (getTemplateOutput, error) {
	template, resp, err := client.Workloads.WorkloadTemplatesAPI.GetTemplateById(ctx, in.TemplateID).Execute()
	if err != nil {
		return getTemplateOutput{}, fmt.Errorf("get template %q: %w", in.TemplateID, runaimcp.MapHTTPError(resp, err))
	}
	if template == nil {
		return getTemplateOutput{}, fmt.Errorf("get template %q: nil response: %w", in.TemplateID, runaimcp.ErrUpstream)
	}
	raw, err := rawTemplateResponse(resp, in.TemplateID)
	if err != nil {
		return getTemplateOutput{}, err
	}
	out := getTemplateOutput{
		TemplateID:   template.Meta.Id,
		Name:         template.Meta.Name,
		WorkloadType: string(template.Meta.WorkloadType),
	}
	if in.ForUpdate {
		out.UpdateRequest, out.ExpectedFingerprint, err = templateUpdateSnapshot(out.WorkloadType, raw)
		if err != nil {
			return getTemplateOutput{}, fmt.Errorf("prepare template %q for update: %w", in.TemplateID, err)
		}
	} else {
		out.Template = templateWithoutSecrets(raw).(map[string]any)
	}
	return out, nil
}

const (
	templateFingerprintPrefix = "sha256:"
	templateFieldDepartmentID = "departmentId"
	templateFieldName         = "name"
	templateFieldScopeType    = "scopeType"
	templateFieldWorkloadType = "workloadType"
)

var knownTemplateResponseFields = map[string]struct{}{
	"assets":                 {},
	"combinedMasterSpec":     {},
	"combinedSpec":           {},
	"complianceIssues":       {},
	"master":                 {},
	"masterSpecSameAsWorker": {},
	"meta":                   {},
	"runnable":               {},
	"spec":                   {},
}

var knownTemplateMetaFields = map[string]struct{}{
	"clusterId":               {},
	"createdAt":               {},
	"createdBy":               {},
	templateFieldDepartmentID: {},
	"description":             {},
	"id":                      {},
	templateFieldName:         {},
	"projectId":               {},
	"scopeId":                 {},
	templateFieldScopeType:    {},
	"updatedAt":               {},
	"updatedBy":               {},
	templateFieldWorkloadType: {},
}

func rawTemplateResponse(resp *http.Response, templateID string) (map[string]any, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("read template %q response: missing response body: %w", templateID, runaimcp.ErrUpstream)
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("read template %q response: %w: %w", templateID, err, runaimcp.ErrUpstream)
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode template %q response: %w: %w", templateID, err, runaimcp.ErrUpstream)
	}
	if raw == nil {
		return nil, fmt.Errorf("decode template %q response: nil response: %w", templateID, runaimcp.ErrUpstream)
	}
	return raw, nil
}

func templateUpdateSnapshot(workloadType string, raw map[string]any) (map[string]any, string, error) {
	if unknown := unknownObjectFields(raw, knownTemplateResponseFields); len(unknown) != 0 {
		return nil, "", unsupportedTemplateFieldsError(unknown)
	}
	if meta, ok := raw["meta"].(map[string]any); ok {
		if unknown := unknownObjectFields(meta, knownTemplateMetaFields); len(unknown) != 0 {
			for i := range unknown {
				unknown[i] = "meta." + unknown[i]
			}
			return nil, "", unsupportedTemplateFieldsError(unknown)
		}
	}
	if unsupported := unsupportedMutableTemplateFields(workloadType, raw); len(unsupported) != 0 {
		return nil, "", unsupportedTemplateFieldsError(unsupported)
	}

	request := buildTemplateUpdateRequest(workloadType, raw)
	if templateContainsSecret(request) {
		return nil, "", fmt.Errorf("template update request contains awsSessionToken, which cannot be exposed or safely "+
			"round-tripped by manage_template; remove or rotate the session token through the API before updating the template: %w",
			runaimcp.ErrConflict)
	}
	// A nullable credential field carries no secret and must not block the
	// snapshot, but it should still be omitted from the caller-visible request.
	request = templateWithoutSecrets(request).(map[string]any)
	adapter, ok := templateAdapters[workloadType]
	if !ok {
		return nil, "", unsupportedTemplateTypeError("workloadType")
	}
	if err := adapter.validateUpdate(request); err != nil {
		return nil, "", fmt.Errorf("template contains fields this MCP version cannot safely preserve during a full-replacement update (%v): %w",
			err, runaimcp.ErrConflict)
	}

	fingerprintValue := map[string]any{
		"request":      request,
		"workloadType": workloadType,
	}
	if meta, ok := raw["meta"].(map[string]any); ok {
		identity := make(map[string]any, 4)
		for _, field := range []string{"id", "scopeId", templateFieldScopeType, templateFieldWorkloadType} {
			if value, exists := meta[field]; exists {
				identity[field] = value
			}
		}
		fingerprintValue["identity"] = identity
	}
	encoded, err := json.Marshal(fingerprintValue)
	if err != nil {
		return nil, "", fmt.Errorf("fingerprint template update request: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return request, templateFingerprintPrefix + hex.EncodeToString(sum[:]), nil
}

func unsupportedMutableTemplateFields(workloadType string, raw map[string]any) []string {
	unsupported := make([]string, 0, 3)
	if workloadType == templateTypeDistributedInference {
		if value, exists := raw["assets"]; exists && value != nil {
			unsupported = append(unsupported, "assets")
		}
	}
	if workloadType != templateTypeDistributed {
		for _, field := range []string{"master", "masterSpecSameAsWorker"} {
			if value, exists := raw[field]; exists && value != nil {
				unsupported = append(unsupported, field)
			}
		}
	}
	return unsupported
}

func templateContainsSecret(value any) bool {
	switch value := value.(type) {
	case map[string]any:
		for key, nested := range value {
			if (strings.EqualFold(key, "awsSessionToken") && nested != nil) || templateContainsSecret(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range value {
			if templateContainsSecret(nested) {
				return true
			}
		}
	}
	return false
}

func templateWithoutSecrets(value any) any {
	switch value := value.(type) {
	case map[string]any:
		clean := make(map[string]any, len(value))
		for key, nested := range value {
			if strings.EqualFold(key, "awsSessionToken") {
				continue
			}
			clean[key] = templateWithoutSecrets(nested)
		}
		return clean
	case []any:
		clean := make([]any, len(value))
		for i, nested := range value {
			clean[i] = templateWithoutSecrets(nested)
		}
		return clean
	default:
		return value
	}
}

func validateTemplateFingerprint(value string) error {
	if !strings.HasPrefix(value, templateFingerprintPrefix) {
		return fmt.Errorf("expectedFingerprint must be the value returned by get_template with forUpdate=true")
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, templateFingerprintPrefix))
	if err != nil || len(decoded) != sha256.Size {
		return fmt.Errorf("expectedFingerprint must be the value returned by get_template with forUpdate=true")
	}
	return nil
}

func unknownObjectFields(value map[string]any, known map[string]struct{}) []string {
	unknown := make([]string, 0)
	for field := range value {
		if _, ok := known[field]; !ok {
			unknown = append(unknown, field)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func unsupportedTemplateFieldsError(fields []string) error {
	return fmt.Errorf("template contains fields this MCP version cannot safely preserve during a full-replacement update (%s); upgrade the MCP server before updating: %w",
		strings.Join(fields, ", "), runaimcp.ErrConflict)
}
