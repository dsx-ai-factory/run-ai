# \NotificationStateAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetDetailedNotificationState**](NotificationStateAPI.md#GetDetailedNotificationState) | **Get** /api/v1/notification-state/detailed | Get detailed notification state
[**GetNotificationState**](NotificationStateAPI.md#GetNotificationState) | **Get** /api/v1/notification-state | Get notification state
[**UpdateNotificationState**](NotificationStateAPI.md#UpdateNotificationState) | **Patch** /api/v1/notification-state | Update notification state



## GetDetailedNotificationState

> DetailedNotificationState GetDetailedNotificationState(ctx).Execute()

Get detailed notification state



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotificationStateAPI.GetDetailedNotificationState(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationStateAPI.GetDetailedNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDetailedNotificationState`: DetailedNotificationState
	fmt.Fprintf(os.Stdout, "Response from `NotificationStateAPI.GetDetailedNotificationState`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetDetailedNotificationStateRequest struct via the builder pattern


### Return type

[**DetailedNotificationState**](DetailedNotificationState.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNotificationState

> NotificationState GetNotificationState(ctx).Execute()

Get notification state



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotificationStateAPI.GetNotificationState(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationStateAPI.GetNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNotificationState`: NotificationState
	fmt.Fprintf(os.Stdout, "Response from `NotificationStateAPI.GetNotificationState`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetNotificationStateRequest struct via the builder pattern


### Return type

[**NotificationState**](NotificationState.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNotificationState

> NotificationState UpdateNotificationState(ctx).NotificationState(notificationState).Execute()

Update notification state



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/pkg/client"
)

func main() {
	notificationState := *openapiclient.NewNotificationState() // NotificationState |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotificationStateAPI.UpdateNotificationState(context.Background()).NotificationState(notificationState).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationStateAPI.UpdateNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNotificationState`: NotificationState
	fmt.Fprintf(os.Stdout, "Response from `NotificationStateAPI.UpdateNotificationState`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNotificationStateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notificationState** | [**NotificationState**](NotificationState.md) |  | 

### Return type

[**NotificationState**](NotificationState.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

