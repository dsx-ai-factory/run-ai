# \PVCsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateClusterWidePvcV1**](PVCsApi.md#CreateClusterWidePvcV1) | **Post** /api/v1/pvcs/{clusterUuid}/{pvcName} | upload the name of cluster-wide pvc to the control plane
[**CreateClusterWidePvcV2**](PVCsApi.md#CreateClusterWidePvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/{pvcName} | upload the details of cluster-wide pvc to the control plane
[**CreateDepartmentWidePvcV1**](PVCsApi.md#CreateDepartmentWidePvcV1) | **Post** /api/v1/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | upload the name of department-wide pvc to the control plane
[**CreateDepartmentWidePvcV2**](PVCsApi.md#CreateDepartmentWidePvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | upload the details of department-wide pvc to the control plane
[**CreateDepartmentsWidePvcV2**](PVCsApi.md#CreateDepartmentsWidePvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/departments/{departmentId}/{pvcName} | upload the details of department-wide pvc to the control plane
[**CreateProjectPvcV1**](PVCsApi.md#CreateProjectPvcV1) | **Post** /api/v1/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | upload the name of project pvc to the control plane
[**CreateProjectPvcV2**](PVCsApi.md#CreateProjectPvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | upload the details of project pvc to the control plane
[**CreateProjectsPvcV2**](PVCsApi.md#CreateProjectsPvcV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespaces/{namespace}/{pvcName} | upload the details of project pvc to the control plane
[**DeleteClusterWidePvcV1**](PVCsApi.md#DeleteClusterWidePvcV1) | **Delete** /api/v1/pvcs/{clusterUuid}/{pvcName} | delete cluster-wide PVC
[**DeleteClusterWidePvcV2**](PVCsApi.md#DeleteClusterWidePvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/{pvcName} | delete cluster-wide PVC
[**DeleteDepartmentWidePvcV1**](PVCsApi.md#DeleteDepartmentWidePvcV1) | **Delete** /api/v1/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | delete department-wide PVC
[**DeleteDepartmentWidePvcV2**](PVCsApi.md#DeleteDepartmentWidePvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/department/{departmentId}/{pvcName} | delete department-wide PVC
[**DeleteDepartmentsWidePvcV2**](PVCsApi.md#DeleteDepartmentsWidePvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/departments/{departmentId}/{pvcName} | delete department-wide PVC
[**DeleteProjectPvcV1**](PVCsApi.md#DeleteProjectPvcV1) | **Delete** /api/v1/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | delete project PVC
[**DeleteProjectPvcV2**](PVCsApi.md#DeleteProjectPvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/namespace/{namespace}/{pvcName} | delete project PVC
[**DeleteProjectsPvcV2**](PVCsApi.md#DeleteProjectsPvcV2) | **Delete** /api/v2/pvcs/{clusterUuid}/namespaces/{namespace}/{pvcName} | delete project PVC
[**ListPvcs**](PVCsApi.md#ListPvcs) | **Get** /api/v1/pvcs | list details of pvcs
[**ObsoleteCreatePvc**](PVCsApi.md#ObsoleteCreatePvc) | **Post** /v1/k8s/clusters/{clusterUuid}/{namespace}/pvcs | upload the details of PVC to the control plane
[**ObsoleteDeletePvc**](PVCsApi.md#ObsoleteDeletePvc) | **Delete** /v1/k8s/clusters/{clusterUuid}/{namespace}/pvcs/{pvcName} | delete PVC
[**ObsoleteSyncPvc**](PVCsApi.md#ObsoleteSyncPvc) | **Post** /v1/k8s/clusters/{clusterUuid}/{namespace}/pvcs/sync | upload the details of PVCs to the control plane
[**SyncClusterWidePvcNamesV1**](PVCsApi.md#SyncClusterWidePvcNamesV1) | **Post** /api/v1/pvcs/{clusterUuid} | sync the names of cluster-wide pvcs to the control plane.
[**SyncClusterWidePvcsV2**](PVCsApi.md#SyncClusterWidePvcsV2) | **Post** /api/v2/pvcs/{clusterUuid} | sync the details of cluster-wide pvcs to the control plane
[**SyncDepartmentWidePvcNamesV1**](PVCsApi.md#SyncDepartmentWidePvcNamesV1) | **Post** /api/v1/pvcs/{clusterUuid}/department/{departmentId} | sync the names of department-wide pvcs to the control plane
[**SyncDepartmentWidePvcsV2**](PVCsApi.md#SyncDepartmentWidePvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/department/{departmentId} | sync the details of department-wide pvcs to the control plane
[**SyncDepartmentsWidePvcsV2**](PVCsApi.md#SyncDepartmentsWidePvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/departments/{departmentId} | sync the details of department-wide pvcs to the control plane
[**SyncProjectPvcNamesV1**](PVCsApi.md#SyncProjectPvcNamesV1) | **Post** /api/v1/pvcs/{clusterUuid}/namespace/{namespace} | sync the names of project pvcs to the control plane
[**SyncProjectPvcsV2**](PVCsApi.md#SyncProjectPvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespace/{namespace} | sync the details of project pvcs to the control plane
[**SyncProjectsPvcsV2**](PVCsApi.md#SyncProjectsPvcsV2) | **Post** /api/v2/pvcs/{clusterUuid}/namespaces/{namespace} | sync the details of project pvcs to the control plane



## CreateClusterWidePvcV1

> string CreateClusterWidePvcV1(ctx, clusterUuid, pvcName).Body(body).Execute()

upload the name of cluster-wide pvc to the control plane



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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    pvcName := "pvcName_example" // string | The pvc name.
    body := interface{}(987) // interface{} |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateClusterWidePvcV1(context.Background(), clusterUuid, pvcName).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateClusterWidePvcV1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateClusterWidePvcV1`: string
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateClusterWidePvcV1`: %v\n", resp)
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


 **body** | **interface{}** |  | 

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

> PVC1 CreateClusterWidePvcV2(ctx, clusterUuid, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()

upload the details of cluster-wide pvc to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    pvcName := "pvcName_example" // string | The pvc name.
    pVCCreationRequest1 := *openapiclient.NewPVCCreationRequest1() // PVCCreationRequest1 |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateClusterWidePvcV2(context.Background(), clusterUuid, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateClusterWidePvcV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateClusterWidePvcV2`: PVC1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateClusterWidePvcV2`: %v\n", resp)
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


 **pVCCreationRequest1** | [**PVCCreationRequest1**](PVCCreationRequest1.md) |  | 

### Return type

[**PVC1**](PVC1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentWidePvcV1

> string CreateDepartmentWidePvcV1(ctx, clusterUuid, departmentId, pvcName).Body(body).Execute()

upload the name of department-wide pvc to the control plane



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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pvcName := "pvcName_example" // string | The pvc name.
    body := interface{}(987) // interface{} | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateDepartmentWidePvcV1(context.Background(), clusterUuid, departmentId, pvcName).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateDepartmentWidePvcV1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDepartmentWidePvcV1`: string
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateDepartmentWidePvcV1`: %v\n", resp)
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



 **body** | **interface{}** |  | 

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

> PVC1 CreateDepartmentWidePvcV2(ctx, clusterUuid, departmentId, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()

upload the details of department-wide pvc to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pvcName := "pvcName_example" // string | The pvc name.
    pVCCreationRequest1 := *openapiclient.NewPVCCreationRequest1() // PVCCreationRequest1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateDepartmentWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateDepartmentWidePvcV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDepartmentWidePvcV2`: PVC1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateDepartmentWidePvcV2`: %v\n", resp)
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



 **pVCCreationRequest1** | [**PVCCreationRequest1**](PVCCreationRequest1.md) |  | 

### Return type

[**PVC1**](PVC1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentsWidePvcV2

> PVC1 CreateDepartmentsWidePvcV2(ctx, clusterUuid, departmentId, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()

upload the details of department-wide pvc to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pvcName := "pvcName_example" // string | The pvc name.
    pVCCreationRequest1 := *openapiclient.NewPVCCreationRequest1() // PVCCreationRequest1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateDepartmentsWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateDepartmentsWidePvcV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDepartmentsWidePvcV2`: PVC1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateDepartmentsWidePvcV2`: %v\n", resp)
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



 **pVCCreationRequest1** | [**PVCCreationRequest1**](PVCCreationRequest1.md) |  | 

### Return type

[**PVC1**](PVC1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectPvcV1

> string CreateProjectPvcV1(ctx, clusterUuid, namespace, pvcName).Body(body).Execute()

upload the name of project pvc to the control plane



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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pvcName := "pvcName_example" // string | The pvc name.
    body := interface{}(987) // interface{} | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateProjectPvcV1(context.Background(), clusterUuid, namespace, pvcName).Body(body).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateProjectPvcV1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateProjectPvcV1`: string
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateProjectPvcV1`: %v\n", resp)
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



 **body** | **interface{}** |  | 

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

> PVC1 CreateProjectPvcV2(ctx, clusterUuid, namespace, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()

upload the details of project pvc to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pvcName := "pvcName_example" // string | The pvc name.
    pVCCreationRequest1 := *openapiclient.NewPVCCreationRequest1() // PVCCreationRequest1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateProjectPvcV2(context.Background(), clusterUuid, namespace, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateProjectPvcV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateProjectPvcV2`: PVC1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateProjectPvcV2`: %v\n", resp)
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



 **pVCCreationRequest1** | [**PVCCreationRequest1**](PVCCreationRequest1.md) |  | 

### Return type

[**PVC1**](PVC1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectsPvcV2

> PVC1 CreateProjectsPvcV2(ctx, clusterUuid, namespace, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()

upload the details of project pvc to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pvcName := "pvcName_example" // string | The pvc name.
    pVCCreationRequest1 := *openapiclient.NewPVCCreationRequest1() // PVCCreationRequest1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.CreateProjectsPvcV2(context.Background(), clusterUuid, namespace, pvcName).PVCCreationRequest1(pVCCreationRequest1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.CreateProjectsPvcV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateProjectsPvcV2`: PVC1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.CreateProjectsPvcV2`: %v\n", resp)
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



 **pVCCreationRequest1** | [**PVCCreationRequest1**](PVCCreationRequest1.md) |  | 

### Return type

[**PVC1**](PVC1.md)

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteClusterWidePvcV1(context.Background(), clusterUuid, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteClusterWidePvcV1``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteClusterWidePvcV2(context.Background(), clusterUuid, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteClusterWidePvcV2``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteDepartmentWidePvcV1(context.Background(), clusterUuid, departmentId, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteDepartmentWidePvcV1``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteDepartmentWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteDepartmentWidePvcV2``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteDepartmentsWidePvcV2(context.Background(), clusterUuid, departmentId, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteDepartmentsWidePvcV2``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteProjectPvcV1(context.Background(), clusterUuid, namespace, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteProjectPvcV1``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteProjectPvcV2(context.Background(), clusterUuid, namespace, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteProjectPvcV2``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.DeleteProjectsPvcV2(context.Background(), clusterUuid, namespace, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.DeleteProjectsPvcV2``: %v\n", err)
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

> PVCListResponse1 ListPvcs(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).Execute()

list details of pvcs

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
    scopeType := "project" // string | The type of resource we want to filter by. (optional)
    scopeId := "scopeId_example" // string | filter by scope id (optional)
    name := "name_example" // string | filter by name (optional)
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
    origin := "origin_example" // string | filter by the origin of the resource. by default, we have to list only customer existing resources (to preserve backward compatibility) to list both existing pvc and assets related pvcs use 'all'. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.ListPvcs(context.Background()).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Origin(origin).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.ListPvcs``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListPvcs`: PVCListResponse1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.ListPvcs`: %v\n", resp)
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

[**PVCListResponse1**](PVCListResponse1.md)

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    singlePVC := *openapiclient.NewSinglePVC("pvc-1") // SinglePVC | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.ObsoleteCreatePvc(context.Background(), clusterUuid, namespace).SinglePVC(singlePVC).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.ObsoleteCreatePvc``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ObsoleteCreatePvc`: SinglePVC
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.ObsoleteCreatePvc`: %v\n", resp)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pvcName := "pvcName_example" // string | The pvc name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.ObsoleteDeletePvc(context.Background(), clusterUuid, namespace, pvcName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.ObsoleteDeletePvc``: %v\n", err)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    singlePVC := []openapiclient.SinglePVC{*openapiclient.NewSinglePVC("pvc-1")} // []SinglePVC | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.ObsoleteSyncPvc(context.Background(), clusterUuid, namespace).SinglePVC(singlePVC).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.ObsoleteSyncPvc``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ObsoleteSyncPvc`: []SinglePVC
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.ObsoleteSyncPvc`: %v\n", resp)
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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    pVCNames := *openapiclient.NewPVCNames([]string{"pvc-1"}) // PVCNames | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncClusterWidePvcNamesV1(context.Background(), clusterUuid).PVCNames(pVCNames).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncClusterWidePvcNamesV1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncClusterWidePvcNamesV1`: PVCNames
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncClusterWidePvcNamesV1`: %v\n", resp)
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

> PVCs1 SyncClusterWidePvcsV2(ctx, clusterUuid).PVCs1(pVCs1).Execute()

sync the details of cluster-wide pvcs to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    pVCs1 := *openapiclient.NewPVCs1([]openapiclient.PVC1{*openapiclient.NewPVC1("pvc-1")}) // PVCs1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncClusterWidePvcsV2(context.Background(), clusterUuid).PVCs1(pVCs1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncClusterWidePvcsV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncClusterWidePvcsV2`: PVCs1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncClusterWidePvcsV2`: %v\n", resp)
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

 **pVCs1** | [**PVCs1**](PVCs1.md) |  | 

### Return type

[**PVCs1**](PVCs1.md)

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pVCNames := *openapiclient.NewPVCNames([]string{"pvc-1"}) // PVCNames | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncDepartmentWidePvcNamesV1(context.Background(), clusterUuid, departmentId).PVCNames(pVCNames).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncDepartmentWidePvcNamesV1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncDepartmentWidePvcNamesV1`: PVCNames
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncDepartmentWidePvcNamesV1`: %v\n", resp)
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

> PVCs1 SyncDepartmentWidePvcsV2(ctx, clusterUuid, departmentId).PVCs1(pVCs1).Execute()

sync the details of department-wide pvcs to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pVCs1 := *openapiclient.NewPVCs1([]openapiclient.PVC1{*openapiclient.NewPVC1("pvc-1")}) // PVCs1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncDepartmentWidePvcsV2(context.Background(), clusterUuid, departmentId).PVCs1(pVCs1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncDepartmentWidePvcsV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncDepartmentWidePvcsV2`: PVCs1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncDepartmentWidePvcsV2`: %v\n", resp)
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


 **pVCs1** | [**PVCs1**](PVCs1.md) |  | 

### Return type

[**PVCs1**](PVCs1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDepartmentsWidePvcsV2

> PVCs1 SyncDepartmentsWidePvcsV2(ctx, clusterUuid, departmentId).PVCs1(pVCs1).Execute()

sync the details of department-wide pvcs to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    departmentId := "1" // string | The id of the department.
    pVCs1 := *openapiclient.NewPVCs1([]openapiclient.PVC1{*openapiclient.NewPVC1("pvc-1")}) // PVCs1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncDepartmentsWidePvcsV2(context.Background(), clusterUuid, departmentId).PVCs1(pVCs1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncDepartmentsWidePvcsV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncDepartmentsWidePvcsV2`: PVCs1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncDepartmentsWidePvcsV2`: %v\n", resp)
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


 **pVCs1** | [**PVCs1**](PVCs1.md) |  | 

### Return type

[**PVCs1**](PVCs1.md)

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
    openapiclient "./openapi"
)

func main() {
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pVCNames := *openapiclient.NewPVCNames([]string{"pvc-1"}) // PVCNames | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncProjectPvcNamesV1(context.Background(), clusterUuid, namespace).PVCNames(pVCNames).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncProjectPvcNamesV1``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncProjectPvcNamesV1`: PVCNames
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncProjectPvcNamesV1`: %v\n", resp)
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

> PVCs1 SyncProjectPvcsV2(ctx, clusterUuid, namespace).PVCs1(pVCs1).Execute()

sync the details of project pvcs to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pVCs1 := *openapiclient.NewPVCs1([]openapiclient.PVC1{*openapiclient.NewPVC1("pvc-1")}) // PVCs1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncProjectPvcsV2(context.Background(), clusterUuid, namespace).PVCs1(pVCs1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncProjectPvcsV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncProjectPvcsV2`: PVCs1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncProjectPvcsV2`: %v\n", resp)
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


 **pVCs1** | [**PVCs1**](PVCs1.md) |  | 

### Return type

[**PVCs1**](PVCs1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectsPvcsV2

> PVCs1 SyncProjectsPvcsV2(ctx, clusterUuid, namespace).PVCs1(pVCs1).Execute()

sync the details of project pvcs to the control plane

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
    clusterUuid := "9f55255e-11ed-47c7-acef-fc4054768dbc" // string | The Universally Unique Identifier (UUID) of the cluster.
    namespace := "namespace_example" // string | The namespace of the project.
    pVCs1 := *openapiclient.NewPVCs1([]openapiclient.PVC1{*openapiclient.NewPVC1("pvc-1")}) // PVCs1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PVCsApi.SyncProjectsPvcsV2(context.Background(), clusterUuid, namespace).PVCs1(pVCs1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PVCsApi.SyncProjectsPvcsV2``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncProjectsPvcsV2`: PVCs1
    fmt.Fprintf(os.Stdout, "Response from `PVCsApi.SyncProjectsPvcsV2`: %v\n", resp)
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


 **pVCs1** | [**PVCs1**](PVCs1.md) |  | 

### Return type

[**PVCs1**](PVCs1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

