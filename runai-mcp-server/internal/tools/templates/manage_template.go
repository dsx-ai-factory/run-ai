/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package templates

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"unicode/utf8"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type manageTemplateInput struct {
	Action              string         `json:"action"                 jsonschema:"What to do. One of: create, update, delete."`
	WorkloadType        string         `json:"workloadType"           jsonschema:"Native workload type. One of: Training, Workspace, Distributed, Inference, DistributedInference."`
	TemplateID          string         `json:"templateId,omitempty"   jsonschema:"Template UUID. Required for update and delete; use list_templates to resolve it. Must not be set for create."`
	Request             map[string]any `json:"request,omitempty"      jsonschema:"Complete API request for create or update. Create requires meta.name and meta.scopeType; update requires meta.name. Supports every field of the selected workload type's template request model, including assets, spec, and distributed master fields where the API accepts them."` //nolint:lll // complete polymorphic request description
	ExpectedFingerprint string         `json:"expectedFingerprint,omitempty" jsonschema:"Required for update. Pass unchanged from get_template with forUpdate=true for a best-effort stale-snapshot check. The API has no atomic conditional-update support."`                                                                                                         //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Confirm             bool           `json:"confirm,omitempty"      jsonschema:"Required for update and delete when the client cannot show an interactive confirmation prompt. Create does not require confirmation."`
}

func (in manageTemplateInput) Validate() error {
	action := common.WriteAction(in.Action)
	if !action.Valid() {
		return fmt.Errorf("action must be one of: create, update, delete")
	}
	adapter, ok := templateAdapters[in.WorkloadType]
	if !ok {
		return unsupportedTemplateTypeError("workloadType")
	}
	if in.TemplateID != "" {
		if err := common.ValidateUUID("templateId", in.TemplateID); err != nil {
			return err
		}
	}

	switch action {
	case common.ActionCreate:
		if in.TemplateID != "" {
			return fmt.Errorf("templateId must not be set for create")
		}
		if in.ExpectedFingerprint != "" {
			return fmt.Errorf("expectedFingerprint must not be set for create")
		}
		if len(in.Request) == 0 {
			return fmt.Errorf("request is required for create")
		}
		if err := validateCreateMeta(in.Request); err != nil {
			return err
		}
		if err := validateTemplateRequestFields(in.WorkloadType, in.Request); err != nil {
			return err
		}
		return adapter.validateCreate(in.Request)
	case common.ActionUpdate:
		if in.TemplateID == "" {
			return fmt.Errorf("templateId is required for update")
		}
		if len(in.Request) == 0 {
			return fmt.Errorf("request is required for update")
		}
		if err := validateUpdateMeta(in.Request); err != nil {
			return err
		}
		if err := validateTemplateRequestFields(in.WorkloadType, in.Request); err != nil {
			return err
		}
		if err := adapter.validateUpdate(in.Request); err != nil {
			return err
		}
		if in.ExpectedFingerprint == "" {
			return fmt.Errorf("expectedFingerprint is required for the best-effort stale-snapshot check; call get_template with forUpdate=true")
		}
		return validateTemplateFingerprint(in.ExpectedFingerprint)
	case common.ActionDelete:
		if in.TemplateID == "" {
			return fmt.Errorf("templateId is required for delete")
		}
		if len(in.Request) != 0 {
			return fmt.Errorf("request must not be set for delete")
		}
		if in.ExpectedFingerprint != "" {
			return fmt.Errorf("expectedFingerprint must not be set for delete")
		}
		return nil
	default:
		return fmt.Errorf("unsupported action %q", in.Action)
	}
}

type manageTemplateOutput struct {
	Action       string `json:"action"               jsonschema:"The action performed: create, update, or delete."`
	Applied      bool   `json:"applied"              jsonschema:"True when the upstream service accepted the change."`
	TemplateID   string `json:"templateId,omitempty" jsonschema:"Created or targeted template UUID."`
	Name         string `json:"name,omitempty"       jsonschema:"Template name."`
	WorkloadType string `json:"workloadType"         jsonschema:"Native workload type."`
}

