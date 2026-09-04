# \NotificationTypesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetNotificationTypes**](NotificationTypesAPI.md#GetNotificationTypes) | **Get** /api/v1/notification-types | Get notification types



## GetNotificationTypes

> []NotificationType GetNotificationTypes(ctx).Execute()

Get notification types



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
	resp, r, err := apiClient.NotificationTypesAPI.GetNotificationTypes(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotificationTypesAPI.GetNotificationTypes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNotificationTypes`: []NotificationType
	fmt.Fprintf(os.Stdout, "Response from `NotificationTypesAPI.GetNotificationTypes`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetNotificationTypesRequest struct via the builder pattern


### Return type

[**[]NotificationType**](NotificationType.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

