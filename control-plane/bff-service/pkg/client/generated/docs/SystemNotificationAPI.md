# \SystemNotificationAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteSystemNotification**](SystemNotificationAPI.md#DeleteSystemNotification) | **Delete** /api/v1/system-notification | Delete system notification
[**GetSystemNotification**](SystemNotificationAPI.md#GetSystemNotification) | **Get** /api/v1/system-notification | Get system notification
[**UpdateSystemNotification**](SystemNotificationAPI.md#UpdateSystemNotification) | **Put** /api/v1/system-notification | Update system notification



## DeleteSystemNotification

> DeleteSystemNotification(ctx).Execute()

Delete system notification

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SystemNotificationAPI.DeleteSystemNotification(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SystemNotificationAPI.DeleteSystemNotification``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSystemNotificationRequest struct via the builder pattern


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


## GetSystemNotification

> GetSystemNotification GetSystemNotification(ctx).Execute()

Get system notification

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SystemNotificationAPI.GetSystemNotification(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SystemNotificationAPI.GetSystemNotification``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSystemNotification`: GetSystemNotification
	fmt.Fprintf(os.Stdout, "Response from `SystemNotificationAPI.GetSystemNotification`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetSystemNotificationRequest struct via the builder pattern


### Return type

[**GetSystemNotification**](GetSystemNotification.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSystemNotification

> UpdateSystemNotification(ctx).UpdateSystemNotification(updateSystemNotification).Execute()

Update system notification

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client"
)

func main() {
	updateSystemNotification := *openapiclient.NewUpdateSystemNotification() // UpdateSystemNotification | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SystemNotificationAPI.UpdateSystemNotification(context.Background()).UpdateSystemNotification(updateSystemNotification).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SystemNotificationAPI.UpdateSystemNotification``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSystemNotificationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateSystemNotification** | [**UpdateSystemNotification**](UpdateSystemNotification.md) |  | 

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

