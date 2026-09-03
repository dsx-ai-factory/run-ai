# \TenantRouteAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRoute**](TenantRouteAPI.md#CreateRoute) | **Post** /api/v1/tenants/{tenantId}/routes | 
[**DeleteRoute**](TenantRouteAPI.md#DeleteRoute) | **Delete** /api/v1/tenants/{tenantId}/routes/{routeId} | 
[**GetAllTenantsWithRoutes**](TenantRouteAPI.md#GetAllTenantsWithRoutes) | **Get** /api/v1/tenants | 
[**GetRoute**](TenantRouteAPI.md#GetRoute) | **Get** /api/v1/tenants/{tenantId}/routes/{routeId} | 
[**GetRoutes**](TenantRouteAPI.md#GetRoutes) | **Get** /api/v1/tenants/{tenantId}/routes | 
[**UpdateRoute**](TenantRouteAPI.md#UpdateRoute) | **Put** /api/v1/tenants/{tenantId}/routes/{routeId} | 



## CreateRoute

> string CreateRoute(ctx, tenantId).RouteData(routeData).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of relevant tenant
	routeData := *openapiclient.NewRouteData("12345", "RouteId_example", map[string]interface{}(123)) // RouteData | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantRouteAPI.CreateRoute(context.Background(), tenantId).RouteData(routeData).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantRouteAPI.CreateRoute``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRoute`: string
	fmt.Fprintf(os.Stdout, "Response from `TenantRouteAPI.CreateRoute`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of relevant tenant | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateRouteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **routeData** | [**RouteData**](RouteData.md) |  | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRoute

> string DeleteRoute(ctx, tenantId, routeId).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of relevant tenant
	routeId := "routeId_example" // string | ID of relevant tenant route

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantRouteAPI.DeleteRoute(context.Background(), tenantId, routeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantRouteAPI.DeleteRoute``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteRoute`: string
	fmt.Fprintf(os.Stdout, "Response from `TenantRouteAPI.DeleteRoute`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of relevant tenant | 
**routeId** | **string** | ID of relevant tenant route | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRouteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAllTenantsWithRoutes

> []string GetAllTenantsWithRoutes(ctx).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantRouteAPI.GetAllTenantsWithRoutes(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantRouteAPI.GetAllTenantsWithRoutes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAllTenantsWithRoutes`: []string
	fmt.Fprintf(os.Stdout, "Response from `TenantRouteAPI.GetAllTenantsWithRoutes`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetAllTenantsWithRoutesRequest struct via the builder pattern


### Return type

**[]string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRoute

> RouteData GetRoute(ctx, tenantId, routeId).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of relevant tenant
	routeId := "routeId_example" // string | ID of relevant tenant route

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantRouteAPI.GetRoute(context.Background(), tenantId, routeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantRouteAPI.GetRoute``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRoute`: RouteData
	fmt.Fprintf(os.Stdout, "Response from `TenantRouteAPI.GetRoute`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of relevant tenant | 
**routeId** | **string** | ID of relevant tenant route | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRouteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RouteData**](RouteData.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRoutes

> []RouteData GetRoutes(ctx, tenantId).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of relevant tenant

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantRouteAPI.GetRoutes(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantRouteAPI.GetRoutes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRoutes`: []RouteData
	fmt.Fprintf(os.Stdout, "Response from `TenantRouteAPI.GetRoutes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of relevant tenant | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRoutesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]RouteData**](RouteData.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateRoute

> string UpdateRoute(ctx, tenantId, routeId).RouteDataForUpdate(routeDataForUpdate).Execute()





### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	tenantId := "tenantId_example" // string | ID of relevant tenant
	routeId := "routeId_example" // string | ID of relevant tenant route
	routeDataForUpdate := *openapiclient.NewRouteDataForUpdate("12345", map[string]interface{}(123)) // RouteDataForUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.TenantRouteAPI.UpdateRoute(context.Background(), tenantId, routeId).RouteDataForUpdate(routeDataForUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TenantRouteAPI.UpdateRoute``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateRoute`: string
	fmt.Fprintf(os.Stdout, "Response from `TenantRouteAPI.UpdateRoute`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of relevant tenant | 
**routeId** | **string** | ID of relevant tenant route | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRouteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **routeDataForUpdate** | [**RouteDataForUpdate**](RouteDataForUpdate.md) |  | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

