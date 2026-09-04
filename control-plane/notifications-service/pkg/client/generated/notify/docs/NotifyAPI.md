# \NotifyAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Notify**](NotifyAPI.md#Notify) | **Post** /notify | 
[**NotifyBatch**](NotifyAPI.md#NotifyBatch) | **Post** /notify_batch | 



## Notify

> string Notify(ctx).Notification(notification).Execute()





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
	notification := *openapiclient.NewNotification("12345", "EventType_example", map[string]interface{}(123)) // Notification | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotifyAPI.Notify(context.Background()).Notification(notification).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotifyAPI.Notify``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Notify`: string
	fmt.Fprintf(os.Stdout, "Response from `NotifyAPI.Notify`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiNotifyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notification** | [**Notification**](Notification.md) |  | 

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


## NotifyBatch

> NotifyBatch200Response NotifyBatch(ctx).Notification(notification).Execute()





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
	notification := []openapiclient.Notification{*openapiclient.NewNotification("12345", "EventType_example", map[string]interface{}(123))} // []Notification | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.NotifyAPI.NotifyBatch(context.Background()).Notification(notification).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `NotifyAPI.NotifyBatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `NotifyBatch`: NotifyBatch200Response
	fmt.Fprintf(os.Stdout, "Response from `NotifyAPI.NotifyBatch`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiNotifyBatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notification** | [**[]Notification**](Notification.md) |  | 

### Return type

[**NotifyBatch200Response**](NotifyBatch200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

