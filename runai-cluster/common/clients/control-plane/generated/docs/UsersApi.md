# \UsersApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ChangeUserPassword**](UsersApi.md#ChangeUserPassword) | **Post** /api/v1/me/password | change user password
[**CreateGroup**](UsersApi.md#CreateGroup) | **Post** /v1/k8s/groups | Create a new group.
[**CreateIntenralUser**](UsersApi.md#CreateIntenralUser) | **Post** /internal/users | create intenral user
[**CreateUser**](UsersApi.md#CreateUser) | **Post** /v1/k8s/users | Create a new user.
[**CreateUser_0**](UsersApi.md#CreateUser_0) | **Post** /api/v1/users | Create a local user.
[**DeleteGroupByName**](UsersApi.md#DeleteGroupByName) | **Delete** /v1/k8s/groups/{groupName} | Delete a group.
[**DeleteUserById**](UsersApi.md#DeleteUserById) | **Delete** /v1/k8s/users/{userId} | Delete a user.
[**DeleteUserById_0**](UsersApi.md#DeleteUserById_0) | **Delete** /api/v1/users/{userId} | Delete a user by id.
[**GeGroupByName**](UsersApi.md#GeGroupByName) | **Get** /v1/k8s/groups/{groupName} | Get group details.
[**GetGroups**](UsersApi.md#GetGroups) | **Get** /v1/k8s/groups | Get groups list.
[**GetRoles**](UsersApi.md#GetRoles) | **Get** /v1/k8s/users/roles | Get all possible permissions.
[**GetUserById**](UsersApi.md#GetUserById) | **Get** /v1/k8s/users/{userId} | Get user details.
[**GetUserById_0**](UsersApi.md#GetUserById_0) | **Get** /api/v1/users/{userId} | Get a user by id.
[**GetUserRoles**](UsersApi.md#GetUserRoles) | **Get** /v1/k8s/users/{userId}/roles | Get user permissions.
[**GetUsers**](UsersApi.md#GetUsers) | **Get** /v1/k8s/users | Get users list.
[**GetUsers_0**](UsersApi.md#GetUsers_0) | **Get** /api/v1/users | Get users.
[**LogoutUser**](UsersApi.md#LogoutUser) | **Post** /api/v1/users/{userId}/logout | Logout a user.
[**ResetUserPassword**](UsersApi.md#ResetUserPassword) | **Post** /api/v1/users/{userId}/password | Reset a user&#39;s password.
[**UpdateGroupByName**](UsersApi.md#UpdateGroupByName) | **Put** /v1/k8s/groups/{groupName} | Update group details.
[**UpdateUserById**](UsersApi.md#UpdateUserById) | **Put** /v1/k8s/users/{userId} | Update user details.



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
    openapiclient "./openapi"
)

func main() {
    userChangePasswordRequest := *openapiclient.NewUserChangePasswordRequest("CurrentPassword_example", "NewPassword_example") // UserChangePasswordRequest | Password to change

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.ChangeUserPassword(context.Background()).UserChangePasswordRequest(userChangePasswordRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.ChangeUserPassword``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ChangeUserPassword`: ChangeUserPassword200Response
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.ChangeUserPassword`: %v\n", resp)
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


## CreateGroup

> GroupCreationResponse CreateGroup(ctx).GroupCreationRequest(groupCreationRequest).Execute()

Create a new group.



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
    groupCreationRequest := *openapiclient.NewGroupCreationRequest("group1", "Email_example", []openapiclient.Role{openapiclient.Role("admin")}) // GroupCreationRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.CreateGroup(context.Background()).GroupCreationRequest(groupCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.CreateGroup``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateGroup`: GroupCreationResponse
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.CreateGroup`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **groupCreationRequest** | [**GroupCreationRequest**](GroupCreationRequest.md) |  | 

### Return type

[**GroupCreationResponse**](GroupCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateIntenralUser

> UserPostResponse CreateIntenralUser(ctx).InternalUserCreationRequest(internalUserCreationRequest).Execute()

create intenral user

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
    internalUserCreationRequest := *openapiclient.NewInternalUserCreationRequest("Email_example", "Realm_example", "Password_example", false) // InternalUserCreationRequest | Internal user object to create

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.CreateIntenralUser(context.Background()).InternalUserCreationRequest(internalUserCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.CreateIntenralUser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateIntenralUser`: UserPostResponse
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.CreateIntenralUser`: %v\n", resp)
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

> UserCreationResponse CreateUser(ctx).UserCreationRequest(userCreationRequest).Execute()

Create a new user.



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
    userCreationRequest := *openapiclient.NewUserCreationRequest("user@email.com", []openapiclient.Role{openapiclient.Role("admin")}, "secret!123") // UserCreationRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.CreateUser(context.Background()).UserCreationRequest(userCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.CreateUser``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateUser`: UserCreationResponse
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.CreateUser`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateUserRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userCreationRequest** | [**UserCreationRequest**](UserCreationRequest.md) |  | 

### Return type

[**UserCreationResponse**](UserCreationResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateUser_0

> UserPostResponse CreateUser_0(ctx).UserCreationRequest1(userCreationRequest1).Execute()

Create a local user.



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
    userCreationRequest1 := *openapiclient.NewUserCreationRequest1("Email_example") // UserCreationRequest1 | User object to create

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.CreateUser_0(context.Background()).UserCreationRequest1(userCreationRequest1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.CreateUser_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateUser_0`: UserPostResponse
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.CreateUser_0`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateUser_1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userCreationRequest1** | [**UserCreationRequest1**](UserCreationRequest1.md) | User object to create | 

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


## DeleteGroupByName

> DeleteGroupByName(ctx, groupName).Execute()

Delete a group.



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
    groupName := "group1" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.DeleteGroupByName(context.Background(), groupName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.DeleteGroupByName``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**groupName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteGroupByNameRequest struct via the builder pattern


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

> DeleteUserById(ctx, userId).UsersType(usersType).Execute()

Delete a user.



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
    userId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
    usersType := []openapiclient.UserType{openapiclient.UserType("sso-user")} // []UserType | Available only when SSO enabled (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.DeleteUserById(context.Background(), userId).UsersType(usersType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.DeleteUserById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteUserByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **usersType** | [**[]UserType**](UserType.md) | Available only when SSO enabled | 

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


## DeleteUserById_0

> DeleteUserById_0(ctx, userId).Execute()

Delete a user by id.



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
    userId := "userId_example" // string | The id of the user

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.DeleteUserById_0(context.Background(), userId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.DeleteUserById_0``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeleteUserById_2Request struct via the builder pattern


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


## GeGroupByName

> Group GeGroupByName(ctx, groupName).Execute()

Get group details.



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
    groupName := "group1" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GeGroupByName(context.Background(), groupName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GeGroupByName``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GeGroupByName`: Group
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GeGroupByName`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**groupName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGeGroupByNameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Group**](Group.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGroups

> []Group GetGroups(ctx).Roles(roles).OnlyCreatedByMe(onlyCreatedByMe).Execute()

Get groups list.



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
    roles := []openapiclient.Role{openapiclient.Role("admin")} // []Role | Filter the returned entities to only those with the required role (optional)
    onlyCreatedByMe := true // bool | Filter the returned entities to only those that created by the requesting subject (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GetGroups(context.Background()).Roles(roles).OnlyCreatedByMe(onlyCreatedByMe).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GetGroups``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetGroups`: []Group
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GetGroups`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **roles** | [**[]Role**](Role.md) | Filter the returned entities to only those with the required role | 
 **onlyCreatedByMe** | **bool** | Filter the returned entities to only those that created by the requesting subject | 

### Return type

[**[]Group**](Group.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRoles

> []Role GetRoles(ctx).Execute()

Get all possible permissions.



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

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GetRoles(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GetRoles``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRoles`: []Role
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GetRoles`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetRolesRequest struct via the builder pattern


### Return type

[**[]Role**](Role.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUserById

> User1 GetUserById(ctx, userId).Execute()

Get user details.



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
    userId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GetUserById(context.Background(), userId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GetUserById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetUserById`: User1
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GetUserById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User1**](User1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUserById_0

> User2 GetUserById_0(ctx, userId).Execute()

Get a user by id.



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
    userId := "userId_example" // string | The id of the user

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GetUserById_0(context.Background(), userId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GetUserById_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetUserById_0`: User2
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GetUserById_0`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | The id of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserById_3Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User2**](User2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUserRoles

> []Role GetUserRoles(ctx, userId).Execute()

Get user permissions.



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
    userId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GetUserRoles(context.Background(), userId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GetUserRoles``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetUserRoles`: []Role
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GetUserRoles`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetUserRolesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]Role**](Role.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUsers

> []User1 GetUsers(ctx).Roles(roles).OnlyCreatedByMe(onlyCreatedByMe).UsersType(usersType).Execute()

Get users list.



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
    roles := []openapiclient.Role{openapiclient.Role("admin")} // []Role | Filter the returned entities to only those with the required role (optional)
    onlyCreatedByMe := true // bool | Filter the returned entities to only those that created by the requesting subject (optional)
    usersType := []openapiclient.UserType{openapiclient.UserType("sso-user")} // []UserType | Available only when SSO enabled (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GetUsers(context.Background()).Roles(roles).OnlyCreatedByMe(onlyCreatedByMe).UsersType(usersType).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GetUsers``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetUsers`: []User1
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GetUsers`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetUsersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **roles** | [**[]Role**](Role.md) | Filter the returned entities to only those with the required role | 
 **onlyCreatedByMe** | **bool** | Filter the returned entities to only those that created by the requesting subject | 
 **usersType** | [**[]UserType**](UserType.md) | Available only when SSO enabled | 

### Return type

[**[]User1**](User1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetUsers_0

> []User2 GetUsers_0(ctx).Filter(filter).Execute()

Get users.



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
    filter := "runai.is_local:true" // string | Filter results by user attribute. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.GetUsers_0(context.Background()).Filter(filter).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.GetUsers_0``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetUsers_0`: []User2
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.GetUsers_0`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetUsers_4Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filter** | **string** | Filter results by user attribute. | 

### Return type

[**[]User2**](User2.md)

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
    openapiclient "./openapi"
)

func main() {
    userId := "userId_example" // string | The id of the user

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.LogoutUser(context.Background(), userId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.LogoutUser``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    userId := "userId_example" // string | The id of the user

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.ResetUserPassword(context.Background(), userId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.ResetUserPassword``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ResetUserPassword`: ResetUserPassword200Response
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.ResetUserPassword`: %v\n", resp)
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


## UpdateGroupByName

> Group UpdateGroupByName(ctx, groupName).GroupWithName(groupWithName).Execute()

Update group details.



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
    groupName := "group1" // string | 
    groupWithName := *openapiclient.NewGroupWithName() // GroupWithName | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.UpdateGroupByName(context.Background(), groupName).GroupWithName(groupWithName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.UpdateGroupByName``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateGroupByName`: Group
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.UpdateGroupByName`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**groupName** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateGroupByNameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **groupWithName** | [**GroupWithName**](GroupWithName.md) |  | 

### Return type

[**Group**](Group.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateUserById

> User1 UpdateUserById(ctx, userId).User1(user1).Execute()

Update user details.



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
    userId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | 
    user1 := *openapiclient.NewUser1("user@email.com") // User1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.UsersApi.UpdateUserById(context.Background(), userId).User1(user1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `UsersApi.UpdateUserById``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateUserById`: User1
    fmt.Fprintf(os.Stdout, "Response from `UsersApi.UpdateUserById`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateUserByIdRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **user1** | [**User1**](User1.md) |  | 

### Return type

[**User1**](User1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