func registerManageTemplate(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[manageTemplateInput, manageTemplateOutput]{
		Name: "manage_template",
		Description: "Create, fully replace, or delete one native Run:ai workload template. workloadType selects the " +
			"Training, Workspace, Distributed, Inference, or DistributedInference template API; request must match that " +
			"API and unknown fields are rejected. Create supports tenant, cluster, department, and project scope through " +
			"request.meta.scopeType and request.meta.scopeId, does not support system scope, and does not require " +
			"confirmation. Update uses a full-replacement PUT and requires the complete update request. System templates " +
			"cannot be updated; call get_template with forUpdate=true first and pass its updateRequest as the complete " +
			"request after making the intended edits, together with expectedFingerprint unchanged. The fingerprint provides " +
			"a best-effort stale-snapshot check before the PUT; the template API does not support an atomic update precondition. " +
			"Update and delete require " +
			"confirmation. Delete permanently removes a non-system template; for a system preset, it hides the preset " +
			"only for the current tenant. Call list_templates when templateId or workloadType is unknown.",
		ReadOnly:    false,
		Destructive: true,
		Handler: func(ctx context.Context, in manageTemplateInput) (manageTemplateOutput, error) {
			return manageTemplateRun(ctx, client, in)
		},
	})
}

func manageTemplateRun(ctx context.Context, client *runaiclient.Client, in manageTemplateInput) (manageTemplateOutput, error) {
	adapter := templateAdapters[in.WorkloadType]
	switch common.WriteAction(in.Action) {
	case common.ActionCreate:
		return createTemplate(ctx, client, adapter, in)
	case common.ActionUpdate, common.ActionDelete:
		target, err := getTemplateMutationTarget(ctx, client, in)
		if err != nil {
			return manageTemplateOutput{}, err
		}
		if common.WriteAction(in.Action) == common.ActionUpdate {
			return updateTemplate(ctx, client, adapter, in, target)
		}
		return deleteTemplate(ctx, client, adapter, in, target)
	default:
		return manageTemplateOutput{}, fmt.Errorf("unsupported action %q: %w", in.Action, runaimcp.ErrValidation)
	}
}

type templateMutationTarget struct {
	template *workloads.SupersetTemplate
	raw      map[string]any
	output   manageTemplateOutput
}

func createTemplate(
	ctx context.Context,
	client *runaiclient.Client,
	adapter templateAdapter,
	in manageTemplateInput,
) (manageTemplateOutput, error) {
	created, resp, err := adapter.create(ctx, client, in.Request)
	if err != nil {
		return manageTemplateOutput{}, templateMutationError("create", in.WorkloadType, "", resp, err)
	}
	if created == nil {
		return manageTemplateOutput{}, nilTemplateResponseError("create", in.WorkloadType)
	}

	out := newManageTemplateOutput(in)
	createdMeta := created.GetMeta()
	out.TemplateID = createdMeta.Id
	out.Name = createdMeta.Name
	out.WorkloadType = string(createdMeta.WorkloadType)
	out.Applied = true
	return out, nil
}

func getTemplateMutationTarget(
	ctx context.Context,
	client *runaiclient.Client,
	in manageTemplateInput,
) (templateMutationTarget, error) {
	current, resp, err := client.Workloads.WorkloadTemplatesAPI.GetTemplateById(ctx, in.TemplateID).Execute()
	if err != nil {
		return templateMutationTarget{}, fmt.Errorf("get template %q: %w", in.TemplateID, runaimcp.MapHTTPError(resp, err))
	}
	if current == nil {
		return templateMutationTarget{}, fmt.Errorf("get template %q: nil response: %w", in.TemplateID, runaimcp.ErrUpstream)
	}
	if string(current.Meta.WorkloadType) != in.WorkloadType {
		return templateMutationTarget{}, fmt.Errorf("template %q has workloadType %s, not %s: %w",
			in.TemplateID, current.Meta.WorkloadType, in.WorkloadType, runaimcp.ErrValidation)
	}

	out := newManageTemplateOutput(in)
	out.TemplateID = current.Meta.Id
	out.Name = current.Meta.Name
	currentRaw, err := rawTemplateResponse(resp, in.TemplateID)
	if err != nil {
		return templateMutationTarget{}, err
	}
	return templateMutationTarget{template: current, raw: currentRaw, output: out}, nil
}

