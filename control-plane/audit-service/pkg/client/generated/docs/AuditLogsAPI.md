# \AuditLogsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DownloadAuditLogs**](AuditLogsAPI.md#DownloadAuditLogs) | **Get** /api/v1/audit/log/file | Download audit logs
[**GetAuditLogs**](AuditLogsAPI.md#GetAuditLogs) | **Get** /api/v1/audit/log | Get audit logs



## DownloadAuditLogs

> DownloadAuditLogs(ctx).Start(start).End(end).FileType(fileType).Offset(offset).NumberOfSamples(numberOfSamples).FilterBy(filterBy).Execute()

Download audit logs



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
)

func main() {
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	fileType := "fileType_example" // string |  (optional) (default to "CSV")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	numberOfSamples := int32(56) // int32 | Maximum number of records to include in the file. The hard cap is 100,000. Use the offset parameter to page through additional records. (optional) (default to 20)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AuditLogsAPI.DownloadAuditLogs(context.Background()).Start(start).End(end).FileType(fileType).Offset(offset).NumberOfSamples(numberOfSamples).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuditLogsAPI.DownloadAuditLogs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDownloadAuditLogsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **start** | **time.Time** | Start date of time range to fetch data in ISO 8601 timestamp format. | 
 **end** | **time.Time** | End date of time range to fetch data in ISO 8601 timestamp format. | 
 **fileType** | **string** |  | [default to &quot;CSV&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **numberOfSamples** | **int32** | Maximum number of records to include in the file. The hard cap is 100,000. Use the offset parameter to page through additional records. | [default to 20]
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

### Return type

 (empty response body)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAuditLogs

> GetAuditLogs200Response GetAuditLogs(ctx).Start(start).End(end).Offset(offset).NumberOfSamples(numberOfSamples).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

Get audit logs



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
)

func main() {
	start := time.Now() // time.Time | Start date of time range to fetch data in ISO 8601 timestamp format.
	end := time.Now() // time.Time | End date of time range to fetch data in ISO 8601 timestamp format.
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	numberOfSamples := int32(20) // int32 | The number of samples to take in the specified time range. (optional) (default to 20)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := "sortBy_example" // string | Sort results by a parameters. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuditLogsAPI.GetAuditLogs(context.Background()).Start(start).End(end).Offset(offset).NumberOfSamples(numberOfSamples).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuditLogsAPI.GetAuditLogs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAuditLogs`: GetAuditLogs200Response
	fmt.Fprintf(os.Stdout, "Response from `AuditLogsAPI.GetAuditLogs`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAuditLogsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **start** | **time.Time** | Start date of time range to fetch data in ISO 8601 timestamp format. | 
 **end** | **time.Time** | End date of time range to fetch data in ISO 8601 timestamp format. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **numberOfSamples** | **int32** | The number of samples to take in the specified time range. | [default to 20]
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | **string** | Sort results by a parameters. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

### Return type

[**GetAuditLogs200Response**](GetAuditLogs200Response.md)

### Authorization

[BearerAuth](../README.md#BearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

