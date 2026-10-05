/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

type AssetsClient struct {
	ApiClient *openapi.APIClient
}

func (c *AssetsClient) ListComputeAssets(ctx context.Context, name string) (*openapi.ComputeListResponse, *http.Response, error) {
	return c.ApiClient.ComputeAPI.ListComputeAssets(ctx).Name(name).Execute()
}

func (c *AssetsClient) CreateComputeAsset(ctx context.Context, request *openapi.ComputeCreationRequest) (*openapi.ComputeAsset, *http.Response, error) {
	return c.ApiClient.ComputeAPI.CreateComputeAsset(ctx).ComputeCreationRequest(*request).Execute()
}

func (c *AssetsClient) ListEnvironmentAssets(ctx context.Context, name string) (*openapi.EnvironmentListResponse, *http.Response, error) {
	return c.ApiClient.EnvironmentAPI.ListEnvironmentAssets(ctx).Name(name).Execute()
}

func (c *AssetsClient) CreateEnvironmentAsset(ctx context.Context, request *openapi.EnvironmentCreationRequest) (*openapi.EnvironmentAsset, *http.Response, error) {
	return c.ApiClient.EnvironmentAPI.CreateEnvironmentAsset(ctx).EnvironmentCreationRequest(*request).Execute()
}

func (c *AssetsClient) ListModelAssets(ctx context.Context, name string) (*openapi.ModelListResponse, *http.Response, error) {
	return c.ApiClient.ModelAPI.ListModelAssets(ctx).Name(name).Execute()
}

func (c *AssetsClient) CreateModelAsset(ctx context.Context, request *openapi.ModelCreationRequest) (*openapi.ModelAsset, *http.Response, error) {
	return c.ApiClient.ModelAPI.CreateModelAsset(ctx).ModelCreationRequest(*request).Execute()
}

func (c *AssetsClient) ListDatasourceAssets(ctx context.Context, ids []string) (*openapi.DatasourceListResponse, *http.Response, error) {
	return c.ApiClient.DatasourcesAPI.ListDatasourceAssets(ctx).AssetIds(strings.Join(ids, ",")).
		Execute()
}

func (c *AssetsClient) RequestAssetFlatFormat(ctx context.Context, request openapi.AssetFlatRequest) (*openapi.AssetFlatResponse, *openapi.Error) {
	resp, httpResp, err := c.ApiClient.AssetsAPI.RequestAssetFlatFormat(ctx).AssetFlatRequest(request).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read asset payload: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		// A non-nil err on the HTTP 200 branch means the generated
		// client hit a decode error after the body was received —
		// resp may be partially populated. Returning (resp, nil)
		// here would silently report that half-decoded response as
		// success and advance the sync counter past whatever the
		// decoder dropped, so surface the decode error instead.
		if err != nil {
			return nil, openapi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("decode response: %v", err))
		}
		if resp != nil {
			return resp, nil
		}
		// Status was OK but the generated client decoded a nil typed
		// response (e.g. empty body on a schema that expected a payload).
		// Surface it as a real 500 rather than letting the final return
		// below format "<nil>" into the 503 message.
		return nil, openapi.NewError(http.StatusInternalServerError,
			"empty response body with status OK")
	}

	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return nil, openapi.NewError(http.StatusServiceUnavailable,
		"no response from assets-service and no transport error")
}

func (c *AssetsClient) CredentialsSyncUpdates(ctx context.Context, clusterId string, fromUpdate int64, request openapi.ClusterSyncReport) (*openapi.CredentialsUpdatesResp, *openapi.Error) {
	resp, httpResp, err := c.ApiClient.SynchronizationAPI.CredentialsSyncUpdates(ctx).ClusterSyncReport(request).After(fromUpdate).ClusterId(clusterId).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		// A non-nil err on the HTTP 200 branch means the generated
		// client hit a decode error after the body was received —
		// resp may be partially populated. Returning (resp, nil)
		// here would silently report that half-decoded response as
		// success and advance the sync counter past whatever the
		// decoder dropped, so surface the decode error instead.
		if err != nil {
			return nil, openapi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("decode response: %v", err))
		}
		if resp != nil {
			return resp, nil
		}
		// Status was OK but the generated client decoded a nil typed
		// response (e.g. empty body on a schema that expected a payload).
		// Surface it as a real 500 rather than letting the final return
		// below format "<nil>" into the 503 message.
		return nil, openapi.NewError(http.StatusInternalServerError,
			"empty response body with status OK")
	}

	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return nil, openapi.NewError(http.StatusServiceUnavailable,
		"no response from assets-service and no transport error")
}

