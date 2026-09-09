# \OpenIDConnectAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetOauthAuthorizationServerMetadata**](OpenIDConnectAPI.md#GetOauthAuthorizationServerMetadata) | **Get** /.well-known/oauth-authorization-server | OAuth 2.0 Authorization Server Metadata
[**GetOpenidConfiguration**](OpenIDConnectAPI.md#GetOpenidConfiguration) | **Get** /.well-known/openid-configuration | OpenID Connect Discovery



## GetOauthAuthorizationServerMetadata

> OAuthAuthorizationServer GetOauthAuthorizationServerMetadata(ctx).Execute()

OAuth 2.0 Authorization Server Metadata



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OpenIDConnectAPI.GetOauthAuthorizationServerMetadata(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OpenIDConnectAPI.GetOauthAuthorizationServerMetadata``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOauthAuthorizationServerMetadata`: OAuthAuthorizationServer
	fmt.Fprintf(os.Stdout, "Response from `OpenIDConnectAPI.GetOauthAuthorizationServerMetadata`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetOauthAuthorizationServerMetadataRequest struct via the builder pattern


### Return type

[**OAuthAuthorizationServer**](OAuthAuthorizationServer.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetOpenidConfiguration

> OpenIDConfiguration GetOpenidConfiguration(ctx).Execute()

OpenID Connect Discovery



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OpenIDConnectAPI.GetOpenidConfiguration(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OpenIDConnectAPI.GetOpenidConfiguration``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOpenidConfiguration`: OpenIDConfiguration
	fmt.Fprintf(os.Stdout, "Response from `OpenIDConnectAPI.GetOpenidConfiguration`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetOpenidConfigurationRequest struct via the builder pattern


### Return type

[**OpenIDConfiguration**](OpenIDConfiguration.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