func updateTemplate(
	ctx context.Context,
	client *runaiclient.Client,
	adapter templateAdapter,
	in manageTemplateInput,
	target templateMutationTarget,
) (manageTemplateOutput, error) {
	currentRequest, err := verifiedTemplateUpdateRequest(in, target)
	if err != nil {
		return target.output, err
	}
	diff := common.DiffMaps(currentRequest, in.Request)
	details := common.RemoveKeys(diff, templateSensitiveKeys).(map[string]any)
	if !templateClientCanElicit(ctx) {
		details = compactTemplateUpdateConfirmation(target.template.Meta, diff)
	}
	if err := common.ConfirmWrite(ctx, in.Confirm, "update", "template",
		target.output.Name, target.output.TemplateID, details); err != nil {
		return target.output, err
	}

	// Re-read after confirmation because an interactive prompt can remain open
	// while another caller changes the full-replacement target.
	confirmedTarget, err := getTemplateMutationTarget(ctx, client, in)
	if err != nil {
		return target.output, err
	}
	if _, err := verifiedTemplateUpdateRequest(in, confirmedTarget); err != nil {
		return confirmedTarget.output, err
	}
	target = confirmedTarget

	// The template API has no ETag, If-Match, resource-version, or equivalent
	// update precondition. Re-reading immediately before PUT reduces stale writes,
	// but cannot close the race if the template changes between these two calls.
	updated, resp, err := adapter.update(ctx, client, target.output.TemplateID, in.Request)
	if err != nil {
		return manageTemplateOutput{}, templateMutationError("update", in.WorkloadType, target.output.TemplateID, resp, err)
	}
	if updated == nil {
		return manageTemplateOutput{}, nilTemplateResponseError("update", in.WorkloadType)
	}
	updatedMeta := updated.GetMeta()
	target.output.TemplateID = updatedMeta.Id
	target.output.Name = updatedMeta.Name
	target.output.WorkloadType = string(updatedMeta.WorkloadType)
	target.output.Applied = true
	return target.output, nil
}

func verifiedTemplateUpdateRequest(in manageTemplateInput, target templateMutationTarget) (map[string]any, error) {
	if target.template.Meta.ScopeType == workloads.SCOPETYPE_SYSTEM {
		return nil, fmt.Errorf("system templates cannot be updated: %w", runaimcp.ErrForbidden)
	}
	currentRequest, currentFingerprint, err := templateUpdateSnapshot(in.WorkloadType, target.raw)
	if err != nil {
		return nil, fmt.Errorf("prepare current template %q for update: %w", target.output.TemplateID, err)
	}
	if currentFingerprint != in.ExpectedFingerprint {
		return nil, fmt.Errorf("template %q changed since get_template returned the update request; retrieve it again before updating: %w",
			target.output.TemplateID, runaimcp.ErrConflict)
	}
	return currentRequest, nil
}

func deleteTemplate(
	ctx context.Context,
	client *runaiclient.Client,
	adapter templateAdapter,
	in manageTemplateInput,
	target templateMutationTarget,
) (manageTemplateOutput, error) {
	warning := "deleting a template is permanent"
	if target.template.Meta.ScopeType == workloads.SCOPETYPE_SYSTEM {
		warning = "deleting a system template hides the preset for the current tenant"
	}
	details := common.WithConfirmWarning(common.RemoveKeys(target.raw, templateSensitiveKeys).(map[string]any), warning)
	if !templateClientCanElicit(ctx) {
		details = compactTemplateDeleteConfirmation(target.template.Meta, warning)
	}
	if err := common.ConfirmWrite(ctx, in.Confirm, "delete", "template",
		target.output.Name, target.output.TemplateID, details); err != nil {
		return target.output, err
	}
	resp, err := adapter.delete(ctx, client, target.output.TemplateID)
	if err != nil {
		return manageTemplateOutput{}, templateMutationError("delete", in.WorkloadType, target.output.TemplateID, resp, err)
	}
	target.output.Applied = true
	return target.output, nil
}

func templateClientCanElicit(ctx context.Context) bool {
	return runaimcp.ClientCanElicitForm(runaimcp.SessionFromContext(ctx))
}

func compactTemplateUpdateConfirmation(meta workloads.TemplateMetaFields, diff map[string]any) map[string]any {
	changedFields := make([]string, 0, len(diff))
	for field := range diff {
		changedFields = append(changedFields, field)
	}
	sort.Strings(changedFields)
	return map[string]any{
		"scope":         templateScopeSummary(meta),
		"changedFields": changedFields,
	}
}

func compactTemplateDeleteConfirmation(meta workloads.TemplateMetaFields, warning string) map[string]any {
	return map[string]any{
		"scope":   templateScopeSummary(meta),
		"warning": warning,
	}
}

func templateScopeSummary(meta workloads.TemplateMetaFields) string {
	scope := string(meta.ScopeType)
	if meta.ScopeId != "" {
		return fmt.Sprintf("%s (%s)", scope, meta.ScopeId)
	}
	return scope
}

func newManageTemplateOutput(in manageTemplateInput) manageTemplateOutput {
	return manageTemplateOutput{
		Action:       in.Action,
		TemplateID:   in.TemplateID,
		WorkloadType: in.WorkloadType,
	}
}

