# \StorageClassConfigurationAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateStorageClassConfiguration**](StorageClassConfigurationAPI.md#CreateStorageClassConfiguration) | **Post** /api/v1/storage-class-configuration | Create storage class configuration
[**DeleteStorageClassConfigurationById**](StorageClassConfigurationAPI.md#DeleteStorageClassConfigurationById) | **Delete** /api/v1/storage-class-configuration/{id} | Delete storage class configuration by id
[**GetStorageClassConfigurationById**](StorageClassConfigurationAPI.md#GetStorageClassConfigurationById) | **Get** /api/v1/storage-class-configuration/{id} | Get storage class configuration by id
[**ListStorageClassConfiguration**](StorageClassConfigurationAPI.md#ListStorageClassConfiguration) | **Get** /api/v1/storage-class-configuration | List storage class configurations
[**ListStorageClassConfigurationInt**](StorageClassConfigurationAPI.md#ListStorageClassConfigurationInt) | **Get** /internal/api/v1/storage-class-configuration | list of storage class configuration (internal)
[**PatchStorageClassConfigurationById**](StorageClassConfigurationAPI.md#PatchStorageClassConfigurationById) | **Patch** /api/v1/storage-class-configuration/{id} | Patch storage class configuration by id
[**UpdateStorageClassConfigurationById**](StorageClassConfigurationAPI.md#UpdateStorageClassConfigurationById) | **Put** /api/v1/storage-class-configuration/{id} | Update storage class configuration by id



## CreateStorageClassConfiguration

> StorageClassConfiguration CreateStorageClassConfiguration(ctx).StorageClassConfigurationCreationRequest(storageClassConfigurationCreationRequest).Execute()

Create storage class configuration



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	storageClassConfigurationCreationRequest := *openapiclient.NewStorageClassConfigurationCreationRequest("cluster", "ScopeId_example", "Standard") // StorageClassConfigurationCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassConfigurationAPI.CreateStorageClassConfiguration(context.Background()).StorageClassConfigurationCreationRequest(storageClassConfigurationCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassConfigurationAPI.CreateStorageClassConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateStorageClassConfiguration`: StorageClassConfiguration
	fmt.Fprintf(os.Stdout, "Response from `StorageClassConfigurationAPI.CreateStorageClassConfiguration`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateStorageClassConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **storageClassConfigurationCreationRequest** | [**StorageClassConfigurationCreationRequest**](StorageClassConfigurationCreationRequest.md) |  | 

### Return type

[**StorageClassConfiguration**](StorageClassConfiguration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteStorageClassConfigurationById

> DeleteStorageClassConfigurationById(ctx, id).TenantId(tenantId).Execute()

Delete storage class configuration by id



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	id := "id_example" // string | unique identifier of the storage class configuration
	tenantId := int32(56) // int32 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.StorageClassConfigurationAPI.DeleteStorageClassConfigurationById(context.Background(), id).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassConfigurationAPI.DeleteStorageClassConfigurationById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | unique identifier of the storage class configuration | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteStorageClassConfigurationByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **int32** |  | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetStorageClassConfigurationById

> StorageClassConfiguration GetStorageClassConfigurationById(ctx, id).TenantId(tenantId).Execute()

Get storage class configuration by id



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	id := "id_example" // string | unique identifier of the storage class configuration
	tenantId := int32(56) // int32 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassConfigurationAPI.GetStorageClassConfigurationById(context.Background(), id).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassConfigurationAPI.GetStorageClassConfigurationById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetStorageClassConfigurationById`: StorageClassConfiguration
	fmt.Fprintf(os.Stdout, "Response from `StorageClassConfigurationAPI.GetStorageClassConfigurationById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | unique identifier of the storage class configuration | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetStorageClassConfigurationByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **tenantId** | **int32** |  | 

### Return type

[**StorageClassConfiguration**](StorageClassConfiguration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListStorageClassConfiguration

> StorageClassConfigurationListResponse ListStorageClassConfiguration(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()

List storage class configurations



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassConfigurationAPI.ListStorageClassConfiguration(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassConfigurationAPI.ListStorageClassConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListStorageClassConfiguration`: StorageClassConfigurationListResponse
	fmt.Fprintf(os.Stdout, "Response from `StorageClassConfigurationAPI.ListStorageClassConfiguration`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListStorageClassConfigurationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | **string** | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**StorageClassConfigurationListResponse**](StorageClassConfigurationListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListStorageClassConfigurationInt

> StorageClassConfigurationListResponse ListStorageClassConfigurationInt(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()

list of storage class configuration (internal)

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassConfigurationAPI.ListStorageClassConfigurationInt(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassConfigurationAPI.ListStorageClassConfigurationInt``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListStorageClassConfigurationInt`: StorageClassConfigurationListResponse
	fmt.Fprintf(os.Stdout, "Response from `StorageClassConfigurationAPI.ListStorageClassConfigurationInt`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListStorageClassConfigurationIntRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **sortBy** | **string** | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**StorageClassConfigurationListResponse**](StorageClassConfigurationListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchStorageClassConfigurationById

> StorageClassConfiguration PatchStorageClassConfigurationById(ctx, id).StorageClassConfigurationPatchRequest(storageClassConfigurationPatchRequest).TenantId(tenantId).Execute()

Patch storage class configuration by id



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	id := "id_example" // string | unique identifier of the storage class configuration
	storageClassConfigurationPatchRequest := *openapiclient.NewStorageClassConfigurationPatchRequest() // StorageClassConfigurationPatchRequest | 
	tenantId := int32(56) // int32 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassConfigurationAPI.PatchStorageClassConfigurationById(context.Background(), id).StorageClassConfigurationPatchRequest(storageClassConfigurationPatchRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassConfigurationAPI.PatchStorageClassConfigurationById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchStorageClassConfigurationById`: StorageClassConfiguration
	fmt.Fprintf(os.Stdout, "Response from `StorageClassConfigurationAPI.PatchStorageClassConfigurationById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | unique identifier of the storage class configuration | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchStorageClassConfigurationByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **storageClassConfigurationPatchRequest** | [**StorageClassConfigurationPatchRequest**](StorageClassConfigurationPatchRequest.md) |  | 
 **tenantId** | **int32** |  | 

### Return type

[**StorageClassConfiguration**](StorageClassConfiguration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateStorageClassConfigurationById

> StorageClassConfiguration UpdateStorageClassConfigurationById(ctx, id).StorageClassConfigurationUpdateRequest(storageClassConfigurationUpdateRequest).TenantId(tenantId).Execute()

Update storage class configuration by id



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	id := "id_example" // string | unique identifier of the storage class configuration
	storageClassConfigurationUpdateRequest := *openapiclient.NewStorageClassConfigurationUpdateRequest() // StorageClassConfigurationUpdateRequest | 
	tenantId := int32(56) // int32 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.StorageClassConfigurationAPI.UpdateStorageClassConfigurationById(context.Background(), id).StorageClassConfigurationUpdateRequest(storageClassConfigurationUpdateRequest).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `StorageClassConfigurationAPI.UpdateStorageClassConfigurationById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateStorageClassConfigurationById`: StorageClassConfiguration
	fmt.Fprintf(os.Stdout, "Response from `StorageClassConfigurationAPI.UpdateStorageClassConfigurationById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | unique identifier of the storage class configuration | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateStorageClassConfigurationByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **storageClassConfigurationUpdateRequest** | [**StorageClassConfigurationUpdateRequest**](StorageClassConfigurationUpdateRequest.md) |  | 
 **tenantId** | **int32** |  | 

### Return type

[**StorageClassConfiguration**](StorageClassConfiguration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

