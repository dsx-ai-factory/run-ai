# \PolicyAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AssetChangedHook**](PolicyAPI.md#AssetChangedHook) | **Put** /internal/policy/asset/{AssetId} | Asset changed hook for asset service
[**DeleteDistributedPolicy**](PolicyAPI.md#DeleteDistributedPolicy) | **Delete** /api/v2/policy/distributed | Delete a distributed policy.
[**DeleteInferencePolicy**](PolicyAPI.md#DeleteInferencePolicy) | **Delete** /api/v2/policy/inferences | Delete an inference policy.
[**DeleteTrainingPolicy**](PolicyAPI.md#DeleteTrainingPolicy) | **Delete** /api/v2/policy/trainings | Delete a training policy.
[**DeleteWorkspacePolicy**](PolicyAPI.md#DeleteWorkspacePolicy) | **Delete** /api/v2/policy/workspaces | Delete a workspace policy.
[**GetDistributedPolicyV2**](PolicyAPI.md#GetDistributedPolicyV2) | **Get** /api/v2/policy/distributed | Get a distributed policy.
[**GetInferencePolicyV2**](PolicyAPI.md#GetInferencePolicyV2) | **Get** /api/v2/policy/inferences | Get an inference policy.
[**GetTrainingPolicyV2**](PolicyAPI.md#GetTrainingPolicyV2) | **Get** /api/v2/policy/trainings | Get a training policy.
[**GetWorkspacePolicyV2**](PolicyAPI.md#GetWorkspacePolicyV2) | **Get** /api/v2/policy/workspaces | Get a workspace policy.
[**ListPolicies**](PolicyAPI.md#ListPolicies) | **Get** /api/v2/policy | List policies
[**OverwriteDistributedPolicyV2**](PolicyAPI.md#OverwriteDistributedPolicyV2) | **Put** /api/v2/policy/distributed | Overwrite a distributed policy.
[**OverwriteInferencePolicyV2**](PolicyAPI.md#OverwriteInferencePolicyV2) | **Put** /api/v2/policy/inferences | Overwrite an inference policy.
[**OverwriteTrainingPolicyV2**](PolicyAPI.md#OverwriteTrainingPolicyV2) | **Put** /api/v2/policy/trainings | Overwrite a training policy.
[**OverwriteWorkspacePolicyV2**](PolicyAPI.md#OverwriteWorkspacePolicyV2) | **Put** /api/v2/policy/workspaces | Overwrite a workspace policy.
[**UpdateDistributedPolicyV2**](PolicyAPI.md#UpdateDistributedPolicyV2) | **Patch** /api/v2/policy/distributed | Update a distributed policy.
[**UpdateInferencePolicyV2**](PolicyAPI.md#UpdateInferencePolicyV2) | **Patch** /api/v2/policy/inferences | Update an inference policy.
[**UpdateTrainingPolicyV2**](PolicyAPI.md#UpdateTrainingPolicyV2) | **Patch** /api/v2/policy/trainings | Update a training policy.
[**UpdateWorkspacePolicyV2**](PolicyAPI.md#UpdateWorkspacePolicyV2) | **Patch** /api/v2/policy/workspaces | Update a workspace policy.



## AssetChangedHook

> AssetChangedHook(ctx, assetId).AssetChangeNotification(assetChangeNotification).Execute()

Asset changed hook for asset service



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	assetId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | Unique identifier of the asset.
	assetChangeNotification := *openapiclient.NewAssetChangeNotification() // AssetChangeNotification |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.AssetChangedHook(context.Background(), assetId).AssetChangeNotification(assetChangeNotification).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.AssetChangedHook``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**assetId** | **string** | Unique identifier of the asset. | 

### Other Parameters

Other parameters are passed through a pointer to a apiAssetChangedHookRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **assetChangeNotification** | [**AssetChangeNotification**](AssetChangeNotification.md) |  | 

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


## DeleteDistributedPolicy

> DeleteDistributedPolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()

Delete a distributed policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteDistributedPolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteDistributedPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDistributedPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

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


## DeleteInferencePolicy

> DeleteInferencePolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()

Delete an inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteInferencePolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteInferencePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteInferencePolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

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


## DeleteTrainingPolicy

> DeleteTrainingPolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()

Delete a training policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteTrainingPolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteTrainingPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteTrainingPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

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


## DeleteWorkspacePolicy

> DeleteWorkspacePolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()

Delete a workspace policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteWorkspacePolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteWorkspacePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteWorkspacePolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

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


## GetDistributedPolicyV2

> DistributedPolicyV2 GetDistributedPolicyV2(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get a distributed policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetDistributedPolicyV2(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetDistributedPolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDistributedPolicyV2`: DistributedPolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetDistributedPolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDistributedPolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**DistributedPolicyV2**](DistributedPolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetInferencePolicyV2

> InferencePolicyV2 GetInferencePolicyV2(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get an inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetInferencePolicyV2(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetInferencePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetInferencePolicyV2`: InferencePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetInferencePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetInferencePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**InferencePolicyV2**](InferencePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetTrainingPolicyV2

> TrainingPolicyV2 GetTrainingPolicyV2(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get a training policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetTrainingPolicyV2(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetTrainingPolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetTrainingPolicyV2`: TrainingPolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetTrainingPolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetTrainingPolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**TrainingPolicyV2**](TrainingPolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetWorkspacePolicyV2

> WorkspacePolicyV2 GetWorkspacePolicyV2(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get a workspace policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetWorkspacePolicyV2(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetWorkspacePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetWorkspacePolicyV2`: WorkspacePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetWorkspacePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetWorkspacePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**WorkspacePolicyV2**](WorkspacePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPolicies

> PolicyListResponse ListPolicies(ctx).WorkloadType(workloadType).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()

List policies



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	workloadType := "workloadType_example" // string | Policy for a specific workload type. (optional)
	scope := "scope_example" // string | filter by this scope. (optional)
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.ListPolicies(context.Background()).WorkloadType(workloadType).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.ListPolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPolicies`: PolicyListResponse
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.ListPolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadType** | **string** | Policy for a specific workload type. | 
 **scope** | **string** | filter by this scope. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

### Return type

[**PolicyListResponse**](PolicyListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OverwriteDistributedPolicyV2

> DistributedPolicyV2 OverwriteDistributedPolicyV2(ctx).ValidateOnly(validateOnly).DistributedPolicyOverwriteRequestV2(distributedPolicyOverwriteRequestV2).Execute()

Overwrite a distributed policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	distributedPolicyOverwriteRequestV2 := *openapiclient.NewDistributedPolicyOverwriteRequestV2() // DistributedPolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteDistributedPolicyV2(context.Background()).ValidateOnly(validateOnly).DistributedPolicyOverwriteRequestV2(distributedPolicyOverwriteRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.OverwriteDistributedPolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OverwriteDistributedPolicyV2`: DistributedPolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.OverwriteDistributedPolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOverwriteDistributedPolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **distributedPolicyOverwriteRequestV2** | [**DistributedPolicyOverwriteRequestV2**](DistributedPolicyOverwriteRequestV2.md) |  | 

### Return type

[**DistributedPolicyV2**](DistributedPolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OverwriteInferencePolicyV2

> InferencePolicyV2 OverwriteInferencePolicyV2(ctx).ValidateOnly(validateOnly).InferencePolicyOverwriteRequestV2(inferencePolicyOverwriteRequestV2).Execute()

Overwrite an inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	inferencePolicyOverwriteRequestV2 := *openapiclient.NewInferencePolicyOverwriteRequestV2() // InferencePolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteInferencePolicyV2(context.Background()).ValidateOnly(validateOnly).InferencePolicyOverwriteRequestV2(inferencePolicyOverwriteRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.OverwriteInferencePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OverwriteInferencePolicyV2`: InferencePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.OverwriteInferencePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOverwriteInferencePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **inferencePolicyOverwriteRequestV2** | [**InferencePolicyOverwriteRequestV2**](InferencePolicyOverwriteRequestV2.md) |  | 

### Return type

[**InferencePolicyV2**](InferencePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OverwriteTrainingPolicyV2

> TrainingPolicyV2 OverwriteTrainingPolicyV2(ctx).ValidateOnly(validateOnly).TrainingPolicyOverwriteRequestV2(trainingPolicyOverwriteRequestV2).Execute()

Overwrite a training policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	trainingPolicyOverwriteRequestV2 := *openapiclient.NewTrainingPolicyOverwriteRequestV2() // TrainingPolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteTrainingPolicyV2(context.Background()).ValidateOnly(validateOnly).TrainingPolicyOverwriteRequestV2(trainingPolicyOverwriteRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.OverwriteTrainingPolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OverwriteTrainingPolicyV2`: TrainingPolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.OverwriteTrainingPolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOverwriteTrainingPolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **trainingPolicyOverwriteRequestV2** | [**TrainingPolicyOverwriteRequestV2**](TrainingPolicyOverwriteRequestV2.md) |  | 

### Return type

[**TrainingPolicyV2**](TrainingPolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OverwriteWorkspacePolicyV2

> WorkspacePolicyV2 OverwriteWorkspacePolicyV2(ctx).ValidateOnly(validateOnly).WorkspacePolicyOverwriteRequestV2(workspacePolicyOverwriteRequestV2).Execute()

Overwrite a workspace policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	workspacePolicyOverwriteRequestV2 := *openapiclient.NewWorkspacePolicyOverwriteRequestV2() // WorkspacePolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteWorkspacePolicyV2(context.Background()).ValidateOnly(validateOnly).WorkspacePolicyOverwriteRequestV2(workspacePolicyOverwriteRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.OverwriteWorkspacePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OverwriteWorkspacePolicyV2`: WorkspacePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.OverwriteWorkspacePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOverwriteWorkspacePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **workspacePolicyOverwriteRequestV2** | [**WorkspacePolicyOverwriteRequestV2**](WorkspacePolicyOverwriteRequestV2.md) |  | 

### Return type

[**WorkspacePolicyV2**](WorkspacePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDistributedPolicyV2

> DistributedPolicyV2 UpdateDistributedPolicyV2(ctx).ValidateOnly(validateOnly).DistributedPolicyChangeRequestV2(distributedPolicyChangeRequestV2).Execute()

Update a distributed policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	distributedPolicyChangeRequestV2 := *openapiclient.NewDistributedPolicyChangeRequestV2() // DistributedPolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateDistributedPolicyV2(context.Background()).ValidateOnly(validateOnly).DistributedPolicyChangeRequestV2(distributedPolicyChangeRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateDistributedPolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDistributedPolicyV2`: DistributedPolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateDistributedPolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDistributedPolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **distributedPolicyChangeRequestV2** | [**DistributedPolicyChangeRequestV2**](DistributedPolicyChangeRequestV2.md) |  | 

### Return type

[**DistributedPolicyV2**](DistributedPolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateInferencePolicyV2

> InferencePolicyV2 UpdateInferencePolicyV2(ctx).ValidateOnly(validateOnly).InferencePolicyChangeRequestV2(inferencePolicyChangeRequestV2).Execute()

Update an inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	inferencePolicyChangeRequestV2 := *openapiclient.NewInferencePolicyChangeRequestV2() // InferencePolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateInferencePolicyV2(context.Background()).ValidateOnly(validateOnly).InferencePolicyChangeRequestV2(inferencePolicyChangeRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateInferencePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateInferencePolicyV2`: InferencePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateInferencePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateInferencePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **inferencePolicyChangeRequestV2** | [**InferencePolicyChangeRequestV2**](InferencePolicyChangeRequestV2.md) |  | 

### Return type

[**InferencePolicyV2**](InferencePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTrainingPolicyV2

> TrainingPolicyV2 UpdateTrainingPolicyV2(ctx).ValidateOnly(validateOnly).TrainingPolicyChangeRequestV2(trainingPolicyChangeRequestV2).Execute()

Update a training policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	trainingPolicyChangeRequestV2 := *openapiclient.NewTrainingPolicyChangeRequestV2() // TrainingPolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateTrainingPolicyV2(context.Background()).ValidateOnly(validateOnly).TrainingPolicyChangeRequestV2(trainingPolicyChangeRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateTrainingPolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateTrainingPolicyV2`: TrainingPolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateTrainingPolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateTrainingPolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **trainingPolicyChangeRequestV2** | [**TrainingPolicyChangeRequestV2**](TrainingPolicyChangeRequestV2.md) |  | 

### Return type

[**TrainingPolicyV2**](TrainingPolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateWorkspacePolicyV2

> WorkspacePolicyV2 UpdateWorkspacePolicyV2(ctx).ValidateOnly(validateOnly).WorkspacePolicyChangeRequestV2(workspacePolicyChangeRequestV2).Execute()

Update a workspace policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-common-packages/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	workspacePolicyChangeRequestV2 := *openapiclient.NewWorkspacePolicyChangeRequestV2() // WorkspacePolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateWorkspacePolicyV2(context.Background()).ValidateOnly(validateOnly).WorkspacePolicyChangeRequestV2(workspacePolicyChangeRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateWorkspacePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateWorkspacePolicyV2`: WorkspacePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateWorkspacePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateWorkspacePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **workspacePolicyChangeRequestV2** | [**WorkspacePolicyChangeRequestV2**](WorkspacePolicyChangeRequestV2.md) |  | 

### Return type

[**WorkspacePolicyV2**](WorkspacePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

