# \EmailAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateEmail**](EmailAPI.md#CreateEmail) | **Post** /api/v1/email | Create email notification channel
[**DeleteEmail**](EmailAPI.md#DeleteEmail) | **Delete** /api/v1/email | Delete email notification channel
[**GetEmail**](EmailAPI.md#GetEmail) | **Get** /api/v1/email | Get email notification channel
[**GetEmailNotificationState**](EmailAPI.md#GetEmailNotificationState) | **Get** /api/v1/email/state | Get email notification state
[**PatchEmail**](EmailAPI.md#PatchEmail) | **Patch** /api/v1/email | Patch email notification channel
[**UpdateEmail**](EmailAPI.md#UpdateEmail) | **Put** /api/v1/email | Update email notification channel
[**UpdateEmailNotificationState**](EmailAPI.md#UpdateEmailNotificationState) | **Patch** /api/v1/email/state | Update email notification state
[**ValidateEmail**](EmailAPI.md#ValidateEmail) | **Post** /api/v1/email/validate | Validate configuration of Email Channel



## CreateEmail

> EmailChannel CreateEmail(ctx).EmailChannel(emailChannel).Execute()

Create email notification channel



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
	emailChannel := *openapiclient.NewEmailChannel(*openapiclient.NewEmailChannelConfig("SmtpHost_example")) // EmailChannel |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EmailAPI.CreateEmail(context.Background()).EmailChannel(emailChannel).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.CreateEmail``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateEmail`: EmailChannel
	fmt.Fprintf(os.Stdout, "Response from `EmailAPI.CreateEmail`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateEmailRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **emailChannel** | [**EmailChannel**](EmailChannel.md) |  | 

### Return type

[**EmailChannel**](EmailChannel.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteEmail

> DeleteEmail(ctx).Execute()

Delete email notification channel



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
	r, err := apiClient.EmailAPI.DeleteEmail(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.DeleteEmail``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteEmailRequest struct via the builder pattern


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


## GetEmail

> EmailChannel GetEmail(ctx).Execute()

Get email notification channel



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
	resp, r, err := apiClient.EmailAPI.GetEmail(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.GetEmail``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetEmail`: EmailChannel
	fmt.Fprintf(os.Stdout, "Response from `EmailAPI.GetEmail`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetEmailRequest struct via the builder pattern


### Return type

[**EmailChannel**](EmailChannel.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetEmailNotificationState

> NotificationStateStatus GetEmailNotificationState(ctx).Execute()

Get email notification state



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
	resp, r, err := apiClient.EmailAPI.GetEmailNotificationState(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.GetEmailNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetEmailNotificationState`: NotificationStateStatus
	fmt.Fprintf(os.Stdout, "Response from `EmailAPI.GetEmailNotificationState`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetEmailNotificationStateRequest struct via the builder pattern


### Return type

[**NotificationStateStatus**](NotificationStateStatus.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchEmail

> EmailChannel PatchEmail(ctx).EmailChannelForPatch(emailChannelForPatch).Execute()

Patch email notification channel



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
	emailChannelForPatch := *openapiclient.NewEmailChannelForPatch(*openapiclient.NewEmailChannelConfigForPatch()) // EmailChannelForPatch |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EmailAPI.PatchEmail(context.Background()).EmailChannelForPatch(emailChannelForPatch).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.PatchEmail``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchEmail`: EmailChannel
	fmt.Fprintf(os.Stdout, "Response from `EmailAPI.PatchEmail`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPatchEmailRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **emailChannelForPatch** | [**EmailChannelForPatch**](EmailChannelForPatch.md) |  | 

### Return type

[**EmailChannel**](EmailChannel.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateEmail

> EmailChannel UpdateEmail(ctx).EmailChannel(emailChannel).Execute()

Update email notification channel



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
	emailChannel := *openapiclient.NewEmailChannel(*openapiclient.NewEmailChannelConfig("SmtpHost_example")) // EmailChannel |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EmailAPI.UpdateEmail(context.Background()).EmailChannel(emailChannel).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.UpdateEmail``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateEmail`: EmailChannel
	fmt.Fprintf(os.Stdout, "Response from `EmailAPI.UpdateEmail`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateEmailRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **emailChannel** | [**EmailChannel**](EmailChannel.md) |  | 

### Return type

[**EmailChannel**](EmailChannel.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateEmailNotificationState

> NotificationStateStatus UpdateEmailNotificationState(ctx).NotificationStateStatus(notificationStateStatus).Execute()

Update email notification state



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
	notificationStateStatus := *openapiclient.NewNotificationStateStatus() // NotificationStateStatus |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EmailAPI.UpdateEmailNotificationState(context.Background()).NotificationStateStatus(notificationStateStatus).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.UpdateEmailNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateEmailNotificationState`: NotificationStateStatus
	fmt.Fprintf(os.Stdout, "Response from `EmailAPI.UpdateEmailNotificationState`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateEmailNotificationStateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notificationStateStatus** | [**NotificationStateStatus**](NotificationStateStatus.md) |  | 

### Return type

[**NotificationStateStatus**](NotificationStateStatus.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ValidateEmail

> ValidateSinkResponse ValidateEmail(ctx).EmailChannelConfig(emailChannelConfig).Execute()

Validate configuration of Email Channel



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
	emailChannelConfig := *openapiclient.NewEmailChannelConfig("SmtpHost_example") // EmailChannelConfig |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EmailAPI.ValidateEmail(context.Background()).EmailChannelConfig(emailChannelConfig).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmailAPI.ValidateEmail``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ValidateEmail`: ValidateSinkResponse
	fmt.Fprintf(os.Stdout, "Response from `EmailAPI.ValidateEmail`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiValidateEmailRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **emailChannelConfig** | [**EmailChannelConfig**](EmailChannelConfig.md) |  | 

### Return type

[**ValidateSinkResponse**](ValidateSinkResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

