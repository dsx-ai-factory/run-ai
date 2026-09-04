# \OrgUnitApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetOrgTree**](OrgUnitApi.md#GetOrgTree) | **Get** /api/v1/org-unit | Get org tree
[**GetOrgUnitsNames**](OrgUnitApi.md#GetOrgUnitsNames) | **Get** /internal/org-unit/names | Get the names of requested org units.



## GetOrgTree

> GetOrgTree200Response GetOrgTree(ctx).Execute()

Get org tree

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
    resp, r, err := apiClient.OrgUnitApi.GetOrgTree(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitApi.GetOrgTree``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetOrgTree`: GetOrgTree200Response
    fmt.Fprintf(os.Stdout, "Response from `OrgUnitApi.GetOrgTree`: %v\n", resp)
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
    openapiclient "./openapi"
)

func main() {
    orgUnitType := "orgUnitType_example" // string | The org unit type.
    orgUnitIds := "7" // string | The org unit unique identifiers, seperated by ','.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.OrgUnitApi.GetOrgUnitsNames(context.Background()).OrgUnitType(orgUnitType).OrgUnitIds(orgUnitIds).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `OrgUnitApi.GetOrgUnitsNames``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetOrgUnitsNames`: []OrgUnit
    fmt.Fprintf(os.Stdout, "Response from `OrgUnitApi.GetOrgUnitsNames`: %v\n", resp)
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

