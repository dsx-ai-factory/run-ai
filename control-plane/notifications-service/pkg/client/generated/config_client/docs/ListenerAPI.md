# \ListenerAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateListener**](ListenerAPI.md#CreateListener) | **Post** /api/v1/listeners | 
[**DeleteListener**](ListenerAPI.md#DeleteListener) | **Delete** /api/v1/listeners/{listenerId} | 
[**GetListener**](ListenerAPI.md#GetListener) | **Get** /api/v1/listeners/{listenerId} | 
[**UpdateListener**](ListenerAPI.md#UpdateListener) | **Put** /api/v1/listeners/{listenerId} | 



## CreateListener

> string CreateListener(ctx).ListenerData(listenerData).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	listenerData := *openapiclient.NewListenerData("ListenerId_example", "ListenerType_example", map[string]interface{}(123)) // ListenerData | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ListenerAPI.CreateListener(context.Background()).ListenerData(listenerData).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ListenerAPI.CreateListener``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateListener`: string
	fmt.Fprintf(os.Stdout, "Response from `ListenerAPI.CreateListener`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateListenerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **listenerData** | [**ListenerData**](ListenerData.md) |  | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteListener

> string DeleteListener(ctx, listenerId).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	listenerId := "listenerId_example" // string | ID of relevant listener to delete

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ListenerAPI.DeleteListener(context.Background(), listenerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ListenerAPI.DeleteListener``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteListener`: string
	fmt.Fprintf(os.Stdout, "Response from `ListenerAPI.DeleteListener`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**listenerId** | **string** | ID of relevant listener to delete | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteListenerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetListener

> ListenerData GetListener(ctx, listenerId).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	listenerId := "listenerId_example" // string | ID of relevant listener to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ListenerAPI.GetListener(context.Background(), listenerId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ListenerAPI.GetListener``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetListener`: ListenerData
	fmt.Fprintf(os.Stdout, "Response from `ListenerAPI.GetListener`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**listenerId** | **string** | ID of relevant listener to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetListenerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ListenerData**](ListenerData.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateListener

> string UpdateListener(ctx, listenerId).ListenerDataForUpdate(listenerDataForUpdate).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	listenerId := "listenerId_example" // string | ID of relevant listener to update
	listenerDataForUpdate := *openapiclient.NewListenerDataForUpdate("ListenerType_example", map[string]interface{}(123)) // ListenerDataForUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ListenerAPI.UpdateListener(context.Background(), listenerId).ListenerDataForUpdate(listenerDataForUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ListenerAPI.UpdateListener``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateListener`: string
	fmt.Fprintf(os.Stdout, "Response from `ListenerAPI.UpdateListener`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**listenerId** | **string** | ID of relevant listener to update | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateListenerRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **listenerDataForUpdate** | [**ListenerDataForUpdate**](ListenerDataForUpdate.md) |  | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

