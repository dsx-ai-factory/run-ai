# \AccessRulesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountAccessRules**](AccessRulesApi.md#CountAccessRules) | **Get** /api/v1/authorization/access-rules/count | Count access rules.
[**CreateAccessRule**](AccessRulesApi.md#CreateAccessRule) | **Post** /api/v1/authorization/access-rules | Create an access rule.
[**DeleteAccessRule**](AccessRulesApi.md#DeleteAccessRule) | **Delete** /api/v1/authorization/access-rules/{accessRuleId} | Delete an access rule.
[**GetAccessRule**](AccessRulesApi.md#GetAccessRule) | **Get** /api/v1/authorization/access-rules/{accessRuleId} | Get an access rule.
[**GetAccessRules**](AccessRulesApi.md#GetAccessRules) | **Get** /api/v1/authorization/access-rules | List the access rules.
[**SyncAccessRules**](AccessRulesApi.md#SyncAccessRules) | **Post** /api/v1/authorization/access-rules/{clusterUuid}/sync | Sync access rules.
[**UpdateAccessRuleStatus**](AccessRulesApi.md#UpdateAccessRuleStatus) | **Put** /api/v1/authorization/access-rules/{accessRuleId}/{clusterUuid}/status | Update an access rule.



## CountAccessRules

> CountAccessRules200Response CountAccessRules(ctx).IncludeDeleted(includeDeleted).FilterBy(filterBy).Execute()

Count access rules.



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
    includeDeleted := false // bool | True to include deleted objects in the result. (optional) (default to false)
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.CountAccessRules(context.Background()).IncludeDeleted(includeDeleted).FilterBy(filterBy).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesApi.CountAccessRules``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CountAccessRules`: CountAccessRules200Response
    fmt.Fprintf(os.Stdout, "Response from `AccessRulesApi.CountAccessRules`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **includeDeleted** | **bool** | True to include deleted objects in the result. | [default to false]
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

### Return type

[**CountAccessRules200Response**](CountAccessRules200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateAccessRule

> AccessRule CreateAccessRule(ctx).AccessRuleCreationFields(accessRuleCreationFields).Execute()

Create an access rule.



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
    accessRuleCreationFields := *openapiclient.NewAccessRuleCreationFields("user@run.ai", openapiclient.SubjectType("user"), int32(53142648), "a418ed33-9399-48c0-a890-122cadd13bfd", openapiclient.ScopeType("system")) // AccessRuleCreationFields | The access rule to create.

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

Delete an access rule.



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

> AccessRuleById GetAccessRule(ctx, accessRuleId).Execute()

Get an access rule.



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
    // response from `GetAccessRule`: AccessRuleById
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

[**AccessRuleById**](AccessRuleById.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetAccessRules

> GetAccessRules200Response GetAccessRules(ctx).SubjectType(subjectType).SubjectIdFilter(subjectIdFilter).SubjectIds(subjectIds).Limit(limit).Offset(offset).LastUpdated(lastUpdated).IncludeDeleted(includeDeleted).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).RoleId(roleId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()

List the access rules.



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
    limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
    offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
    lastUpdated := "2021-12-14T16:04:15.099Z" // string | Filter by last update time. (optional)
    includeDeleted := false // bool | True to include deleted objects in the result. (optional) (default to false)
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
    scopeType := "project" // string | The type of resource we want to filter by. (optional)
    scopeId := "2" // string | The scope resource id that we want to filter by. (optional)
    roleId := int32(56) // int32 | The role id we want to filter by. (optional)
    sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
    sortBy := openapiclient.AccessRulesSortFilterFields("subjectId") // AccessRulesSortFilterFields | Sort results by a parameter. (optional)
    filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.GetAccessRules(context.Background()).SubjectType(subjectType).SubjectIdFilter(subjectIdFilter).SubjectIds(subjectIds).Limit(limit).Offset(offset).LastUpdated(lastUpdated).IncludeDeleted(includeDeleted).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).RoleId(roleId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Execute()
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
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **lastUpdated** | **string** | Filter by last update time. | 
 **includeDeleted** | **bool** | True to include deleted objects in the result. | [default to false]
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | The scope resource id that we want to filter by. | 
 **roleId** | **int32** | The role id we want to filter by. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | [**AccessRulesSortFilterFields**](AccessRulesSortFilterFields.md) | Sort results by a parameter. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 

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

> SyncAccessRules(ctx, clusterUuid).AccessRuleForSync(accessRuleForSync).Execute()

Sync access rules.



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
    accessRuleForSync := []openapiclient.AccessRuleForSync{*openapiclient.NewAccessRuleForSync("Ready", "Failed to enforce department 'department-a'", int32(32))} // []AccessRuleForSync | The statuses to sync. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.SyncAccessRules(context.Background(), clusterUuid).AccessRuleForSync(accessRuleForSync).Execute()
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
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

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

> UpdateAccessRuleStatus(ctx, accessRuleId, clusterUuid).AccessRuleStatusRequest(accessRuleStatusRequest).Execute()

Update an access rule.



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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    accessRuleStatusRequest := *openapiclient.NewAccessRuleStatusRequest("Ready", "Failed to enforce department 'department-a'") // AccessRuleStatusRequest | The status of the access rule in the cluster.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.AccessRulesApi.UpdateAccessRuleStatus(context.Background(), accessRuleId, clusterUuid).AccessRuleStatusRequest(accessRuleStatusRequest).Execute()
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
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

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

