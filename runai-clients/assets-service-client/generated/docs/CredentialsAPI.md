# \CredentialsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAccessKey**](CredentialsAPI.md#CreateAccessKey) | **Post** /api/v1/asset/credentials/access-key | Create an access key.
[**CreateDockerRegistry**](CredentialsAPI.md#CreateDockerRegistry) | **Post** /api/v1/asset/credentials/docker-registry | Create a docker registry credential.
[**CreateGenericSecret**](CredentialsAPI.md#CreateGenericSecret) | **Post** /api/v1/asset/credentials/generic-secret | Create a generic-secret.
[**CreateNgcApiKey**](CredentialsAPI.md#CreateNgcApiKey) | **Post** /api/v1/asset/credentials/ngc-api-key | Create a ngc-api-key.
[**CreatePassword**](CredentialsAPI.md#CreatePassword) | **Post** /api/v1/asset/credentials/password | Create a userid / password credential.
[**DeleteAccessKey**](CredentialsAPI.md#DeleteAccessKey) | **Delete** /api/v1/asset/credentials/access-key/{AssetId} | Delete an access key.
[**DeleteDockerRegistry**](CredentialsAPI.md#DeleteDockerRegistry) | **Delete** /api/v1/asset/credentials/docker-registry/{AssetId} | Delete a docker registry credential.
[**DeleteGenericSecret**](CredentialsAPI.md#DeleteGenericSecret) | **Delete** /api/v1/asset/credentials/generic-secret/{AssetId} | Delete a generic-secret.
[**DeleteNgcApiKey**](CredentialsAPI.md#DeleteNgcApiKey) | **Delete** /api/v1/asset/credentials/ngc-api-key/{AssetId} | Delete a ngc-api-key.
[**DeletePassword**](CredentialsAPI.md#DeletePassword) | **Delete** /api/v1/asset/credentials/password/{AssetId} | Delete a password asset.
[**GetAccessKeyById**](CredentialsAPI.md#GetAccessKeyById) | **Get** /api/v1/asset/credentials/access-key/{AssetId} | Get an access key.
[**GetDockerRegistryById**](CredentialsAPI.md#GetDockerRegistryById) | **Get** /api/v1/asset/credentials/docker-registry/{AssetId} | Get a docker registry credential.
[**GetGenericSecretById**](CredentialsAPI.md#GetGenericSecretById) | **Get** /api/v1/asset/credentials/generic-secret/{AssetId} | Get a generic-secret.
[**GetNgcApiKeyById**](CredentialsAPI.md#GetNgcApiKeyById) | **Get** /api/v1/asset/credentials/ngc-api-key/{AssetId} | Get a ngc-api-key.
[**GetNgcApiKeyByIdInternal**](CredentialsAPI.md#GetNgcApiKeyByIdInternal) | **Get** /internal/api/v1/asset/ngc-api-key/{AssetId} | Get a NGC API key by id (internal).
[**GetPasswordById**](CredentialsAPI.md#GetPasswordById) | **Get** /api/v1/asset/credentials/password/{AssetId} | Get a userid / password credential.
[**ListAccessKeys**](CredentialsAPI.md#ListAccessKeys) | **Get** /api/v1/asset/credentials/access-key | List access keys.
[**ListCredentialsAssets**](CredentialsAPI.md#ListCredentialsAssets) | **Get** /api/v1/asset/credentials | List credentials.
[**ListDockerRegistries**](CredentialsAPI.md#ListDockerRegistries) | **Get** /api/v1/asset/credentials/docker-registry | List docker registry credentials.
[**ListGenericSecret**](CredentialsAPI.md#ListGenericSecret) | **Get** /api/v1/asset/credentials/generic-secret | List generic-secrets.
[**ListNgcApiKey**](CredentialsAPI.md#ListNgcApiKey) | **Get** /api/v1/asset/credentials/ngc-api-key | List ngc-api-keys.
[**ListPasswords**](CredentialsAPI.md#ListPasswords) | **Get** /api/v1/asset/credentials/password | List password credentials.
[**UpdateAccessKey**](CredentialsAPI.md#UpdateAccessKey) | **Put** /api/v1/asset/credentials/access-key/{AssetId} | Update an access key.
[**UpdateDockerRegistry**](CredentialsAPI.md#UpdateDockerRegistry) | **Put** /api/v1/asset/credentials/docker-registry/{AssetId} | Update a docker registry credential.
[**UpdateGenericSecret**](CredentialsAPI.md#UpdateGenericSecret) | **Put** /api/v1/asset/credentials/generic-secret/{AssetId} | Update a generic-secret.
[**UpdateNgcApiKey**](CredentialsAPI.md#UpdateNgcApiKey) | **Put** /api/v1/asset/credentials/ngc-api-key/{AssetId} | Update a ngc-api-key.
[**UpdatePassword**](CredentialsAPI.md#UpdatePassword) | **Put** /api/v1/asset/credentials/password/{AssetId} | Update a password credential.



## CreateAccessKey

> AccessKey CreateAccessKey(ctx).AccessKeyCreationRequest(accessKeyCreationRequest).Execute()

Create an access key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	accessKeyCreationRequest := *openapiclient.NewAccessKeyCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewAccessKeyCreationSpec()) // AccessKeyCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.CreateAccessKey(context.Background()).AccessKeyCreationRequest(accessKeyCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.CreateAccessKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateAccessKey`: AccessKey
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.CreateAccessKey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateAccessKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **accessKeyCreationRequest** | [**AccessKeyCreationRequest**](AccessKeyCreationRequest.md) |  | 

### Return type

[**AccessKey**](AccessKey.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDockerRegistry

> DockerRegistry CreateDockerRegistry(ctx).DockerRegistryCreationRequest(dockerRegistryCreationRequest).Execute()

Create a docker registry credential.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	dockerRegistryCreationRequest := *openapiclient.NewDockerRegistryCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewDockerRegistryCreationSpec()) // DockerRegistryCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.CreateDockerRegistry(context.Background()).DockerRegistryCreationRequest(dockerRegistryCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.CreateDockerRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDockerRegistry`: DockerRegistry
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.CreateDockerRegistry`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateDockerRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **dockerRegistryCreationRequest** | [**DockerRegistryCreationRequest**](DockerRegistryCreationRequest.md) |  | 

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateGenericSecret

> GenericSecret CreateGenericSecret(ctx).GenericSecretCreationRequest(genericSecretCreationRequest).Execute()

Create a generic-secret.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	genericSecretCreationRequest := *openapiclient.NewGenericSecretCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewGenericSecretCreationSpec()) // GenericSecretCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.CreateGenericSecret(context.Background()).GenericSecretCreationRequest(genericSecretCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.CreateGenericSecret``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateGenericSecret`: GenericSecret
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.CreateGenericSecret`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateGenericSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **genericSecretCreationRequest** | [**GenericSecretCreationRequest**](GenericSecretCreationRequest.md) |  | 

### Return type

[**GenericSecret**](GenericSecret.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateNgcApiKey

> NgcApiKey CreateNgcApiKey(ctx).NgcApiKeyCreationRequest(ngcApiKeyCreationRequest).Execute()

Create a ngc-api-key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	ngcApiKeyCreationRequest := *openapiclient.NewNgcApiKeyCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewNgcApiKeyCreationSpec()) // NgcApiKeyCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.CreateNgcApiKey(context.Background()).NgcApiKeyCreationRequest(ngcApiKeyCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.CreateNgcApiKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateNgcApiKey`: NgcApiKey
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.CreateNgcApiKey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNgcApiKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ngcApiKeyCreationRequest** | [**NgcApiKeyCreationRequest**](NgcApiKeyCreationRequest.md) |  | 

### Return type

[**NgcApiKey**](NgcApiKey.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreatePassword

> Password CreatePassword(ctx).PasswordCreationRequest(passwordCreationRequest).Execute()

Create a userid / password credential.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	passwordCreationRequest := *openapiclient.NewPasswordCreationRequest(*openapiclient.NewAssetCreationRequest("my-asset", openapiclient.Scope("system")), *openapiclient.NewPasswordCreationSpec()) // PasswordCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.CreatePassword(context.Background()).PasswordCreationRequest(passwordCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.CreatePassword``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreatePassword`: Password
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.CreatePassword`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreatePasswordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **passwordCreationRequest** | [**PasswordCreationRequest**](PasswordCreationRequest.md) |  | 

### Return type

[**Password**](Password.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteAccessKey

> HttpResponse DeleteAccessKey(ctx, assetId).Execute()

Delete an access key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.DeleteAccessKey(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.DeleteAccessKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteAccessKey`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.DeleteAccessKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteAccessKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDockerRegistry

> HttpResponse DeleteDockerRegistry(ctx, assetId).Execute()

Delete a docker registry credential.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.DeleteDockerRegistry(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.DeleteDockerRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteDockerRegistry`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.DeleteDockerRegistry`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDockerRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteGenericSecret

> DeleteGenericSecret(ctx, assetId).Execute()

Delete a generic-secret.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.CredentialsAPI.DeleteGenericSecret(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.DeleteGenericSecret``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteGenericSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteNgcApiKey

> DeleteNgcApiKey(ctx, assetId).Execute()

Delete a ngc-api-key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.CredentialsAPI.DeleteNgcApiKey(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.DeleteNgcApiKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNgcApiKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePassword

> HttpResponse DeletePassword(ctx, assetId).Execute()

Delete a password asset.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.DeletePassword(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.DeletePassword``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeletePassword`: HttpResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.DeletePassword`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeletePasswordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**HttpResponse**](HttpResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAccessKeyById

> AccessKey GetAccessKeyById(ctx, assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()

Get an access key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.GetAccessKeyById(context.Background(), assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.GetAccessKeyById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAccessKeyById`: AccessKey
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.GetAccessKeyById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAccessKeyByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**AccessKey**](AccessKey.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetDockerRegistryById

> DockerRegistry GetDockerRegistryById(ctx, assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()

Get a docker registry credential.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.GetDockerRegistryById(context.Background(), assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.GetDockerRegistryById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDockerRegistryById`: DockerRegistry
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.GetDockerRegistryById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetDockerRegistryByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGenericSecretById

> GenericSecret GetGenericSecretById(ctx, assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()

Get a generic-secret.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.GetGenericSecretById(context.Background(), assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.GetGenericSecretById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGenericSecretById`: GenericSecret
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.GetGenericSecretById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetGenericSecretByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**GenericSecret**](GenericSecret.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNgcApiKeyById

> NgcApiKey GetNgcApiKeyById(ctx, assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()

Get a ngc-api-key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.GetNgcApiKeyById(context.Background(), assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.GetNgcApiKeyById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNgcApiKeyById`: NgcApiKey
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.GetNgcApiKeyById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNgcApiKeyByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**NgcApiKey**](NgcApiKey.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNgcApiKeyByIdInternal

> NgcApiKeyInternal GetNgcApiKeyByIdInternal(ctx, assetId).Execute()

Get a NGC API key by id (internal).



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.GetNgcApiKeyByIdInternal(context.Background(), assetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.GetNgcApiKeyByIdInternal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNgcApiKeyByIdInternal`: NgcApiKeyInternal
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.GetNgcApiKeyByIdInternal`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetNgcApiKeyByIdInternalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**NgcApiKeyInternal**](NgcApiKeyInternal.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPasswordById

> Password GetPasswordById(ctx, assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()

Get a userid / password credential.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.GetPasswordById(context.Background(), assetId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.GetPasswordById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPasswordById`: Password
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.GetPasswordById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPasswordByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**Password**](Password.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListAccessKeys

> AccessKeyListResponse ListAccessKeys(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()

List access keys.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.ListAccessKeys(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.ListAccessKeys``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListAccessKeys`: AccessKeyListResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.ListAccessKeys`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListAccessKeysRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**AccessKeyListResponse**](AccessKeyListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListCredentialsAssets

> CredentialsListResponse ListCredentialsAssets(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).IncludeDescendants(includeDescendants).Type_(type_).Execute()

List credentials.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	privateOnly := true // bool | Whether the query should include private assets. mutually exclusive with `includePrivate` parameter. (optional)
	includePrivate := true // bool | Whether the query should include private assets. mutually exclusive with `privateOnly` parameter. (optional)
	includeDescendants := true // bool | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. (optional)
	type_ := "type__example" // string | Filter credentials by type. If provided, only credentials of the specified type will be returned. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.ListCredentialsAssets(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).IncludeDescendants(includeDescendants).Type_(type_).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.ListCredentialsAssets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListCredentialsAssets`: CredentialsListResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.ListCredentialsAssets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListCredentialsAssetsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **privateOnly** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;includePrivate&#x60; parameter. | 
 **includePrivate** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;privateOnly&#x60; parameter. | 
 **includeDescendants** | **bool** | Indicates whether to include assets from all descendant scopes under the specified scope. When set to true, and a specific scope identifier (for example, clusterId) is provided, the query also returns assets from all nested scopes within it (for example, departments and projects).  Mutually exclusive with scope. | 
 **type_** | **string** | Filter credentials by type. If provided, only credentials of the specified type will be returned. | 

### Return type

[**CredentialsListResponse**](CredentialsListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListDockerRegistries

> DockerRegistryListResponse ListDockerRegistries(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).Execute()

List docker registry credentials.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	privateOnly := true // bool | Whether the query should include private assets. mutually exclusive with `includePrivate` parameter. (optional)
	includePrivate := true // bool | Whether the query should include private assets. mutually exclusive with `privateOnly` parameter. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.ListDockerRegistries(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.ListDockerRegistries``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListDockerRegistries`: DockerRegistryListResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.ListDockerRegistries`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListDockerRegistriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **privateOnly** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;includePrivate&#x60; parameter. | 
 **includePrivate** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;privateOnly&#x60; parameter. | 

### Return type

[**DockerRegistryListResponse**](DockerRegistryListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListGenericSecret

> GenericSecretListResponse ListGenericSecret(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).Execute()

List generic-secrets.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	privateOnly := true // bool | Whether the query should include private assets. mutually exclusive with `includePrivate` parameter. (optional)
	includePrivate := true // bool | Whether the query should include private assets. mutually exclusive with `privateOnly` parameter. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.ListGenericSecret(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.ListGenericSecret``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListGenericSecret`: GenericSecretListResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.ListGenericSecret`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListGenericSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **privateOnly** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;includePrivate&#x60; parameter. | 
 **includePrivate** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;privateOnly&#x60; parameter. | 

### Return type

[**GenericSecretListResponse**](GenericSecretListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListNgcApiKey

> NgcApiKeyListResponse ListNgcApiKey(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).Execute()

List ngc-api-keys.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)
	privateOnly := true // bool | Whether the query should include private assets. mutually exclusive with `includePrivate` parameter. (optional)
	includePrivate := true // bool | Whether the query should include private assets. mutually exclusive with `privateOnly` parameter. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.ListNgcApiKey(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).PrivateOnly(privateOnly).IncludePrivate(includePrivate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.ListNgcApiKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListNgcApiKey`: NgcApiKeyListResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.ListNgcApiKey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListNgcApiKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 
 **privateOnly** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;includePrivate&#x60; parameter. | 
 **includePrivate** | **bool** | Whether the query should include private assets. mutually exclusive with &#x60;privateOnly&#x60; parameter. | 

### Return type

[**NgcApiKeyListResponse**](NgcApiKeyListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPasswords

> PasswordListResponse ListPasswords(ctx).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()

List password credentials.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	name := "name_example" // string | Filter results by name. (optional)
	scope := "scope_example" // string | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. (optional)
	projectId := int32(56) // int32 | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. (optional)
	departmentId := "1" // string | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. (optional)
	usageInfo := true // bool | Whether the query should include asset usage information as part of the response. (optional)
	statusInfo := true // bool | Whether the query should include asset status information as part of the response. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.ListPasswords(context.Background()).Name(name).Scope(scope).ProjectId(projectId).DepartmentId(departmentId).ClusterId(clusterId).UsageInfo(usageInfo).StatusInfo(statusInfo).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.ListPasswords``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPasswords`: PasswordListResponse
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.ListPasswords`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPasswordsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **string** | Filter results by name. | 
 **scope** | **string** | Filters results by scope. Returns only assets that belong to the specified scope. Mutually exclusive with includeDescendants. Valid values: tenant, cluster, department, project. | 
 **projectId** | **int32** | Filter results by project id. If scope filter is project, only assets from the specific project will be included in the response. Otherwise, the response will include project, department, cluster, tenant and system assets. | 
 **departmentId** | **string** | Filter results by department id. If scope filter is department, only assets from the specific department will be included in the response. Otherwise, the response will include department, cluster, tenant and system assets. | 
 **clusterId** | **string** | Filter results by Universally Unique Identifier (UUID) of the cluster. If scope filter is cluster, only assets from the specific cluster will be included in the response. Otherwise, the response will include cluster, tenant and system assets. | 
 **usageInfo** | **bool** | Whether the query should include asset usage information as part of the response. | 
 **statusInfo** | **bool** | Whether the query should include asset status information as part of the response. | 

### Return type

[**PasswordListResponse**](PasswordListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateAccessKey

> AccessKey UpdateAccessKey(ctx, assetId).AccessKeyUpdateRequest(accessKeyUpdateRequest).Execute()

Update an access key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	accessKeyUpdateRequest := *openapiclient.NewAccessKeyUpdateRequest() // AccessKeyUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.UpdateAccessKey(context.Background(), assetId).AccessKeyUpdateRequest(accessKeyUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.UpdateAccessKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateAccessKey`: AccessKey
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.UpdateAccessKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateAccessKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **accessKeyUpdateRequest** | [**AccessKeyUpdateRequest**](AccessKeyUpdateRequest.md) |  | 

### Return type

[**AccessKey**](AccessKey.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDockerRegistry

> DockerRegistry UpdateDockerRegistry(ctx, assetId).DockerRegistryUpdateRequest(dockerRegistryUpdateRequest).Execute()

Update a docker registry credential.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	dockerRegistryUpdateRequest := *openapiclient.NewDockerRegistryUpdateRequest() // DockerRegistryUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.UpdateDockerRegistry(context.Background(), assetId).DockerRegistryUpdateRequest(dockerRegistryUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.UpdateDockerRegistry``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDockerRegistry`: DockerRegistry
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.UpdateDockerRegistry`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDockerRegistryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **dockerRegistryUpdateRequest** | [**DockerRegistryUpdateRequest**](DockerRegistryUpdateRequest.md) |  | 

### Return type

[**DockerRegistry**](DockerRegistry.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateGenericSecret

> GenericSecret UpdateGenericSecret(ctx, assetId).GenericSecretUpdateRequest(genericSecretUpdateRequest).Execute()

Update a generic-secret.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	genericSecretUpdateRequest := *openapiclient.NewGenericSecretUpdateRequest() // GenericSecretUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.UpdateGenericSecret(context.Background(), assetId).GenericSecretUpdateRequest(genericSecretUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.UpdateGenericSecret``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateGenericSecret`: GenericSecret
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.UpdateGenericSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateGenericSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **genericSecretUpdateRequest** | [**GenericSecretUpdateRequest**](GenericSecretUpdateRequest.md) |  | 

### Return type

[**GenericSecret**](GenericSecret.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNgcApiKey

> NgcApiKey UpdateNgcApiKey(ctx, assetId).NgcApiKeyUpdateRequest(ngcApiKeyUpdateRequest).Execute()

Update a ngc-api-key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	ngcApiKeyUpdateRequest := *openapiclient.NewNgcApiKeyUpdateRequest() // NgcApiKeyUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.UpdateNgcApiKey(context.Background(), assetId).NgcApiKeyUpdateRequest(ngcApiKeyUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.UpdateNgcApiKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNgcApiKey`: NgcApiKey
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.UpdateNgcApiKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNgcApiKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **ngcApiKeyUpdateRequest** | [**NgcApiKeyUpdateRequest**](NgcApiKeyUpdateRequest.md) |  | 

### Return type

[**NgcApiKey**](NgcApiKey.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdatePassword

> Password UpdatePassword(ctx, assetId).PasswordUpdateRequest(passwordUpdateRequest).Execute()

Update a password credential.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/assets-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	passwordUpdateRequest := *openapiclient.NewPasswordUpdateRequest() // PasswordUpdateRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.CredentialsAPI.UpdatePassword(context.Background(), assetId).PasswordUpdateRequest(passwordUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `CredentialsAPI.UpdatePassword``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdatePassword`: Password
	fmt.Fprintf(os.Stdout, "Response from `CredentialsAPI.UpdatePassword`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdatePasswordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **passwordUpdateRequest** | [**PasswordUpdateRequest**](PasswordUpdateRequest.md) |  | 

### Return type

[**Password**](Password.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

