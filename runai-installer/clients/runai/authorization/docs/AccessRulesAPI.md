# \AccessRulesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AccessRulesBatch**](AccessRulesAPI.md#AccessRulesBatch) | **Post** /api/v1/authorization/access-rules/batch | Delete Multiple Access Rules
[**AccessRulesBatchCreate**](AccessRulesAPI.md#AccessRulesBatchCreate) | **Post** /api/v1/authorization/access-rules/batch-create | Create Multiple Access Rules
[**CountAccessRules**](AccessRulesAPI.md#CountAccessRules) | **Get** /api/v1/authorization/access-rules/count | Count access rules.
[**CreateAccessRule**](AccessRulesAPI.md#CreateAccessRule) | **Post** /api/v1/authorization/access-rules | Create an access rule.
[**DeleteAccessRule**](AccessRulesAPI.md#DeleteAccessRule) | **Delete** /api/v1/authorization/access-rules/{accessRuleId} | Delete an access rule.
[**GetAccessRule**](AccessRulesAPI.md#GetAccessRule) | **Get** /api/v1/authorization/access-rules/{accessRuleId} | Get an access rule.
[**GetAccessRules**](AccessRulesAPI.md#GetAccessRules) | **Get** /api/v1/authorization/access-rules | List the access rules.
[**GetAccessRulesSync**](AccessRulesAPI.md#GetAccessRulesSync) | **Get** /api/v1/authorization/access-rules/sync | List the access rules.
[**SyncAccessRules**](AccessRulesAPI.md#SyncAccessRules) | **Post** /api/v1/authorization/access-rules/{clusterUuid}/sync | Sync access rules.
[**UpdateAccessRuleStatus**](AccessRulesAPI.md#UpdateAccessRuleStatus) | **Put** /api/v1/authorization/access-rules/{accessRuleId}/{clusterUuid}/status | Update an access rule.



## AccessRulesBatch

> AccessRuleDeletionBatchResponse AccessRulesBatch(ctx).AccessRulesBatchFields(accessRulesBatchFields).Execute()

Delete Multiple Access Rules

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
	accessRulesBatchFields := *openapiclient.NewAccessRulesBatchFields([]string{"Ids_example"}, openapiclient.AccessRulesBatchAction("delete")) // AccessRulesBatchFields |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessRulesAPI.AccessRulesBatch(context.Background()).AccessRulesBatchFields(accessRulesBatchFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.AccessRulesBatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccessRulesBatch`: AccessRuleDeletionBatchResponse
	fmt.Fprintf(os.Stdout, "Response from `AccessRulesAPI.AccessRulesBatch`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAccessRulesBatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **accessRulesBatchFields** | [**AccessRulesBatchFields**](AccessRulesBatchFields.md) |  | 

### Return type

[**AccessRuleDeletionBatchResponse**](AccessRuleDeletionBatchResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AccessRulesBatchCreate

> AccessRulesBatchCreateResponse AccessRulesBatchCreate(ctx).AccessRulesBatchCreateFields(accessRulesBatchCreateFields).Execute()

Create Multiple Access Rules



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
	accessRulesBatchCreateFields := *openapiclient.NewAccessRulesBatchCreateFields(*openapiclient.NewAccessRulesBatchCreateFieldsPayload(int32(53142648), []openapiclient.AccessRuleSubjectsInner{*openapiclient.NewAccessRuleSubjectsInner("user@run.ai", openapiclient.SubjectType("user"))}, []openapiclient.AccessRuleScopesInner{*openapiclient.NewAccessRuleScopesInner("ScopeId_example", openapiclient.ScopeType("system"))})) // AccessRulesBatchCreateFields |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessRulesAPI.AccessRulesBatchCreate(context.Background()).AccessRulesBatchCreateFields(accessRulesBatchCreateFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.AccessRulesBatchCreate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AccessRulesBatchCreate`: AccessRulesBatchCreateResponse
	fmt.Fprintf(os.Stdout, "Response from `AccessRulesAPI.AccessRulesBatchCreate`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAccessRulesBatchCreateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **accessRulesBatchCreateFields** | [**AccessRulesBatchCreateFields**](AccessRulesBatchCreateFields.md) |  | 

### Return type

[**AccessRulesBatchCreateResponse**](AccessRulesBatchCreateResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CountAccessRules

> CountAccessRules200Response CountAccessRules(ctx).IncludeDeleted(includeDeleted).FilterBy(filterBy).Search(search).Execute()

Count access rules.



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
	includeDeleted := false // bool | True to include deleted objects in the result. (optional) (default to false)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessRulesAPI.CountAccessRules(context.Background()).IncludeDeleted(includeDeleted).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.CountAccessRules``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountAccessRules`: CountAccessRules200Response
	fmt.Fprintf(os.Stdout, "Response from `AccessRulesAPI.CountAccessRules`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **includeDeleted** | **bool** | True to include deleted objects in the result. | [default to false]
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **search** | **string** | Filter results by a free text search. | 

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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	accessRuleCreationFields := *openapiclient.NewAccessRuleCreationFields("user@run.ai", openapiclient.SubjectType("user"), int32(53142648), "a418ed33-9399-48c0-a890-122cadd13bfd", openapiclient.ScopeType("system")) // AccessRuleCreationFields | The access rule to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessRulesAPI.CreateAccessRule(context.Background()).AccessRuleCreationFields(accessRuleCreationFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.CreateAccessRule``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateAccessRule`: AccessRule
	fmt.Fprintf(os.Stdout, "Response from `AccessRulesAPI.CreateAccessRule`: %v\n", resp)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	accessRuleId := int32(32) // int32 | The id of the access rule to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AccessRulesAPI.DeleteAccessRule(context.Background(), accessRuleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.DeleteAccessRule``: %v\n", err)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	accessRuleId := int32(32) // int32 | The id of the access rule to retrieve

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessRulesAPI.GetAccessRule(context.Background(), accessRuleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.GetAccessRule``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAccessRule`: AccessRuleById
	fmt.Fprintf(os.Stdout, "Response from `AccessRulesAPI.GetAccessRule`: %v\n", resp)
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

> GetAccessRules200Response GetAccessRules(ctx).SubjectType(subjectType).SubjectIdFilter(subjectIdFilter).SubjectIds(subjectIds).Limit(limit).Offset(offset).LastUpdated(lastUpdated).IncludeDeleted(includeDeleted).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).RoleId(roleId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).Execute()

List the access rules.



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
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessRulesAPI.GetAccessRules(context.Background()).SubjectType(subjectType).SubjectIdFilter(subjectIdFilter).SubjectIds(subjectIds).Limit(limit).Offset(offset).LastUpdated(lastUpdated).IncludeDeleted(includeDeleted).ClusterId(clusterId).ScopeType(scopeType).ScopeId(scopeId).RoleId(roleId).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.GetAccessRules``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAccessRules`: GetAccessRules200Response
	fmt.Fprintf(os.Stdout, "Response from `AccessRulesAPI.GetAccessRules`: %v\n", resp)
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
 **search** | **string** | Filter results by a free text search. | 

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


## GetAccessRulesSync

> GetAccessRulesSync200Response GetAccessRulesSync(ctx).Limit(limit).Offset(offset).LastUpdated(lastUpdated).ClusterId(clusterId).Execute()

List the access rules.



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
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	lastUpdated := "2021-12-14T16:04:15.099Z" // string | Filter by last update time. (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AccessRulesAPI.GetAccessRulesSync(context.Background()).Limit(limit).Offset(offset).LastUpdated(lastUpdated).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.GetAccessRulesSync``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetAccessRulesSync`: GetAccessRulesSync200Response
	fmt.Fprintf(os.Stdout, "Response from `AccessRulesAPI.GetAccessRulesSync`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetAccessRulesSyncRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **lastUpdated** | **string** | Filter by last update time. | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

### Return type

[**GetAccessRulesSync200Response**](GetAccessRulesSync200Response.md)

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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	accessRuleForSync := []openapiclient.AccessRuleForSync{*openapiclient.NewAccessRuleForSync(openapiclient.AccessRuleClusterPhase("Ready"), "Failed to enforce department 'department-a'", int32(32))} // []AccessRuleForSync | The statuses to sync. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AccessRulesAPI.SyncAccessRules(context.Background(), clusterUuid).AccessRuleForSync(accessRuleForSync).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.SyncAccessRules``: %v\n", err)
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
	openapiclient "github.com/GIT_USER_ID/GIT_REPO_ID"
)

func main() {
	accessRuleId := int32(32) // int32 | The id of the access rule to retrieve
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	accessRuleStatusRequest := *openapiclient.NewAccessRuleStatusRequest(openapiclient.AccessRuleClusterPhase("Ready"), "Failed to enforce department 'department-a'") // AccessRuleStatusRequest | The status of the access rule in the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.AccessRulesAPI.UpdateAccessRuleStatus(context.Background(), accessRuleId, clusterUuid).AccessRuleStatusRequest(accessRuleStatusRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AccessRulesAPI.UpdateAccessRuleStatus``: %v\n", err)
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

