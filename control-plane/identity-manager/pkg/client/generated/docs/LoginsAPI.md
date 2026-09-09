# \LoginsAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetLoginScreenURL**](LoginsAPI.md#GetLoginScreenURL) | **Get** /api/v1/login | Get login url



## GetLoginScreenURL

> GetLoginScreenURL(ctx).UserAgent(userAgent).AuthenticationFlow(authenticationFlow).RedirectUrl(redirectUrl).CodeChallenge(codeChallenge).CodeChallengeMethod(codeChallengeMethod).Execute()

Get login url



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
	userAgent := "userAgent_example" // string | 
	authenticationFlow := "authenticationFlow_example" // string |  (optional)
	redirectUrl := "redirectUrl_example" // string |  (optional)
	codeChallenge := "codeChallenge_example" // string |  (optional)
	codeChallengeMethod := openapiclient.CodeChallengeMethod("plain") // CodeChallengeMethod |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.LoginsAPI.GetLoginScreenURL(context.Background()).UserAgent(userAgent).AuthenticationFlow(authenticationFlow).RedirectUrl(redirectUrl).CodeChallenge(codeChallenge).CodeChallengeMethod(codeChallengeMethod).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LoginsAPI.GetLoginScreenURL``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetLoginScreenURLRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userAgent** | **string** |  | 
 **authenticationFlow** | **string** |  | 
 **redirectUrl** | **string** |  | 
 **codeChallenge** | **string** |  | 
 **codeChallengeMethod** | [**CodeChallengeMethod**](CodeChallengeMethod.md) |  | 

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

