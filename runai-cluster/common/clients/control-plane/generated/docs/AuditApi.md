# \AuditApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetAuditLogs**](AuditApi.md#GetAuditLogs) | **Get** /v1/k8s/audit | Get audit logs.



## GetAuditLogs

> []AuditLogRecord GetAuditLogs(ctx).Start(start).End(end).ClusterUuid(clusterUuid).Action(action).SourceType(sourceType).SourceId(sourceId).SourceName(sourceName).EntityType(entityType).EntityId(entityId).Limit(limit).Offset(offset).Success(success).Download(download).Execute()

Get audit logs.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    openapiclient "./openapi"
)

func main() {
    start := time.Now() // time.Time | Starting date for audit logs retrieval. format yyyy-MM-dd for date yyyy-MM-ddTHH:mm:ss for date-time. (optional)
    end := time.Now() // time.Time | Ending date for audit logs retrieval. format yyyy-MM-dd for date yyyy-MM-ddTHH:mm:ss for date-time. (optional)
    clusterUuid := "123e4567-e89b-12d3-a456-426614174000" // string | The cluster uuid of the logged operation (for cluster related operations) (optional)
    action := "Create" // string | The action of the logged operation. (optional)
    sourceType := "User" // string | The type of the source of the action. (optional)
    sourceId := "96a4382e-afa5-4604-9eb1-c3071aa021fc" // string | The id of the source of the action. (optional)
    sourceName := "test@run.ai" // string | The name of the source of the action. (optional)
    entityType := "Department" // string | The type of the action related entity. (optional)
    entityId := "51" // string | The id of the action related entity. (optional)
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    success := "true" // string | enter true for success audits and false for failures (leave blank for all) (optional)
    download := "true" // string | enter true to download the logs into logs.json file (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AuditApi.GetAuditLogs(context.Background()).Start(start).End(end).ClusterUuid(clusterUuid).Action(action).SourceType(sourceType).SourceId(sourceId).SourceName(sourceName).EntityType(entityType).EntityId(entityId).Limit(limit).Offset(offset).Success(success).Download(download).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AuditApi.GetAuditLogs``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAuditLogs`: []AuditLogRecord
    fmt.Fprintf(os.Stdout, "Response from `AuditApi.GetAuditLogs`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAuditLogsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **start** | **time.Time** | Starting date for audit logs retrieval. format yyyy-MM-dd for date yyyy-MM-ddTHH:mm:ss for date-time. | 
 **end** | **time.Time** | Ending date for audit logs retrieval. format yyyy-MM-dd for date yyyy-MM-ddTHH:mm:ss for date-time. | 
 **clusterUuid** | **string** | The cluster uuid of the logged operation (for cluster related operations) | 
 **action** | **string** | The action of the logged operation. | 
 **sourceType** | **string** | The type of the source of the action. | 
 **sourceId** | **string** | The id of the source of the action. | 
 **sourceName** | **string** | The name of the source of the action. | 
 **entityType** | **string** | The type of the action related entity. | 
 **entityId** | **string** | The id of the action related entity. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **success** | **string** | enter true for success audits and false for failures (leave blank for all) | 
 **download** | **string** | enter true to download the logs into logs.json file | 

### Return type

[**[]AuditLogRecord**](AuditLogRecord.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