func (c *AssetsClient) DatasourceSyncUpdates(ctx context.Context, clusterId string, fromUpdate int64, request openapi.ClusterSyncReport) (*openapi.DatasourceUpdatesResp, *openapi.Error) {
	resp, httpResp, err := c.ApiClient.SynchronizationAPI.DatasourceSyncUpdates(ctx).ClusterSyncReport(request).After(fromUpdate).ClusterId(clusterId).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		// A non-nil err on the HTTP 200 branch means the generated
		// client hit a decode error after the body was received —
		// resp may be partially populated. Returning (resp, nil)
		// here would silently report that half-decoded response as
		// success and advance the sync counter past whatever the
		// decoder dropped, so surface the decode error instead.
		if err != nil {
			return nil, openapi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("decode response: %v", err))
		}
		if resp != nil {
			return resp, nil
		}
		// Status was OK but the generated client decoded a nil typed
		// response (e.g. empty body on a schema that expected a payload).
		// Surface it as a real 500 rather than letting the final return
		// below format "<nil>" into the 503 message.
		return nil, openapi.NewError(http.StatusInternalServerError,
			"empty response body with status OK")
	}

	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return nil, openapi.NewError(http.StatusServiceUnavailable,
		"no response from assets-service and no transport error")
}

// K8sObjectsSyncUpdates wraps the k8s-objects sync endpoint. The endpoint is a
// single family-wide feed for every k8s-object asset kind (resource-claim-template
// today, more to follow) — one call fetches updates for all kinds the control
// plane knows about and the agent dispatches per-kind on the response.
func (c *AssetsClient) K8sObjectsSyncUpdates(ctx context.Context, clusterId string, fromUpdate int64, request openapi.ClusterSyncReport) (*openapi.K8sObjectsUpdatesResp, *openapi.Error) {
	resp, httpResp, err := c.ApiClient.SynchronizationAPI.K8sObjectsSyncUpdates(ctx).ClusterSyncReport(request).After(fromUpdate).ClusterId(clusterId).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		// A non-nil err on the HTTP 200 branch means the generated
		// client hit a decode error after the body was received —
		// resp may be partially populated. Returning (resp, nil)
		// here would silently report that half-decoded response as
		// success and advance the sync counter past whatever the
		// decoder dropped, so surface the decode error instead.
		if err != nil {
			return nil, openapi.NewError(http.StatusInternalServerError,
				fmt.Sprintf("decode response: %v", err))
		}
		if resp != nil {
			return resp, nil
		}
		// Status was OK but the generated client decoded a nil typed
		// response (e.g. empty body on a schema that expected a payload).
		// Surface it as a real 500 rather than letting the final return
		// below format "<nil>" into the 503 message.
		return nil, openapi.NewError(http.StatusInternalServerError,
			"empty response body with status OK")
	}

	if err != nil {
		return nil, openapi.NewError(http.StatusServiceUnavailable, err.Error())
	}
	return nil, openapi.NewError(http.StatusServiceUnavailable,
		"no response from assets-service and no transport error")
}

func (c *AssetsClient) ListTemplatesOfTenant(ctx context.Context) (*openapi.WorkloadTemplateListResponse, *openapi.Error) {
	resp, httpResp, err := c.ApiClient.TemplateAPI.ListTemplates(ctx).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload of list templates: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		if resp != nil {
			return resp, nil
		}
	}

	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request list of templates: %v", err))
}

func (c *AssetsClient) ListCredentialsAssets(ctx context.Context, privateOnly bool) (*openapi.CredentialsListResponse, *openapi.Error) {
	resp, httpResp, err := c.ApiClient.CredentialsAPI.ListCredentialsAssets(ctx).PrivateOnly(privateOnly).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload of list credentials: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		if resp != nil {
			return resp, nil
		}
	}

	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request list of credentials: %v", err))
}

func (c *AssetsClient) SyncEvents(ctx context.Context, request openapi.EventsSyncRequest) (*http.Response, error) {
	httpResp, err := c.ApiClient.EventsAPI.SyncEvents(ctx).EventsSyncRequest(request).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if err != nil {
			return httpResp, err
		}

		return httpResp, nil
	}

	return nil, err
}

func (c *AssetsClient) SyncAssetsResourcesIds(ctx context.Context, request openapi.AssetResourcesIdsSyncRequest) (*http.Response, error) {
	httpResp, err := c.ApiClient.SynchronizationAPI.SyncAssetsResourcesIds(ctx).AssetResourcesIdsSyncRequest(request).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}
	}

	return httpResp, err
}

func (s *AssetsClient) GetNgcApiKeyByIdInternal(ctx context.Context, assetId string) (*openapi.NgcApiKeyInternal, *openapi.Error) {
	resp, httpResp, err := s.ApiClient.CredentialsAPI.GetNgcApiKeyByIdInternal(ctx, assetId).Execute()
	if httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		if httpResp.StatusCode != http.StatusOK {
			// Read the body
			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				return nil, openapi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to read response payload of list credentials: %v", err.Error()))
			}

			var apiError openapi.Error

			jsonErr := json.Unmarshal(body, &apiError)
			if jsonErr != nil {
				return nil, openapi.NewError(int32(httpResp.StatusCode), string(body))
			} else {
				return nil, &apiError
			}
		}

		if resp != nil {
			return resp, nil
		}
	}

	return nil, openapi.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to request ngc api key by id internal: %v", err))
}
