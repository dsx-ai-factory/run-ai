# \SinksAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateSink**](SinksAPI.md#CreateSink) | **Post** /api/v1/sinks | 
[**DeleteSink**](SinksAPI.md#DeleteSink) | **Delete** /api/v1/sinks/{sinkId} | 
[**GetSink**](SinksAPI.md#GetSink) | **Get** /api/v1/sinks/{sinkId} | 
[**GetSinks**](SinksAPI.md#GetSinks) | **Get** /api/v1/sinks | 
[**UpdateSink**](SinksAPI.md#UpdateSink) | **Put** /api/v1/sinks/{sinkId} | 
[**ValidateSink**](SinksAPI.md#ValidateSink) | **Post** /api/v1/sinks/validate | 



## CreateSink

> string CreateSink(ctx).SinkData(sinkData).Execute()





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
	sinkData := *openapiclient.NewSinkData("Id_example", "Type_example") // SinkData | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SinksAPI.CreateSink(context.Background()).SinkData(sinkData).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SinksAPI.CreateSink``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateSink`: string
	fmt.Fprintf(os.Stdout, "Response from `SinksAPI.CreateSink`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateSinkRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **sinkData** | [**SinkData**](SinkData.md) |  | 

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


## DeleteSink

> string DeleteSink(ctx, sinkId).Execute()





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
	sinkId := "sinkId_example" // string | ID of the sink

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SinksAPI.DeleteSink``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteSink`: string
	fmt.Fprintf(os.Stdout, "Response from `SinksAPI.DeleteSink`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sinkId** | **string** | ID of the sink | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSinkRequest struct via the builder pattern


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


## GetSink

> SinkData GetSink(ctx, sinkId).Execute()





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
	sinkId := "sinkId_example" // string | ID of the sink

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SinksAPI.GetSink``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSink`: SinkData
	fmt.Fprintf(os.Stdout, "Response from `SinksAPI.GetSink`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sinkId** | **string** | ID of the sink | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSinkRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**SinkData**](SinkData.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSinks

> []SinkData GetSinks(ctx).Execute()





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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SinksAPI.GetSinks(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SinksAPI.GetSinks``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSinks`: []SinkData
	fmt.Fprintf(os.Stdout, "Response from `SinksAPI.GetSinks`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetSinksRequest struct via the builder pattern


### Return type

[**[]SinkData**](SinkData.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSink

> string UpdateSink(ctx, sinkId).SinkDataForUpdate(sinkDataForUpdate).Execute()





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
	sinkId := "sinkId_example" // string | ID of the sink
	sinkDataForUpdate := *openapiclient.NewSinkDataForUpdate("Type_example") // SinkDataForUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SinksAPI.UpdateSink(context.Background(), sinkId).SinkDataForUpdate(sinkDataForUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SinksAPI.UpdateSink``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateSink`: string
	fmt.Fprintf(os.Stdout, "Response from `SinksAPI.UpdateSink`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**sinkId** | **string** | ID of the sink | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSinkRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **sinkDataForUpdate** | [**SinkDataForUpdate**](SinkDataForUpdate.md) |  | 

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


## ValidateSink

> ValidateSink200Response ValidateSink(ctx).SinkData(sinkData).Execute()





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
	sinkData := *openapiclient.NewSinkData("Id_example", "Type_example") // SinkData | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SinksAPI.ValidateSink(context.Background()).SinkData(sinkData).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SinksAPI.ValidateSink``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ValidateSink`: ValidateSink200Response
	fmt.Fprintf(os.Stdout, "Response from `SinksAPI.ValidateSink`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiValidateSinkRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **sinkData** | [**SinkData**](SinkData.md) |  | 

### Return type

[**ValidateSink200Response**](ValidateSink200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

