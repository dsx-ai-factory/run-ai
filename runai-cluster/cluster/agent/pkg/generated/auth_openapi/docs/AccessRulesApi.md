# \AccessRulesApi

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateAccessRule**](AccessRulesApi.md#CreateAccessRule) | **Post** /api/v1/authorization/access-rules | Create an access rule. Bind a predefined role to a subject (user, group or app) in some scope.
[**DeleteAccessRule**](AccessRulesApi.md#DeleteAccessRule) | **Delete** /api/v1/authorization/access-rules/{accessRuleId} | Delete access rule. This operation will delete the subject permissions that assigned to it by this access rule.
[**GetAccessRule**](AccessRulesApi.md#GetAccessRule) | **Get** /api/v1/authorization/access-rules/{accessRuleId} | Get access rule.
[**GetAccessRules**](AccessRulesApi.md#GetAccessRules) | **Get** /api/v1/authorization/access-rules | List access rules.
[**SyncAccessRules**](AccessRulesApi.md#SyncAccessRules) | **Post** /api/v1/authorization/access-rules/{clusterId}/sync | Sync the statuses of the access rules in the cluster.
[**UpdateAccessRuleStatus**](AccessRulesApi.md#UpdateAccessRuleStatus) | **Put** /api/v1/authorization/access-rules/{accessRuleId}/{clusterId}/status | Update the status of the access rule in the cluster.



## CreateAccessRule

> AccessRule CreateAccessRule(ctx).AccessRuleCreationFields(accessRuleCreationFields).Execute()

Create an access rule. Bind a predefined role to a subject (user, group or app) in some scope.

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
    accessRuleCreationFields := *openapiclient.NewAccessRuleCreationFields("user@run.ai", openapiclient.SubjectType("user"), int32(53142648), "a418ed33-9399-48c0-a890-122cadd13bfd", openapiclient.ScopeType("tenant")) // AccessRuleCreationFields | The access rule to create.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.CreateAccessRule(context.Background()).AccessRuleCreationFields(accessRuleCreationFields).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesApi.CreateAccessRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateAccessRule`: AccessRule
    fmt.Fprintf(os.Stdout, "Response from `AccessRulesApi.CreateAccessRule`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateAccessRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **accessRuleCreationFields** | [**AccessRuleCreationFields**](AccessRuleCreationFields.md) | The access rule to create. | 

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


## DeleteAccessRule

> DeleteAccessRule(ctx, accessRuleId).Execute()

Delete access rule. This operation will delete the subject permissions that assigned to it by this access rule.

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
    accessRuleId := int32(32) // int32 | The id of the access rule to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.DeleteAccessRule(context.Background(), accessRuleId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesApi.DeleteAccessRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessRuleId** | **int32** | The id of the access rule to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteAccessRuleRequest struct via the builder pattern


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


## GetAccessRule

> AccessRule GetAccessRule(ctx, accessRuleId).Execute()

Get access rule.

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
    accessRuleId := int32(32) // int32 | The id of the access rule to retrieve

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.GetAccessRule(context.Background(), accessRuleId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesApi.GetAccessRule``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAccessRule`: AccessRule
    fmt.Fprintf(os.Stdout, "Response from `AccessRulesApi.GetAccessRule`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessRuleId** | **int32** | The id of the access rule to retrieve | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetAccessRuleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**AccessRule**](AccessRule.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAccessRules

> GetAccessRules200Response GetAccessRules(ctx).SubjectType(subjectType).SubjectIdFilter(subjectIdFilter).SubjectIds(subjectIds).Limit(limit).Offset(offset).LastUpdated(lastUpdated).IncludeDeleted(includeDeleted).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).RoleId(roleId).Execute()

List access rules.

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
    subjectType := "user" // string | The type of resource we want to filter by. (optional)
    subjectIdFilter := "some.user" // string | Part of the subject id that we want to filter by. (optional)
    subjectIds := []string{"some.user@run.ai"} // []string | The ids of the subjects to filter the response for. (optional)
    limit := int32(1000) // int32 | The maximum number of entries to return. (optional)
    offset := int32(4000) // int32 | The offset of the first item returned in the collection. (optional)
    lastUpdated := "2021-12-14T16:04:15.099Z" // string | Filter by last update time. (optional)
    includeDeleted := false // bool | True to include deleted objects in the result. (optional) (default to false)
    clusterId := "7daba6ee-2aa9-4ab3-b735-a3d4bcdff72f" // string | The cluster id to filter the response for. (optional)
    scopeType := "project" // string | The type of resource we want to filter by. (optional)
    scopeId := "2" // string | The scope resource id that we want to filter by. (optional)
    roleId := int32(56) // int32 | The role id we want to filter by. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.GetAccessRules(context.Background()).SubjectType(subjectType).SubjectIdFilter(subjectIdFilter).SubjectIds(subjectIds).Limit(limit).Offset(offset).LastUpdated(lastUpdated).IncludeDeleted(includeDeleted).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).RoleId(roleId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesApi.GetAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetAccessRules`: GetAccessRules200Response
    fmt.Fprintf(os.Stdout, "Response from `AccessRulesApi.GetAccessRules`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **subjectType** | **string** | The type of resource we want to filter by. | 
 **subjectIdFilter** | **string** | Part of the subject id that we want to filter by. | 
 **subjectIds** | **[]string** | The ids of the subjects to filter the response for. | 
 **limit** | **int32** | The maximum number of entries to return. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **lastUpdated** | **string** | Filter by last update time. | 
 **includeDeleted** | **bool** | True to include deleted objects in the result. | [default to false]
 **clusterId** | **string** | The cluster id to filter the response for. | 
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | The scope resource id that we want to filter by. | 
 **roleId** | **int32** | The role id we want to filter by. | 

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


## SyncAccessRules

> SyncAccessRules(ctx, clusterId).AccessRuleForSync(accessRuleForSync).Execute()

Sync the statuses of the access rules in the cluster.



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
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The clusterId of the current status update
    accessRuleForSync := []openapiclient.AccessRuleForSync{*openapiclient.NewAccessRuleForSync("Ready", "Failed to enforce department 'department-a'", int32(32))} // []AccessRuleForSync | The statuses to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.SyncAccessRules(context.Background(), clusterId).AccessRuleForSync(accessRuleForSync).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesApi.SyncAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The clusterId of the current status update | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **accessRuleForSync** | [**[]AccessRuleForSync**](AccessRuleForSync.md) | The statuses to sync. | 

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


## UpdateAccessRuleStatus

> UpdateAccessRuleStatus(ctx, accessRuleId, clusterId).AccessRuleStatusRequest(accessRuleStatusRequest).Execute()

Update the status of the access rule in the cluster.

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
    accessRuleId := int32(32) // int32 | The id of the access rule to retrieve
    clusterId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The clusterId of the current status update
    accessRuleStatusRequest := *openapiclient.NewAccessRuleStatusRequest("Ready", "Failed to enforce department 'department-a'") // AccessRuleStatusRequest | The status of the access rule in the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.UpdateAccessRuleStatus(context.Background(), accessRuleId, clusterId).AccessRuleStatusRequest(accessRuleStatusRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesApi.UpdateAccessRuleStatus``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**accessRuleId** | **int32** | The id of the access rule to retrieve | 
**clusterId** | **string** | The clusterId of the current status update | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateAccessRuleStatusRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **accessRuleStatusRequest** | [**AccessRuleStatusRequest**](AccessRuleStatusRequest.md) | The status of the access rule in the cluster. | 

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

