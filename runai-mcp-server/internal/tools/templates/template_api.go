/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package templates

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

const (
	templateTypeTraining             = "Training"
	templateTypeWorkspace            = "Workspace"
	templateTypeDistributed          = "Distributed"
	templateTypeInference            = "Inference"
	templateTypeDistributedInference = "DistributedInference"
)

type templateResult interface {
	GetMeta() workloads.TemplateMetaFields
}

type templateAdapter struct {
	validateCreate func(map[string]any) error
	validateUpdate func(map[string]any) error
	create         func(context.Context, *runaiclient.Client, map[string]any) (templateResult, *http.Response, error)
	update         func(context.Context, *runaiclient.Client, string, map[string]any) (templateResult, *http.Response, error)
	delete         func(context.Context, *runaiclient.Client, string) (*http.Response, error)
}

var templateAdapters = map[string]templateAdapter{
	templateTypeWorkspace: templateAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r workloads.WorkspaceTemplateCreationRequest) (*workloads.WorkspaceTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.CreateWorkspaceTemplate(ctx).WorkspaceTemplateCreationRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r workloads.WorkspaceTemplateUpdateRequest) (*workloads.WorkspaceTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.UpdateWorkspaceTemplate(ctx, id).WorkspaceTemplateUpdateRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.DeleteWorkspaceTemplate(ctx, id).Execute()
		}),
	templateTypeTraining: templateAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r workloads.TrainingTemplateCreationRequest) (*workloads.TrainingTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.CreateTrainingTemplate(ctx).TrainingTemplateCreationRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r workloads.TrainingTemplateUpdateRequest) (*workloads.TrainingTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.UpdateTrainingTemplate(ctx, id).TrainingTemplateUpdateRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.DeleteTrainingTemplate(ctx, id).Execute()
		}),
	templateTypeDistributed: templateAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r workloads.DistributedTemplateCreationRequest) (*workloads.DistributedTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.CreateDistributedTemplate(ctx).DistributedTemplateCreationRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r workloads.DistributedTemplateUpdateRequest) (*workloads.DistributedTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.UpdateDistributedTemplate(ctx, id).DistributedTemplateUpdateRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.DeleteDistributedTemplate(ctx, id).Execute()
		}),
	templateTypeInference: templateAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r workloads.InferenceTemplateCreationRequest) (*workloads.InferenceTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.CreateInferenceTemplate(ctx).InferenceTemplateCreationRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r workloads.InferenceTemplateUpdateRequest) (*workloads.InferenceTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.UpdateInferenceTemplate(ctx, id).InferenceTemplateUpdateRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.DeleteInferenceTemplate(ctx, id).Execute()
		}),
	templateTypeDistributedInference: templateAdapterFor(
		func(ctx context.Context, c *runaiclient.Client, r workloads.DistributedInferenceTemplateCreationRequest) (*workloads.DistributedInferenceTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.CreateDistributedInferenceTemplate(ctx).DistributedInferenceTemplateCreationRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string, r workloads.DistributedInferenceTemplateUpdateRequest) (*workloads.DistributedInferenceTemplate, *http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.UpdateDistributedInferenceTemplate(ctx, id).DistributedInferenceTemplateUpdateRequest(r).Execute()
		},
		func(ctx context.Context, c *runaiclient.Client, id string) (*http.Response, error) {
			return c.Workloads.WorkloadTemplatesAPI.DeleteDistributedInferenceTemplate(ctx, id).Execute()
		}),
}

func templateAdapterFor[C any, U any, R templateResult](
	create func(context.Context, *runaiclient.Client, C) (R, *http.Response, error),
	update func(context.Context, *runaiclient.Client, string, U) (R, *http.Response, error),
	deleteCall func(context.Context, *runaiclient.Client, string) (*http.Response, error),
) templateAdapter {
	decodeCreate := func(raw map[string]any) (C, error) { return decodeTemplateRequest[C](raw, "create") }
	decodeUpdate := func(raw map[string]any) (U, error) { return decodeTemplateRequest[U](raw, "update") }
	return templateAdapter{
		validateCreate: func(raw map[string]any) error { _, err := decodeCreate(raw); return err },
		validateUpdate: func(raw map[string]any) error { _, err := decodeUpdate(raw); return err },
		create: func(ctx context.Context, client *runaiclient.Client, raw map[string]any) (templateResult, *http.Response, error) {
			req, err := decodeCreate(raw)
			if err != nil {
				return nil, nil, err
			}
			result, resp, err := create(ctx, client, req)
			if isNilTemplateResult(result) {
				return nil, resp, err
			}
			return result, resp, err
		},
		update: func(ctx context.Context, client *runaiclient.Client, id string, raw map[string]any) (templateResult, *http.Response, error) {
			req, err := decodeUpdate(raw)
			if err != nil {
				return nil, nil, err
			}
			result, resp, err := update(ctx, client, id, req)
			if isNilTemplateResult(result) {
				return nil, resp, err
			}
			return result, resp, err
		},
		delete: deleteCall,
	}
}

func decodeTemplateRequest[T any](raw map[string]any, action string) (T, error) {
	var out T
	if err := common.DecodeTypedRequest(raw, &out); err != nil {
		return out, fmt.Errorf("decode %s request: %v", action, err)
	}
	return out, nil
}

func isNilTemplateResult(value any) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func templateMutationError(action, workloadType, id string, resp *http.Response, err error) error {
	target := templateKindLabel(workloadType) + " template"
	if id == "" {
		return fmt.Errorf("%s %s: %w", action, target, runaimcp.MapHTTPError(resp, err))
	}
	return fmt.Errorf("%s %s %q: %w", action, target, id, runaimcp.MapHTTPError(resp, err))
}

func nilTemplateResponseError(action, workloadType string) error {
	return fmt.Errorf("%s %s template: nil response: %w", action, templateKindLabel(workloadType), runaimcp.ErrUpstream)
}

func templateKindLabel(workloadType string) string {
	if workloadType == templateTypeDistributedInference {
		return "distributed inference"
	}
	return strings.ToLower(workloadType)
}

func unsupportedTemplateTypeError(field string) error {
	return fmt.Errorf("%s must be one of: %s, %s, %s, %s, %s",
		field, templateTypeTraining, templateTypeWorkspace, templateTypeDistributed,
		templateTypeInference, templateTypeDistributedInference)
}
