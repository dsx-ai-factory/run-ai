# \ResearcherCommandLineInterfaceDeprecatedAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetBinaryDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetBinaryDeprecated) | **Get** /api/v1/cli/dist/{operatingSystem}/{architecture}/runai | Download RunAI Researcher command line binary
[**GetInstallerLinuxCommandsDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetInstallerLinuxCommandsDeprecated) | **Get** /api/v1/cli/installer/linux/commands | Get Linux installer script commands
[**GetInstallerLinuxDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetInstallerLinuxDeprecated) | **Get** /api/v1/cli/installer/linux | Download Linux installer script
[**GetInstallerMacCommandsDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetInstallerMacCommandsDeprecated) | **Get** /api/v1/cli/installer/mac/commands | Get Mac installer script commands
[**GetInstallerMacDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetInstallerMacDeprecated) | **Get** /api/v1/cli/installer/mac | Download Mac installer script
[**GetInstallerUnixCommandsDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetInstallerUnixCommandsDeprecated) | **Get** /api/v1/cli/installer/unix/commands | Get Unix installer script commands
[**GetInstallerUnixDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetInstallerUnixDeprecated) | **Get** /api/v1/cli/installer/unix | Download Unix installer script
[**GetInstallerWindowsCommandsDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetInstallerWindowsCommandsDeprecated) | **Get** /api/v1/cli/installer/windows/commands | Get Windows MSI installer script commands
[**GetManualDocumentDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetManualDocumentDeprecated) | **Get** /api/v1/cli/docs/{documentName} | Get CLI document by name
[**GetVersionDeprecated**](ResearcherCommandLineInterfaceDeprecatedAPI.md#GetVersionDeprecated) | **Get** /api/v1/cli/version | Get CLI latest version



## GetBinaryDeprecated

> *os.File GetBinaryDeprecated(ctx, operatingSystem, architecture).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetBinaryDeprecated(context.Background(), operatingSystem, architecture).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetBinaryDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetBinaryDeprecated`: *os.File
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetBinaryDeprecated`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**operatingSystem** | **string** | The operating system name. | 
**architecture** | **string** | The architecture type. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetBinaryDeprecatedRequest struct via the builder pattern


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


## GetInstallerLinuxCommandsDeprecated

> Command GetInstallerLinuxCommandsDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerLinuxCommandsDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerLinuxCommandsDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerLinuxCommandsDeprecated`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerLinuxCommandsDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerLinuxCommandsDeprecatedRequest struct via the builder pattern


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


## GetInstallerLinuxDeprecated

> string GetInstallerLinuxDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerLinuxDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerLinuxDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerLinuxDeprecated`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerLinuxDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerLinuxDeprecatedRequest struct via the builder pattern


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


## GetInstallerMacCommandsDeprecated

> Command GetInstallerMacCommandsDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerMacCommandsDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerMacCommandsDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerMacCommandsDeprecated`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerMacCommandsDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerMacCommandsDeprecatedRequest struct via the builder pattern


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


## GetInstallerMacDeprecated

> string GetInstallerMacDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerMacDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerMacDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerMacDeprecated`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerMacDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerMacDeprecatedRequest struct via the builder pattern


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


## GetInstallerUnixCommandsDeprecated

> Command GetInstallerUnixCommandsDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerUnixCommandsDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerUnixCommandsDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerUnixCommandsDeprecated`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerUnixCommandsDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerUnixCommandsDeprecatedRequest struct via the builder pattern


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


## GetInstallerUnixDeprecated

> string GetInstallerUnixDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerUnixDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerUnixDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerUnixDeprecated`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerUnixDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerUnixDeprecatedRequest struct via the builder pattern


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


## GetInstallerWindowsCommandsDeprecated

> Command GetInstallerWindowsCommandsDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerWindowsCommandsDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerWindowsCommandsDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerWindowsCommandsDeprecated`: Command
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetInstallerWindowsCommandsDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerWindowsCommandsDeprecatedRequest struct via the builder pattern


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


## GetManualDocumentDeprecated

> string GetManualDocumentDeprecated(ctx, documentName).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetManualDocumentDeprecated(context.Background(), documentName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetManualDocumentDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetManualDocumentDeprecated`: string
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetManualDocumentDeprecated`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**documentName** | **string** | The manual document name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetManualDocumentDeprecatedRequest struct via the builder pattern


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


## GetVersionDeprecated

> Version GetVersionDeprecated(ctx).Execute()

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
	resp, r, err := apiClient.ResearcherCommandLineInterfaceDeprecatedAPI.GetVersionDeprecated(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ResearcherCommandLineInterfaceDeprecatedAPI.GetVersionDeprecated``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetVersionDeprecated`: Version
	fmt.Fprintf(os.Stdout, "Response from `ResearcherCommandLineInterfaceDeprecatedAPI.GetVersionDeprecated`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetVersionDeprecatedRequest struct via the builder pattern


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

