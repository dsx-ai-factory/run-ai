# \InternalApi

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeleteScopeAccessRules**](InternalApi.md#DeleteScopeAccessRules) | **Delete** /internal/authorization/scope-access-rules | Delete access rules that assigned to the requested scope resource
[**DeleteSubjectAccessRules**](InternalApi.md#DeleteSubjectAccessRules) | **Delete** /internal/authorization/subject-access-rules | Delete access rules that assigned to the requested subject.
[**EvaluatePermission**](InternalApi.md#EvaluatePermission) | **Post** /internal/authorization/evaluate | Evaluate user permission to the requested resource. The returned permitted scopes are the scopes that the action on the resource is permitted, with OR between the multiple scopes.
[**GetScopeAccessRules**](InternalApi.md#GetScopeAccessRules) | **Get** /internal/authorization/scope-access-rules | Get access rules that assigned to the requested scope resource
[**GetSubjectAccessRules**](InternalApi.md#GetSubjectAccessRules) | **Get** /internal/authorization/subject-access-rules | List specific subject access rules.
[**SetAdminUser**](InternalApi.md#SetAdminUser) | **Post** /internal/authorization/admin-user | Sets user as an admin user.
[**SetClusterApp**](InternalApi.md#SetClusterApp) | **Post** /internal/authorization/cluster-app | Create an access rule with a cluster app role.
[**SetInstallerApp**](InternalApi.md#SetInstallerApp) | **Post** /internal/authorization/installer-app | Create an access rule with an installer app role.
[**UpdateScopeAccessRules**](InternalApi.md#UpdateScopeAccessRules) | **Put** /internal/authorization/scope-access-rules | Update access rules that assigned to the requested scope resource



## DeleteScopeAccessRules

> DeleteScopeAccessRules(ctx).ScopeType(scopeType).ScopeId(scopeId).LastUpdated(lastUpdated).Roles(roles).Execute()

Delete access rules that assigned to the requested scope resource

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
    scopeType := "project" // string | The type of resource we want to filter by.
    scopeId := "2" // string | The scope resource id that we want to filter by.
    lastUpdated := "2021-12-14T16:04:15.099Z" // string | Filter by last update time. (optional)
    roles := "ML engineer,L1 researcher" // string | The roles we want to filter by. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.DeleteScopeAccessRules(context.Background()).ScopeType(scopeType).ScopeId(scopeId).LastUpdated(lastUpdated).Roles(roles).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.DeleteScopeAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteScopeAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | The scope resource id that we want to filter by. | 
 **lastUpdated** | **string** | Filter by last update time. | 
 **roles** | **string** | The roles we want to filter by. | 

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


## DeleteSubjectAccessRules

> DeleteSubjectAccessRules(ctx).SubjectType(subjectType).SubjectId(subjectId).LastUpdated(lastUpdated).Roles(roles).Execute()

Delete access rules that assigned to the requested subject.

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
    subjectType := "user" // string | The type of resource we want to filter by.
    subjectId := "some.user@run.ai" // string | The subject id that we want to filter by.
    lastUpdated := "2021-12-14T16:04:15.099Z" // string | Filter by last update time. (optional)
    roles := "ML engineer,L1 researcher" // string | The roles we want to filter by. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.DeleteSubjectAccessRules(context.Background()).SubjectType(subjectType).SubjectId(subjectId).LastUpdated(lastUpdated).Roles(roles).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.DeleteSubjectAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteSubjectAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **subjectType** | **string** | The type of resource we want to filter by. | 
 **subjectId** | **string** | The subject id that we want to filter by. | 
 **lastUpdated** | **string** | Filter by last update time. | 
 **roles** | **string** | The roles we want to filter by. | 

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
    openapiclient "./openapi"
)

func main() {
    backendEvaluationRequest := *openapiclient.NewBackendEvaluationRequest(openapiclient.ResourceType("department"), openapiclient.Action("create")) // BackendEvaluationRequest | The request parameters.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.EvaluatePermission(context.Background()).BackendEvaluationRequest(backendEvaluationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.EvaluatePermission``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `EvaluatePermission`: PermittedScopes
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.EvaluatePermission`: %v\n", resp)
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


## GetScopeAccessRules

> GetAccessRules200Response GetScopeAccessRules(ctx).ClusterId(clusterId).Roles(roles).DepartmentIds(departmentIds).ProjectIds(projectIds).Execute()

Get access rules that assigned to the requested scope resource

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
    clusterId := "7daba6ee-2aa9-4ab3-b735-a3d4bcdff72f" // string | The cluster id to filter the response for. (optional)
    roles := "ML engineer,L1 researcher" // string | The roles we want to filter by. (optional)
    departmentIds := "1,2" // string | The department scope resource ids that we want to filter by. (optional)
    projectIds := "1,2" // string | The project scope resource ids that we want to filter by. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.GetScopeAccessRules(context.Background()).ClusterId(clusterId).Roles(roles).DepartmentIds(departmentIds).ProjectIds(projectIds).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.GetScopeAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetScopeAccessRules`: GetAccessRules200Response
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.GetScopeAccessRules`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetScopeAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **string** | The cluster id to filter the response for. | 
 **roles** | **string** | The roles we want to filter by. | 
 **departmentIds** | **string** | The department scope resource ids that we want to filter by. | 
 **projectIds** | **string** | The project scope resource ids that we want to filter by. | 

### Return type

[**GetAccessRules200Response**](GetAccessRules200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSubjectAccessRules

> []AccessRule GetSubjectAccessRules(ctx).SubjectType(subjectType).SubjectIds(subjectIds).Execute()

List specific subject access rules.

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
    subjectType := "user" // string | The type of resource we want to filter by.
    subjectIds := []string{"some.user@run.ai"} // []string | The ids of the subjects to filter the response for. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.GetSubjectAccessRules(context.Background()).SubjectType(subjectType).SubjectIds(subjectIds).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.GetSubjectAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetSubjectAccessRules`: []AccessRule
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.GetSubjectAccessRules`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetSubjectAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **subjectType** | **string** | The type of resource we want to filter by. | 
 **subjectIds** | **[]string** | The ids of the subjects to filter the response for. | 

### Return type

[**[]AccessRule**](AccessRule.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetAdminUser

> AccessRule SetAdminUser(ctx).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()

Sets user as an admin user.

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
    internalCreateAccessRuleRequest := *openapiclient.NewInternalCreateAccessRuleRequest("user@run.ai", openapiclient.SubjectType("user")) // InternalCreateAccessRuleRequest | The request parameters.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.SetAdminUser(context.Background()).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.SetAdminUser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SetAdminUser`: AccessRule
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.SetAdminUser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetAdminUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **internalCreateAccessRuleRequest** | [**InternalCreateAccessRuleRequest**](InternalCreateAccessRuleRequest.md) | The request parameters. | 

### Return type

[**AccessRule**](AccessRule.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetClusterApp

> AccessRule SetClusterApp(ctx).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()

Create an access rule with a cluster app role.

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
    internalCreateAccessRuleRequest := *openapiclient.NewInternalCreateAccessRuleRequest("user@run.ai", openapiclient.SubjectType("user")) // InternalCreateAccessRuleRequest | The request parameters.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.SetClusterApp(context.Background()).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.SetClusterApp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SetClusterApp`: AccessRule
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.SetClusterApp`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetClusterAppRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **internalCreateAccessRuleRequest** | [**InternalCreateAccessRuleRequest**](InternalCreateAccessRuleRequest.md) | The request parameters. | 

### Return type

[**AccessRule**](AccessRule.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SetInstallerApp

> AccessRule SetInstallerApp(ctx).XRunaiTenantId(xRunaiTenantId).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()

Create an access rule with an installer app role.

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
    xRunaiTenantId := int32(1) // int32 | The id of the tenant
    internalCreateAccessRuleRequest := *openapiclient.NewInternalCreateAccessRuleRequest("user@run.ai", openapiclient.SubjectType("user")) // InternalCreateAccessRuleRequest | The request parameters.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.SetInstallerApp(context.Background()).XRunaiTenantId(xRunaiTenantId).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.SetInstallerApp``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SetInstallerApp`: AccessRule
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.SetInstallerApp`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSetInstallerAppRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xRunaiTenantId** | **int32** | The id of the tenant | 
 **internalCreateAccessRuleRequest** | [**InternalCreateAccessRuleRequest**](InternalCreateAccessRuleRequest.md) | The request parameters. | 

### Return type

[**AccessRule**](AccessRule.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateScopeAccessRules

> UpdateScopeAccessRules(ctx).ScopeType(scopeType).ScopeId(scopeId).UpdateScopeAccessRulesRequest(updateScopeAccessRulesRequest).Execute()

Update access rules that assigned to the requested scope resource

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
    scopeType := "project" // string | The type of resource we want to filter by.
    scopeId := "2" // string | The scope resource id that we want to filter by.
    updateScopeAccessRulesRequest := *openapiclient.NewUpdateScopeAccessRulesRequest("team-a") // UpdateScopeAccessRulesRequest | The access rule fields to update.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.UpdateScopeAccessRules(context.Background()).ScopeType(scopeType).ScopeId(scopeId).UpdateScopeAccessRulesRequest(updateScopeAccessRulesRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.UpdateScopeAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateScopeAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | The scope resource id that we want to filter by. | 
 **updateScopeAccessRulesRequest** | [**UpdateScopeAccessRulesRequest**](UpdateScopeAccessRulesRequest.md) | The access rule fields to update. | 

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

