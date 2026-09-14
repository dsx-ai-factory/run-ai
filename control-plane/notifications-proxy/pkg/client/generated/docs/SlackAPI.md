# \SlackAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddSlackAppDeprecated**](SlackAPI.md#AddSlackAppDeprecated) | **Get** /api/v1/notification-channels/slack/add-app | Add Slack app
[**AuthorizeSlackApp**](SlackAPI.md#AuthorizeSlackApp) | **Get** /api/v1/slack/auth | Authorize a Slack app
[**AuthorizeSlackAppDeprecated**](SlackAPI.md#AuthorizeSlackAppDeprecated) | **Get** /api/v1/notification-channels/slack/auth-app | Authorize a Slack app
[**ConnectSlackWorkspace**](SlackAPI.md#ConnectSlackWorkspace) | **Post** /api/v1/slack/workspace | Connect Slack workspace
[**CreateSlackApp**](SlackAPI.md#CreateSlackApp) | **Post** /api/v1/slack/app | Create Slack app
[**CreateSlackAppDeprecated**](SlackAPI.md#CreateSlackAppDeprecated) | **Post** /api/v1/notification-channels/slack/create-app | Create Slack app
[**DeleteSlackApp**](SlackAPI.md#DeleteSlackApp) | **Delete** /api/v1/slack/app | Delete Slack app
[**DisconnectSlackWorkspace**](SlackAPI.md#DisconnectSlackWorkspace) | **Delete** /api/v1/slack/workspace | Disconnect Slack workspace
[**GetDetailedSlackNotificationState**](SlackAPI.md#GetDetailedSlackNotificationState) | **Get** /api/v1/slack/state/detailed | Get detailed slack notification state
[**GetSlack**](SlackAPI.md#GetSlack) | **Get** /api/v1/slack | Get slack notification channel
[**GetSlackNotificationState**](SlackAPI.md#GetSlackNotificationState) | **Get** /api/v1/slack/state | Get slack notification state
[**PatchSlack**](SlackAPI.md#PatchSlack) | **Patch** /api/v1/slack | Patch slack notification channel
[**UpdateSlackNotificationState**](SlackAPI.md#UpdateSlackNotificationState) | **Patch** /api/v1/slack/state | Update slack notification state
[**ValidateSlack**](SlackAPI.md#ValidateSlack) | **Post** /api/v1/slack/validate | Validate configuration of Slack Channel



## AddSlackAppDeprecated

> ConnectSlackWorkspace200Response AddSlackAppDeprecated(ctx).Execute()

Add Slack app



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
	resp, r, err := apiClient.SlackAPI.AddSlackAppDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.AddSlackAppDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AddSlackAppDeprecated`: ConnectSlackWorkspace200Response
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.AddSlackAppDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAddSlackAppDeprecatedRequest struct via the builder pattern


### Return type

[**ConnectSlackWorkspace200Response**](ConnectSlackWorkspace200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthorizeSlackApp

> AuthorizeSlackApp(ctx).Code(code).State(state).Execute()

Authorize a Slack app



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
	code := "code_example" // string | Handles the OAuth redirect from Slack. Exchanges the authorization code for an access token, new tenant associated sink for notifications created.
	state := "state_example" // string | A CSRF token or UI redirect hint sent by the frontend.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SlackAPI.AuthorizeSlackApp(context.Background()).Code(code).State(state).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.AuthorizeSlackApp``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthorizeSlackAppRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **code** | **string** | Handles the OAuth redirect from Slack. Exchanges the authorization code for an access token, new tenant associated sink for notifications created. | 
 **state** | **string** | A CSRF token or UI redirect hint sent by the frontend. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthorizeSlackAppDeprecated

> AuthorizeSlackAppDeprecated(ctx).Code(code).State(state).Execute()

Authorize a Slack app



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
	code := "code_example" // string | Handles the OAuth redirect from Slack. Exchanges the authorization code for an access token, new tenant associated sink for notifications created.
	state := "state_example" // string | A CSRF token or UI redirect hint sent by the frontend.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SlackAPI.AuthorizeSlackAppDeprecated(context.Background()).Code(code).State(state).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.AuthorizeSlackAppDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthorizeSlackAppDeprecatedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **code** | **string** | Handles the OAuth redirect from Slack. Exchanges the authorization code for an access token, new tenant associated sink for notifications created. | 
 **state** | **string** | A CSRF token or UI redirect hint sent by the frontend. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ConnectSlackWorkspace

> ConnectSlackWorkspace200Response ConnectSlackWorkspace(ctx).Execute()

Connect Slack workspace



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
	resp, r, err := apiClient.SlackAPI.ConnectSlackWorkspace(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.ConnectSlackWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ConnectSlackWorkspace`: ConnectSlackWorkspace200Response
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.ConnectSlackWorkspace`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiConnectSlackWorkspaceRequest struct via the builder pattern


### Return type

[**ConnectSlackWorkspace200Response**](ConnectSlackWorkspace200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateSlackApp

> CreateSlackAppResponse CreateSlackApp(ctx).CreateSlackAppRequestBody(createSlackAppRequestBody).Execute()

Create Slack app



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
	createSlackAppRequestBody := *openapiclient.NewCreateSlackAppRequestBody("AccessToken_example") // CreateSlackAppRequestBody |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SlackAPI.CreateSlackApp(context.Background()).CreateSlackAppRequestBody(createSlackAppRequestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.CreateSlackApp``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateSlackApp`: CreateSlackAppResponse
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.CreateSlackApp`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSlackAppRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createSlackAppRequestBody** | [**CreateSlackAppRequestBody**](CreateSlackAppRequestBody.md) |  | 

### Return type

[**CreateSlackAppResponse**](CreateSlackAppResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateSlackAppDeprecated

> CreateSlackAppResponse CreateSlackAppDeprecated(ctx).CreateSlackAppRequestBody(createSlackAppRequestBody).Execute()

Create Slack app



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
	createSlackAppRequestBody := *openapiclient.NewCreateSlackAppRequestBody("AccessToken_example") // CreateSlackAppRequestBody |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SlackAPI.CreateSlackAppDeprecated(context.Background()).CreateSlackAppRequestBody(createSlackAppRequestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.CreateSlackAppDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateSlackAppDeprecated`: CreateSlackAppResponse
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.CreateSlackAppDeprecated`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSlackAppDeprecatedRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createSlackAppRequestBody** | [**CreateSlackAppRequestBody**](CreateSlackAppRequestBody.md) |  | 

### Return type

[**CreateSlackAppResponse**](CreateSlackAppResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteSlackApp

> DeleteSlackApp(ctx).Execute()

Delete Slack app



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
	r, err := apiClient.SlackAPI.DeleteSlackApp(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.DeleteSlackApp``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSlackAppRequest struct via the builder pattern


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


## DisconnectSlackWorkspace

> DisconnectSlackWorkspace(ctx).Execute()

Disconnect Slack workspace



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
	r, err := apiClient.SlackAPI.DisconnectSlackWorkspace(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.DisconnectSlackWorkspace``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDisconnectSlackWorkspaceRequest struct via the builder pattern


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


## GetDetailedSlackNotificationState

> DetailedSlackNotificationsState GetDetailedSlackNotificationState(ctx).Execute()

Get detailed slack notification state



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
	resp, r, err := apiClient.SlackAPI.GetDetailedSlackNotificationState(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.GetDetailedSlackNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDetailedSlackNotificationState`: DetailedSlackNotificationsState
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.GetDetailedSlackNotificationState`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetDetailedSlackNotificationStateRequest struct via the builder pattern


### Return type

[**DetailedSlackNotificationsState**](DetailedSlackNotificationsState.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSlack

> SlackNotificationChannel GetSlack(ctx).Execute()

Get slack notification channel



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
	resp, r, err := apiClient.SlackAPI.GetSlack(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.GetSlack``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSlack`: SlackNotificationChannel
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.GetSlack`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetSlackRequest struct via the builder pattern


### Return type

[**SlackNotificationChannel**](SlackNotificationChannel.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSlackNotificationState

> NotificationStateStatus GetSlackNotificationState(ctx).Execute()

Get slack notification state



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
	resp, r, err := apiClient.SlackAPI.GetSlackNotificationState(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.GetSlackNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSlackNotificationState`: NotificationStateStatus
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.GetSlackNotificationState`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetSlackNotificationStateRequest struct via the builder pattern


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


## PatchSlack

> SlackNotificationChannel PatchSlack(ctx).SlackNotificationChannelPatchRequest(slackNotificationChannelPatchRequest).Execute()

Patch slack notification channel



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
	slackNotificationChannelPatchRequest := *openapiclient.NewSlackNotificationChannelPatchRequest(*openapiclient.NewSlackNotificationChannelPatchRequestConfig()) // SlackNotificationChannelPatchRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SlackAPI.PatchSlack(context.Background()).SlackNotificationChannelPatchRequest(slackNotificationChannelPatchRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.PatchSlack``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchSlack`: SlackNotificationChannel
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.PatchSlack`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPatchSlackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **slackNotificationChannelPatchRequest** | [**SlackNotificationChannelPatchRequest**](SlackNotificationChannelPatchRequest.md) |  | 

### Return type

[**SlackNotificationChannel**](SlackNotificationChannel.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSlackNotificationState

> NotificationStateStatus UpdateSlackNotificationState(ctx).NotificationStateStatus(notificationStateStatus).Execute()

Update slack notification state



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
	resp, r, err := apiClient.SlackAPI.UpdateSlackNotificationState(context.Background()).NotificationStateStatus(notificationStateStatus).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.UpdateSlackNotificationState``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateSlackNotificationState`: NotificationStateStatus
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.UpdateSlackNotificationState`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSlackNotificationStateRequest struct via the builder pattern


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


## ValidateSlack

> ValidateSinkResponse ValidateSlack(ctx).SlackNotificationChannelConfig(slackNotificationChannelConfig).Execute()

Validate configuration of Slack Channel



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
	slackNotificationChannelConfig := *openapiclient.NewSlackNotificationChannelConfig() // SlackNotificationChannelConfig |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SlackAPI.ValidateSlack(context.Background()).SlackNotificationChannelConfig(slackNotificationChannelConfig).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SlackAPI.ValidateSlack``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ValidateSlack`: ValidateSinkResponse
	fmt.Fprintf(os.Stdout, "Response from `SlackAPI.ValidateSlack`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiValidateSlackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **slackNotificationChannelConfig** | [**SlackNotificationChannelConfig**](SlackNotificationChannelConfig.md) |  | 

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

