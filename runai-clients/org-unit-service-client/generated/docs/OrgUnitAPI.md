# \OrgUnitAPI

All URIs are relative to *http://localhost:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetInternalDeletedOrgUnits**](OrgUnitAPI.md#GetInternalDeletedOrgUnits) | **Get** /internal/org-unit/deleted-org-units | Get internal deleted org-units
[**GetInternalOrgTree**](OrgUnitAPI.md#GetInternalOrgTree) | **Get** /internal/org-unit/org-tree | Get internal org tree
[**GetOrgTree**](OrgUnitAPI.md#GetOrgTree) | **Get** /api/v1/org-unit | Get org tree
[**GetOrgUnitsNames**](OrgUnitAPI.md#GetOrgUnitsNames) | **Get** /internal/org-unit/names | Get the names of requested org units.
[**GetRanks**](OrgUnitAPI.md#GetRanks) | **Get** /api/v1/org-unit/ranks | Get ranks
[**NodePoolResetResources**](OrgUnitAPI.md#NodePoolResetResources) | **Post** /internal/org-unit/node-pools/resources/reset | Reset resources of a specific nodepool for all departments and projects in the cluster.
[**OrgsNodePoolSummary**](OrgUnitAPI.md#OrgsNodePoolSummary) | **Get** /internal/org-unit/node-pools/{nodepoolId}/summary | Get org units summary per nodepool
[**UpdateOrgUnitSettings**](OrgUnitAPI.md#UpdateOrgUnitSettings) | **Post** /internal/org-unit/settings | 



## GetInternalDeletedOrgUnits

> GetInternalDeletedOrgUnits200Response GetInternalDeletedOrgUnits(ctx).DeletedAfter(deletedAfter).Execute()

Get internal deleted org-units

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
    "time"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	deletedAfter := time.Now() // time.Time | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OrgUnitAPI.GetInternalDeletedOrgUnits(context.Background()).DeletedAfter(deletedAfter).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.GetInternalDeletedOrgUnits``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInternalDeletedOrgUnits`: GetInternalDeletedOrgUnits200Response
	fmt.Fprintf(os.Stdout, "Response from `OrgUnitAPI.GetInternalDeletedOrgUnits`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetInternalDeletedOrgUnitsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deletedAfter** | **time.Time** |  | 

### Return type

[**GetInternalDeletedOrgUnits200Response**](GetInternalDeletedOrgUnits200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInternalOrgTree

> GetInternalOrgTree200Response GetInternalOrgTree(ctx).Execute()

Get internal org tree

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OrgUnitAPI.GetInternalOrgTree(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.GetInternalOrgTree``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInternalOrgTree`: GetInternalOrgTree200Response
	fmt.Fprintf(os.Stdout, "Response from `OrgUnitAPI.GetInternalOrgTree`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetInternalOrgTreeRequest struct via the builder pattern


### Return type

[**GetInternalOrgTree200Response**](GetInternalOrgTree200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetOrgTree

> GetInternalOrgTree200Response GetOrgTree(ctx).Execute()

Get org tree

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OrgUnitAPI.GetOrgTree(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.GetOrgTree``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOrgTree`: GetInternalOrgTree200Response
	fmt.Fprintf(os.Stdout, "Response from `OrgUnitAPI.GetOrgTree`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetOrgTreeRequest struct via the builder pattern


### Return type

[**GetInternalOrgTree200Response**](GetInternalOrgTree200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetOrgUnitsNames

> []OrgUnit GetOrgUnitsNames(ctx).OrgUnitType(orgUnitType).OrgUnitIds(orgUnitIds).Execute()

Get the names of requested org units.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	orgUnitType := "orgUnitType_example" // string | The org unit type.
	orgUnitIds := "7" // string | The org unit unique identifiers, seperated by ','.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OrgUnitAPI.GetOrgUnitsNames(context.Background()).OrgUnitType(orgUnitType).OrgUnitIds(orgUnitIds).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.GetOrgUnitsNames``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetOrgUnitsNames`: []OrgUnit
	fmt.Fprintf(os.Stdout, "Response from `OrgUnitAPI.GetOrgUnitsNames`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetOrgUnitsNamesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **orgUnitType** | **string** | The org unit type. | 
 **orgUnitIds** | **string** | The org unit unique identifiers, seperated by &#39;,&#39;. | 

### Return type

[**[]OrgUnit**](OrgUnit.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRanks

> GetRanks200Response GetRanks(ctx).Execute()

Get ranks



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OrgUnitAPI.GetRanks(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.GetRanks``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetRanks`: GetRanks200Response
	fmt.Fprintf(os.Stdout, "Response from `OrgUnitAPI.GetRanks`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGetRanksRequest struct via the builder pattern


### Return type

[**GetRanks200Response**](GetRanks200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## NodePoolResetResources

> NodePoolResetResources(ctx).OrgUnitsNodepoolResourcesResetRequest(orgUnitsNodepoolResourcesResetRequest).Execute()

Reset resources of a specific nodepool for all departments and projects in the cluster.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	orgUnitsNodepoolResourcesResetRequest := *openapiclient.NewOrgUnitsNodepoolResourcesResetRequest("71f69d83-ba66-4822-adf5-55ce55efd210", "22", openapiclient.NodepoolResourcesResetAction("CPU_QUOTAS_TO_DEFAULT")) // OrgUnitsNodepoolResourcesResetRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.OrgUnitAPI.NodePoolResetResources(context.Background()).OrgUnitsNodepoolResourcesResetRequest(orgUnitsNodepoolResourcesResetRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.NodePoolResetResources``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiNodePoolResetResourcesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **orgUnitsNodepoolResourcesResetRequest** | [**OrgUnitsNodepoolResourcesResetRequest**](OrgUnitsNodepoolResourcesResetRequest.md) |  | 

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


## OrgsNodePoolSummary

> OrgsNodePoolSummary200Response OrgsNodePoolSummary(ctx, nodepoolId).NodepoolSummaryType(nodepoolSummaryType).Execute()

Get org units summary per nodepool

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	nodepoolId := "17" // string | The unique id identifying the Node Pool.
	nodepoolSummaryType := openapiclient.OrgUnitsNodepoolSummaryType("HAS_QUOTA") // OrgUnitsNodepoolSummaryType | Specify which summary to request.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.OrgUnitAPI.OrgsNodePoolSummary(context.Background(), nodepoolId).NodepoolSummaryType(nodepoolSummaryType).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.OrgsNodePoolSummary``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OrgsNodePoolSummary`: OrgsNodePoolSummary200Response
	fmt.Fprintf(os.Stdout, "Response from `OrgUnitAPI.OrgsNodePoolSummary`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**nodepoolId** | **string** | The unique id identifying the Node Pool. | 

### Other Parameters

Other parameters are passed through a pointer to a apiOrgsNodePoolSummaryRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nodepoolSummaryType** | [**OrgUnitsNodepoolSummaryType**](OrgUnitsNodepoolSummaryType.md) | Specify which summary to request. | 

### Return type

[**OrgsNodePoolSummary200Response**](OrgsNodePoolSummary200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateOrgUnitSettings

> UpdateOrgUnitSettings(ctx).OrgUnitSingleSetting(orgUnitSingleSetting).Execute()



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client"
)

func main() {
	orgUnitSingleSetting := *openapiclient.NewOrgUnitSingleSetting() // OrgUnitSingleSetting | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.OrgUnitAPI.UpdateOrgUnitSettings(context.Background()).OrgUnitSingleSetting(orgUnitSingleSetting).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitAPI.UpdateOrgUnitSettings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateOrgUnitSettingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **orgUnitSingleSetting** | [**OrgUnitSingleSetting**](OrgUnitSingleSetting.md) |  | 

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

