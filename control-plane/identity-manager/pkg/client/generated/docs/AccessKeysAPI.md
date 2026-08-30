# \AccessKeysAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAccessKey**](AccessKeysAPI.md#CreateAccessKey) | **Post** /api/v1/access-keys | Create an access key.
[**DeleteAccessKeyById**](AccessKeysAPI.md#DeleteAccessKeyById) | **Delete** /api/v1/access-keys/{accessKeyId} | Delete an access key by id.
[**DeleteAccessKeyByIdAdministration**](AccessKeysAPI.md#DeleteAccessKeyByIdAdministration) | **Delete** /api/v1/administration/access-keys/{accessKeyId} | Delete an access key by id for administrations.
[**GetAccessKeyById**](AccessKeysAPI.md#GetAccessKeyById) | **Get** /api/v1/access-keys/{accessKeyId} | Get access key by id.
[**GetAccessKeys**](AccessKeysAPI.md#GetAccessKeys) | **Get** /api/v1/access-keys | Get a list of access keys.
[**GetAccessKeysAdministration**](AccessKeysAPI.md#GetAccessKeysAdministration) | **Get** /api/v1/administration/access-keys | Get a list of all access keys.
[**RegenerateAccessKeySecret**](AccessKeysAPI.md#RegenerateAccessKeySecret) | **Post** /api/v1/access-keys/{accessKeyId}/secret | Regenerate an access key secret.



## CreateAccessKey

> AccessKeyPostResponse CreateAccessKey(ctx).AccessKeyCreationRequest(accessKeyCreationRequest).Execute()

Create an access key.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	accessKeyCreationRequest := *openapiclient.NewAccessKeyCreationRequest("Name_example") // AccessKeyCreationRequest | Access key object to create

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessKeysAPI.CreateAccessKey(context.Background()).AccessKeyCreationRequest(accessKeyCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessKeysAPI.CreateAccessKey``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateAccessKey`: AccessKeyPostResponse
	fmt.Fprintf(os.Stdout, "Response from `AccessKeysAPI.CreateAccessKey`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateAccessKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **accessKeyCreationRequest** | [**AccessKeyCreationRequest**](AccessKeyCreationRequest.md) | Access key object to create | 

### Return type

[**AccessKeyPostResponse**](AccessKeyPostResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteAccessKeyById

> DeleteAccessKeyById(ctx, accessKeyId).Execute()

Delete an access key by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	accessKeyId := "accessKeyId_example" // string | The access key ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AccessKeysAPI.DeleteAccessKeyById(context.Background(), accessKeyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessKeysAPI.DeleteAccessKeyById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessKeyId** | **string** | The access key ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteAccessKeyByIdRequest struct via the builder pattern


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


## DeleteAccessKeyByIdAdministration

> DeleteAccessKeyByIdAdministration(ctx, accessKeyId).Execute()

Delete an access key by id for administrations.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	accessKeyId := "accessKeyId_example" // string | The access key ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AccessKeysAPI.DeleteAccessKeyByIdAdministration(context.Background(), accessKeyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessKeysAPI.DeleteAccessKeyByIdAdministration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessKeyId** | **string** | The access key ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteAccessKeyByIdAdministrationRequest struct via the builder pattern


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


## GetAccessKeyById

> AccessKey GetAccessKeyById(ctx, accessKeyId).Execute()

Get access key by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	accessKeyId := "accessKeyId_example" // string | The access key ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessKeysAPI.GetAccessKeyById(context.Background(), accessKeyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessKeysAPI.GetAccessKeyById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAccessKeyById`: AccessKey
	fmt.Fprintf(os.Stdout, "Response from `AccessKeysAPI.GetAccessKeyById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessKeyId** | **string** | The access key ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAccessKeyByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## GetAccessKeys

> AccessKeys GetAccessKeys(ctx).Execute()

Get a list of access keys.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessKeysAPI.GetAccessKeys(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessKeysAPI.GetAccessKeys``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAccessKeys`: AccessKeys
	fmt.Fprintf(os.Stdout, "Response from `AccessKeysAPI.GetAccessKeys`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetAccessKeysRequest struct via the builder pattern


### Return type

[**AccessKeys**](AccessKeys.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAccessKeysAdministration

> AccessKeysAdministration GetAccessKeysAdministration(ctx).ClientId(clientId).CreatedBy(createdBy).Execute()

Get a list of all access keys.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	clientId := "3434c-12323ab-4ce2ea6-b44fc-db344d" // string | Filter results by clientId (optional)
	createdBy := "test@run.ai" // string | Filter results by creatdBy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessKeysAPI.GetAccessKeysAdministration(context.Background()).ClientId(clientId).CreatedBy(createdBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessKeysAPI.GetAccessKeysAdministration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAccessKeysAdministration`: AccessKeysAdministration
	fmt.Fprintf(os.Stdout, "Response from `AccessKeysAPI.GetAccessKeysAdministration`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAccessKeysAdministrationRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clientId** | **string** | Filter results by clientId | 
 **createdBy** | **string** | Filter results by creatdBy | 

### Return type

[**AccessKeysAdministration**](AccessKeysAdministration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RegenerateAccessKeySecret

> RegenerateAccessKeySecret200Response RegenerateAccessKeySecret(ctx, accessKeyId).Execute()

Regenerate an access key secret.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

func main() {
	accessKeyId := "accessKeyId_example" // string | The access key ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessKeysAPI.RegenerateAccessKeySecret(context.Background(), accessKeyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessKeysAPI.RegenerateAccessKeySecret``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RegenerateAccessKeySecret`: RegenerateAccessKeySecret200Response
	fmt.Fprintf(os.Stdout, "Response from `AccessKeysAPI.RegenerateAccessKeySecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessKeyId** | **string** | The access key ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiRegenerateAccessKeySecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RegenerateAccessKeySecret200Response**](RegenerateAccessKeySecret200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

