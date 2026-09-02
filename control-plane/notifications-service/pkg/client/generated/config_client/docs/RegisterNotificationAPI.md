# \RegisterNotificationAPI

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateNotificationKind**](RegisterNotificationAPI.md#CreateNotificationKind) | **Post** /api/v1/notification-kinds | 
[**DeleteNotificationKind**](RegisterNotificationAPI.md#DeleteNotificationKind) | **Delete** /api/v1/notification-kinds | 
[**GetNotificationKinds**](RegisterNotificationAPI.md#GetNotificationKinds) | **Get** /api/v1/notification-kinds | 
[**UpdateNotificationKind**](RegisterNotificationAPI.md#UpdateNotificationKind) | **Put** /api/v1/notification-kinds | 



## CreateNotificationKind

> string CreateNotificationKind(ctx).NotificationKind(notificationKind).Execute()





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
	notificationKind := *openapiclient.NewNotificationKind("Origin_example", "Category_example", "Type_example", "Description_example", "Template_example") // NotificationKind | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegisterNotificationAPI.CreateNotificationKind(context.Background()).NotificationKind(notificationKind).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegisterNotificationAPI.CreateNotificationKind``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateNotificationKind`: string
	fmt.Fprintf(os.Stdout, "Response from `RegisterNotificationAPI.CreateNotificationKind`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateNotificationKindRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notificationKind** | [**NotificationKind**](NotificationKind.md) |  | 

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


## DeleteNotificationKind

> string DeleteNotificationKind(ctx).Category(category).Type_(type_).Execute()





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
	category := "category_example" // string | 
	type_ := "type__example" // string | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegisterNotificationAPI.DeleteNotificationKind(context.Background()).Category(category).Type_(type_).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegisterNotificationAPI.DeleteNotificationKind``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteNotificationKind`: string
	fmt.Fprintf(os.Stdout, "Response from `RegisterNotificationAPI.DeleteNotificationKind`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNotificationKindRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **category** | **string** |  | 
 **type_** | **string** |  | 

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


## GetNotificationKinds

> []NotificationKind GetNotificationKinds(ctx).Origin(origin).Category(category).Type_(type_).Execute()





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
	origin := "origin_example" // string |  (optional)
	category := "category_example" // string |  (optional)
	type_ := "type__example" // string |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegisterNotificationAPI.GetNotificationKinds(context.Background()).Origin(origin).Category(category).Type_(type_).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegisterNotificationAPI.GetNotificationKinds``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNotificationKinds`: []NotificationKind
	fmt.Fprintf(os.Stdout, "Response from `RegisterNotificationAPI.GetNotificationKinds`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNotificationKindsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **origin** | **string** |  | 
 **category** | **string** |  | 
 **type_** | **string** |  | 

### Return type

[**[]NotificationKind**](NotificationKind.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateNotificationKind

> string UpdateNotificationKind(ctx).NotificationKind(notificationKind).Execute()





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
	notificationKind := *openapiclient.NewNotificationKind("Origin_example", "Category_example", "Type_example", "Description_example", "Template_example") // NotificationKind | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.RegisterNotificationAPI.UpdateNotificationKind(context.Background()).NotificationKind(notificationKind).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `RegisterNotificationAPI.UpdateNotificationKind``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNotificationKind`: string
	fmt.Fprintf(os.Stdout, "Response from `RegisterNotificationAPI.UpdateNotificationKind`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNotificationKindRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **notificationKind** | [**NotificationKind**](NotificationKind.md) |  | 

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

