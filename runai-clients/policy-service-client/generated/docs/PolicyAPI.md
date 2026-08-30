# \PolicyAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApplyGvkPolicy**](PolicyAPI.md#ApplyGvkPolicy) | **Put** /api/v1/policies/{PolicyId} | Replace a policy. [experimental]
[**AssetChangedHook**](PolicyAPI.md#AssetChangedHook) | **Put** /internal/policy/asset/{AssetId} | Asset changed hook for asset service
[**CreateGvkPolicy**](PolicyAPI.md#CreateGvkPolicy) | **Post** /api/v1/policies | Create a policy. [experimental]
[**DeleteDistributedInferencePolicy**](PolicyAPI.md#DeleteDistributedInferencePolicy) | **Delete** /api/v2/policy/distributed-inferences | Delete a distributed inference policy.
[**DeleteDistributedPolicy**](PolicyAPI.md#DeleteDistributedPolicy) | **Delete** /api/v2/policy/distributed | Delete a distributed policy.
[**DeleteFallbackPolicy**](PolicyAPI.md#DeleteFallbackPolicy) | **Delete** /api/v2/policy/fallback | Delete a fallback policy.
[**DeleteGvkPolicy**](PolicyAPI.md#DeleteGvkPolicy) | **Delete** /api/v1/policies/{PolicyId} | Delete a policy. [experimental]
[**DeleteGvkPolicyRules**](PolicyAPI.md#DeleteGvkPolicyRules) | **Delete** /api/v1/policies/{PolicyId}/rules | Delete rules from a policy. [experimental]
[**DeleteInferencePolicy**](PolicyAPI.md#DeleteInferencePolicy) | **Delete** /api/v2/policy/inferences | Delete an inference policy.
[**DeleteNimServicePolicy**](PolicyAPI.md#DeleteNimServicePolicy) | **Delete** /api/v2/policy/nim-services | Delete NVIDIA NIM service policy.
[**DeleteTrainingPolicy**](PolicyAPI.md#DeleteTrainingPolicy) | **Delete** /api/v2/policy/trainings | Delete a training policy.
[**DeleteWorkspacePolicy**](PolicyAPI.md#DeleteWorkspacePolicy) | **Delete** /api/v2/policy/workspaces | Delete a workspace policy.
[**GetDistributedInferencePolicyV2**](PolicyAPI.md#GetDistributedInferencePolicyV2) | **Get** /api/v2/policy/distributed-inferences | Get a distributed inference policy.
[**GetDistributedPolicyV2**](PolicyAPI.md#GetDistributedPolicyV2) | **Get** /api/v2/policy/distributed | Get a distributed policy.
[**GetEffectivePolicy**](PolicyAPI.md#GetEffectivePolicy) | **Post** /api/v1/policies/effective | Get the effective policy by type and scope. The effective policy is a combination of the policy provided for the given scope (if exist) and the effective policies of its parent scopes.
[**GetFallbackPolicy**](PolicyAPI.md#GetFallbackPolicy) | **Get** /api/v2/policy/fallback | Get a fallback policy.
[**GetGvkPolicy**](PolicyAPI.md#GetGvkPolicy) | **Get** /api/v1/policies/{PolicyId} | Get a policy. [experimental]
[**GetGvkPolicyByTypeAndScope**](PolicyAPI.md#GetGvkPolicyByTypeAndScope) | **Post** /internal/policies | Get a policy by type and scope
[**GetInferencePolicyV2**](PolicyAPI.md#GetInferencePolicyV2) | **Get** /api/v2/policy/inferences | Get an inference policy.
[**GetMappingFieldInfo**](PolicyAPI.md#GetMappingFieldInfo) | **Post** /api/v1/policies/mappings/field-info | Get mapping field info. [experimental]
[**GetMappingMinimal**](PolicyAPI.md#GetMappingMinimal) | **Post** /api/v1/policies/mappings/minimal | Get minimal mapping info. [experimental]
[**GetNimServicePolicy**](PolicyAPI.md#GetNimServicePolicy) | **Get** /api/v2/policy/nim-services | Get NVIDIA NIM service policy.
[**GetPolicyCrossType**](PolicyAPI.md#GetPolicyCrossType) | **Get** /internal/policy/cross-type | Get policy cross workload types.
[**GetSyncConfig**](PolicyAPI.md#GetSyncConfig) | **Get** /api/v2/policy/sync-config/{ClusterId} | Get sync config for a given cluster.
[**GetTrainingPolicyV2**](PolicyAPI.md#GetTrainingPolicyV2) | **Get** /api/v2/policy/trainings | Get a training policy.
[**GetWorkspacePolicyV2**](PolicyAPI.md#GetWorkspacePolicyV2) | **Get** /api/v2/policy/workspaces | Get a workspace policy.
[**ListGvkPolicies**](PolicyAPI.md#ListGvkPolicies) | **Get** /api/v1/policies | List policies. [experimental]
[**ListPolicies**](PolicyAPI.md#ListPolicies) | **Get** /api/v2/policy | List policies
[**OverwriteDistributedInferencePolicyV2**](PolicyAPI.md#OverwriteDistributedInferencePolicyV2) | **Put** /api/v2/policy/distributed-inferences | Overwrite a distributed inference policy.
[**OverwriteDistributedPolicyV2**](PolicyAPI.md#OverwriteDistributedPolicyV2) | **Put** /api/v2/policy/distributed | Overwrite a distributed policy.
[**OverwriteFallbackPolicy**](PolicyAPI.md#OverwriteFallbackPolicy) | **Put** /api/v2/policy/fallback | Overwrite a fallback policy.
[**OverwriteInferencePolicyV2**](PolicyAPI.md#OverwriteInferencePolicyV2) | **Put** /api/v2/policy/inferences | Overwrite an inference policy.
[**OverwriteNimServicePolicy**](PolicyAPI.md#OverwriteNimServicePolicy) | **Put** /api/v2/policy/nim-services | Overwrite NVIDIA NIM service policy.
[**OverwriteTrainingPolicyV2**](PolicyAPI.md#OverwriteTrainingPolicyV2) | **Put** /api/v2/policy/trainings | Overwrite a training policy.
[**OverwriteWorkspacePolicyV2**](PolicyAPI.md#OverwriteWorkspacePolicyV2) | **Put** /api/v2/policy/workspaces | Overwrite a workspace policy.
[**PatchGvkPolicy**](PolicyAPI.md#PatchGvkPolicy) | **Patch** /api/v1/policies/{PolicyId} | Patch a policy. [experimental]
[**UpdateDistributedInferencePolicyV2**](PolicyAPI.md#UpdateDistributedInferencePolicyV2) | **Patch** /api/v2/policy/distributed-inferences | Update a distributed inference policy.
[**UpdateDistributedPolicyV2**](PolicyAPI.md#UpdateDistributedPolicyV2) | **Patch** /api/v2/policy/distributed | Update a distributed policy.
[**UpdateFallbackPolicy**](PolicyAPI.md#UpdateFallbackPolicy) | **Patch** /api/v2/policy/fallback | Update workloads fallback policy.
[**UpdateInferencePolicyV2**](PolicyAPI.md#UpdateInferencePolicyV2) | **Patch** /api/v2/policy/inferences | Update an inference policy.
[**UpdateNimServicePolicy**](PolicyAPI.md#UpdateNimServicePolicy) | **Patch** /api/v2/policy/nim-services | Update NVIDIA NIM service policy.
[**UpdateSyncConfig**](PolicyAPI.md#UpdateSyncConfig) | **Patch** /api/v2/policy/sync-config | update sync configuration for a given cluster
[**UpdateTrainingPolicyV2**](PolicyAPI.md#UpdateTrainingPolicyV2) | **Patch** /api/v2/policy/trainings | Update a training policy.
[**UpdateWorkspacePolicyV2**](PolicyAPI.md#UpdateWorkspacePolicyV2) | **Patch** /api/v2/policy/workspaces | Update a workspace policy.



## ApplyGvkPolicy

> Policy ApplyGvkPolicy(ctx, policyId).UpdateGvkPolicyReq(updateGvkPolicyReq).Execute()

Replace a policy. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier for the policy.
	updateGvkPolicyReq := *openapiclient.NewUpdateGvkPolicyReq() // UpdateGvkPolicyReq |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.ApplyGvkPolicy(context.Background(), policyId).UpdateGvkPolicyReq(updateGvkPolicyReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.ApplyGvkPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ApplyGvkPolicy`: Policy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.ApplyGvkPolicy`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** | The unique identifier for the policy. | 

### Other Parameters

Other parameters are passed through a pointer to a apiApplyGvkPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateGvkPolicyReq** | [**UpdateGvkPolicyReq**](UpdateGvkPolicyReq.md) |  | 

### Return type

[**Policy**](Policy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


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
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
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


## CreateGvkPolicy

> Policy CreateGvkPolicy(ctx).CreateGvkPolicyReq(createGvkPolicyReq).Execute()

Create a policy. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	createGvkPolicyReq := *openapiclient.NewCreateGvkPolicyReq(*openapiclient.NewGVK("v1", "Deployment"), openapiclient.ScopeType("system")) // CreateGvkPolicyReq |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.CreateGvkPolicy(context.Background()).CreateGvkPolicyReq(createGvkPolicyReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.CreateGvkPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateGvkPolicy`: Policy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.CreateGvkPolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiCreateGvkPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createGvkPolicyReq** | [**CreateGvkPolicyReq**](CreateGvkPolicyReq.md) |  | 

### Return type

[**Policy**](Policy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteDistributedInferencePolicy

> DeleteDistributedInferencePolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()

Delete a distributed inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteDistributedInferencePolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteDistributedInferencePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDistributedInferencePolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 

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


## DeleteDistributedPolicy

> DeleteDistributedPolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()

Delete a distributed policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteDistributedPolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 

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


## DeleteFallbackPolicy

> DeleteFallbackPolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()

Delete a fallback policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteFallbackPolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteFallbackPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteFallbackPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 

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


## DeleteGvkPolicy

> DeleteGvkPolicy(ctx, policyId).Execute()

Delete a policy. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier for the policy.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteGvkPolicy(context.Background(), policyId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteGvkPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** | The unique identifier for the policy. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteGvkPolicyRequest struct via the builder pattern


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


## DeleteGvkPolicyRules

> Policy DeleteGvkPolicyRules(ctx, policyId).DeleteGvkPolicyRulesReq(deleteGvkPolicyRulesReq).Execute()

Delete rules from a policy. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier for the policy.
	deleteGvkPolicyRulesReq := *openapiclient.NewDeleteGvkPolicyRulesReq([]openapiclient.PolicyRuleRef{*openapiclient.NewPolicyRuleRef("containerPort", "containerByName")}) // DeleteGvkPolicyRulesReq |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.DeleteGvkPolicyRules(context.Background(), policyId).DeleteGvkPolicyRulesReq(deleteGvkPolicyRulesReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteGvkPolicyRules``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeleteGvkPolicyRules`: Policy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.DeleteGvkPolicyRules`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** | The unique identifier for the policy. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteGvkPolicyRulesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **deleteGvkPolicyRulesReq** | [**DeleteGvkPolicyRulesReq**](DeleteGvkPolicyRulesReq.md) |  | 

### Return type

[**Policy**](Policy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteInferencePolicy

> DeleteInferencePolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()

Delete an inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteInferencePolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 

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


## DeleteNimServicePolicy

> DeleteNimServicePolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()

Delete NVIDIA NIM service policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteNimServicePolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.DeleteNimServicePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeleteNimServicePolicyRequest struct via the builder pattern


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

> DeleteTrainingPolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()

Delete a training policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteTrainingPolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 

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

> DeleteWorkspacePolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()

Delete a workspace policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PolicyAPI.DeleteWorkspacePolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).ApproveClusterDeletion(approveClusterDeletion).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 

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


## GetDistributedInferencePolicyV2

> DistributedInferencePolicyV2 GetDistributedInferencePolicyV2(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get a distributed inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetDistributedInferencePolicyV2(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetDistributedInferencePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetDistributedInferencePolicyV2`: DistributedInferencePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetDistributedInferencePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetDistributedInferencePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**DistributedInferencePolicyV2**](DistributedInferencePolicyV2.md)

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
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
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


## GetEffectivePolicy

> EffectiveRules GetEffectivePolicy(ctx).WorkloadTypeAndScope(workloadTypeAndScope).Execute()

Get the effective policy by type and scope. The effective policy is a combination of the policy provided for the given scope (if exist) and the effective policies of its parent scopes.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	workloadTypeAndScope := *openapiclient.NewWorkloadTypeAndScope(*openapiclient.NewGVK("v1", "Deployment"), openapiclient.ScopeType("system")) // WorkloadTypeAndScope | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetEffectivePolicy(context.Background()).WorkloadTypeAndScope(workloadTypeAndScope).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetEffectivePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetEffectivePolicy`: EffectiveRules
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetEffectivePolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetEffectivePolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadTypeAndScope** | [**WorkloadTypeAndScope**](WorkloadTypeAndScope.md) |  | 

### Return type

[**EffectiveRules**](EffectiveRules.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetFallbackPolicy

> FallbackPolicy GetFallbackPolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get a fallback policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetFallbackPolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetFallbackPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetFallbackPolicy`: FallbackPolicy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetFallbackPolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetFallbackPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**FallbackPolicy**](FallbackPolicy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGvkPolicy

> GetGvkPolicyResp GetGvkPolicy(ctx, policyId).Breakdown(breakdown).Execute()

Get a policy. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier for the policy.
	breakdown := true // bool | When true, includes a breakdown object with resolved fields and spec selectors from the Mapping subsystem. Most callers should leave this off. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetGvkPolicy(context.Background(), policyId).Breakdown(breakdown).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetGvkPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGvkPolicy`: GetGvkPolicyResp
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetGvkPolicy`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** | The unique identifier for the policy. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetGvkPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **breakdown** | **bool** | When true, includes a breakdown object with resolved fields and spec selectors from the Mapping subsystem. Most callers should leave this off. | 

### Return type

[**GetGvkPolicyResp**](GetGvkPolicyResp.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetGvkPolicyByTypeAndScope

> GetGvkPolicyResp GetGvkPolicyByTypeAndScope(ctx).WorkloadTypeAndScope(workloadTypeAndScope).Execute()

Get a policy by type and scope

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	workloadTypeAndScope := *openapiclient.NewWorkloadTypeAndScope(*openapiclient.NewGVK("v1", "Deployment"), openapiclient.ScopeType("system")) // WorkloadTypeAndScope |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetGvkPolicyByTypeAndScope(context.Background()).WorkloadTypeAndScope(workloadTypeAndScope).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetGvkPolicyByTypeAndScope``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetGvkPolicyByTypeAndScope`: GetGvkPolicyResp
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetGvkPolicyByTypeAndScope`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetGvkPolicyByTypeAndScopeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workloadTypeAndScope** | [**WorkloadTypeAndScope**](WorkloadTypeAndScope.md) |  | 

### Return type

[**GetGvkPolicyResp**](GetGvkPolicyResp.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
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
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
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


## GetMappingFieldInfo

> Breakdown GetMappingFieldInfo(ctx).GetMappingFieldInfoReq(getMappingFieldInfoReq).Execute()

Get mapping field info. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	getMappingFieldInfoReq := *openapiclient.NewGetMappingFieldInfoReq() // GetMappingFieldInfoReq |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetMappingFieldInfo(context.Background()).GetMappingFieldInfoReq(getMappingFieldInfoReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetMappingFieldInfo``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMappingFieldInfo`: Breakdown
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetMappingFieldInfo`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetMappingFieldInfoRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **getMappingFieldInfoReq** | [**GetMappingFieldInfoReq**](GetMappingFieldInfoReq.md) |  | 

### Return type

[**Breakdown**](Breakdown.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetMappingMinimal

> GetMappingMinimalResp GetMappingMinimal(ctx).GetMappingMinimalInfoReq(getMappingMinimalInfoReq).Execute()

Get minimal mapping info. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	getMappingMinimalInfoReq := *openapiclient.NewGetMappingMinimalInfoReq(*openapiclient.NewGVK("v1", "Deployment")) // GetMappingMinimalInfoReq |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetMappingMinimal(context.Background()).GetMappingMinimalInfoReq(getMappingMinimalInfoReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetMappingMinimal``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetMappingMinimal`: GetMappingMinimalResp
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetMappingMinimal`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetMappingMinimalRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **getMappingMinimalInfoReq** | [**GetMappingMinimalInfoReq**](GetMappingMinimalInfoReq.md) |  | 

### Return type

[**GetMappingMinimalResp**](GetMappingMinimalResp.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetNimServicePolicy

> NimServicePolicy GetNimServicePolicy(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get NVIDIA NIM service policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetNimServicePolicy(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetNimServicePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetNimServicePolicy`: NimServicePolicy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetNimServicePolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetNimServicePolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**NimServicePolicy**](NimServicePolicy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPolicyCrossType

> PolicyCrossType GetPolicyCrossType(ctx).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()

Get policy cross workload types.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	scope := "scope_example" // string | The scope that the policy relates to.
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	noCreationMode := true // bool | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetPolicyCrossType(context.Background()).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).NoCreationMode(noCreationMode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetPolicyCrossType``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetPolicyCrossType`: PolicyCrossType
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetPolicyCrossType`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetPolicyCrossTypeRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scope** | **string** | The scope that the policy relates to. | 
 **departmentId** | **string** | Filter using the department id. | 
 **projectId** | **string** | project id to filter by | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **noCreationMode** | **bool** | Indicates that the policy should be returned only if the tenant has at least one given policy, otherwise it should not be created and instead return error 404. This is done for the frontend which needs the policy info also for tenants that are not using policy | 

### Return type

[**PolicyCrossType**](PolicyCrossType.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetSyncConfig

> SyncConfig GetSyncConfig(ctx, clusterId).Execute()

Get sync config for a given cluster.

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	clusterId := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.GetSyncConfig(context.Background(), clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.GetSyncConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GetSyncConfig`: SyncConfig
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.GetSyncConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterId** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetSyncConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**SyncConfig**](SyncConfig.md)

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
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
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
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
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


## ListGvkPolicies

> ListGvkPoliciesResp ListGvkPolicies(ctx).FilterBy(filterBy).SortBy(sortBy).Offset(offset).Limit(limit).SortOrder(sortOrder).Execute()

List policies. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	filterBy := []string{"Inner_example"} // []string | Filter results by a parameter. Use the format field-name operator value. Filterable fields are `id`, `type` and `scope`. Operators are `==` Equals, `!=` Not equals, `<=` Less than or equal, `>=` Greater than or equal, `=@` contains, `!@` Does not contain, `=^` Starts with and `=$` Ends with. Dates are in ISO 8601 timestamp format and available for operators `==`, `!=`, `<=` and `>=`. (optional)
	sortBy := "sortBy_example" // string | The field to sort the results by. (optional)
	offset := int32(100) // int32 | The offset of the first item returned in the collection. (optional)
	limit := int32(56) // int32 | The maximum number of entries to return. (optional) (default to 50)
	sortOrder := "sortOrder_example" // string | Sort results in descending or ascending order. (optional) (default to "asc")

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.ListGvkPolicies(context.Background()).FilterBy(filterBy).SortBy(sortBy).Offset(offset).Limit(limit).SortOrder(sortOrder).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.ListGvkPolicies``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListGvkPolicies`: ListGvkPoliciesResp
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.ListGvkPolicies`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListGvkPoliciesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **filterBy** | **[]string** | Filter results by a parameter. Use the format field-name operator value. Filterable fields are &#x60;id&#x60;, &#x60;type&#x60; and &#x60;scope&#x60;. Operators are &#x60;&#x3D;&#x3D;&#x60; Equals, &#x60;!&#x3D;&#x60; Not equals, &#x60;&lt;&#x3D;&#x60; Less than or equal, &#x60;&gt;&#x3D;&#x60; Greater than or equal, &#x60;&#x3D;@&#x60; contains, &#x60;!@&#x60; Does not contain, &#x60;&#x3D;^&#x60; Starts with and &#x60;&#x3D;$&#x60; Ends with. Dates are in ISO 8601 timestamp format and available for operators &#x60;&#x3D;&#x3D;&#x60;, &#x60;!&#x3D;&#x60;, &#x60;&lt;&#x3D;&#x60; and &#x60;&gt;&#x3D;&#x60;. | 
 **sortBy** | **string** | The field to sort the results by. | 
 **offset** | **int32** | The offset of the first item returned in the collection. | 
 **limit** | **int32** | The maximum number of entries to return. | [default to 50]
 **sortOrder** | **string** | Sort results in descending or ascending order. | [default to &quot;asc&quot;]

### Return type

[**ListGvkPoliciesResp**](ListGvkPoliciesResp.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPolicies

> PolicyListResponse ListPolicies(ctx).WorkloadType(workloadType).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).IncludeFallbackPolicies(includeFallbackPolicies).IncludeGvkPolicies(includeGvkPolicies).Execute()

List policies



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	workloadType := "workloadType_example" // string | Policy for a specific workload type. (optional)
	scope := "scope_example" // string | filter by this scope. (optional)
	departmentId := "1" // string | Filter using the department id. (optional)
	projectId := "1" // string | project id to filter by (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	includeFallbackPolicies := true // bool | whether to include fallback policies in the list. Default to false. (optional)
	includeGvkPolicies := true // bool | whether to include GVK-based policies in the list. Default to false. See /api/v1/policies for more details. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.ListPolicies(context.Background()).WorkloadType(workloadType).Scope(scope).DepartmentId(departmentId).ProjectId(projectId).ClusterId(clusterId).IncludeFallbackPolicies(includeFallbackPolicies).IncludeGvkPolicies(includeGvkPolicies).Execute()
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
 **includeFallbackPolicies** | **bool** | whether to include fallback policies in the list. Default to false. | 
 **includeGvkPolicies** | **bool** | whether to include GVK-based policies in the list. Default to false. See /api/v1/policies for more details. | 

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


## OverwriteDistributedInferencePolicyV2

> DistributedInferencePolicyV2 OverwriteDistributedInferencePolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedInferencePolicyOverwriteRequestV2(distributedInferencePolicyOverwriteRequestV2).Execute()

Overwrite a distributed inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	distributedInferencePolicyOverwriteRequestV2 := *openapiclient.NewDistributedInferencePolicyOverwriteRequestV2() // DistributedInferencePolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteDistributedInferencePolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedInferencePolicyOverwriteRequestV2(distributedInferencePolicyOverwriteRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.OverwriteDistributedInferencePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OverwriteDistributedInferencePolicyV2`: DistributedInferencePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.OverwriteDistributedInferencePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOverwriteDistributedInferencePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
 **distributedInferencePolicyOverwriteRequestV2** | [**DistributedInferencePolicyOverwriteRequestV2**](DistributedInferencePolicyOverwriteRequestV2.md) |  | 

### Return type

[**DistributedInferencePolicyV2**](DistributedInferencePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OverwriteDistributedPolicyV2

> DistributedPolicyV2 OverwriteDistributedPolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedPolicyOverwriteRequestV2(distributedPolicyOverwriteRequestV2).Execute()

Overwrite a distributed policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	distributedPolicyOverwriteRequestV2 := *openapiclient.NewDistributedPolicyOverwriteRequestV2() // DistributedPolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteDistributedPolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedPolicyOverwriteRequestV2(distributedPolicyOverwriteRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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


## OverwriteFallbackPolicy

> FallbackPolicy OverwriteFallbackPolicy(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).FallbackPolicyOverwriteRequest(fallbackPolicyOverwriteRequest).Execute()

Overwrite a fallback policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	fallbackPolicyOverwriteRequest := *openapiclient.NewFallbackPolicyOverwriteRequest() // FallbackPolicyOverwriteRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteFallbackPolicy(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).FallbackPolicyOverwriteRequest(fallbackPolicyOverwriteRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.OverwriteFallbackPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OverwriteFallbackPolicy`: FallbackPolicy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.OverwriteFallbackPolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOverwriteFallbackPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
 **fallbackPolicyOverwriteRequest** | [**FallbackPolicyOverwriteRequest**](FallbackPolicyOverwriteRequest.md) |  | 

### Return type

[**FallbackPolicy**](FallbackPolicy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OverwriteInferencePolicyV2

> InferencePolicyV2 OverwriteInferencePolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).InferencePolicyOverwriteRequestV2(inferencePolicyOverwriteRequestV2).Execute()

Overwrite an inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	inferencePolicyOverwriteRequestV2 := *openapiclient.NewInferencePolicyOverwriteRequestV2() // InferencePolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteInferencePolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).InferencePolicyOverwriteRequestV2(inferencePolicyOverwriteRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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


## OverwriteNimServicePolicy

> NimServicePolicy OverwriteNimServicePolicy(ctx).ValidateOnly(validateOnly).NimServicePolicyOverwriteRequest(nimServicePolicyOverwriteRequest).Execute()

Overwrite NVIDIA NIM service policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	nimServicePolicyOverwriteRequest := *openapiclient.NewNimServicePolicyOverwriteRequest() // NimServicePolicyOverwriteRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteNimServicePolicy(context.Background()).ValidateOnly(validateOnly).NimServicePolicyOverwriteRequest(nimServicePolicyOverwriteRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.OverwriteNimServicePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `OverwriteNimServicePolicy`: NimServicePolicy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.OverwriteNimServicePolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOverwriteNimServicePolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **nimServicePolicyOverwriteRequest** | [**NimServicePolicyOverwriteRequest**](NimServicePolicyOverwriteRequest.md) |  | 

### Return type

[**NimServicePolicy**](NimServicePolicy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OverwriteTrainingPolicyV2

> TrainingPolicyV2 OverwriteTrainingPolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).TrainingPolicyOverwriteRequestV2(trainingPolicyOverwriteRequestV2).Execute()

Overwrite a training policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	trainingPolicyOverwriteRequestV2 := *openapiclient.NewTrainingPolicyOverwriteRequestV2() // TrainingPolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteTrainingPolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).TrainingPolicyOverwriteRequestV2(trainingPolicyOverwriteRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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

> WorkspacePolicyV2 OverwriteWorkspacePolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).WorkspacePolicyOverwriteRequestV2(workspacePolicyOverwriteRequestV2).Execute()

Overwrite a workspace policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	workspacePolicyOverwriteRequestV2 := *openapiclient.NewWorkspacePolicyOverwriteRequestV2() // WorkspacePolicyOverwriteRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.OverwriteWorkspacePolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).WorkspacePolicyOverwriteRequestV2(workspacePolicyOverwriteRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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


## PatchGvkPolicy

> Policy PatchGvkPolicy(ctx, policyId).UpdateGvkPolicyReq(updateGvkPolicyReq).Execute()

Patch a policy. [experimental]



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	policyId := "38400000-8cf0-11bd-b23e-10b96e4ef00d" // string | The unique identifier for the policy.
	updateGvkPolicyReq := *openapiclient.NewUpdateGvkPolicyReq() // UpdateGvkPolicyReq |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.PatchGvkPolicy(context.Background(), policyId).UpdateGvkPolicyReq(updateGvkPolicyReq).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.PatchGvkPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `PatchGvkPolicy`: Policy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.PatchGvkPolicy`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**policyId** | **string** | The unique identifier for the policy. | 

### Other Parameters

Other parameters are passed through a pointer to a apiPatchGvkPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **updateGvkPolicyReq** | [**UpdateGvkPolicyReq**](UpdateGvkPolicyReq.md) |  | 

### Return type

[**Policy**](Policy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDistributedInferencePolicyV2

> DistributedInferencePolicyV2 UpdateDistributedInferencePolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedInferencePolicyChangeRequestV2(distributedInferencePolicyChangeRequestV2).Execute()

Update a distributed inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	distributedInferencePolicyChangeRequestV2 := *openapiclient.NewDistributedInferencePolicyChangeRequestV2() // DistributedInferencePolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateDistributedInferencePolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedInferencePolicyChangeRequestV2(distributedInferencePolicyChangeRequestV2).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateDistributedInferencePolicyV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateDistributedInferencePolicyV2`: DistributedInferencePolicyV2
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateDistributedInferencePolicyV2`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateDistributedInferencePolicyV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
 **distributedInferencePolicyChangeRequestV2** | [**DistributedInferencePolicyChangeRequestV2**](DistributedInferencePolicyChangeRequestV2.md) |  | 

### Return type

[**DistributedInferencePolicyV2**](DistributedInferencePolicyV2.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateDistributedPolicyV2

> DistributedPolicyV2 UpdateDistributedPolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedPolicyChangeRequestV2(distributedPolicyChangeRequestV2).Execute()

Update a distributed policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	distributedPolicyChangeRequestV2 := *openapiclient.NewDistributedPolicyChangeRequestV2() // DistributedPolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateDistributedPolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).DistributedPolicyChangeRequestV2(distributedPolicyChangeRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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


## UpdateFallbackPolicy

> FallbackPolicy UpdateFallbackPolicy(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).FallbackPolicyChangeRequest(fallbackPolicyChangeRequest).Execute()

Update workloads fallback policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	fallbackPolicyChangeRequest := *openapiclient.NewFallbackPolicyChangeRequest() // FallbackPolicyChangeRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateFallbackPolicy(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).FallbackPolicyChangeRequest(fallbackPolicyChangeRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateFallbackPolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateFallbackPolicy`: FallbackPolicy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateFallbackPolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateFallbackPolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
 **fallbackPolicyChangeRequest** | [**FallbackPolicyChangeRequest**](FallbackPolicyChangeRequest.md) |  | 

### Return type

[**FallbackPolicy**](FallbackPolicy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateInferencePolicyV2

> InferencePolicyV2 UpdateInferencePolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).InferencePolicyChangeRequestV2(inferencePolicyChangeRequestV2).Execute()

Update an inference policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	inferencePolicyChangeRequestV2 := *openapiclient.NewInferencePolicyChangeRequestV2() // InferencePolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateInferencePolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).InferencePolicyChangeRequestV2(inferencePolicyChangeRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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


## UpdateNimServicePolicy

> NimServicePolicy UpdateNimServicePolicy(ctx).ValidateOnly(validateOnly).NimServicePolicyChangeRequest(nimServicePolicyChangeRequest).Execute()

Update NVIDIA NIM service policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	nimServicePolicyChangeRequest := *openapiclient.NewNimServicePolicyChangeRequest() // NimServicePolicyChangeRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateNimServicePolicy(context.Background()).ValidateOnly(validateOnly).NimServicePolicyChangeRequest(nimServicePolicyChangeRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateNimServicePolicy``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateNimServicePolicy`: NimServicePolicy
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateNimServicePolicy`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateNimServicePolicyRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **validateOnly** | **bool** | Validate the given policy payload without applying it | 
 **nimServicePolicyChangeRequest** | [**NimServicePolicyChangeRequest**](NimServicePolicyChangeRequest.md) |  | 

### Return type

[**NimServicePolicy**](NimServicePolicy.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateSyncConfig

> SyncConfig UpdateSyncConfig(ctx).SyncConfig(syncConfig).Execute()

update sync configuration for a given cluster

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	syncConfig := *openapiclient.NewSyncConfig() // SyncConfig |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateSyncConfig(context.Background()).SyncConfig(syncConfig).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PolicyAPI.UpdateSyncConfig``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UpdateSyncConfig`: SyncConfig
	fmt.Fprintf(os.Stdout, "Response from `PolicyAPI.UpdateSyncConfig`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUpdateSyncConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **syncConfig** | [**SyncConfig**](SyncConfig.md) |  | 

### Return type

[**SyncConfig**](SyncConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateTrainingPolicyV2

> TrainingPolicyV2 UpdateTrainingPolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).TrainingPolicyChangeRequestV2(trainingPolicyChangeRequestV2).Execute()

Update a training policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	trainingPolicyChangeRequestV2 := *openapiclient.NewTrainingPolicyChangeRequestV2() // TrainingPolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateTrainingPolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).TrainingPolicyChangeRequestV2(trainingPolicyChangeRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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

> WorkspacePolicyV2 UpdateWorkspacePolicyV2(ctx).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).WorkspacePolicyChangeRequestV2(workspacePolicyChangeRequestV2).Execute()

Update a workspace policy.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/runai-clients/policy-service-client"
)

func main() {
	validateOnly := true // bool | Validate the given policy payload without applying it (optional)
	approveClusterDeletion := true // bool | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. (optional)
	workspacePolicyChangeRequestV2 := *openapiclient.NewWorkspacePolicyChangeRequestV2() // WorkspacePolicyChangeRequestV2 |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PolicyAPI.UpdateWorkspacePolicyV2(context.Background()).ValidateOnly(validateOnly).ApproveClusterDeletion(approveClusterDeletion).WorkspacePolicyChangeRequestV2(workspacePolicyChangeRequestV2).Execute()
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
 **approveClusterDeletion** | **bool** | Required when modifying or deleting a policy if the tenant includes one or more clusters running version 2.20 or earlier. In such cases, the platform must delete the policy from the affected clusters. This flag indicates that the administrator approves this deletion. | 
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

