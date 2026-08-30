# \ResearcherCommandLineInterfaceAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetBinary**](ResearcherCommandLineInterfaceAPI.md#GetBinary) | **Get** /api/v1/cli-exposer/dist/{operatingSystem}/{architecture}/runai | Download RunAI Researcher command line binary
[**GetInstallerLinux**](ResearcherCommandLineInterfaceAPI.md#GetInstallerLinux) | **Get** /api/v1/cli-exposer/installer/linux | Download Linux installer script
[**GetInstallerLinuxCommands**](ResearcherCommandLineInterfaceAPI.md#GetInstallerLinuxCommands) | **Get** /api/v1/cli-exposer/installer/linux/commands | Get Linux installer script commands
[**GetInstallerMac**](ResearcherCommandLineInterfaceAPI.md#GetInstallerMac) | **Get** /api/v1/cli-exposer/installer/mac | Download Mac installer script
[**GetInstallerMacCommands**](ResearcherCommandLineInterfaceAPI.md#GetInstallerMacCommands) | **Get** /api/v1/cli-exposer/installer/mac/commands | Get Mac installer script commands
[**GetInstallerUnix**](ResearcherCommandLineInterfaceAPI.md#GetInstallerUnix) | **Get** /api/v1/cli-exposer/installer/unix | Download Unix installer script
[**GetInstallerUnixCommands**](ResearcherCommandLineInterfaceAPI.md#GetInstallerUnixCommands) | **Get** /api/v1/cli-exposer/installer/unix/commands | Get Unix installer script commands
[**GetInstallerWindowsCommands**](ResearcherCommandLineInterfaceAPI.md#GetInstallerWindowsCommands) | **Get** /api/v1/cli-exposer/installer/windows/commands | Get Windows MSI installer script commands
[**GetManualDocument**](ResearcherCommandLineInterfaceAPI.md#GetManualDocument) | **Get** /api/v1/cli-exposer/docs/{documentName} | Get CLI document by name
[**GetVersion**](ResearcherCommandLineInterfaceAPI.md#GetVersion) | **Get** /api/v1/cli-exposer/version | Get CLI latest version



## GetBinary

> *os.File GetBinary(ctx, operatingSystem, architecture).Execute()

Download RunAI Researcher command line binary



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {
	operatingSystem := "linux" // string | The operating system name.
	architecture := "arm64" // string | The architecture type.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetBinary(context.Background(), operatingSystem, architecture).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetBinary``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetBinary`: *os.File
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetBinary`: %v\n", resp)
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


## GetInstallerLinux

> string GetInstallerLinux(ctx).Execute()

Download Linux installer script



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerLinux(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetInstallerLinux``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerLinux`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetInstallerLinux`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerLinuxRequest struct via the builder pattern


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


## GetInstallerLinuxCommands

> Command GetInstallerLinuxCommands(ctx).Execute()

Get Linux installer script commands



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerLinuxCommands(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetInstallerLinuxCommands``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerLinuxCommands`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetInstallerLinuxCommands`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerLinuxCommandsRequest struct via the builder pattern


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


## GetInstallerMac

> string GetInstallerMac(ctx).Execute()

Download Mac installer script



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerMac(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetInstallerMac``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerMac`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetInstallerMac`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerMacRequest struct via the builder pattern


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


## GetInstallerMacCommands

> Command GetInstallerMacCommands(ctx).Execute()

Get Mac installer script commands



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerMacCommands(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetInstallerMacCommands``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerMacCommands`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetInstallerMacCommands`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerMacCommandsRequest struct via the builder pattern


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


## GetInstallerUnix

> string GetInstallerUnix(ctx).Execute()

Download Unix installer script



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerUnix(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetInstallerUnix``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerUnix`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetInstallerUnix`: %v\n", resp)
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

Get Unix installer script commands



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerUnixCommands(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetInstallerUnixCommands``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerUnixCommands`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetInstallerUnixCommands`: %v\n", resp)
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

Get Windows MSI installer script commands



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetInstallerWindowsCommands(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetInstallerWindowsCommands``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerWindowsCommands`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetInstallerWindowsCommands`: %v\n", resp)
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

Get CLI document by name



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {
	documentName := "runai.md" // string | The manual document name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetManualDocument(context.Background(), documentName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetManualDocument``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetManualDocument`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetManualDocument`: %v\n", resp)
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

Get CLI latest version



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/cli-exposer-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ResearcherCommandLineInterfaceAPI.GetVersion(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceAPI.GetVersion``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVersion`: Version
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceAPI.GetVersion`: %v\n", resp)
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

