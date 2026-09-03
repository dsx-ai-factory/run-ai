# \InternalApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateInternalAccessRule**](InternalApi.md#CreateInternalAccessRule) | **Post** /internal/authorization/access-rules | Create internal access rule with out privileged escalation validation.
[**DeleteScopeAccessRules**](InternalApi.md#DeleteScopeAccessRules) | **Delete** /internal/authorization/scope-access-rules | Delete access rules that assigned to the requested scope resource
[**DeleteSubjectAccessRules**](InternalApi.md#DeleteSubjectAccessRules) | **Delete** /internal/authorization/subject-access-rules | Delete access rules that assigned to the requested subject.
[**EvaluatePermission**](InternalApi.md#EvaluatePermission) | **Post** /internal/authorization/evaluate | Evaluate user permission to the requested resource. The returned permitted scopes are the scopes that the action on the resource is permitted, with OR between the multiple scopes.
[**GetClustersHistories**](InternalApi.md#GetClustersHistories) | **Get** /internal/cluster-histories | List cluster histories.
[**GetScopeAccessRules**](InternalApi.md#GetScopeAccessRules) | **Get** /internal/authorization/scope-access-rules | Get access rules that assigned to the requested scope resource
[**GetSubjectAccessRules**](InternalApi.md#GetSubjectAccessRules) | **Get** /internal/authorization/subject-access-rules | List specific subject access rules.
[**SetClusterAccessControl**](InternalApi.md#SetClusterAccessControl) | **Put** /internal/clusters/{clusterUuid}/access-control | Set the cluster access control.
[**UpdateScopeAccessRules**](InternalApi.md#UpdateScopeAccessRules) | **Put** /internal/authorization/scope-access-rules | Update access rules that assigned to the requested scope resource



## CreateInternalAccessRule

> AccessRule CreateInternalAccessRule(ctx).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()

Create internal access rule with out privileged escalation validation.

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
    internalCreateAccessRuleRequest := *openapiclient.NewInternalCreateAccessRuleRequest("user@run.ai", openapiclient.SubjectType("user"), openapiclient.AllowedInternalCreateRoles("System administrator")) // InternalCreateAccessRuleRequest | The request parameters.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.CreateInternalAccessRule(context.Background()).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.CreateInternalAccessRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateInternalAccessRule`: AccessRule
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.CreateInternalAccessRule`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateInternalAccessRuleRequest struct via the builder pattern


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
    scopeType := "scopeType_example" // string | The resource scope type to filter by.
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
 **scopeType** | **string** | The resource scope type to filter by. | 
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


## GetClustersHistories

> []ClusterHistory GetClustersHistories(ctx).ClusterId(clusterId).Execute()

List cluster histories.

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
    clusterId := []string{"Inner_example"} // []string | The clusters Ids of the current request.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.GetClustersHistories(context.Background()).ClusterId(clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.GetClustersHistories``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetClustersHistories`: []ClusterHistory
    fmt.Fprintf(os.Stdout, "Response from `InternalApi.GetClustersHistories`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetClustersHistoriesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterId** | **[]string** | The clusters Ids of the current request. | 

### Return type

[**[]ClusterHistory**](ClusterHistory.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
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
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
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
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
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


## SetClusterAccessControl

> SetClusterAccessControl(ctx, clusterUuid).Execute()

Set the cluster access control.

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.InternalApi.SetClusterAccessControl(context.Background(), clusterUuid).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `InternalApi.SetClusterAccessControl``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSetClusterAccessControlRequest struct via the builder pattern


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
    scopeType := "scopeType_example" // string | The resource scope type to filter by.
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
 **scopeType** | **string** | The resource scope type to filter by. | 
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

