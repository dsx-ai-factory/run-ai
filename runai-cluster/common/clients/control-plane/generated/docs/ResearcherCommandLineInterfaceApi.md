# \ResearcherCommandLineInterfaceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetBinary**](ResearcherCommandLineInterfaceApi.md#GetBinary) | **Get** /api/v1/cli/dist/{operatingSystem}/{architecture}/runai | Download RunAI Researcher command line binary [Experimental]
[**GetInstallerUnix**](ResearcherCommandLineInterfaceApi.md#GetInstallerUnix) | **Get** /api/v1/cli/installer/unix | Download Unix installer script [Experimental]
[**GetInstallerUnixCommands**](ResearcherCommandLineInterfaceApi.md#GetInstallerUnixCommands) | **Get** /api/v1/cli/installer/unix/commands | Get Unix installer script commands [Experimental]
[**GetInstallerWindowsCommands**](ResearcherCommandLineInterfaceApi.md#GetInstallerWindowsCommands) | **Get** /api/v1/cli/installer/windows/commands | Get Windows MSI installer script commands [Experimental]
[**GetManualDocument**](ResearcherCommandLineInterfaceApi.md#GetManualDocument) | **Get** /api/v1/cli/docs/{documentName} | Get CLI document by name [Experimental]
[**GetVersion**](ResearcherCommandLineInterfaceApi.md#GetVersion) | **Get** /api/v1/cli/version | Get CLI latest version [Experimental]



## GetBinary

> *os.File GetBinary(ctx, operatingSystem, architecture).Execute()

Download RunAI Researcher command line binary [Experimental]



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
    operatingSystem := "linux" // string | The operating system name.
    architecture := "arm64" // string | The architecture type.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ResearcherCommandLineInterfaceApi.GetBinary(context.Background(), operatingSystem, architecture).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceApi.GetBinary``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetBinary`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceApi.GetBinary`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**operatingSystem** | **string** | The operating system name. | 
**architecture** | **string** | The architecture type. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetBinaryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[***os.File**](*os.File.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/octet-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInstallerUnix

> string GetInstallerUnix(ctx).Execute()

Download Unix installer script [Experimental]



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ResearcherCommandLineInterfaceApi.GetInstallerUnix(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceApi.GetInstallerUnix``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInstallerUnix`: string
    fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceApi.GetInstallerUnix`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerUnixRequest struct via the builder pattern


### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInstallerUnixCommands

> Command GetInstallerUnixCommands(ctx).Execute()

Get Unix installer script commands [Experimental]



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ResearcherCommandLineInterfaceApi.GetInstallerUnixCommands(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceApi.GetInstallerUnixCommands``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInstallerUnixCommands`: Command
    fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceApi.GetInstallerUnixCommands`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerUnixCommandsRequest struct via the builder pattern


### Return type

[**Command**](Command.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInstallerWindowsCommands

> Command GetInstallerWindowsCommands(ctx).Execute()

Get Windows MSI installer script commands [Experimental]



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ResearcherCommandLineInterfaceApi.GetInstallerWindowsCommands(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceApi.GetInstallerWindowsCommands``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetInstallerWindowsCommands`: Command
    fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceApi.GetInstallerWindowsCommands`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerWindowsCommandsRequest struct via the builder pattern


### Return type

[**Command**](Command.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetManualDocument

> string GetManualDocument(ctx, documentName).Execute()

Get CLI document by name [Experimental]



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
    documentName := "runai.md" // string | The manual document name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ResearcherCommandLineInterfaceApi.GetManualDocument(context.Background(), documentName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceApi.GetManualDocument``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetManualDocument`: string
    fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceApi.GetManualDocument`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**documentName** | **string** | The manual document name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetManualDocumentRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/html, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetVersion

> Version GetVersion(ctx).Execute()

Get CLI latest version [Experimental]



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ResearcherCommandLineInterfaceApi.GetVersion(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceApi.GetVersion``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetVersion`: Version
    fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceApi.GetVersion`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetVersionRequest struct via the builder pattern


### Return type

[**Version**](Version.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

