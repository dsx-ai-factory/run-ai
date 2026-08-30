# \ServiceAccountsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateServiceAccount**](ServiceAccountsAPI.md#CreateServiceAccount) | **Post** /api/v1/service-accounts | Create a service account.
[**CreateServiceAccountInternally**](ServiceAccountsAPI.md#CreateServiceAccountInternally) | **Post** /internal/service-accounts | Create a service account internally
[**DeleteServiceAccountById**](ServiceAccountsAPI.md#DeleteServiceAccountById) | **Delete** /api/v1/service-accounts/{serviceAccountId} | Delete a service account by id.
[**GetInstallerServiceAccount**](ServiceAccountsAPI.md#GetInstallerServiceAccount) | **Get** /api/v1/service-accounts/installer | Get the installer service account.
[**GetServiceAccountById**](ServiceAccountsAPI.md#GetServiceAccountById) | **Get** /api/v1/service-accounts/{serviceAccountId} | Get service account by id.
[**GetServiceAccounts**](ServiceAccountsAPI.md#GetServiceAccounts) | **Get** /api/v1/service-accounts | Get a list of service accounts.
[**MintInstallerBootstrapToken**](ServiceAccountsAPI.md#MintInstallerBootstrapToken) | **Post** /internal/installer-bootstrap | Mint a short-lived bootstrap JWT for the installer service account.
[**RegenerateServiceAccountSecret**](ServiceAccountsAPI.md#RegenerateServiceAccountSecret) | **Post** /api/v1/service-accounts/{serviceAccountId}/secret | Regenerate a service account secret.
[**UpdateServiceAccountById**](ServiceAccountsAPI.md#UpdateServiceAccountById) | **Patch** /api/v1/service-accounts/{serviceAccountId} | Update service account details by id.



## CreateServiceAccount

> ServiceAccountPostResponse CreateServiceAccount(ctx).ServiceAccountCreationRequest(serviceAccountCreationRequest).Execute()

Create a service account.



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
	serviceAccountCreationRequest := *openapiclient.NewServiceAccountCreationRequest("Name_example") // ServiceAccountCreationRequest | Service account object to create

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ServiceAccountsAPI.CreateServiceAccount(context.Background()).ServiceAccountCreationRequest(serviceAccountCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.CreateServiceAccount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateServiceAccount`: ServiceAccountPostResponse
	fmt.Fprintf(os.Stdout, "Response from `ServiceAccountsAPI.CreateServiceAccount`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateServiceAccountRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **serviceAccountCreationRequest** | [**ServiceAccountCreationRequest**](ServiceAccountCreationRequest.md) | Service account object to create | 

### Return type

[**ServiceAccountPostResponse**](ServiceAccountPostResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateServiceAccountInternally

> ServiceAccountPostResponse CreateServiceAccountInternally(ctx).InternalServiceAccountCreationRequest(internalServiceAccountCreationRequest).Execute()

Create a service account internally



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
	internalServiceAccountCreationRequest := *openapiclient.NewInternalServiceAccountCreationRequest("Name_example", "Realm_example") // InternalServiceAccountCreationRequest | Service account object to create

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ServiceAccountsAPI.CreateServiceAccountInternally(context.Background()).InternalServiceAccountCreationRequest(internalServiceAccountCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.CreateServiceAccountInternally``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateServiceAccountInternally`: ServiceAccountPostResponse
	fmt.Fprintf(os.Stdout, "Response from `ServiceAccountsAPI.CreateServiceAccountInternally`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateServiceAccountInternallyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **internalServiceAccountCreationRequest** | [**InternalServiceAccountCreationRequest**](InternalServiceAccountCreationRequest.md) | Service account object to create | 

### Return type

[**ServiceAccountPostResponse**](ServiceAccountPostResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteServiceAccountById

> DeleteServiceAccountById(ctx, serviceAccountId).Execute()

Delete a service account by id.



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
	serviceAccountId := "6d2894ba-f998-4039-bba1-caba57caf681" // string | Service account id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ServiceAccountsAPI.DeleteServiceAccountById(context.Background(), serviceAccountId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.DeleteServiceAccountById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**serviceAccountId** | **string** | Service account id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteServiceAccountByIdRequest struct via the builder pattern


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


## GetInstallerServiceAccount

> ServiceAccountPostResponse GetInstallerServiceAccount(ctx).TenantId(tenantId).Execute()

Get the installer service account.



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
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ServiceAccountsAPI.GetInstallerServiceAccount(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.GetInstallerServiceAccount``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInstallerServiceAccount`: ServiceAccountPostResponse
	fmt.Fprintf(os.Stdout, "Response from `ServiceAccountsAPI.GetInstallerServiceAccount`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetInstallerServiceAccountRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**ServiceAccountPostResponse**](ServiceAccountPostResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetServiceAccountById

> ServiceAccount GetServiceAccountById(ctx, serviceAccountId).Execute()

Get service account by id.



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
	serviceAccountId := "6d2894ba-f998-4039-bba1-caba57caf681" // string | Service account id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ServiceAccountsAPI.GetServiceAccountById(context.Background(), serviceAccountId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.GetServiceAccountById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetServiceAccountById`: ServiceAccount
	fmt.Fprintf(os.Stdout, "Response from `ServiceAccountsAPI.GetServiceAccountById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**serviceAccountId** | **string** | Service account id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetServiceAccountByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ServiceAccount**](ServiceAccount.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetServiceAccounts

> []ServiceAccount GetServiceAccounts(ctx).Execute()

Get a list of service accounts.



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
	resp, r, err := apiClient.ServiceAccountsAPI.GetServiceAccounts(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.GetServiceAccounts``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetServiceAccounts`: []ServiceAccount
	fmt.Fprintf(os.Stdout, "Response from `ServiceAccountsAPI.GetServiceAccounts`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetServiceAccountsRequest struct via the builder pattern


### Return type

[**[]ServiceAccount**](ServiceAccount.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MintInstallerBootstrapToken

> InstallerBootstrapResponse MintInstallerBootstrapToken(ctx).TenantId(tenantId).Execute()

Mint a short-lived bootstrap JWT for the installer service account.



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
	tenantId := "1008" // string | If exists will return the relevant resource for the tenant ID (optional) (default to "")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ServiceAccountsAPI.MintInstallerBootstrapToken(context.Background()).TenantId(tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.MintInstallerBootstrapToken``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `MintInstallerBootstrapToken`: InstallerBootstrapResponse
	fmt.Fprintf(os.Stdout, "Response from `ServiceAccountsAPI.MintInstallerBootstrapToken`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiMintInstallerBootstrapTokenRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantId** | **string** | If exists will return the relevant resource for the tenant ID | [default to &quot;&quot;]

### Return type

[**InstallerBootstrapResponse**](InstallerBootstrapResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RegenerateServiceAccountSecret

> RegenerateServiceAccountSecretResponse RegenerateServiceAccountSecret(ctx, serviceAccountId).Execute()

Regenerate a service account secret.



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
	serviceAccountId := "6d2894ba-f998-4039-bba1-caba57caf681" // string | Service account id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ServiceAccountsAPI.RegenerateServiceAccountSecret(context.Background(), serviceAccountId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.RegenerateServiceAccountSecret``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RegenerateServiceAccountSecret`: RegenerateServiceAccountSecretResponse
	fmt.Fprintf(os.Stdout, "Response from `ServiceAccountsAPI.RegenerateServiceAccountSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**serviceAccountId** | **string** | Service account id | 

### Other Parameters

Other parameters are passed through a pointer to a apiRegenerateServiceAccountSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RegenerateServiceAccountSecretResponse**](RegenerateServiceAccountSecretResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateServiceAccountById

> UpdateServiceAccountById(ctx, serviceAccountId).ServiceAccountPatchRequest(serviceAccountPatchRequest).Execute()

Update service account details by id.



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
	serviceAccountId := "6d2894ba-f998-4039-bba1-caba57caf681" // string | Service account id
	serviceAccountPatchRequest := *openapiclient.NewServiceAccountPatchRequest(false) // ServiceAccountPatchRequest | Service account object that needs to be updated. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ServiceAccountsAPI.UpdateServiceAccountById(context.Background(), serviceAccountId).ServiceAccountPatchRequest(serviceAccountPatchRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ServiceAccountsAPI.UpdateServiceAccountById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**serviceAccountId** | **string** | Service account id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateServiceAccountByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **serviceAccountPatchRequest** | [**ServiceAccountPatchRequest**](ServiceAccountPatchRequest.md) | Service account object that needs to be updated. | 

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

