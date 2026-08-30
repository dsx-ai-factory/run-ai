# \SettingsAPI

All URIs are relative to *https://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetClusterSettingByKey**](SettingsAPI.md#GetClusterSettingByKey) | **Get** /api/v1/clusters/settings/{key} | Get cluster setting by key
[**GetClusterSettings**](SettingsAPI.md#GetClusterSettings) | **Get** /api/v1/clusters/settings | Get cluster settings
[**GetNodePoolSettingByKey**](SettingsAPI.md#GetNodePoolSettingByKey) | **Get** /api/v1/node-pools/settings/{key} | Get node pool setting by key
[**GetNodePoolSettings**](SettingsAPI.md#GetNodePoolSettings) | **Get** /api/v1/node-pools/settings | Get node pool settings
[**UpdateClusterSettingByKey**](SettingsAPI.md#UpdateClusterSettingByKey) | **Put** /api/v1/clusters/settings/{key} | Update cluster setting by key
[**UpdateNodePoolSettingByKey**](SettingsAPI.md#UpdateNodePoolSettingByKey) | **Put** /api/v1/node-pools/settings/{key} | Update node pool setting by key



## GetClusterSettingByKey

> EnableAdvancedGpuMetricsSetting GetClusterSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get cluster setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	key := openapiclient.ClusterSettingsKeyEnum("enableAdvancedGpuMetrics") // ClusterSettingsKeyEnum | The settings key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetClusterSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetClusterSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterSettingByKey`: EnableAdvancedGpuMetricsSetting
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetClusterSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**ClusterSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**EnableAdvancedGpuMetricsSetting**](EnableAdvancedGpuMetricsSetting.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetClusterSettings

> ClusterSettingsGetResponse GetClusterSettings(ctx).TenantId(tenantId).Execute()

Get cluster settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetClusterSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetClusterSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetClusterSettings`: ClusterSettingsGetResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetClusterSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetClusterSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**ClusterSettingsGetResponse**](ClusterSettingsGetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodePoolSettingByKey

> NodePoolSetting GetNodePoolSettingByKey(ctx, key).TenantId(tenantId).Execute()

Get node pool setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	key := openapiclient.NodePoolSettingsKeyEnum("projectManualOverQuota") // NodePoolSettingsKeyEnum | The settings key
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetNodePoolSettingByKey(context.Background(), key).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetNodePoolSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodePoolSettingByKey`: NodePoolSetting
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetNodePoolSettingByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**NodePoolSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNodePoolSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**NodePoolSetting**](NodePoolSetting.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNodePoolSettings

> NodePoolSettingsGetResponse GetNodePoolSettings(ctx).TenantId(tenantId).Execute()

Get node pool settings

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SettingsAPI.GetNodePoolSettings(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.GetNodePoolSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNodePoolSettings`: NodePoolSettingsGetResponse
	fmt.Fprintf(os.Stdout, "Response from `SettingsAPI.GetNodePoolSettings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNodePoolSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**NodePoolSettingsGetResponse**](NodePoolSettingsGetResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateClusterSettingByKey

> UpdateClusterSettingByKey(ctx, key).UpdateClusterSettingByKeyRequest(updateClusterSettingByKeyRequest).TenantId(tenantId).Execute()

Update cluster setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	key := openapiclient.ClusterSettingsKeyEnum("enableAdvancedGpuMetrics") // ClusterSettingsKeyEnum | The settings key
	updateClusterSettingByKeyRequest := *openapiclient.NewUpdateClusterSettingByKeyRequest() // UpdateClusterSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateClusterSettingByKey(context.Background(), key).UpdateClusterSettingByKeyRequest(updateClusterSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateClusterSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**ClusterSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateClusterSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateClusterSettingByKeyRequest** | [**UpdateClusterSettingByKeyRequest**](UpdateClusterSettingByKeyRequest.md) |  | 
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNodePoolSettingByKey

> UpdateNodePoolSettingByKey(ctx, key).UpdateNodePoolSettingByKeyRequest(updateNodePoolSettingByKeyRequest).TenantId(tenantId).Execute()

Update node pool setting by key

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
)

func main() {
	key := openapiclient.NodePoolSettingsKeyEnum("projectManualOverQuota") // NodePoolSettingsKeyEnum | The settings key
	updateNodePoolSettingByKeyRequest := *openapiclient.NewUpdateNodePoolSettingByKeyRequest() // UpdateNodePoolSettingByKeyRequest | 
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SettingsAPI.UpdateNodePoolSettingByKey(context.Background(), key).UpdateNodePoolSettingByKeyRequest(updateNodePoolSettingByKeyRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SettingsAPI.UpdateNodePoolSettingByKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**key** | [**NodePoolSettingsKeyEnum**](.md) | The settings key | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNodePoolSettingByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateNodePoolSettingByKeyRequest** | [**UpdateNodePoolSettingByKeyRequest**](UpdateNodePoolSettingByKeyRequest.md) |  | 
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