func validateCreateMeta(request map[string]any) error {
	meta, err := requestMeta(request, "create")
	if err != nil {
		return err
	}
	if _, err := common.RequiredString("request.meta.name", meta["name"]); err != nil {
		return err
	}
	scopeType, err := common.RequiredString("request.meta.scopeType", meta["scopeType"])
	if err != nil {
		return err
	}
	scopeID, _ := meta["scopeId"].(string)
	switch workloads.ScopeType(scopeType) {
	case workloads.SCOPETYPE_TENANT:
		if scopeID != "" {
			return fmt.Errorf("request.meta.scopeId must not be set for tenant scope; the API derives it")
		}
		return nil
	case workloads.SCOPETYPE_CLUSTER:
		return common.ValidateUUID("request.meta.scopeId", scopeID)
	case workloads.SCOPETYPE_DEPARTMENT:
		_, err := common.RequiredString("request.meta.scopeId", scopeID)
		return err
	case workloads.SCOPETYPE_PROJECT:
		projectID, err := strconv.ParseInt(scopeID, 10, 32)
		if err != nil || projectID <= 0 {
			return fmt.Errorf("request.meta.scopeId must be a positive 32-bit project ID (got %q)", scopeID)
		}
		return nil
	case workloads.SCOPETYPE_SYSTEM:
		return fmt.Errorf("request.meta.scopeType=system is not supported by the workload-template create API")
	default:
		return fmt.Errorf("request.meta.scopeType must be one of: tenant, cluster, department, project")
	}
}

func validateUpdateMeta(request map[string]any) error {
	meta, err := requestMeta(request, "update")
	if err != nil {
		return err
	}
	_, err = common.RequiredString("request.meta.name", meta["name"])
	return err
}

func validateTemplateRequestFields(workloadType string, request map[string]any) error {
	meta, _ := request["meta"].(map[string]any)
	if err := validateTemplateDescription(meta); err != nil {
		return err
	}
	if err := validateTemplateAssets("request.assets", request["assets"]); err != nil {
		return err
	}
	if workloadType != templateTypeDistributed {
		return nil
	}

	master, hasMaster := request["master"]
	sameAsWorker, _ := request["masterSpecSameAsWorker"].(bool)
	if sameAsWorker && hasMaster && master != nil {
		return fmt.Errorf("request.master must not be set when request.masterSpecSameAsWorker=true")
	}
	masterObject, _ := master.(map[string]any)
	return validateTemplateAssets("request.master.assets", masterObject["assets"])
}

func validateTemplateDescription(meta map[string]any) error {
	if meta == nil {
		return nil
	}
	value, exists := meta["description"]
	if !exists || value == nil {
		return nil
	}
	description, ok := value.(string)
	if !ok {
		return nil
	}
	length := utf8.RuneCountInString(description)
	if length == 0 || length > 250 {
		return fmt.Errorf("request.meta.description must contain between 1 and 250 characters, or be null to clear it")
	}
	return nil
}

func validateTemplateAssets(path string, value any) error {
	if value == nil {
		return nil
	}
	assets, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	for _, field := range []string{"environment", "compute"} {
		assetID, exists := assets[field]
		if !exists || assetID == nil {
			continue
		}
		if id, ok := assetID.(string); ok {
			if err := common.ValidateUUID(path+"."+field, id); err != nil {
				return err
			}
		}
	}
	datasources, ok := assets["datasources"].([]any)
	if !ok {
		return nil
	}
	for i, value := range datasources {
		if id, ok := value.(string); ok {
			if err := common.ValidateUUID(fmt.Sprintf("%s.datasources[%d]", path, i), id); err != nil {
				return err
			}
		}
	}
	return nil
}

func requestMeta(request map[string]any, action string) (map[string]any, error) {
	value, ok := request["meta"]
	if !ok {
		return nil, fmt.Errorf("request.meta is required for %s", action)
	}
	meta, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("request.meta must be an object")
	}
	return meta, nil
}

func buildTemplateUpdateRequest(workloadType string, raw map[string]any) map[string]any {
	request := map[string]any{}
	meta := map[string]any{}
	if current, ok := raw["meta"].(map[string]any); ok {
		meta["name"] = current["name"]
		if description, exists := current["description"]; exists {
			meta["description"] = description
		}
	}
	request["meta"] = meta
	fields := []string{"spec"}
	if workloadType != templateTypeDistributedInference {
		fields = append(fields, "assets")
	}
	if workloadType == templateTypeDistributed {
		fields = append(fields, "masterSpecSameAsWorker", "master")
	}
	for _, field := range fields {
		if value, exists := raw[field]; exists {
			request[field] = value
		}
	}
	return request
}
