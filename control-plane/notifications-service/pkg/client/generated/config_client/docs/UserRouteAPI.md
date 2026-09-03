# \UserRouteAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateUserRoute**](UserRouteAPI.md#CreateUserRoute) | **Post** /api/v1/tenants/{tenantId}/users/{userId}/routes | 
[**DeleteAllUserRoutes**](UserRouteAPI.md#DeleteAllUserRoutes) | **Delete** /api/v1/tenants/{tenantId}/users/routes | 
[**DeleteNotifyIdFromAllUserRoutes**](UserRouteAPI.md#DeleteNotifyIdFromAllUserRoutes) | **Delete** /api/v1/tenants/{tenantId}/users/routes/notify/{notifyId} | 
[**DeleteUserRoute**](UserRouteAPI.md#DeleteUserRoute) | **Delete** /api/v1/tenants/{tenantId}/users/{userId}/routes/{routeId} | 
[**GetUserRoutes**](UserRouteAPI.md#GetUserRoutes) | **Get** /api/v1/tenants/{tenantId}/users/{userId}/routes | 
[**UpdateUserRoute**](UserRouteAPI.md#UpdateUserRoute) | **Put** /api/v1/tenants/{tenantId}/users/{userId}/routes/{routeId} | 



## CreateUserRoute

> string CreateUserRoute(ctx, tenantId, userId).UserRouteData(userRouteData).Execute()





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
	tenantId := "tenantId_example" // string | ID of the relevant tenant
	userId := "userId_example" // string | ID of the relevant user
	userRouteData := *openapiclient.NewUserRouteData("TenantId_example", "UserId_example", "RouteId_example", map[string]interface{}(123)) // UserRouteData | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserRouteAPI.CreateUserRoute(context.Background(), tenantId, userId).UserRouteData(userRouteData).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserRouteAPI.CreateUserRoute``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateUserRoute`: string
	fmt.Fprintf(os.Stdout, "Response from `UserRouteAPI.CreateUserRoute`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of the relevant tenant | 
**userId** | **string** | ID of the relevant user | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateUserRouteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **userRouteData** | [**UserRouteData**](UserRouteData.md) |  | 

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


## DeleteAllUserRoutes

> string DeleteAllUserRoutes(ctx, tenantId).Execute()





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
	resp, r, err := apiClient.UserRouteAPI.DeleteAllUserRoutes(context.Background(), tenantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserRouteAPI.DeleteAllUserRoutes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteAllUserRoutes`: string
	fmt.Fprintf(os.Stdout, "Response from `UserRouteAPI.DeleteAllUserRoutes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of relevant tenant | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteAllUserRoutesRequest struct via the builder pattern


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


## DeleteNotifyIdFromAllUserRoutes

> string DeleteNotifyIdFromAllUserRoutes(ctx, tenantId, notifyId).Execute()





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
	tenantId := "tenantId_example" // string | ID of the relevant tenant
	notifyId := "notifyId_example" // string | ID of the relevant notify/sink

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserRouteAPI.DeleteNotifyIdFromAllUserRoutes(context.Background(), tenantId, notifyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserRouteAPI.DeleteNotifyIdFromAllUserRoutes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteNotifyIdFromAllUserRoutes`: string
	fmt.Fprintf(os.Stdout, "Response from `UserRouteAPI.DeleteNotifyIdFromAllUserRoutes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of the relevant tenant | 
**notifyId** | **string** | ID of the relevant notify/sink | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNotifyIdFromAllUserRoutesRequest struct via the builder pattern


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


## DeleteUserRoute

> string DeleteUserRoute(ctx, tenantId, userId, routeId).Execute()





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
	tenantId := "tenantId_example" // string | ID of the relevant tenant
	userId := "userId_example" // string | ID of the relevant user
	routeId := "routeId_example" // string | ID of the relevant route

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserRouteAPI.DeleteUserRoute(context.Background(), tenantId, userId, routeId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserRouteAPI.DeleteUserRoute``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteUserRoute`: string
	fmt.Fprintf(os.Stdout, "Response from `UserRouteAPI.DeleteUserRoute`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of the relevant tenant | 
**userId** | **string** | ID of the relevant user | 
**routeId** | **string** | ID of the relevant route | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteUserRouteRequest struct via the builder pattern


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


## GetUserRoutes

> []RouteData GetUserRoutes(ctx, tenantId, userId).Execute()





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
	tenantId := "tenantId_example" // string | ID of the relevant tenant
	userId := "userId_example" // string | ID of the relevant user

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserRouteAPI.GetUserRoutes(context.Background(), tenantId, userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserRouteAPI.GetUserRoutes``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUserRoutes`: []RouteData
	fmt.Fprintf(os.Stdout, "Response from `UserRouteAPI.GetUserRoutes`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of the relevant tenant | 
**userId** | **string** | ID of the relevant user | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserRoutesRequest struct via the builder pattern


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


## UpdateUserRoute

> string UpdateUserRoute(ctx, tenantId, userId, routeId).UserRouteDataForUpdate(userRouteDataForUpdate).Execute()





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
	tenantId := "tenantId_example" // string | ID of the relevant tenant
	userId := "userId_example" // string | ID of the relevant user
	routeId := "routeId_example" // string | ID of the relevant route
	userRouteDataForUpdate := *openapiclient.NewUserRouteDataForUpdate("TenantId_example", "UserId_example", map[string]interface{}(123)) // UserRouteDataForUpdate | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserRouteAPI.UpdateUserRoute(context.Background(), tenantId, userId, routeId).UserRouteDataForUpdate(userRouteDataForUpdate).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserRouteAPI.UpdateUserRoute``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateUserRoute`: string
	fmt.Fprintf(os.Stdout, "Response from `UserRouteAPI.UpdateUserRoute`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**tenantId** | **string** | ID of the relevant tenant | 
**userId** | **string** | ID of the relevant user | 
**routeId** | **string** | ID of the relevant route | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateUserRouteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **userRouteDataForUpdate** | [**UserRouteDataForUpdate**](UserRouteDataForUpdate.md) |  | 

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

