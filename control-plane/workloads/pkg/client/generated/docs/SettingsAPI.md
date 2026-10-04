# \SettingsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWorkloadsServiceSettingByKey**](SettingsAPI.md#GetWorkloadsServiceSettingByKey) | **Get** /api/v1/workloads/settings/{key} | Get a workloads setting by key.
[**GetWorkloadsServiceSettings**](SettingsAPI.md#GetWorkloadsServiceSettings) | **Get** /api/v1/workloads/settings | Get all workloads settings.
[**UpdateWorkloadsServiceSettingByKey**](SettingsAPI.md#UpdateWorkloadsServiceSettingByKey) | **Put** /api/v1/workloads/settings/{key} | Update a workloads setting by key.



## GetWorkloadsServiceSettingByKey

> GetWorkloadsSetting GetWorkloadsServiceSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get a workloads setting by key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	key := openapiclient.WorkloadsSettingsKeyEnum("enablePolicies") // WorkloadsSettingsKeyEnum | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetWorkloadsServiceSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetWorkloadsServiceSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadsServiceSettingByKey`: GetWorkloadsSetting
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetWorkloadsServiceSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**WorkloadsSettingsKeyEnum**](.md) |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsServiceSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**GetWorkloadsSetting**](GetWorkloadsSetting.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkloadsServiceSettings

> WorkloadsSettingsGetResponse GetWorkloadsServiceSettings(ctx).TenantId(tenantId).Execute()

Get all workloads settings.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetWorkloadsServiceSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetWorkloadsServiceSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkloadsServiceSettings`: WorkloadsSettingsGetResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetWorkloadsServiceSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkloadsServiceSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**WorkloadsSettingsGetResponse**](WorkloadsSettingsGetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateWorkloadsServiceSettingByKey

> UpdateWorkloadsSettingByKeyRequest UpdateWorkloadsServiceSettingByKey(ctx, key).UpdateWorkloadsSettingByKeyRequest(updateWorkloadsSettingByKeyRequest).TenantId(tenantId).Execute()

Update a workloads setting by key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client"
)

func main() {
	key := openapiclient.WorkloadsSettingsKeyEnum("enablePolicies") // WorkloadsSettingsKeyEnum | 
	updateWorkloadsSettingByKeyRequest := *openapiclient.NewUpdateWorkloadsSettingByKeyRequest() // UpdateWorkloadsSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.UpdateWorkloadsServiceSettingByKey(context.Background(), key).UpdateWorkloadsSettingByKeyRequest(updateWorkloadsSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateWorkloadsServiceSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateWorkloadsServiceSettingByKey`: UpdateWorkloadsSettingByKeyRequest
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.UpdateWorkloadsServiceSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**WorkloadsSettingsKeyEnum**](.md) |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateWorkloadsServiceSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateWorkloadsSettingByKeyRequest** | [**UpdateWorkloadsSettingByKeyRequest**](UpdateWorkloadsSettingByKeyRequest.md) |  | 
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**UpdateWorkloadsSettingByKeyRequest**](UpdateWorkloadsSettingByKeyRequest.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

