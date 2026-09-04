# \RolesAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRole**](RolesAPI.md#CreateRole) | **Post** /api/v2/authorization/roles | Create a role.
[**DeleteRole**](RolesAPI.md#DeleteRole) | **Delete** /api/v2/authorization/roles/{roleIdPath} | Delete a role by id.
[**GetRole**](RolesAPI.md#GetRole) | **Get** /api/v2/authorization/roles/{roleIdPath} | Get a role by id.
[**GetRoleV1**](RolesAPI.md#GetRoleV1) | **Get** /api/v1/authorization/roles/{roleIdPath} | Get a role by id.
[**GetRoles**](RolesAPI.md#GetRoles) | **Get** /api/v2/authorization/roles | Get a list of roles.
[**GetRolesV1**](RolesAPI.md#GetRolesV1) | **Get** /api/v1/authorization/roles | Get a list of roles.
[**RoleActive**](RolesAPI.md#RoleActive) | **Post** /api/v2/authorization/roles/{roleIdPath}/active | Enable or Disable a role by id.
[**RoleDisable**](RolesAPI.md#RoleDisable) | **Post** /api/v2/authorization/roles/{roleIdPath}/disable | Disable a role by id.
[**RoleEnable**](RolesAPI.md#RoleEnable) | **Post** /api/v2/authorization/roles/{roleIdPath}/enable | Enable a role by id.
[**UpdateRole**](RolesAPI.md#UpdateRole) | **Put** /api/v2/authorization/roles/{roleIdPath} | Update a role by id.



## CreateRole

> Role CreateRole(ctx).RoleCreationFieldsInternal(roleCreationFieldsInternal).Execute()

Create a role.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleCreationFieldsInternal := *openapiclient.NewRoleCreationFieldsInternal("admin", "can manage all resources") // RoleCreationFieldsInternal | The role to create.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RolesAPI.CreateRole(context.Background()).RoleCreationFieldsInternal(roleCreationFieldsInternal).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.CreateRole``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateRole`: Role
	fmt.Fprintf(os.Stdout, "Response from `RolesAPI.CreateRole`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateRoleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **roleCreationFieldsInternal** | [**RoleCreationFieldsInternal**](RoleCreationFieldsInternal.md) | The role to create. | 

### Return type

[**Role**](Role.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteRole

> DeleteRole(ctx, roleIdPath).Execute()

Delete a role by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleIdPath := int32(32) // int32 | The ID of the role

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RolesAPI.DeleteRole(context.Background(), roleIdPath).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.DeleteRole``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleIdPath** | **int32** | The ID of the role | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteRoleRequest struct via the builder pattern


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


## GetRole

> Role GetRole(ctx, roleIdPath).Execute()

Get a role by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleIdPath := int32(32) // int32 | The ID of the role

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RolesAPI.GetRole(context.Background(), roleIdPath).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.GetRole``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRole`: Role
	fmt.Fprintf(os.Stdout, "Response from `RolesAPI.GetRole`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleIdPath** | **int32** | The ID of the role | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRoleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Role**](Role.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRoleV1

> RoleV1 GetRoleV1(ctx, roleIdPath).Execute()

Get a role by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleIdPath := int32(32) // int32 | The ID of the role

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RolesAPI.GetRoleV1(context.Background(), roleIdPath).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.GetRoleV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRoleV1`: RoleV1
	fmt.Fprintf(os.Stdout, "Response from `RolesAPI.GetRoleV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleIdPath** | **int32** | The ID of the role | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetRoleV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RoleV1**](RoleV1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRoles

> GetRoles200Response GetRoles(ctx).Limit(limit).Offset(offset).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).IncludeInternal(includeInternal).Execute()

Get a list of roles.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	sortBy := openapiclient.RolesSortFilterFields("name") // RolesSortFilterFields | Sort results by a parameter. (optional)
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are == Equals, != Not equals, <= Less than or equal, >= Greater than or equal, =@ contains, !@ Does not contains, =^ Starts with and =$ Ends with. Dates are in ISO 8601 timestamp format and available for operators ==, !=, <= and >=. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)
	includeInternal := false // bool |  (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RolesAPI.GetRoles(context.Background()).Limit(limit).Offset(offset).SortOrder(sortOrder).SortBy(sortBy).FilterBy(filterBy).Search(search).IncludeInternal(includeInternal).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.GetRoles``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRoles`: GetRoles200Response
	fmt.Fprintf(os.Stdout, "Response from `RolesAPI.GetRoles`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetRolesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **sortBy** | [**RolesSortFilterFields**](RolesSortFilterFields.md) | Sort results by a parameter. | 
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &#x3D;&#x3D; Equals, !&#x3D; Not equals, &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains, !@ Does not contains, &#x3D;^ Starts with and &#x3D;$ Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x3D;&#x3D;, !&#x3D;, &lt;&#x3D; and &gt;&#x3D;. | 
 **search** | **string** | Filter results by a free text search. | 
 **includeInternal** | **bool** |  | [default to false]

### Return type

[**GetRoles200Response**](GetRoles200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRolesV1

> []RoleV1 GetRolesV1(ctx).IncludeInternal(includeInternal).Execute()

Get a list of roles.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	includeInternal := false // bool |  (optional) (default to false)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RolesAPI.GetRolesV1(context.Background()).IncludeInternal(includeInternal).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.GetRolesV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRolesV1`: []RoleV1
	fmt.Fprintf(os.Stdout, "Response from `RolesAPI.GetRolesV1`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetRolesV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **includeInternal** | **bool** |  | [default to false]

### Return type

[**[]RoleV1**](RoleV1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## RoleActive

> RoleActive(ctx, roleIdPath).RoleActiveFields(roleActiveFields).Execute()

Enable or Disable a role by id.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleIdPath := int32(32) // int32 | The ID of the role
	roleActiveFields := *openapiclient.NewRoleActiveFields(false) // RoleActiveFields | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RolesAPI.RoleActive(context.Background(), roleIdPath).RoleActiveFields(roleActiveFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.RoleActive``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleIdPath** | **int32** | The ID of the role | 

### Other Parameters

Other parameters are passed through a pointer to a apiRoleActiveRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **roleActiveFields** | [**RoleActiveFields**](RoleActiveFields.md) |  | 

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


## RoleDisable

> RoleDisable(ctx, roleIdPath).RoleEnabledDisabledFields(roleEnabledDisabledFields).Execute()

Disable a role by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleIdPath := int32(32) // int32 | The ID of the role
	roleEnabledDisabledFields := *openapiclient.NewRoleEnabledDisabledFields() // RoleEnabledDisabledFields | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RolesAPI.RoleDisable(context.Background(), roleIdPath).RoleEnabledDisabledFields(roleEnabledDisabledFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.RoleDisable``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleIdPath** | **int32** | The ID of the role | 

### Other Parameters

Other parameters are passed through a pointer to a apiRoleDisableRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **roleEnabledDisabledFields** | [**RoleEnabledDisabledFields**](RoleEnabledDisabledFields.md) |  | 

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


## RoleEnable

> RoleEnable(ctx, roleIdPath).RoleEnabledDisabledFields(roleEnabledDisabledFields).Execute()

Enable a role by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleIdPath := int32(32) // int32 | The ID of the role
	roleEnabledDisabledFields := *openapiclient.NewRoleEnabledDisabledFields() // RoleEnabledDisabledFields | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.RolesAPI.RoleEnable(context.Background(), roleIdPath).RoleEnabledDisabledFields(roleEnabledDisabledFields).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.RoleEnable``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleIdPath** | **int32** | The ID of the role | 

### Other Parameters

Other parameters are passed through a pointer to a apiRoleEnableRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **roleEnabledDisabledFields** | [**RoleEnabledDisabledFields**](RoleEnabledDisabledFields.md) |  | 

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


## UpdateRole

> Role UpdateRole(ctx, roleIdPath).RoleCreationFieldsInternal(roleCreationFieldsInternal).Execute()

Update a role by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
)

func main() {
	roleIdPath := int32(32) // int32 | The ID of the role
	roleCreationFieldsInternal := *openapiclient.NewRoleCreationFieldsInternal("admin", "can manage all resources") // RoleCreationFieldsInternal | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RolesAPI.UpdateRole(context.Background(), roleIdPath).RoleCreationFieldsInternal(roleCreationFieldsInternal).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RolesAPI.UpdateRole``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateRole`: Role
	fmt.Fprintf(os.Stdout, "Response from `RolesAPI.UpdateRole`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**roleIdPath** | **int32** | The ID of the role | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateRoleRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **roleCreationFieldsInternal** | [**RoleCreationFieldsInternal**](RoleCreationFieldsInternal.md) |  | 

### Return type

[**Role**](Role.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

