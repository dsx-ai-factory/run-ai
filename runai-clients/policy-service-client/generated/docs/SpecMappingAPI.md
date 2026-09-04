# \SpecMappingAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateFieldGroup**](SpecMappingAPI.md#CreateFieldGroup) | **Post** /api/v1/policies/field-groups | Create a field group. [experimental]
[**CreateFormatter**](SpecMappingAPI.md#CreateFormatter) | **Post** /api/v1/policies/formatter | Create a formatter. [experimental]
[**CreateMapping**](SpecMappingAPI.md#CreateMapping) | **Post** /api/v1/policies/mappings | Create a mapping. [experimental]
[**DeleteFieldGroup**](SpecMappingAPI.md#DeleteFieldGroup) | **Delete** /api/v1/policies/field-groups/{FieldGroupName} | Delete a field group. [experimental]
[**DeleteFieldGroupFields**](SpecMappingAPI.md#DeleteFieldGroupFields) | **Delete** /api/v1/policies/field-groups/{FieldGroupName}/fields | Delete fields from a field group. [experimental]
[**DeleteFieldGroupItemSelectors**](SpecMappingAPI.md#DeleteFieldGroupItemSelectors) | **Delete** /api/v1/policies/field-groups/{FieldGroupName}/item-selectors | Delete item selectors from a field group. [experimental]
[**DeleteFormatter**](SpecMappingAPI.md#DeleteFormatter) | **Delete** /api/v1/policies/formatter/{FormatterName} | Delete a formatter. [experimental]
[**DeleteFormatterFields**](SpecMappingAPI.md#DeleteFormatterFields) | **Delete** /api/v1/policies/formatter/{FormatterName}/fields | Delete fields from a formatter. [experimental]
[**DeleteMapping**](SpecMappingAPI.md#DeleteMapping) | **Delete** /api/v1/policies/mappings | Delete a mapping. [experimental]
[**DeleteMappingSpecSelectors**](SpecMappingAPI.md#DeleteMappingSpecSelectors) | **Delete** /api/v1/policies/mappings/specSelectors | Delete spec selectors from a mapping. [experimental]
[**GetFieldGroup**](SpecMappingAPI.md#GetFieldGroup) | **Get** /api/v1/policies/field-groups/{FieldGroupName} | Get a field group. [experimental]
[**GetFormatter**](SpecMappingAPI.md#GetFormatter) | **Get** /api/v1/policies/formatter/{FormatterName} | Get a formatter. [experimental]
[**ListFieldGroups**](SpecMappingAPI.md#ListFieldGroups) | **Get** /api/v1/policies/field-groups | List field groups. [experimental]
[**ListFormatters**](SpecMappingAPI.md#ListFormatters) | **Get** /api/v1/policies/formatter | List formatters. [experimental]
[**ListMappings**](SpecMappingAPI.md#ListMappings) | **Get** /api/v1/policies/mappings | List mappings. [experimental]
[**PatchFieldGroup**](SpecMappingAPI.md#PatchFieldGroup) | **Patch** /api/v1/policies/field-groups/{FieldGroupName} | Patch a field group. [experimental]
[**PatchFormatter**](SpecMappingAPI.md#PatchFormatter) | **Patch** /api/v1/policies/formatter/{FormatterName} | Patch a formatter. [experimental]
[**UpdateMapping**](SpecMappingAPI.md#UpdateMapping) | **Patch** /api/v1/policies/mappings | Update a mapping. [experimental]



## CreateFieldGroup

> FieldGroup CreateFieldGroup(ctx).FieldGroupCreationRequest(fieldGroupCreationRequest).Execute()

Create a field group. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	fieldGroupCreationRequest := *openapiclient.NewFieldGroupCreationRequest("podSpec") // FieldGroupCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.CreateFieldGroup(context.Background()).FieldGroupCreationRequest(fieldGroupCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.CreateFieldGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateFieldGroup`: FieldGroup
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.CreateFieldGroup`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateFieldGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fieldGroupCreationRequest** | [**FieldGroupCreationRequest**](FieldGroupCreationRequest.md) |  | 

### Return type

[**FieldGroup**](FieldGroup.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateFormatter

> Formatter CreateFormatter(ctx).FormatterCreationRequest(formatterCreationRequest).Execute()

Create a formatter. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	formatterCreationRequest := *openapiclient.NewFormatterCreationRequest("podSpec") // FormatterCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.CreateFormatter(context.Background()).FormatterCreationRequest(formatterCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.CreateFormatter``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateFormatter`: Formatter
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.CreateFormatter`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateFormatterRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **formatterCreationRequest** | [**FormatterCreationRequest**](FormatterCreationRequest.md) |  | 

### Return type

[**Formatter**](Formatter.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateMapping

> Mapping CreateMapping(ctx).MappingCreUpdRequest(mappingCreUpdRequest).Execute()

Create a mapping. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	mappingCreUpdRequest := *openapiclient.NewMappingCreUpdRequest(*openapiclient.NewGVK("v1", "Deployment")) // MappingCreUpdRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.CreateMapping(context.Background()).MappingCreUpdRequest(mappingCreUpdRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.CreateMapping``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateMapping`: Mapping
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.CreateMapping`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateMappingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **mappingCreUpdRequest** | [**MappingCreUpdRequest**](MappingCreUpdRequest.md) |  | 

### Return type

[**Mapping**](Mapping.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteFieldGroup

> DeleteFieldGroup(ctx, fieldGroupName).Execute()

Delete a field group. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	fieldGroupName := "podSpec" // string | The unique name of the field group.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SpecMappingAPI.DeleteFieldGroup(context.Background(), fieldGroupName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.DeleteFieldGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldGroupName** | **string** | The unique name of the field group. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFieldGroupRequest struct via the builder pattern


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


## DeleteFieldGroupFields

> FieldGroup DeleteFieldGroupFields(ctx, fieldGroupName).FieldGroupDeleteFieldsRequest(fieldGroupDeleteFieldsRequest).Execute()

Delete fields from a field group. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	fieldGroupName := "podSpec" // string | The unique name of the field group.
	fieldGroupDeleteFieldsRequest := *openapiclient.NewFieldGroupDeleteFieldsRequest([]string{"containerPort"}) // FieldGroupDeleteFieldsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.DeleteFieldGroupFields(context.Background(), fieldGroupName).FieldGroupDeleteFieldsRequest(fieldGroupDeleteFieldsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.DeleteFieldGroupFields``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteFieldGroupFields`: FieldGroup
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.DeleteFieldGroupFields`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldGroupName** | **string** | The unique name of the field group. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFieldGroupFieldsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fieldGroupDeleteFieldsRequest** | [**FieldGroupDeleteFieldsRequest**](FieldGroupDeleteFieldsRequest.md) |  | 

### Return type

[**FieldGroup**](FieldGroup.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteFieldGroupItemSelectors

> FieldGroup DeleteFieldGroupItemSelectors(ctx, fieldGroupName).FieldGroupDeleteItemSelectorsRequest(fieldGroupDeleteItemSelectorsRequest).Execute()

Delete item selectors from a field group. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	fieldGroupName := "podSpec" // string | The unique name of the field group.
	fieldGroupDeleteItemSelectorsRequest := *openapiclient.NewFieldGroupDeleteItemSelectorsRequest([]string{"allPorts"}) // FieldGroupDeleteItemSelectorsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.DeleteFieldGroupItemSelectors(context.Background(), fieldGroupName).FieldGroupDeleteItemSelectorsRequest(fieldGroupDeleteItemSelectorsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.DeleteFieldGroupItemSelectors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteFieldGroupItemSelectors`: FieldGroup
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.DeleteFieldGroupItemSelectors`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldGroupName** | **string** | The unique name of the field group. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFieldGroupItemSelectorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fieldGroupDeleteItemSelectorsRequest** | [**FieldGroupDeleteItemSelectorsRequest**](FieldGroupDeleteItemSelectorsRequest.md) |  | 

### Return type

[**FieldGroup**](FieldGroup.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteFormatter

> DeleteFormatter(ctx, formatterName).Execute()

Delete a formatter. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	formatterName := "podSpec" // string | The unique name of the formatter.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SpecMappingAPI.DeleteFormatter(context.Background(), formatterName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.DeleteFormatter``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**formatterName** | **string** | The unique name of the formatter. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFormatterRequest struct via the builder pattern


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


## DeleteFormatterFields

> Formatter DeleteFormatterFields(ctx, formatterName).FormatterDeleteFieldsRequest(formatterDeleteFieldsRequest).Execute()

Delete fields from a formatter. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	formatterName := "podSpec" // string | The unique name of the formatter.
	formatterDeleteFieldsRequest := *openapiclient.NewFormatterDeleteFieldsRequest([]string{"containerPort"}) // FormatterDeleteFieldsRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.DeleteFormatterFields(context.Background(), formatterName).FormatterDeleteFieldsRequest(formatterDeleteFieldsRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.DeleteFormatterFields``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteFormatterFields`: Formatter
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.DeleteFormatterFields`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**formatterName** | **string** | The unique name of the formatter. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFormatterFieldsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **formatterDeleteFieldsRequest** | [**FormatterDeleteFieldsRequest**](FormatterDeleteFieldsRequest.md) |  | 

### Return type

[**Formatter**](Formatter.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteMapping

> DeleteMapping(ctx).MappingDeletionRequest(mappingDeletionRequest).Execute()

Delete a mapping. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	mappingDeletionRequest := *openapiclient.NewMappingDeletionRequest(*openapiclient.NewGVK("v1", "Deployment")) // MappingDeletionRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.SpecMappingAPI.DeleteMapping(context.Background()).MappingDeletionRequest(mappingDeletionRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.DeleteMapping``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteMappingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **mappingDeletionRequest** | [**MappingDeletionRequest**](MappingDeletionRequest.md) |  | 

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


## DeleteMappingSpecSelectors

> Mapping DeleteMappingSpecSelectors(ctx).MappingDeleteSpecSelectorsReq(mappingDeleteSpecSelectorsReq).Execute()

Delete spec selectors from a mapping. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	mappingDeleteSpecSelectorsReq := *openapiclient.NewMappingDeleteSpecSelectorsReq(*openapiclient.NewGVK("v1", "Deployment"), []string{"containerByName"}) // MappingDeleteSpecSelectorsReq |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.DeleteMappingSpecSelectors(context.Background()).MappingDeleteSpecSelectorsReq(mappingDeleteSpecSelectorsReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.DeleteMappingSpecSelectors``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteMappingSpecSelectors`: Mapping
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.DeleteMappingSpecSelectors`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteMappingSpecSelectorsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **mappingDeleteSpecSelectorsReq** | [**MappingDeleteSpecSelectorsReq**](MappingDeleteSpecSelectorsReq.md) |  | 

### Return type

[**Mapping**](Mapping.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFieldGroup

> FieldGroup GetFieldGroup(ctx, fieldGroupName).Execute()

Get a field group. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	fieldGroupName := "podSpec" // string | The unique name of the field group.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.GetFieldGroup(context.Background(), fieldGroupName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.GetFieldGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFieldGroup`: FieldGroup
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.GetFieldGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldGroupName** | **string** | The unique name of the field group. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFieldGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**FieldGroup**](FieldGroup.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFormatter

> Formatter GetFormatter(ctx, formatterName).Execute()

Get a formatter. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	formatterName := "podSpec" // string | The unique name of the formatter.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.GetFormatter(context.Background(), formatterName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.GetFormatter``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFormatter`: Formatter
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.GetFormatter`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**formatterName** | **string** | The unique name of the formatter. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetFormatterRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Formatter**](Formatter.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListFieldGroups

> FieldGroupListResponse ListFieldGroups(ctx).Offset(offset).Limit(limit).Execute()

List field groups. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.ListFieldGroups(context.Background()).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.ListFieldGroups``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListFieldGroups`: FieldGroupListResponse
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.ListFieldGroups`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListFieldGroupsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**FieldGroupListResponse**](FieldGroupListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListFormatters

> FormatterListResponse ListFormatters(ctx).Offset(offset).Limit(limit).Execute()

List formatters. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.ListFormatters(context.Background()).Offset(offset).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.ListFormatters``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListFormatters`: FormatterListResponse
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.ListFormatters`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListFormattersRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]

### Return type

[**FormatterListResponse**](FormatterListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListMappings

> MappingListResponse ListMappings(ctx).FilterBy(filterBy).Execute()

List mappings. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter mappings by workload type. Each expression uses the format field-name operator value, where field-name is one of `group`, `kind`, or `version`. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` Contains, `!@` Does not contain, `=^` Starts with, and `=$` Ends with. Up to 10 expressions may be supplied, and up to 1000 mappings are returned. For example, [\"kind==dynamoGraphDeployment\",\"version==v1alpha1\"]. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.ListMappings(context.Background()).FilterBy(filterBy).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.ListMappings``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListMappings`: MappingListResponse
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.ListMappings`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListMappingsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter mappings by workload type. Each expression uses the format field-name operator value, where field-name is one of &#x60;group&#x60;, &#x60;kind&#x60;, or &#x60;version&#x60;. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; Contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with, and &#x60;&#x3D;$&#x60; Ends with. Up to 10 expressions may be supplied, and up to 1000 mappings are returned. For example, [\&quot;kind&#x3D;&#x3D;dynamoGraphDeployment\&quot;,\&quot;version&#x3D;&#x3D;v1alpha1\&quot;]. | 

### Return type

[**MappingListResponse**](MappingListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchFieldGroup

> FieldGroup PatchFieldGroup(ctx, fieldGroupName).FieldGroupUpdateRequest(fieldGroupUpdateRequest).Execute()

Patch a field group. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	fieldGroupName := "podSpec" // string | The unique name of the field group.
	fieldGroupUpdateRequest := *openapiclient.NewFieldGroupUpdateRequest() // FieldGroupUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.PatchFieldGroup(context.Background(), fieldGroupName).FieldGroupUpdateRequest(fieldGroupUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.PatchFieldGroup``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchFieldGroup`: FieldGroup
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.PatchFieldGroup`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**fieldGroupName** | **string** | The unique name of the field group. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchFieldGroupRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fieldGroupUpdateRequest** | [**FieldGroupUpdateRequest**](FieldGroupUpdateRequest.md) |  | 

### Return type

[**FieldGroup**](FieldGroup.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## PatchFormatter

> Formatter PatchFormatter(ctx, formatterName).FormatterUpdateRequest(formatterUpdateRequest).Execute()

Patch a formatter. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	formatterName := "podSpec" // string | The unique name of the formatter.
	formatterUpdateRequest := *openapiclient.NewFormatterUpdateRequest() // FormatterUpdateRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.PatchFormatter(context.Background(), formatterName).FormatterUpdateRequest(formatterUpdateRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.PatchFormatter``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchFormatter`: Formatter
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.PatchFormatter`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**formatterName** | **string** | The unique name of the formatter. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchFormatterRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **formatterUpdateRequest** | [**FormatterUpdateRequest**](FormatterUpdateRequest.md) |  | 

### Return type

[**Formatter**](Formatter.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateMapping

> Mapping UpdateMapping(ctx).MappingCreUpdRequest(mappingCreUpdRequest).Execute()

Update a mapping. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
)

func main() {
	mappingCreUpdRequest := *openapiclient.NewMappingCreUpdRequest(*openapiclient.NewGVK("v1", "Deployment")) // MappingCreUpdRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SpecMappingAPI.UpdateMapping(context.Background()).MappingCreUpdRequest(mappingCreUpdRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SpecMappingAPI.UpdateMapping``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateMapping`: Mapping
	fmt.Fprintf(os.Stdout, "Response from `SpecMappingAPI.UpdateMapping`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateMappingRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **mappingCreUpdRequest** | [**MappingCreUpdRequest**](MappingCreUpdRequest.md) |  | 

### Return type

[**Mapping**](Mapping.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

