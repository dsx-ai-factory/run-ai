# \InternalAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateInternalAccessRule**](InternalAPI.md#CreateInternalAccessRule) | **Post** /internal/authorization/access-rules | Create internal access rule with out privileged escalation validation.
[**DeleteScopeAccessRules**](InternalAPI.md#DeleteScopeAccessRules) | **Delete** /internal/authorization/scope-access-rules | Delete access rules that assigned to the requested scope resource
[**DeleteSubjectAccessRules**](InternalAPI.md#DeleteSubjectAccessRules) | **Delete** /internal/authorization/subject-access-rules | Delete access rules that assigned to the requested subject.
[**GetGroups**](InternalAPI.md#GetGroups) | **Get** /internal/authorization/groups | Get all group names of a tenant
[**GetScopeAccessRules**](InternalAPI.md#GetScopeAccessRules) | **Get** /internal/authorization/scope-access-rules | Get access rules that assigned to the requested scope resource
[**GetSubjectAccessRules**](InternalAPI.md#GetSubjectAccessRules) | **Get** /internal/authorization/subject-access-rules | List specific subject access rules.
[**PrivilegeEscalationEvaluate**](InternalAPI.md#PrivilegeEscalationEvaluate) | **Post** /internal/authorization/privilege-escalation/evaluate | Evaluate privilege escalation.
[**UpdateScopeAccessRules**](InternalAPI.md#UpdateScopeAccessRules) | **Put** /internal/authorization/scope-access-rules | Update access rules that assigned to the requested scope resource



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
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {
	internalCreateAccessRuleRequest := *openapiclient.NewInternalCreateAccessRuleRequest("user@run.ai", openapiclient.SubjectType("user"), openapiclient.AllowedInternalCreateRoles("System administrator")) // InternalCreateAccessRuleRequest | The request parameters.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.CreateInternalAccessRule(context.Background()).InternalCreateAccessRuleRequest(internalCreateAccessRuleRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.CreateInternalAccessRule``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateInternalAccessRule`: AccessRule
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.CreateInternalAccessRule`: %v\n", resp)
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
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {
	scopeType := "scopeType_example" // string | The resource scope type to filter by.
	scopeId := "2" // string | The scope resource id that we want to filter by.
	lastUpdated := "2021-12-14T16:04:15.099Z" // string | Filter by last update time. (optional)
	roles := "ML engineer,L1 researcher" // string | The roles we want to filter by. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.InternalAPI.DeleteScopeAccessRules(context.Background()).ScopeType(scopeType).ScopeId(scopeId).LastUpdated(lastUpdated).Roles(roles).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.DeleteScopeAccessRules``: %v\n", err)
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

> DeleteSubjectAccessRules(ctx).SubjectType(subjectType).SubjectId(subjectId).DryRun(dryRun).Execute()

Delete access rules that assigned to the requested subject.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {
	subjectType := "user" // string | The type of resource we want to filter by.
	subjectId := "some.user@run.ai" // string | The subject id that we want to filter by.
	dryRun := false // bool | True to perform validations only, without side effects. (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.InternalAPI.DeleteSubjectAccessRules(context.Background()).SubjectType(subjectType).SubjectId(subjectId).DryRun(dryRun).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.DeleteSubjectAccessRules``: %v\n", err)
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
 **dryRun** | **bool** | True to perform validations only, without side effects. | [default to false]

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


## GetGroups

> GetGroups200Response GetGroups(ctx).Execute()

Get all group names of a tenant

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetGroups(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetGroups``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGroups`: GetGroups200Response
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetGroups`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetGroupsRequest struct via the builder pattern


### Return type

[**GetGroups200Response**](GetGroups200Response.md)

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
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	roles := "ML engineer,L1 researcher" // string | The roles we want to filter by. (optional)
	departmentIds := "1,2" // string | The department scope resource ids that we want to filter by. (optional)
	projectIds := "1,2" // string | The project scope resource ids that we want to filter by. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetScopeAccessRules(context.Background()).ClusterId(clusterId).Roles(roles).DepartmentIds(departmentIds).ProjectIds(projectIds).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetScopeAccessRules``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetScopeAccessRules`: GetAccessRules200Response
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetScopeAccessRules`: %v\n", resp)
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
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {
	subjectType := "user" // string | The type of resource we want to filter by.
	subjectIds := []string{"some.user@run.ai"} // []string | The ids of the subjects to filter the response for. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.InternalAPI.GetSubjectAccessRules(context.Background()).SubjectType(subjectType).SubjectIds(subjectIds).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.GetSubjectAccessRules``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSubjectAccessRules`: []AccessRule
	fmt.Fprintf(os.Stdout, "Response from `InternalAPI.GetSubjectAccessRules`: %v\n", resp)
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


## PrivilegeEscalationEvaluate

> PrivilegeEscalationEvaluate(ctx).SubjectId(subjectId).GroupIds(groupIds).Execute()

Evaluate privilege escalation.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {
	subjectId := "some.user@run.ai" // string | The subject id that we want to filter by.
	groupIds := []string{"Inner_example"} // []string | The ids of the subject's groups. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.InternalAPI.PrivilegeEscalationEvaluate(context.Background()).SubjectId(subjectId).GroupIds(groupIds).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.PrivilegeEscalationEvaluate``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiPrivilegeEscalationEvaluateRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **subjectId** | **string** | The subject id that we want to filter by. | 
 **groupIds** | **[]string** | The ids of the subject&#39;s groups. | 

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
	openapiclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

func main() {
	scopeType := "scopeType_example" // string | The resource scope type to filter by.
	scopeId := "2" // string | The scope resource id that we want to filter by.
	updateScopeAccessRulesRequest := *openapiclient.NewUpdateScopeAccessRulesRequest("team-a") // UpdateScopeAccessRulesRequest | The access rule fields to update.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.InternalAPI.UpdateScopeAccessRules(context.Background()).ScopeType(scopeType).ScopeId(scopeId).UpdateScopeAccessRulesRequest(updateScopeAccessRulesRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `InternalAPI.UpdateScopeAccessRules``: %v\n", err)
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

