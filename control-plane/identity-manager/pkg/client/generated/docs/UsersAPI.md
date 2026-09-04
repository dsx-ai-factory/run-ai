# \UsersAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BatchCreateUsers**](UsersAPI.md#BatchCreateUsers) | **Post** /api/v1/users/batch-create | Create multiple local users in batch.
[**ChangeUserPassword**](UsersAPI.md#ChangeUserPassword) | **Post** /api/v1/me/password | change user password
[**CountUsers**](UsersAPI.md#CountUsers) | **Get** /api/v1/users/count | Count users
[**CreateIntenralUser**](UsersAPI.md#CreateIntenralUser) | **Post** /internal/users | create internal user
[**CreateUser**](UsersAPI.md#CreateUser) | **Post** /api/v1/users | Create a local user.
[**DeleteInternalUserByEmail**](UsersAPI.md#DeleteInternalUserByEmail) | **Delete** /internal/realms/{realm}/users/{email} | Delete a user by email (internal).
[**DeleteUserById**](UsersAPI.md#DeleteUserById) | **Delete** /api/v1/users/{userId} | Delete a user by id.
[**GetInternalUserById**](UsersAPI.md#GetInternalUserById) | **Get** /internal/users/{userId} | Get a user by id (internal).
[**GetUserAccessGroups**](UsersAPI.md#GetUserAccessGroups) | **Get** /internal/users/{email}/accessGroups | Get a users access groups.
[**GetUserById**](UsersAPI.md#GetUserById) | **Get** /api/v1/users/{userId} | Get a user by id.
[**GetUsers**](UsersAPI.md#GetUsers) | **Get** /api/v1/users | Get users.
[**GetUsersMinimal**](UsersAPI.md#GetUsersMinimal) | **Get** /api/v1/users-minimal | Get minimal user display information.
[**LogoutUser**](UsersAPI.md#LogoutUser) | **Post** /api/v1/users/{userId}/logout | Logout a user.
[**PatchUserById**](UsersAPI.md#PatchUserById) | **Patch** /api/v1/users/{userId} | Update a user by id.
[**ResetUserPassword**](UsersAPI.md#ResetUserPassword) | **Post** /api/v1/users/{userId}/password | Reset a user&#39;s password.



## BatchCreateUsers

> BulkUserCreationResponse BatchCreateUsers(ctx).BulkUserCreationRequest(bulkUserCreationRequest).Execute()

Create multiple local users in batch.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	bulkUserCreationRequest := *openapiclient.NewBulkUserCreationRequest([]openapiclient.UserCreationRequestBulk{*openapiclient.NewUserCreationRequestBulk("Email_example")}) // BulkUserCreationRequest | Bulk user creation request containing array of max 10 users

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.BatchCreateUsers(context.Background()).BulkUserCreationRequest(bulkUserCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.BatchCreateUsers``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BatchCreateUsers`: BulkUserCreationResponse
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.BatchCreateUsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiBatchCreateUsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **bulkUserCreationRequest** | [**BulkUserCreationRequest**](BulkUserCreationRequest.md) | Bulk user creation request containing array of max 10 users | 

### Return type

[**BulkUserCreationResponse**](BulkUserCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ChangeUserPassword

> ChangeUserPassword200Response ChangeUserPassword(ctx).UserChangePasswordRequest(userChangePasswordRequest).Execute()

change user password

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userChangePasswordRequest := *openapiclient.NewUserChangePasswordRequest("CurrentPassword_example", "NewPassword_example") // UserChangePasswordRequest | Password to change

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.ChangeUserPassword(context.Background()).UserChangePasswordRequest(userChangePasswordRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.ChangeUserPassword``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ChangeUserPassword`: ChangeUserPassword200Response
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.ChangeUserPassword`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiChangeUserPasswordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userChangePasswordRequest** | [**UserChangePasswordRequest**](UserChangePasswordRequest.md) | Password to change | 

### Return type

[**ChangeUserPassword200Response**](ChangeUserPassword200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CountUsers

> CountUsers200Response CountUsers(ctx).FilterBy(filterBy).Search(search).Execute()

Count users



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are <= Less than or equal, >= Greater than or equal, =@ contains. Dates are in ISO 8601 timestamp format and available for operators <=, >= and ==. (optional)
	search := "test@run" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.CountUsers(context.Background()).FilterBy(filterBy).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.CountUsers``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountUsers`: CountUsers200Response
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.CountUsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountUsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains. Dates are in ISO 8601 timestamp format and available for operators &lt;&#x3D;, &gt;&#x3D; and &#x3D;&#x3D;. | 
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**CountUsers200Response**](CountUsers200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateIntenralUser

> UserPostResponse CreateIntenralUser(ctx).InternalUserCreationRequest(internalUserCreationRequest).Execute()

create internal user

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	internalUserCreationRequest := *openapiclient.NewInternalUserCreationRequest("Email_example", "Realm_example", false) // InternalUserCreationRequest | Internal user object to create

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.CreateIntenralUser(context.Background()).InternalUserCreationRequest(internalUserCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.CreateIntenralUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateIntenralUser`: UserPostResponse
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.CreateIntenralUser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateIntenralUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **internalUserCreationRequest** | [**InternalUserCreationRequest**](InternalUserCreationRequest.md) | Internal user object to create | 

### Return type

[**UserPostResponse**](UserPostResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateUser

> UserPostResponse CreateUser(ctx).UserCreationRequest(userCreationRequest).Execute()

Create a local user.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userCreationRequest := *openapiclient.NewUserCreationRequest("Email_example") // UserCreationRequest | User object to create

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.CreateUser(context.Background()).UserCreationRequest(userCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.CreateUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateUser`: UserPostResponse
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.CreateUser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userCreationRequest** | [**UserCreationRequest**](UserCreationRequest.md) | User object to create | 

### Return type

[**UserPostResponse**](UserPostResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteInternalUserByEmail

> DeleteInternalUserByEmail(ctx, realm, email).Execute()

Delete a user by email (internal).



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	realm := "realm_example" // string | The name of the realm
	email := "email_example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UsersAPI.DeleteInternalUserByEmail(context.Background(), realm, email).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.DeleteInternalUserByEmail``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**realm** | **string** | The name of the realm | 
**email** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteInternalUserByEmailRequest struct via the builder pattern


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


## DeleteUserById

> DeleteUserById(ctx, userId).Execute()

Delete a user by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userId := "userId_example" // string | The id of the user

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UsersAPI.DeleteUserById(context.Background(), userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.DeleteUserById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | The id of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteUserByIdRequest struct via the builder pattern


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


## GetInternalUserById

> User GetInternalUserById(ctx, userId).Execute()

Get a user by id (internal).



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userId := "userId_example" // string | The id of the user

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.GetInternalUserById(context.Background(), userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.GetInternalUserById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInternalUserById`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.GetInternalUserById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | The id of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetInternalUserByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User**](User.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUserAccessGroups

> []string GetUserAccessGroups(ctx, email).Execute()

Get a users access groups.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	email := "email_example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.GetUserAccessGroups(context.Background(), email).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.GetUserAccessGroups``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUserAccessGroups`: []string
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.GetUserAccessGroups`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**email** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserAccessGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

**[]string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUserById

> User GetUserById(ctx, userId).Execute()

Get a user by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userId := "userId_example" // string | The id of the user

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.GetUserById(context.Background(), userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.GetUserById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUserById`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.GetUserById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | The id of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User**](User.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUsers

> []User GetUsers(ctx).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()

Get users.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Operators are <= Less than or equal, >= Greater than or equal, =@ contains. Dates are in ISO 8601 timestamp format and available for operators <=, >= and ==. (optional)
	sortBy := openapiclient.UsersFilterSortFields("username") // UsersFilterSortFields | Sort results by a parameters. (optional)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 500)
	search := "test@run" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.GetUsers(context.Background()).FilterBy(filterBy).SortBy(sortBy).SortOrder(sortOrder).Offset(offset).Limit(limit).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.GetUsers``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUsers`: []User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.GetUsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetUsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Operators are &lt;&#x3D; Less than or equal, &gt;&#x3D; Greater than or equal, &#x3D;@ contains. Dates are in ISO 8601 timestamp format and available for operators &lt;&#x3D;, &gt;&#x3D; and &#x3D;&#x3D;. | 
 **sortBy** | [**UsersFilterSortFields**](UsersFilterSortFields.md) | Sort results by a parameters. | 
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 500]
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**[]User**](User.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUsersMinimal

> []UserMinimal GetUsersMinimal(ctx).FilterBy(filterBy).Search(search).Id(id).Limit(limit).Execute()

Get minimal user display information.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by username. Use the format field-name operator value. Supported operators are == (exact match) and =@ (contains). Multiple entries (up to 25) are allowed only with the == operator.  (optional)
	search := "john" // string | Filter results by a free text search across username, firstName and lastName (OR logic). Minimum 2 characters. (optional)
	id := []string{"Inner_example"} // []string | Bulk lookup by user IDs (comma-separated). Returns the matching users in addition to any results from `search` or `filterBy`. Up to 25 IDs per request. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 500)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.GetUsersMinimal(context.Background()).FilterBy(filterBy).Search(search).Id(id).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.GetUsersMinimal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetUsersMinimal`: []UserMinimal
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.GetUsersMinimal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetUsersMinimalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by username. Use the format field-name operator value. Supported operators are &#x3D;&#x3D; (exact match) and &#x3D;@ (contains). Multiple entries (up to 25) are allowed only with the &#x3D;&#x3D; operator.  | 
 **search** | **string** | Filter results by a free text search across username, firstName and lastName (OR logic). Minimum 2 characters. | 
 **id** | **[]string** | Bulk lookup by user IDs (comma-separated). Returns the matching users in addition to any results from &#x60;search&#x60; or &#x60;filterBy&#x60;. Up to 25 IDs per request. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 500]

### Return type

[**[]UserMinimal**](UserMinimal.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## LogoutUser

> LogoutUser(ctx, userId).Execute()

Logout a user.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userId := "userId_example" // string | The id of the user

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.UsersAPI.LogoutUser(context.Background(), userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.LogoutUser``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | The id of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiLogoutUserRequest struct via the builder pattern


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


## PatchUserById

> User PatchUserById(ctx, userId).UserPatchRequest(userPatchRequest).Execute()

Update a user by id.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userId := "userId_example" // string | The id of the user
	userPatchRequest := *openapiclient.NewUserPatchRequest() // UserPatchRequest | User patch object with fields to update. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.PatchUserById(context.Background(), userId).UserPatchRequest(userPatchRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.PatchUserById``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchUserById`: User
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.PatchUserById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | The id of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchUserByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **userPatchRequest** | [**UserPatchRequest**](UserPatchRequest.md) | User patch object with fields to update. | 

### Return type

[**User**](User.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ResetUserPassword

> ResetUserPassword200Response ResetUserPassword(ctx, userId).Execute()

Reset a user's password.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

func main() {
	userId := "userId_example" // string | The id of the user

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UsersAPI.ResetUserPassword(context.Background(), userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UsersAPI.ResetUserPassword``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ResetUserPassword`: ResetUserPassword200Response
	fmt.Fprintf(os.Stdout, "Response from `UsersAPI.ResetUserPassword`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | The id of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiResetUserPasswordRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ResetUserPassword200Response**](ResetUserPassword200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

