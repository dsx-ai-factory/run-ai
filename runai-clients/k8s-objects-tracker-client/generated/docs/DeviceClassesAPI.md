# \DeviceClassesAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateDeviceClass**](DeviceClassesAPI.md#CreateDeviceClass) | **Post** /api/v1/device-classes | create a Device Class
[**DeleteDeviceClass**](DeviceClassesAPI.md#DeleteDeviceClass) | **Delete** /api/v1/device-classes | delete a Device Class
[**GetDeviceClasses**](DeviceClassesAPI.md#GetDeviceClasses) | **Get** /api/v1/device-classes | get device classes for a given cluster
[**SyncDeviceClass**](DeviceClassesAPI.md#SyncDeviceClass) | **Post** /api/v1/device-classes/sync | Sync cluster Device Classes
[**UpdateDeviceClass**](DeviceClassesAPI.md#UpdateDeviceClass) | **Put** /api/v1/device-classes | update a Device Class



## CreateDeviceClass

> DeviceClass CreateDeviceClass(ctx).DeviceClass(deviceClass).Execute()

create a Device Class

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	deviceClass := *openapiclient.NewDeviceClass("gpu.nvidia.com") // DeviceClass | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceClassesAPI.CreateDeviceClass(context.Background()).DeviceClass(deviceClass).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceClassesAPI.CreateDeviceClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDeviceClass`: DeviceClass
	fmt.Fprintf(os.Stdout, "Response from `DeviceClassesAPI.CreateDeviceClass`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDeviceClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deviceClass** | [**DeviceClass**](DeviceClass.md) |  | 

### Return type

[**DeviceClass**](DeviceClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDeviceClass

> DeleteDeviceClass(ctx).ClusterId(clusterId).Name(name).Execute()

delete a Device Class

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	name := "name_example" // string | filter by name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeviceClassesAPI.DeleteDeviceClass(context.Background()).ClusterId(clusterId).Name(name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceClassesAPI.DeleteDeviceClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDeviceClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **name** | **string** | filter by name | 

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


## GetDeviceClasses

> DeviceClasses GetDeviceClasses(ctx).ClusterId(clusterId).Name(name).Execute()

get device classes for a given cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	name := "name_example" // string | filter by name (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceClassesAPI.GetDeviceClasses(context.Background()).ClusterId(clusterId).Name(name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceClassesAPI.GetDeviceClasses``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDeviceClasses`: DeviceClasses
	fmt.Fprintf(os.Stdout, "Response from `DeviceClassesAPI.GetDeviceClasses`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDeviceClassesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **name** | **string** | filter by name | 

### Return type

[**DeviceClasses**](DeviceClasses.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDeviceClass

> DeviceClasses SyncDeviceClass(ctx).ClusterId(clusterId).DeviceClasses(deviceClasses).Execute()

Sync cluster Device Classes



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | The id of the cluster
	deviceClasses := *openapiclient.NewDeviceClasses([]openapiclient.DeviceClass{*openapiclient.NewDeviceClass("gpu.nvidia.com")}) // DeviceClasses | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceClassesAPI.SyncDeviceClass(context.Background()).ClusterId(clusterId).DeviceClasses(deviceClasses).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceClassesAPI.SyncDeviceClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDeviceClass`: DeviceClasses
	fmt.Fprintf(os.Stdout, "Response from `DeviceClassesAPI.SyncDeviceClass`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncDeviceClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The id of the cluster | 
 **deviceClasses** | [**DeviceClasses**](DeviceClasses.md) |  | 

### Return type

[**DeviceClasses**](DeviceClasses.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDeviceClass

> DeviceClass UpdateDeviceClass(ctx).DeviceClass(deviceClass).Execute()

update a Device Class

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
)

func main() {
	deviceClass := *openapiclient.NewDeviceClass("gpu.nvidia.com") // DeviceClass | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeviceClassesAPI.UpdateDeviceClass(context.Background()).DeviceClass(deviceClass).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeviceClassesAPI.UpdateDeviceClass``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDeviceClass`: DeviceClass
	fmt.Fprintf(os.Stdout, "Response from `DeviceClassesAPI.UpdateDeviceClass`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDeviceClassRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deviceClass** | [**DeviceClass**](DeviceClass.md) |  | 

### Return type

[**DeviceClass**](DeviceClass.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

