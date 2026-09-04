# \AssetsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**RequestAssetFlatFormat**](AssetsAPI.md#RequestAssetFlatFormat) | **Post** /api/v1/asset/flat | Request the flat representation of an asset



## RequestAssetFlatFormat

> AssetFlatResponse RequestAssetFlatFormat(ctx).AssetFlatRequest(assetFlatRequest).Execute()

Request the flat representation of an asset



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
)

func main() {
	assetFlatRequest := *openapiclient.NewAssetFlatRequest("AssetId_example", "71f69d83-ba66-4822-adf5-55ce55efd210", "1", openapiclient.WorkloadType("Workspace")) // AssetFlatRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AssetsAPI.RequestAssetFlatFormat(context.Background()).AssetFlatRequest(assetFlatRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AssetsAPI.RequestAssetFlatFormat``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `RequestAssetFlatFormat`: AssetFlatResponse
	fmt.Fprintf(os.Stdout, "Response from `AssetsAPI.RequestAssetFlatFormat`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiRequestAssetFlatFormatRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **assetFlatRequest** | [**AssetFlatRequest**](AssetFlatRequest.md) |  | 

### Return type

[**AssetFlatResponse**](AssetFlatResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

