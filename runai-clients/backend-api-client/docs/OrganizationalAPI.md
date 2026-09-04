# \OrganizationalAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetOrgTree**](OrganizationalAPI.md#GetOrgTree) | **Get** /v1/k8s/org-tree | Get tenant organizational tree



## GetOrgTree

> GetOrgTree200Response GetOrgTree(ctx).Execute()

Get tenant organizational tree

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/backend-api-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OrganizationalAPI.GetOrgTree(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrganizationalAPI.GetOrgTree``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOrgTree`: GetOrgTree200Response
	fmt.Fprintf(os.Stdout, "Response from `OrganizationalAPI.GetOrgTree`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetOrgTreeRequest struct via the builder pattern


### Return type

[**GetOrgTree200Response**](GetOrgTree200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

