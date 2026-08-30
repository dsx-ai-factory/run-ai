# \MeAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CountMeAccessRules**](MeAPI.md#CountMeAccessRules) | **Get** /api/v1/authorization/me/access-rules/count | Count the access rules assigned to the requesting user.
[**GetMeAccessRules**](MeAPI.md#GetMeAccessRules) | **Get** /api/v1/authorization/me/access-rules | List the access rules assigned to the requesting user.



## CountMeAccessRules

> CountAccessRules200Response CountMeAccessRules(ctx).Search(search).Execute()

Count the access rules assigned to the requesting user.



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
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MeAPI.CountMeAccessRules(context.Background()).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MeAPI.CountMeAccessRules``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CountMeAccessRules`: CountAccessRules200Response
	fmt.Fprintf(os.Stdout, "Response from `MeAPI.CountMeAccessRules`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCountMeAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
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


## GetMeAccessRules

> MeAccessRulesResponse GetMeAccessRules(ctx).Limit(limit).Offset(offset).Search(search).Execute()

List the access rules assigned to the requesting user.



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
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	search := "test project" // string | Filter results by a free text search. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MeAPI.GetMeAccessRules(context.Background()).Limit(limit).Offset(offset).Search(search).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MeAPI.GetMeAccessRules``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMeAccessRules`: MeAccessRulesResponse
	fmt.Fprintf(os.Stdout, "Response from `MeAPI.GetMeAccessRules`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetMeAccessRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **search** | **string** | Filter results by a free text search. | 

### Return type

[**MeAccessRulesResponse**](MeAccessRulesResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

