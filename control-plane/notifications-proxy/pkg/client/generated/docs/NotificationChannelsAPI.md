# \NotificationChannelsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNotificationChannel**](NotificationChannelsAPI.md#CreateNotificationChannel) | **Post** /api/v1/notification-channels | Create notification channel
[**DeleteNotificationChannel**](NotificationChannelsAPI.md#DeleteNotificationChannel) | **Delete** /api/v1/notification-channels/{name} | Delete notification channel
[**GetNotificationChannel**](NotificationChannelsAPI.md#GetNotificationChannel) | **Get** /api/v1/notification-channels/{name} | Get notification channel
[**GetNotificationChannels**](NotificationChannelsAPI.md#GetNotificationChannels) | **Get** /api/v1/notification-channels | Get supported Notification Channels
[**PatchNotificationChannel**](NotificationChannelsAPI.md#PatchNotificationChannel) | **Patch** /api/v1/notification-channels/{name} | Patch notification channel
[**UpdateNotificationChannel**](NotificationChannelsAPI.md#UpdateNotificationChannel) | **Put** /api/v1/notification-channels/{name} | Update notification channel
[**ValidateNotificationChannel**](NotificationChannelsAPI.md#ValidateNotificationChannel) | **Post** /api/v1/validate-notification-channel | Validate configuration of Notification Channel



## CreateNotificationChannel

> CreateNotificationChannel(ctx).NotificationChannel(notificationChannel).Execute()

Create notification channel



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client"
)

func main() {
	notificationChannel := *openapiclient.NewNotificationChannel(openapiclient.NotificationChannelType("email"), *openapiclient.NewNotificationChannelConfig()) // NotificationChannel |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NotificationChannelsAPI.CreateNotificationChannel(context.Background()).NotificationChannel(notificationChannel).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationChannelsAPI.CreateNotificationChannel``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNotificationChannelRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notificationChannel** | [**NotificationChannel**](NotificationChannel.md) |  | 

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


## DeleteNotificationChannel

> DeleteNotificationChannel(ctx, name).Execute()

Delete notification channel



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client"
)

func main() {
	name := "name_example" // string | Notification Channel name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NotificationChannelsAPI.DeleteNotificationChannel(context.Background(), name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationChannelsAPI.DeleteNotificationChannel``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Notification Channel name | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNotificationChannelRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## GetNotificationChannel

> NotificationChannelResponse GetNotificationChannel(ctx, name).Execute()

Get notification channel



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client"
)

func main() {
	name := "name_example" // string | Notification Channel name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotificationChannelsAPI.GetNotificationChannel(context.Background(), name).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationChannelsAPI.GetNotificationChannel``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNotificationChannel`: NotificationChannelResponse
	fmt.Fprintf(os.Stdout, "Response from `NotificationChannelsAPI.GetNotificationChannel`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Notification Channel name | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNotificationChannelRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**NotificationChannelResponse**](NotificationChannelResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNotificationChannels

> []NotificationChannelMeta GetNotificationChannels(ctx).Execute()

Get supported Notification Channels



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotificationChannelsAPI.GetNotificationChannels(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationChannelsAPI.GetNotificationChannels``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNotificationChannels`: []NotificationChannelMeta
	fmt.Fprintf(os.Stdout, "Response from `NotificationChannelsAPI.GetNotificationChannels`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetNotificationChannelsRequest struct via the builder pattern


### Return type

[**[]NotificationChannelMeta**](NotificationChannelMeta.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchNotificationChannel

> PatchNotificationChannel(ctx, name).NotificationChannelForPatch(notificationChannelForPatch).Execute()

Patch notification channel



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client"
)

func main() {
	name := "name_example" // string | Notification Channel name
	notificationChannelForPatch := *openapiclient.NewNotificationChannelForPatch() // NotificationChannelForPatch |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NotificationChannelsAPI.PatchNotificationChannel(context.Background(), name).NotificationChannelForPatch(notificationChannelForPatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationChannelsAPI.PatchNotificationChannel``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Notification Channel name | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchNotificationChannelRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **notificationChannelForPatch** | [**NotificationChannelForPatch**](NotificationChannelForPatch.md) |  | 

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


## UpdateNotificationChannel

> UpdateNotificationChannel(ctx, name).NotificationChannel(notificationChannel).Execute()

Update notification channel



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client"
)

func main() {
	name := "name_example" // string | Notification Channel name
	notificationChannel := *openapiclient.NewNotificationChannel(openapiclient.NotificationChannelType("email"), *openapiclient.NewNotificationChannelConfig()) // NotificationChannel |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.NotificationChannelsAPI.UpdateNotificationChannel(context.Background(), name).NotificationChannel(notificationChannel).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationChannelsAPI.UpdateNotificationChannel``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**name** | **string** | Notification Channel name | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNotificationChannelRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **notificationChannel** | [**NotificationChannel**](NotificationChannel.md) |  | 

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


## ValidateNotificationChannel

> bool ValidateNotificationChannel(ctx).NotificationChannelForValidate(notificationChannelForValidate).Execute()

Validate configuration of Notification Channel



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client"
)

func main() {
	notificationChannelForValidate := *openapiclient.NewNotificationChannelForValidate(openapiclient.NotificationChannelType("email"), *openapiclient.NewNotificationChannelConfig()) // NotificationChannelForValidate |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotificationChannelsAPI.ValidateNotificationChannel(context.Background()).NotificationChannelForValidate(notificationChannelForValidate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationChannelsAPI.ValidateNotificationChannel``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ValidateNotificationChannel`: bool
	fmt.Fprintf(os.Stdout, "Response from `NotificationChannelsAPI.ValidateNotificationChannel`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiValidateNotificationChannelRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notificationChannelForValidate** | [**NotificationChannelForValidate**](NotificationChannelForValidate.md) |  | 

### Return type

**bool**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

