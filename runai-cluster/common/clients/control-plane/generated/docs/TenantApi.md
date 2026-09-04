# \TenantApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateTenant**](TenantApi.md#CreateTenant) | **Post** /api/v1/tenants | Create tenant
[**DeleteTenantById**](TenantApi.md#DeleteTenantById) | **Delete** /api/v1/tenants/{tenantId} | soft delete tenant
[**GetTenant**](TenantApi.md#GetTenant) | **Get** /v1/k8s/tenant | Get tenant
[**GetTenantById**](TenantApi.md#GetTenantById) | **Get** /api/v1/tenants/{tenantId} | Get tenant
[**GetTenantConfig**](TenantApi.md#GetTenantConfig) | **Get** /v1/k8s/config | Get tenant auth config
[**GetTenantSettings**](TenantApi.md#GetTenantSettings) | **Get** /v1/k8s/setting | Get all tenant settings.
[**GetTenantSettingsByKey**](TenantApi.md#GetTenantSettingsByKey) | **Get** /v1/k8s/setting/{settingKey} | Get a tenant setting by key.
[**GetTenants**](TenantApi.md#GetTenants) | **Get** /api/v1/tenants | Get a list of tenants.
[**Logo**](TenantApi.md#Logo) | **Get** /v1/k8s/tenant/{tenantId}/logo | Get tenant logo.
[**PostTenantEula**](TenantApi.md#PostTenantEula) | **Post** /api/v1/sign-eula | Post tenant eula
[**UpdateTenantSetting**](TenantApi.md#UpdateTenantSetting) | **Put** /v1/k8s/setting | Update a tenant setting.
[**UploadTenantLogo**](TenantApi.md#UploadTenantLogo) | **Post** /v1/k8s/tenant/{tenantId}/logo | Upload a tenant logo.



## CreateTenant

> TenantCreationResponse CreateTenant(ctx).TenantCreationRequest(tenantCreationRequest).Execute()

Create tenant

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
    tenantCreationRequest := *openapiclient.NewTenantCreationRequest("Name_example", "first@run.ai") // TenantCreationRequest | The tenant to create.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.CreateTenant(context.Background()).TenantCreationRequest(tenantCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.CreateTenant``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateTenant`: TenantCreationResponse
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.CreateTenant`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantCreationRequest** | [**TenantCreationRequest**](TenantCreationRequest.md) | The tenant to create. | 

### Return type

[**TenantCreationResponse**](TenantCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteTenantById

> DeleteTenantById200Response DeleteTenantById(ctx, tenantId).Execute()

soft delete tenant

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
    tenantId := int32(56) // int32 | The id of the tenant to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.DeleteTenantById(context.Background(), tenantId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.DeleteTenantById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeleteTenantById`: DeleteTenantById200Response
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.DeleteTenantById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** | The id of the tenant to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTenantByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DeleteTenantById200Response**](DeleteTenantById200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenant

> Tenant GetTenant(ctx).TenantName(tenantName).Execute()

Get tenant



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
    tenantName := "tenantName_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.GetTenant(context.Background()).TenantName(tenantName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.GetTenant``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTenant`: Tenant
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.GetTenant`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantName** | **string** |  | 

### Return type

[**Tenant**](Tenant.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantById

> Tenant1 GetTenantById(ctx, tenantId).Execute()

Get tenant

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
    tenantId := int32(56) // int32 | The id of the tenant to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.GetTenantById(context.Background(), tenantId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.GetTenantById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTenantById`: Tenant1
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.GetTenantById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** | The id of the tenant to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Tenant1**](Tenant1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantConfig

> Config GetTenantConfig(ctx).TenantName(tenantName).Execute()

Get tenant auth config



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
    tenantName := "tenantName_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.GetTenantConfig(context.Background()).TenantName(tenantName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.GetTenantConfig``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTenantConfig`: Config
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.GetTenantConfig`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantName** | **string** |  | 

### Return type

[**Config**](Config.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantSettings

> []Setting GetTenantSettings(ctx).Execute()

Get all tenant settings.



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
    resp, r, err := apiClient.TenantApi.GetTenantSettings(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.GetTenantSettings``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTenantSettings`: []Setting
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.GetTenantSettings`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantSettingsRequest struct via the builder pattern


### Return type

[**[]Setting**](Setting.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenantSettingsByKey

> bool GetTenantSettingsByKey(ctx, settingKey).Execute()

Get a tenant setting by key.



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
    settingKey := "settingKey_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.GetTenantSettingsByKey(context.Background(), settingKey).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.GetTenantSettingsByKey``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTenantSettingsByKey`: bool
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.GetTenantSettingsByKey`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**settingKey** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantSettingsByKeyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**bool**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTenants

> []Tenant1 GetTenants(ctx).Execute()

Get a list of tenants.



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
    resp, r, err := apiClient.TenantApi.GetTenants(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.GetTenants``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetTenants`: []Tenant1
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.GetTenants`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetTenantsRequest struct via the builder pattern


### Return type

[**[]Tenant1**](Tenant1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Logo

> Logo200Response Logo(ctx, tenantId).Execute()

Get tenant logo.



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
    tenantId := int32(56) // int32 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.Logo(context.Background(), tenantId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.Logo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `Logo`: Logo200Response
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.Logo`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiLogoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Logo200Response**](Logo200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PostTenantEula

> PostTenantEula(ctx).PostTenantEulaRequest(postTenantEulaRequest).Execute()

Post tenant eula

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
    postTenantEulaRequest := *openapiclient.NewPostTenantEulaRequest("Email_example") // PostTenantEulaRequest | The user email that signing the eula.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.PostTenantEula(context.Background()).PostTenantEulaRequest(postTenantEulaRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.PostTenantEula``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPostTenantEulaRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **postTenantEulaRequest** | [**PostTenantEulaRequest**](PostTenantEulaRequest.md) | The user email that signing the eula. | 

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


## UpdateTenantSetting

> TenantSettingCreationResponse UpdateTenantSetting(ctx).TenantSettingCreationRequest(tenantSettingCreationRequest).Execute()

Update a tenant setting.



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
    tenantSettingCreationRequest := *openapiclient.NewTenantSettingCreationRequest() // TenantSettingCreationRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.UpdateTenantSetting(context.Background()).TenantSettingCreationRequest(tenantSettingCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.UpdateTenantSetting``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateTenantSetting`: TenantSettingCreationResponse
    fmt.Fprintf(os.Stdout, "Response from `TenantApi.UpdateTenantSetting`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTenantSettingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **tenantSettingCreationRequest** | [**TenantSettingCreationRequest**](TenantSettingCreationRequest.md) |  | 

### Return type

[**TenantSettingCreationResponse**](TenantSettingCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UploadTenantLogo

> UploadTenantLogo(ctx, tenantId).Logo200Response(logo200Response).Execute()

Upload a tenant logo.



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
    tenantId := int32(56) // int32 | 
    logo200Response := *openapiclient.NewLogo200Response("Logo_example") // Logo200Response |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.TenantApi.UploadTenantLogo(context.Background(), tenantId).Logo200Response(logo200Response).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `TenantApi.UploadTenantLogo``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **int32** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadTenantLogoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **logo200Response** | [**Logo200Response**](Logo200Response.md) |  | 

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

