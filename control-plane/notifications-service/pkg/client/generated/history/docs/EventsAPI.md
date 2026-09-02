# \EventsAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteEvent**](EventsAPI.md#DeleteEvent) | **Delete** /api/v1/events/{eventId} | 
[**DeleteEvents**](EventsAPI.md#DeleteEvents) | **Delete** /api/v1/events | 
[**GetEvents**](EventsAPI.md#GetEvents) | **Get** /api/v1/events | 
[**UpdateEventRead**](EventsAPI.md#UpdateEventRead) | **Patch** /api/v1/events/{eventId} | 



## DeleteEvent

> string DeleteEvent(ctx, eventId).Execute()





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
	eventId := "eventId_example" // string | Unique identifier of the event

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EventsAPI.DeleteEvent(context.Background(), eventId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EventsAPI.DeleteEvent``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteEvent`: string
	fmt.Fprintf(os.Stdout, "Response from `EventsAPI.DeleteEvent`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**eventId** | **string** | Unique identifier of the event | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteEventRequest struct via the builder pattern


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


## DeleteEvents

> string DeleteEvents(ctx).TenantId(tenantId).UserId(userId).Read(read).Execute()





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
	tenantId := "tenantId_example" // string | Tenant that owns the events
	userId := "userId_example" // string | User within the tenant whose events are returned or deleted; only supported when tenantId is provided
	read := true // bool | When set, return or delete only events with this read status (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EventsAPI.DeleteEvents(context.Background()).TenantId(tenantId).UserId(userId).Read(read).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EventsAPI.DeleteEvents``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteEvents`: string
	fmt.Fprintf(os.Stdout, "Response from `EventsAPI.DeleteEvents`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | Tenant that owns the events | 
 **userId** | **string** | User within the tenant whose events are returned or deleted; only supported when tenantId is provided | 
 **read** | **bool** | When set, return or delete only events with this read status | 

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


## GetEvents

> []Event GetEvents(ctx).TenantId(tenantId).UserId(userId).Read(read).Execute()





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
	tenantId := "tenantId_example" // string | Tenant that owns the events (optional)
	userId := "userId_example" // string | User within the tenant whose events are returned or deleted; only supported when tenantId is provided (optional)
	read := true // bool | When set, return or delete only events with this read status (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EventsAPI.GetEvents(context.Background()).TenantId(tenantId).UserId(userId).Read(read).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EventsAPI.GetEvents``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetEvents`: []Event
	fmt.Fprintf(os.Stdout, "Response from `EventsAPI.GetEvents`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | Tenant that owns the events | 
 **userId** | **string** | User within the tenant whose events are returned or deleted; only supported when tenantId is provided | 
 **read** | **bool** | When set, return or delete only events with this read status | 

### Return type

[**[]Event**](Event.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateEventRead

> string UpdateEventRead(ctx, eventId).EventReadUpdate(eventReadUpdate).Execute()





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
	eventId := "eventId_example" // string | Unique identifier of the event
	eventReadUpdate := *openapiclient.NewEventReadUpdate() // EventReadUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EventsAPI.UpdateEventRead(context.Background(), eventId).EventReadUpdate(eventReadUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EventsAPI.UpdateEventRead``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateEventRead`: string
	fmt.Fprintf(os.Stdout, "Response from `EventsAPI.UpdateEventRead`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**eventId** | **string** | Unique identifier of the event | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateEventReadRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **eventReadUpdate** | [**EventReadUpdate**](EventReadUpdate.md) |  | 

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

