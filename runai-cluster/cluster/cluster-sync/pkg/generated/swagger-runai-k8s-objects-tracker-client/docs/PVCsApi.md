# \PVCsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateClusterWidePvcV1**](PVCsAPI.md#CreateClusterWidePvcV1) | **Post** /api/v1/pvcs/{clusterUuid}/{pvcName} | upload the name of cluster-wide pvc to the control plane
[**CreateClusterWidePvcV2**](PVCsAPI.md#CreateClusterWidePvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/{pvcName} | upload the details of cluster-wide pvc to the control plane
[**CreateDepartmentWidePvcV1**](PVCsAPI.md#CreateDepartmentWidePvcV1) | **Post** /api/v1/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | upload the name of department-wide pvc to the control plane
[**CreateDepartmentWidePvcV2**](PVCsAPI.md#CreateDepartmentWidePvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | upload the details of department-wide pvc to the control plane
[**CreateDepartmentsWidePvcV2**](PVCsAPI.md#CreateDepartmentsWidePvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/departments/{departmentId}/{pvcName} | upload the details of department-wide pvc to the control plane
[**CreateProjectPvcV1**](PVCsAPI.md#CreateProjectPvcV1) | **Post** /api/v1/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | upload the name of project pvc to the control plane
[**CreateProjectPvcV2**](PVCsAPI.md#CreateProjectPvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | upload the details of project pvc to the control plane
[**CreateProjectsPvcV2**](PVCsAPI.md#CreateProjectsPvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespaces/{namespace}/{pvcName} | upload the details of project pvc to the control plane
[**DeleteClusterWidePvcV1**](PVCsAPI.md#DeleteClusterWidePvcV1) | **Delete** /api/v1/pvcs/{clusterUuid}/{pvcName} | delete cluster-wide PVC
[**DeleteClusterWidePvcV2**](PVCsAPI.md#DeleteClusterWidePvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/{pvcName} | delete cluster-wide PVC
[**DeleteDepartmentWidePvcV1**](PVCsAPI.md#DeleteDepartmentWidePvcV1) | **Delete** /api/v1/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | delete department-wide PVC
[**DeleteDepartmentWidePvcV2**](PVCsAPI.md#DeleteDepartmentWidePvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | delete department-wide PVC
[**DeleteDepartmentsWidePvcV2**](PVCsAPI.md#DeleteDepartmentsWidePvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/departments/{departmentId}/{pvcName} | delete department-wide PVC
[**DeleteProjectPvcV1**](PVCsAPI.md#DeleteProjectPvcV1) | **Delete** /api/v1/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | delete project PVC
[**DeleteProjectPvcV2**](PVCsAPI.md#DeleteProjectPvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | delete project PVC
[**DeleteProjectsPvcV2**](PVCsAPI.md#DeleteProjectsPvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/namespaces/{namespace}/{pvcName} | delete project PVC
[**ListPvcs**](PVCsAPI.md#ListPvcs) | **Get** /api/v1/pvcs | list details of pvcs
[**ObsoleteCreatePvc**](PVCsAPI.md#ObsoleteCreatePvc) | **Post** /v1/k8s/clusters/{clusterUuid}/{namespace}/pvcs | upload the details of PVC to the control plane
[**ObsoleteDeletePvc**](PVCsAPI.md#ObsoleteDeletePvc) | **Delete** /v1/k8s/clusters/{clusterUuid}/{namespace}/pvcs/{pvcName} | delete PVC
[**ObsoleteSyncPvc**](PVCsAPI.md#ObsoleteSyncPvc) | **Post** /v1/k8s/clusters/{clusterUuid}/{namespace}/pvcs/sync | upload the details of PVCs to the control plane
[**SyncClusterWidePvcNamesV1**](PVCsAPI.md#SyncClusterWidePvcNamesV1) | **Post** /api/v1/pvcs/{clusterUuid} | sync the names of cluster-wide pvcs to the control plane.
[**SyncClusterWidePvcsV2**](PVCsAPI.md#SyncClusterWidePvcsV2) | **Post** /api/v2/pvcs/{clusterUuid} | sync the details of cluster-wide pvcs to the control plane
[**SyncDepartmentWidePvcNamesV1**](PVCsAPI.md#SyncDepartmentWidePvcNamesV1) | **Post** /api/v1/pvcs/{clusterUuid}/department/{departmentId} | sync the names of department-wide pvcs to the control plane
[**SyncDepartmentWidePvcsV2**](PVCsAPI.md#SyncDepartmentWidePvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/department/{departmentId} | sync the details of department-wide pvcs to the control plane
[**SyncDepartmentsWidePvcsV2**](PVCsAPI.md#SyncDepartmentsWidePvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/departments/{departmentId} | sync the details of department-wide pvcs to the control plane
[**SyncProjectPvcNamesV1**](PVCsAPI.md#SyncProjectPvcNamesV1) | **Post** /api/v1/pvcs/{clusterUuid}/namespace/{namespace} | sync the names of project pvcs to the control plane
[**SyncProjectPvcsV2**](PVCsAPI.md#SyncProjectPvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespace/{namespace} | sync the details of project pvcs to the control plane
[**SyncProjectsPvcsV2**](PVCsAPI.md#SyncProjectsPvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespaces/{namespace} | sync the details of project pvcs to the control plane



## CreateClusterWidePvcV1

> string CreateClusterWidePvcV1(ctx, clusterUuid, pvcName).CreateClusterWidePvcV1Request(createClusterWidePvcV1Request).Execute()

upload the name of cluster-wide pvc to the control plane



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	pvcName := "pvcName_example" // string | The pvc name.
	createClusterWidePvcV1Request := *openapiclient.NewCreateClusterWidePvcV1Request() // CreateClusterWidePvcV1Request |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateClusterWidePvcV1(context.Background(), clusterUuid, pvcName).CreateClusterWidePvcV1Request(createClusterWidePvcV1Request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateClusterWidePvcV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateClusterWidePvcV1`: string
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateClusterWidePvcV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateClusterWidePvcV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **createClusterWidePvcV1Request** | [**CreateClusterWidePvcV1Request**](CreateClusterWidePvcV1Request.md) |  | 

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateClusterWidePvcV2

> PVC CreateClusterWidePvcV2(ctx, clusterUuid, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()

upload the details of cluster-wide pvc to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	pvcName := "pvcName_example" // string | The pvc name.
	pVCCreationRequest := *openapiclient.NewPVCCreationRequest() // PVCCreationRequest |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateClusterWidePvcV2(context.Background(), clusterUuid, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateClusterWidePvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateClusterWidePvcV2`: PVC
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateClusterWidePvcV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateClusterWidePvcV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pVCCreationRequest** | [**PVCCreationRequest**](PVCCreationRequest.md) |  | 

### Return type

[**PVC**](PVC.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentWidePvcV1

> string CreateDepartmentWidePvcV1(ctx, clusterUuid, departmentId, pvcName).CreateClusterWidePvcV1Request(createClusterWidePvcV1Request).Execute()

upload the name of department-wide pvc to the control plane



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pvcName := "pvcName_example" // string | The pvc name.
	createClusterWidePvcV1Request := *openapiclient.NewCreateClusterWidePvcV1Request() // CreateClusterWidePvcV1Request |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateDepartmentWidePvcV1(context.Background(), clusterUuid, departmentId, pvcName).CreateClusterWidePvcV1Request(createClusterWidePvcV1Request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateDepartmentWidePvcV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDepartmentWidePvcV1`: string
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateDepartmentWidePvcV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentWidePvcV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **createClusterWidePvcV1Request** | [**CreateClusterWidePvcV1Request**](CreateClusterWidePvcV1Request.md) |  | 

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentWidePvcV2

> PVC CreateDepartmentWidePvcV2(ctx, clusterUuid, departmentId, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()

upload the details of department-wide pvc to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pvcName := "pvcName_example" // string | The pvc name.
	pVCCreationRequest := *openapiclient.NewPVCCreationRequest() // PVCCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateDepartmentWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateDepartmentWidePvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDepartmentWidePvcV2`: PVC
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateDepartmentWidePvcV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentWidePvcV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **pVCCreationRequest** | [**PVCCreationRequest**](PVCCreationRequest.md) |  | 

### Return type

[**PVC**](PVC.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentsWidePvcV2

> PVC CreateDepartmentsWidePvcV2(ctx, clusterUuid, departmentId, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()

upload the details of department-wide pvc to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pvcName := "pvcName_example" // string | The pvc name.
	pVCCreationRequest := *openapiclient.NewPVCCreationRequest() // PVCCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateDepartmentsWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateDepartmentsWidePvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDepartmentsWidePvcV2`: PVC
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateDepartmentsWidePvcV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentsWidePvcV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **pVCCreationRequest** | [**PVCCreationRequest**](PVCCreationRequest.md) |  | 

### Return type

[**PVC**](PVC.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectPvcV1

> string CreateProjectPvcV1(ctx, clusterUuid, namespace, pvcName).CreateClusterWidePvcV1Request(createClusterWidePvcV1Request).Execute()

upload the name of project pvc to the control plane



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pvcName := "pvcName_example" // string | The pvc name.
	createClusterWidePvcV1Request := *openapiclient.NewCreateClusterWidePvcV1Request() // CreateClusterWidePvcV1Request |  (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateProjectPvcV1(context.Background(), clusterUuid, namespace, pvcName).CreateClusterWidePvcV1Request(createClusterWidePvcV1Request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateProjectPvcV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateProjectPvcV1`: string
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateProjectPvcV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectPvcV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **createClusterWidePvcV1Request** | [**CreateClusterWidePvcV1Request**](CreateClusterWidePvcV1Request.md) |  | 

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectPvcV2

> PVC CreateProjectPvcV2(ctx, clusterUuid, namespace, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()

upload the details of project pvc to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pvcName := "pvcName_example" // string | The pvc name.
	pVCCreationRequest := *openapiclient.NewPVCCreationRequest() // PVCCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateProjectPvcV2(context.Background(), clusterUuid, namespace, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateProjectPvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateProjectPvcV2`: PVC
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateProjectPvcV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectPvcV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **pVCCreationRequest** | [**PVCCreationRequest**](PVCCreationRequest.md) |  | 

### Return type

[**PVC**](PVC.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectsPvcV2

> PVC CreateProjectsPvcV2(ctx, clusterUuid, namespace, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()

upload the details of project pvc to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pvcName := "pvcName_example" // string | The pvc name.
	pVCCreationRequest := *openapiclient.NewPVCCreationRequest() // PVCCreationRequest | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.CreateProjectsPvcV2(context.Background(), clusterUuid, namespace, pvcName).PVCCreationRequest(pVCCreationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.CreateProjectsPvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateProjectsPvcV2`: PVC
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.CreateProjectsPvcV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectsPvcV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **pVCCreationRequest** | [**PVCCreationRequest**](PVCCreationRequest.md) |  | 

### Return type

[**PVC**](PVC.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteClusterWidePvcV1

> DeleteClusterWidePvcV1(ctx, clusterUuid, pvcName).Execute()

delete cluster-wide PVC



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteClusterWidePvcV1(context.Background(), clusterUuid, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteClusterWidePvcV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteClusterWidePvcV1Request struct via the builder pattern


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


## DeleteClusterWidePvcV2

> DeleteClusterWidePvcV2(ctx, clusterUuid, pvcName).Execute()

delete cluster-wide PVC

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteClusterWidePvcV2(context.Background(), clusterUuid, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteClusterWidePvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteClusterWidePvcV2Request struct via the builder pattern


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


## DeleteDepartmentWidePvcV1

> DeleteDepartmentWidePvcV1(ctx, clusterUuid, departmentId, pvcName).Execute()

delete department-wide PVC



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteDepartmentWidePvcV1(context.Background(), clusterUuid, departmentId, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteDepartmentWidePvcV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentWidePvcV1Request struct via the builder pattern


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


## DeleteDepartmentWidePvcV2

> DeleteDepartmentWidePvcV2(ctx, clusterUuid, departmentId, pvcName).Execute()

delete department-wide PVC

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteDepartmentWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteDepartmentWidePvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentWidePvcV2Request struct via the builder pattern


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


## DeleteDepartmentsWidePvcV2

> DeleteDepartmentsWidePvcV2(ctx, clusterUuid, departmentId, pvcName).Execute()

delete department-wide PVC

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteDepartmentsWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteDepartmentsWidePvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentsWidePvcV2Request struct via the builder pattern


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


## DeleteProjectPvcV1

> DeleteProjectPvcV1(ctx, clusterUuid, namespace, pvcName).Execute()

delete project PVC



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteProjectPvcV1(context.Background(), clusterUuid, namespace, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteProjectPvcV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectPvcV1Request struct via the builder pattern


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


## DeleteProjectPvcV2

> DeleteProjectPvcV2(ctx, clusterUuid, namespace, pvcName).Execute()

delete project PVC

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteProjectPvcV2(context.Background(), clusterUuid, namespace, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteProjectPvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectPvcV2Request struct via the builder pattern


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


## DeleteProjectsPvcV2

> DeleteProjectsPvcV2(ctx, clusterUuid, namespace, pvcName).Execute()

delete project PVC

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.DeleteProjectsPvcV2(context.Background(), clusterUuid, namespace, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.DeleteProjectsPvcV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectsPvcV2Request struct via the builder pattern


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


## ListPvcs

> PVCListResponse ListPvcs(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).Execute()

list details of pvcs

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	scopeType := "project" // string | The type of resource we want to filter by. (optional)
	scopeId := "scopeId_example" // string | filter by scope id (optional)
	name := "name_example" // string | filter by name (optional)
	clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
	origin := "origin_example" // string | filter by the origin of the resource. by default, we have to list only customer existing resources (to preserve backward compatibility) to list both existing pvc and assets related pvcs use 'all'. (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.ListPvcs(context.Background()).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.ListPvcs``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListPvcs`: PVCListResponse
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.ListPvcs`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListPvcsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | filter by scope id | 
 **name** | **string** | filter by name | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **origin** | **string** | filter by the origin of the resource. by default, we have to list only customer existing resources (to preserve backward compatibility) to list both existing pvc and assets related pvcs use &#39;all&#39;. | 

### Return type

[**PVCListResponse**](PVCListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ObsoleteCreatePvc

> SinglePVC ObsoleteCreatePvc(ctx, clusterUuid, namespace).SinglePVC(singlePVC).Execute()

upload the details of PVC to the control plane



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	singlePVC := *openapiclient.NewSinglePVC("pvc-1") // SinglePVC | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.ObsoleteCreatePvc(context.Background(), clusterUuid, namespace).SinglePVC(singlePVC).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.ObsoleteCreatePvc``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ObsoleteCreatePvc`: SinglePVC
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.ObsoleteCreatePvc`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiObsoleteCreatePvcRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **singlePVC** | [**SinglePVC**](SinglePVC.md) |  | 

### Return type

[**SinglePVC**](SinglePVC.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ObsoleteDeletePvc

> ObsoleteDeletePvc(ctx, clusterUuid, namespace, pvcName).Execute()

delete PVC



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pvcName := "pvcName_example" // string | The pvc name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.PVCsAPI.ObsoleteDeletePvc(context.Background(), clusterUuid, namespace, pvcName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.ObsoleteDeletePvc``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**pvcName** | **string** | The pvc name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiObsoleteDeletePvcRequest struct via the builder pattern


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


## ObsoleteSyncPvc

> []SinglePVC ObsoleteSyncPvc(ctx, clusterUuid, namespace).SinglePVC(singlePVC).Execute()

upload the details of PVCs to the control plane



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	singlePVC := []openapiclient.SinglePVC{*openapiclient.NewSinglePVC("pvc-1")} // []SinglePVC | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.ObsoleteSyncPvc(context.Background(), clusterUuid, namespace).SinglePVC(singlePVC).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.ObsoleteSyncPvc``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ObsoleteSyncPvc`: []SinglePVC
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.ObsoleteSyncPvc`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiObsoleteSyncPvcRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **singlePVC** | [**[]SinglePVC**](SinglePVC.md) |  | 

### Return type

[**[]SinglePVC**](SinglePVC.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncClusterWidePvcNamesV1

> PVCNames SyncClusterWidePvcNamesV1(ctx, clusterUuid).PVCNames(pVCNames).Execute()

sync the names of cluster-wide pvcs to the control plane.



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	pVCNames := *openapiclient.NewPVCNames([]string{"pvc-1"}) // PVCNames | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncClusterWidePvcNamesV1(context.Background(), clusterUuid).PVCNames(pVCNames).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncClusterWidePvcNamesV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncClusterWidePvcNamesV1`: PVCNames
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncClusterWidePvcNamesV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncClusterWidePvcNamesV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pVCNames** | [**PVCNames**](PVCNames.md) |  | 

### Return type

[**PVCNames**](PVCNames.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncClusterWidePvcsV2

> PVCs SyncClusterWidePvcsV2(ctx, clusterUuid).PVCs(pVCs).Execute()

sync the details of cluster-wide pvcs to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	pVCs := *openapiclient.NewPVCs([]openapiclient.PVC{*openapiclient.NewPVC("pvc-1")}) // PVCs | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncClusterWidePvcsV2(context.Background(), clusterUuid).PVCs(pVCs).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncClusterWidePvcsV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncClusterWidePvcsV2`: PVCs
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncClusterWidePvcsV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncClusterWidePvcsV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **pVCs** | [**PVCs**](PVCs.md) |  | 

### Return type

[**PVCs**](PVCs.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDepartmentWidePvcNamesV1

> PVCNames SyncDepartmentWidePvcNamesV1(ctx, clusterUuid, departmentId).PVCNames(pVCNames).Execute()

sync the names of department-wide pvcs to the control plane



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pVCNames := *openapiclient.NewPVCNames([]string{"pvc-1"}) // PVCNames | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncDepartmentWidePvcNamesV1(context.Background(), clusterUuid, departmentId).PVCNames(pVCNames).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncDepartmentWidePvcNamesV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDepartmentWidePvcNamesV1`: PVCNames
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncDepartmentWidePvcNamesV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDepartmentWidePvcNamesV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pVCNames** | [**PVCNames**](PVCNames.md) |  | 

### Return type

[**PVCNames**](PVCNames.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDepartmentWidePvcsV2

> PVCs SyncDepartmentWidePvcsV2(ctx, clusterUuid, departmentId).PVCs(pVCs).Execute()

sync the details of department-wide pvcs to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pVCs := *openapiclient.NewPVCs([]openapiclient.PVC{*openapiclient.NewPVC("pvc-1")}) // PVCs | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncDepartmentWidePvcsV2(context.Background(), clusterUuid, departmentId).PVCs(pVCs).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncDepartmentWidePvcsV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDepartmentWidePvcsV2`: PVCs
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncDepartmentWidePvcsV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDepartmentWidePvcsV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pVCs** | [**PVCs**](PVCs.md) |  | 

### Return type

[**PVCs**](PVCs.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDepartmentsWidePvcsV2

> PVCs SyncDepartmentsWidePvcsV2(ctx, clusterUuid, departmentId).PVCs(pVCs).Execute()

sync the details of department-wide pvcs to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	departmentId := "1" // string | The id of the department.
	pVCs := *openapiclient.NewPVCs([]openapiclient.PVC{*openapiclient.NewPVC("pvc-1")}) // PVCs | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncDepartmentsWidePvcsV2(context.Background(), clusterUuid, departmentId).PVCs(pVCs).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncDepartmentsWidePvcsV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDepartmentsWidePvcsV2`: PVCs
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncDepartmentsWidePvcsV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDepartmentsWidePvcsV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pVCs** | [**PVCs**](PVCs.md) |  | 

### Return type

[**PVCs**](PVCs.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectPvcNamesV1

> PVCNames SyncProjectPvcNamesV1(ctx, clusterUuid, namespace).PVCNames(pVCNames).Execute()

sync the names of project pvcs to the control plane



### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pVCNames := *openapiclient.NewPVCNames([]string{"pvc-1"}) // PVCNames | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncProjectPvcNamesV1(context.Background(), clusterUuid, namespace).PVCNames(pVCNames).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncProjectPvcNamesV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncProjectPvcNamesV1`: PVCNames
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncProjectPvcNamesV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectPvcNamesV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pVCNames** | [**PVCNames**](PVCNames.md) |  | 

### Return type

[**PVCNames**](PVCNames.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectPvcsV2

> PVCs SyncProjectPvcsV2(ctx, clusterUuid, namespace).PVCs(pVCs).Execute()

sync the details of project pvcs to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pVCs := *openapiclient.NewPVCs([]openapiclient.PVC{*openapiclient.NewPVC("pvc-1")}) // PVCs | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncProjectPvcsV2(context.Background(), clusterUuid, namespace).PVCs(pVCs).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncProjectPvcsV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncProjectPvcsV2`: PVCs
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncProjectPvcsV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectPvcsV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pVCs** | [**PVCs**](PVCs.md) |  | 

### Return type

[**PVCs**](PVCs.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectsPvcsV2

> PVCs SyncProjectsPvcsV2(ctx, clusterUuid, namespace).PVCs(pVCs).Execute()

sync the details of project pvcs to the control plane

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/run-ai/runai/control-plane/runai-common-packages/k8s-objects-tracker-client"
)

func main() {
	clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
	namespace := "namespace_example" // string | The namespace of the project.
	pVCs := *openapiclient.NewPVCs([]openapiclient.PVC{*openapiclient.NewPVC("pvc-1")}) // PVCs | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.PVCsAPI.SyncProjectsPvcsV2(context.Background(), clusterUuid, namespace).PVCs(pVCs).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `PVCsAPI.SyncProjectsPvcsV2``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncProjectsPvcsV2`: PVCs
	fmt.Fprintf(os.Stdout, "Response from `PVCsAPI.SyncProjectsPvcsV2`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectsPvcsV2Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **pVCs** | [**PVCs**](PVCs.md) |  | 

### Return type

[**PVCs**](PVCs.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

