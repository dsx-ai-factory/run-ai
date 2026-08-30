# \RealmAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRealm**](RealmAPI.md#CreateRealm) | **Post** /internal/realms | Create runai realm
[**DeleteRealm**](RealmAPI.md#DeleteRealm) | **Delete** /internal/realms/{realm} | delete realm
[**GetOrCreateRealm**](RealmAPI.md#GetOrCreateRealm) | **Put** /internal/realms | Get or create runai realm
[**ReconcileRealm**](RealmAPI.md#ReconcileRealm) | **Post** /internal/realms/{realm}/reconcile | Reconcile runai realm



## CreateRealm

> RealmCreationResponse CreateRealm(ctx).RealmCreationRequest(realmCreationRequest).Execute()

Create runai realm



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
	realmCreationRequest := *openapiclient.NewRealmCreationRequest("RealmName_example", "TenantId_example") // RealmCreationRequest | The tenant on behalf realm will be created.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RealmAPI.CreateRealm(context.Background()).RealmCreationRequest(realmCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RealmAPI.CreateRealm``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRealm`: RealmCreationResponse
	fmt.Fprintf(os.Stdout, "Response from `RealmAPI.CreateRealm`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRealmRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **realmCreationRequest** | [**RealmCreationRequest**](RealmCreationRequest.md) | The tenant on behalf realm will be created. | 

### Return type

[**RealmCreationResponse**](RealmCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRealm

> DeleteRealm(ctx, realm).IsHardDelete(isHardDelete).Execute()

delete realm

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
	realm := "realm_example" // string | The name of the realm
	isHardDelete := true // bool | Boolean indicator for soft/hard delete (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RealmAPI.DeleteRealm(context.Background(), realm).IsHardDelete(isHardDelete).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RealmAPI.DeleteRealm``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**realm** | **string** | The name of the realm | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRealmRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **isHardDelete** | **bool** | Boolean indicator for soft/hard delete | 

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


## GetOrCreateRealm

> RealmCreationResponse GetOrCreateRealm(ctx).RealmCreationRequest(realmCreationRequest).Execute()

Get or create runai realm



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
	realmCreationRequest := *openapiclient.NewRealmCreationRequest("RealmName_example", "TenantId_example") // RealmCreationRequest | The tenant on behalf realm will be created.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RealmAPI.GetOrCreateRealm(context.Background()).RealmCreationRequest(realmCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RealmAPI.GetOrCreateRealm``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOrCreateRealm`: RealmCreationResponse
	fmt.Fprintf(os.Stdout, "Response from `RealmAPI.GetOrCreateRealm`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetOrCreateRealmRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **realmCreationRequest** | [**RealmCreationRequest**](RealmCreationRequest.md) | The tenant on behalf realm will be created. | 

### Return type

[**RealmCreationResponse**](RealmCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ReconcileRealm

> ReconcileRealm(ctx, realm).RealmReconcileRequest(realmReconcileRequest).Execute()

Reconcile runai realm



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
	realm := "realm_example" // string | The name of the realm
	realmReconcileRequest := *openapiclient.NewRealmReconcileRequest() // RealmReconcileRequest | how to run reconcile (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RealmAPI.ReconcileRealm(context.Background(), realm).RealmReconcileRequest(realmReconcileRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RealmAPI.ReconcileRealm``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**realm** | **string** | The name of the realm | 

### Other Parameters

Other parameters are passed through a pointer to a apiReconcileRealmRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **realmReconcileRequest** | [**RealmReconcileRequest**](RealmReconcileRequest.md) | how to run reconcile | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

