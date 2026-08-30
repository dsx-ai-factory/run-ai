# \EventsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SyncEvents**](EventsAPI.md#SyncEvents) | **Post** /api/v1/asset/sync/events | Sync events.



## SyncEvents

> SyncEvents(ctx).EventsSyncRequest(eventsSyncRequest).Execute()

Sync events.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	eventsSyncRequest := *openapiclient.NewEventsSyncRequest([]openapiclient.EventSyncFields{*openapiclient.NewEventSyncFields("Id_example", "71f69d83-ba66-4822-adf5-55ce55efd210", "Normal", "Started container z", "Started", time.Now(), "kubelet", *openapiclient.NewInvolvedObject("Uid_example", "Pod", "test-0-1", "runai-test"))}) // EventsSyncRequest | The events to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.EventsAPI.SyncEvents(context.Background()).EventsSyncRequest(eventsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EventsAPI.SyncEvents``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **eventsSyncRequest** | [**EventsSyncRequest**](EventsSyncRequest.md) | The events to sync. | 

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

