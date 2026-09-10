# \PodGroupsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SyncPodGroups**](PodGroupsAPI.md#SyncPodGroups) | **Post** /api/v1/workloads/podgroups/sync | Sync pod groups.



## SyncPodGroups

> SyncPodGroups(ctx).PodGroupsSyncRequest(podGroupsSyncRequest).Execute()

Sync pod groups.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client"
)

func main() {
	podGroupsSyncRequest := *openapiclient.NewPodGroupsSyncRequest([]openapiclient.PodGroupsSyncFields{*openapiclient.NewPodGroupsSyncFields("pod-group-a", "Id_example", "WorkloadId_example")}) // PodGroupsSyncRequest | The pod groups to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PodGroupsAPI.SyncPodGroups(context.Background()).PodGroupsSyncRequest(podGroupsSyncRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PodGroupsAPI.SyncPodGroups``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSyncPodGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **podGroupsSyncRequest** | [**PodGroupsSyncRequest**](PodGroupsSyncRequest.md) | The pod groups to sync. | 

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

