# \PermissionsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**EvaluatePermission**](PermissionsAPI.md#EvaluatePermission) | **Post** /api/v1/authorization/evaluate | Evaluate user permission to the requested resource. The returned permitted scopes are the scopes that the action on the resource is permitted, with OR between the multiple scopes.
[**GetPermissionSet**](PermissionsAPI.md#GetPermissionSet) | **Get** /api/v1/authorization/permission-sets/{permissionSetId} | Get permission set by id.
[**GetPermissionSets**](PermissionsAPI.md#GetPermissionSets) | **Get** /api/v1/authorization/permission-sets | Get permission sets.
[**GetPermissions**](PermissionsAPI.md#GetPermissions) | **Get** /api/v1/authorization/permissions | Get a summary of user permissions.
[**GetPermittedScopes**](PermissionsAPI.md#GetPermittedScopes) | **Post** /api/v1/authorization/permitted-scopes | Calculate permitted scopes.



## EvaluatePermission

> PermittedScopes EvaluatePermission(ctx).BackendEvaluationRequest(backendEvaluationRequest).Execute()

Evaluate user permission to the requested resource. The returned permitted scopes are the scopes that the action on the resource is permitted, with OR between the multiple scopes.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	backendEvaluationRequest := *openapiclient.NewBackendEvaluationRequest() // BackendEvaluationRequest | The request parameters.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PermissionsAPI.EvaluatePermission(context.Background()).BackendEvaluationRequest(backendEvaluationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PermissionsAPI.EvaluatePermission``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `EvaluatePermission`: PermittedScopes
	fmt.Fprintf(os.Stdout, "Response from `PermissionsAPI.EvaluatePermission`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiEvaluatePermissionRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **backendEvaluationRequest** | [**BackendEvaluationRequest**](BackendEvaluationRequest.md) | The request parameters. | 

### Return type

[**PermittedScopes**](PermittedScopes.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPermissionSet

> PermissionSet GetPermissionSet(ctx, permissionSetId).Execute()

Get permission set by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	permissionSetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PermissionsAPI.GetPermissionSet(context.Background(), permissionSetId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PermissionsAPI.GetPermissionSet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPermissionSet`: PermissionSet
	fmt.Fprintf(os.Stdout, "Response from `PermissionsAPI.GetPermissionSet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**permissionSetId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetPermissionSetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**PermissionSet**](PermissionSet.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPermissionSets

> GetPermissionSets200Response GetPermissionSets(ctx).Execute()

Get permission sets.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PermissionsAPI.GetPermissionSets(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PermissionsAPI.GetPermissionSets``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPermissionSets`: GetPermissionSets200Response
	fmt.Fprintf(os.Stdout, "Response from `PermissionsAPI.GetPermissionSets`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetPermissionSetsRequest struct via the builder pattern


### Return type

[**GetPermissionSets200Response**](GetPermissionSets200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPermissions

> []Permission GetPermissions(ctx).Execute()

Get a summary of user permissions.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PermissionsAPI.GetPermissions(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PermissionsAPI.GetPermissions``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPermissions`: []Permission
	fmt.Fprintf(os.Stdout, "Response from `PermissionsAPI.GetPermissions`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetPermissionsRequest struct via the builder pattern


### Return type

[**[]Permission**](Permission.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPermittedScopes

> PermittedScopesActions GetPermittedScopes(ctx).GetPermittedScopesRequest(getPermittedScopesRequest).Execute()

Calculate permitted scopes.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	getPermittedScopesRequest := *openapiclient.NewGetPermittedScopesRequest(openapiclient.ResourceType("department")) // GetPermittedScopesRequest | The request parameters.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PermissionsAPI.GetPermittedScopes(context.Background()).GetPermittedScopesRequest(getPermittedScopesRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PermissionsAPI.GetPermittedScopes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPermittedScopes`: PermittedScopesActions
	fmt.Fprintf(os.Stdout, "Response from `PermissionsAPI.GetPermittedScopes`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetPermittedScopesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **getPermittedScopesRequest** | [**GetPermittedScopesRequest**](GetPermittedScopesRequest.md) | The request parameters. | 

### Return type

[**PermittedScopesActions**](PermittedScopesActions.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

