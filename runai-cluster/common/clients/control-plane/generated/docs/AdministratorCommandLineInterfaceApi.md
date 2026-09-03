# \AdministratorCommandLineInterfaceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetAdminCLIRelease**](AdministratorCommandLineInterfaceApi.md#GetAdminCLIRelease) | **Get** /v1/k8s/admin-cli/{os} | Get Administrator Command Line Interface release.
[**GetAdminCLIReleaseByVersion**](AdministratorCommandLineInterfaceApi.md#GetAdminCLIReleaseByVersion) | **Get** /v1/k8s/admin-cli/{version}/{os} | Get Administrator Command Line Interface release by version.
[**GetAdminCLIReleaseByVersionChecksum**](AdministratorCommandLineInterfaceApi.md#GetAdminCLIReleaseByVersionChecksum) | **Get** /v1/k8s/admin-cli/{version}/{os}/checksum | Get Administrator Command Line Interface release checksums.
[**GetAdminCLIReleaseChecksum**](AdministratorCommandLineInterfaceApi.md#GetAdminCLIReleaseChecksum) | **Get** /v1/k8s/admin-cli/{os}/checksum | Get Administrator Command Line Interface release details.



## GetAdminCLIRelease

> *os.File GetAdminCLIRelease(ctx, os).Execute()

Get Administrator Command Line Interface release.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    os := "os_example" // string | The operating system

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdministratorCommandLineInterfaceApi.GetAdminCLIRelease(context.Background(), os).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdministratorCommandLineInterfaceApi.GetAdminCLIRelease``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAdminCLIRelease`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `AdministratorCommandLineInterfaceApi.GetAdminCLIRelease`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**os** | **string** | The operating system | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAdminCLIReleaseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[***os.File**](*os.File.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/x-tar, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAdminCLIReleaseByVersion

> *os.File GetAdminCLIReleaseByVersion(ctx, os, version).Execute()

Get Administrator Command Line Interface release by version.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    os := "os_example" // string | The operating system
    version := "version_example" // string | run:ai version (semantic version)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseByVersion(context.Background(), os, version).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseByVersion``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAdminCLIReleaseByVersion`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseByVersion`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**os** | **string** | The operating system | 
**version** | **string** | run:ai version (semantic version) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAdminCLIReleaseByVersionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[***os.File**](*os.File.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/x-tar, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAdminCLIReleaseByVersionChecksum

> ChecksumResponse GetAdminCLIReleaseByVersionChecksum(ctx, os, version).Execute()

Get Administrator Command Line Interface release checksums.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    os := "os_example" // string | The operating system
    version := "version_example" // string | run:ai version (semantic version)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseByVersionChecksum(context.Background(), os, version).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseByVersionChecksum``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAdminCLIReleaseByVersionChecksum`: ChecksumResponse
    fmt.Fprintf(os.Stdout, "Response from `AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseByVersionChecksum`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**os** | **string** | The operating system | 
**version** | **string** | run:ai version (semantic version) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAdminCLIReleaseByVersionChecksumRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ChecksumResponse**](ChecksumResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAdminCLIReleaseChecksum

> ChecksumResponse GetAdminCLIReleaseChecksum(ctx, os).Execute()

Get Administrator Command Line Interface release details.



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    os := "os_example" // string | The operating system

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseChecksum(context.Background(), os).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseChecksum``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAdminCLIReleaseChecksum`: ChecksumResponse
    fmt.Fprintf(os.Stdout, "Response from `AdministratorCommandLineInterfaceApi.GetAdminCLIReleaseChecksum`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**os** | **string** | The operating system | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAdminCLIReleaseChecksumRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ChecksumResponse**](ChecksumResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

