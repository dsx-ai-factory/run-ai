# \SecretsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateClusterWideSecret**](SecretsApi.md#CreateClusterWideSecret) | **Post** /api/v1/secrets/{clusterUuid}/{secretName} | upload the details of cluster-wide secret to the control plane
[**CreateDepartmentWideSecret**](SecretsApi.md#CreateDepartmentWideSecret) | **Post** /api/v1/secrets/{clusterUuid}/department/{departmentId}/{secretName} | upload the details of department-wide secret to the control plane
[**CreateDepartmentsWideSecret**](SecretsApi.md#CreateDepartmentsWideSecret) | **Post** /api/v1/secrets/{clusterUuid}/departments/{departmentId}/{secretName} | upload the details of department-wide secret to the control plane
[**CreateProjectSecret**](SecretsApi.md#CreateProjectSecret) | **Post** /api/v1/secrets/{clusterUuid}/namespace/{namespace}/{secretName} | upload the details of project-wide secret to the control plane
[**CreateProjectsSecret**](SecretsApi.md#CreateProjectsSecret) | **Post** /api/v1/secrets/{clusterUuid}/namespaces/{namespace}/{secretName} | upload the details of project-wide secret to the control plane
[**DeleteClusterWideSecret**](SecretsApi.md#DeleteClusterWideSecret) | **Delete** /api/v1/secrets/{clusterUuid}/{secretName} | delete cluster-wide secret
[**DeleteDepartmentWideSecret**](SecretsApi.md#DeleteDepartmentWideSecret) | **Delete** /api/v1/secrets/{clusterUuid}/department/{departmentId}/{secretName} | delete department-wide secret
[**DeleteDepartmentsWideSecret**](SecretsApi.md#DeleteDepartmentsWideSecret) | **Delete** /api/v1/secrets/{clusterUuid}/departments/{departmentId}/{secretName} | delete department-wide secret
[**DeleteProjectSecret**](SecretsApi.md#DeleteProjectSecret) | **Delete** /api/v1/secrets/{clusterUuid}/namespace/{namespace}/{secretName} | delete project secret
[**DeleteProjectsSecret**](SecretsApi.md#DeleteProjectsSecret) | **Delete** /api/v1/secrets/{clusterUuid}/namespaces/{namespace}/{secretName} | delete project secret
[**ListSecrets**](SecretsApi.md#ListSecrets) | **Get** /api/v1/secrets | list details of secrets
[**ObsoleteCreateSecret**](SecretsApi.md#ObsoleteCreateSecret) | **Post** /v1/k8s/clusters/{clusterUuid}/{namespace}/secrets | upload the details of a secret to the control plane
[**ObsoleteDeleteSecret**](SecretsApi.md#ObsoleteDeleteSecret) | **Delete** /v1/k8s/clusters/{clusterUuid}/{namespace}/secrets/{secretName} | delete secret
[**ObsoleteSyncSecret**](SecretsApi.md#ObsoleteSyncSecret) | **Post** /v1/k8s/clusters/{clusterUuid}/{namespace}/secrets/sync | upload the details of secrets to the control plane
[**SyncClusterWideSecrets**](SecretsApi.md#SyncClusterWideSecrets) | **Post** /api/v1/secrets/{clusterUuid} | sync the details of cluster-wide secrets to the control plane
[**SyncDepartmentWideSecrets**](SecretsApi.md#SyncDepartmentWideSecrets) | **Post** /api/v1/secrets/{clusterUuid}/department/{departmentId} | sync the details of department-wide secrets to the control plane
[**SyncDepartmentsWideSecrets**](SecretsApi.md#SyncDepartmentsWideSecrets) | **Post** /api/v1/secrets/{clusterUuid}/departments/{departmentId} | sync the details of department-wide secrets to the control plane
[**SyncProjectSecrets**](SecretsApi.md#SyncProjectSecrets) | **Post** /api/v1/secrets/{clusterUuid}/namespace/{namespace} | sync the details of project secrets to the control plane
[**SyncProjectsSecrets**](SecretsApi.md#SyncProjectsSecrets) | **Post** /api/v1/secrets/{clusterUuid}/namespaces/{namespace} | sync the details of project secrets to the control plane



## CreateClusterWideSecret

> SecretCreationRequest CreateClusterWideSecret(ctx, clusterUuid, secretName).SecretCreationRequest(secretCreationRequest).Execute()

upload the details of cluster-wide secret to the control plane

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
    secretName := "secretName_example" // string | The secret name.
    secretCreationRequest := *openapiclient.NewSecretCreationRequest(openapiclient.SecretType("access-key")) // SecretCreationRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.CreateClusterWideSecret(context.Background(), clusterUuid, secretName).SecretCreationRequest(secretCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.CreateClusterWideSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateClusterWideSecret`: SecretCreationRequest
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.CreateClusterWideSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateClusterWideSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **secretCreationRequest** | [**SecretCreationRequest**](SecretCreationRequest.md) |  | 

### Return type

[**SecretCreationRequest**](SecretCreationRequest.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentWideSecret

> Secret3 CreateDepartmentWideSecret(ctx, clusterUuid, departmentId, secretName).SecretCreationRequest(secretCreationRequest).Execute()

upload the details of department-wide secret to the control plane

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
    secretName := "secretName_example" // string | The secret name.
    secretCreationRequest := *openapiclient.NewSecretCreationRequest(openapiclient.SecretType("access-key")) // SecretCreationRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.CreateDepartmentWideSecret(context.Background(), clusterUuid, departmentId, secretName).SecretCreationRequest(secretCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.CreateDepartmentWideSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDepartmentWideSecret`: Secret3
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.CreateDepartmentWideSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentWideSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **secretCreationRequest** | [**SecretCreationRequest**](SecretCreationRequest.md) |  | 

### Return type

[**Secret3**](Secret3.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentsWideSecret

> Secret3 CreateDepartmentsWideSecret(ctx, clusterUuid, departmentId, secretName).SecretCreationRequest(secretCreationRequest).Execute()

upload the details of department-wide secret to the control plane

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
    secretName := "secretName_example" // string | The secret name.
    secretCreationRequest := *openapiclient.NewSecretCreationRequest(openapiclient.SecretType("access-key")) // SecretCreationRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.CreateDepartmentsWideSecret(context.Background(), clusterUuid, departmentId, secretName).SecretCreationRequest(secretCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.CreateDepartmentsWideSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateDepartmentsWideSecret`: Secret3
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.CreateDepartmentsWideSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentsWideSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **secretCreationRequest** | [**SecretCreationRequest**](SecretCreationRequest.md) |  | 

### Return type

[**Secret3**](Secret3.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectSecret

> Secret3 CreateProjectSecret(ctx, clusterUuid, namespace, secretName).SecretCreationRequest(secretCreationRequest).Execute()

upload the details of project-wide secret to the control plane

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
    secretName := "secretName_example" // string | The secret name.
    secretCreationRequest := *openapiclient.NewSecretCreationRequest(openapiclient.SecretType("access-key")) // SecretCreationRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.CreateProjectSecret(context.Background(), clusterUuid, namespace, secretName).SecretCreationRequest(secretCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.CreateProjectSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateProjectSecret`: Secret3
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.CreateProjectSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **secretCreationRequest** | [**SecretCreationRequest**](SecretCreationRequest.md) |  | 

### Return type

[**Secret3**](Secret3.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectsSecret

> Secret3 CreateProjectsSecret(ctx, clusterUuid, namespace, secretName).SecretCreationRequest(secretCreationRequest).Execute()

upload the details of project-wide secret to the control plane

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
    secretName := "secretName_example" // string | The secret name.
    secretCreationRequest := *openapiclient.NewSecretCreationRequest(openapiclient.SecretType("access-key")) // SecretCreationRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.CreateProjectsSecret(context.Background(), clusterUuid, namespace, secretName).SecretCreationRequest(secretCreationRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.CreateProjectsSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateProjectsSecret`: Secret3
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.CreateProjectsSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectsSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **secretCreationRequest** | [**SecretCreationRequest**](SecretCreationRequest.md) |  | 

### Return type

[**Secret3**](Secret3.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteClusterWideSecret

> DeleteClusterWideSecret(ctx, clusterUuid, secretName).Execute()

delete cluster-wide secret

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
    secretName := "secretName_example" // string | The secret name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.DeleteClusterWideSecret(context.Background(), clusterUuid, secretName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.DeleteClusterWideSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteClusterWideSecretRequest struct via the builder pattern


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


## DeleteDepartmentWideSecret

> DeleteDepartmentWideSecret(ctx, clusterUuid, departmentId, secretName).Execute()

delete department-wide secret

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
    secretName := "secretName_example" // string | The secret name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.DeleteDepartmentWideSecret(context.Background(), clusterUuid, departmentId, secretName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.DeleteDepartmentWideSecret``: %v\n", err)
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
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentWideSecretRequest struct via the builder pattern


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


## DeleteDepartmentsWideSecret

> DeleteDepartmentsWideSecret(ctx, clusterUuid, departmentId, secretName).Execute()

delete department-wide secret

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
    secretName := "secretName_example" // string | The secret name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.DeleteDepartmentsWideSecret(context.Background(), clusterUuid, departmentId, secretName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.DeleteDepartmentsWideSecret``: %v\n", err)
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
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentsWideSecretRequest struct via the builder pattern


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


## DeleteProjectSecret

> DeleteProjectSecret(ctx, clusterUuid, namespace, secretName).Execute()

delete project secret

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
    secretName := "secretName_example" // string | The secret name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.DeleteProjectSecret(context.Background(), clusterUuid, namespace, secretName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.DeleteProjectSecret``: %v\n", err)
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
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectSecretRequest struct via the builder pattern


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


## DeleteProjectsSecret

> DeleteProjectsSecret(ctx, clusterUuid, namespace, secretName).Execute()

delete project secret

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
    secretName := "secretName_example" // string | The secret name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.DeleteProjectsSecret(context.Background(), clusterUuid, namespace, secretName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.DeleteProjectsSecret``: %v\n", err)
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
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectsSecretRequest struct via the builder pattern


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


## ListSecrets

> SecretListResponse ListSecrets(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).SecretType(secretType).ClusterId(clusterId).Origin(origin).Execute()

list details of secrets

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
    secretType := "secretType_example" // string | filter by secret type. supported values - password, access-key, docker-registry. (optional)
    clusterId := "d73a738f-fab3-430a-8fa3-5241493d7128" // string | Filter using the Universally Unique Identifier (UUID) of the cluster. (optional)
    origin := "origin_example" // string | filter by the origin of the resource. by default, we have to list only customer existing resources (to preserve backward compatibility) to list both existing pvc and assets related pvcs use 'all'. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.ListSecrets(context.Background()).ScopeType(scopeType).ScopeId(scopeId).Name(name).SecretType(secretType).ClusterId(clusterId).Origin(origin).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.ListSecrets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListSecrets`: SecretListResponse
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.ListSecrets`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListSecretsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | filter by scope id | 
 **name** | **string** | filter by name | 
 **secretType** | **string** | filter by secret type. supported values - password, access-key, docker-registry. | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 
 **origin** | **string** | filter by the origin of the resource. by default, we have to list only customer existing resources (to preserve backward compatibility) to list both existing pvc and assets related pvcs use &#39;all&#39;. | 

### Return type

[**SecretListResponse**](SecretListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ObsoleteCreateSecret

> SingleSecret ObsoleteCreateSecret(ctx, clusterUuid, namespace).SingleSecret(singleSecret).Execute()

upload the details of a secret to the control plane



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
    singleSecret := *openapiclient.NewSingleSecret("secret-1") // SingleSecret | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.ObsoleteCreateSecret(context.Background(), clusterUuid, namespace).SingleSecret(singleSecret).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.ObsoleteCreateSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ObsoleteCreateSecret`: SingleSecret
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.ObsoleteCreateSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiObsoleteCreateSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **singleSecret** | [**SingleSecret**](SingleSecret.md) |  | 

### Return type

[**SingleSecret**](SingleSecret.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ObsoleteDeleteSecret

> ObsoleteDeleteSecret(ctx, clusterUuid, namespace, secretName).Execute()

delete secret



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
    secretName := "secretName_example" // string | The secret name.

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.ObsoleteDeleteSecret(context.Background(), clusterUuid, namespace, secretName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.ObsoleteDeleteSecret``: %v\n", err)
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
**secretName** | **string** | The secret name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiObsoleteDeleteSecretRequest struct via the builder pattern


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


## ObsoleteSyncSecret

> []SingleSecret ObsoleteSyncSecret(ctx, clusterUuid, namespace).SingleSecret(singleSecret).Execute()

upload the details of secrets to the control plane



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
    singleSecret := []openapiclient.SingleSecret{*openapiclient.NewSingleSecret("secret-1")} // []SingleSecret | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.ObsoleteSyncSecret(context.Background(), clusterUuid, namespace).SingleSecret(singleSecret).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.ObsoleteSyncSecret``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ObsoleteSyncSecret`: []SingleSecret
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.ObsoleteSyncSecret`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiObsoleteSyncSecretRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **singleSecret** | [**[]SingleSecret**](SingleSecret.md) |  | 

### Return type

[**[]SingleSecret**](SingleSecret.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncClusterWideSecrets

> Secrets1 SyncClusterWideSecrets(ctx, clusterUuid).Secrets1(secrets1).Execute()

sync the details of cluster-wide secrets to the control plane

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
    secrets1 := *openapiclient.NewSecrets1([]openapiclient.Secret3{*openapiclient.NewSecret3("secret-1", openapiclient.SecretType("access-key"))}) // Secrets1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.SyncClusterWideSecrets(context.Background(), clusterUuid).Secrets1(secrets1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.SyncClusterWideSecrets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncClusterWideSecrets`: Secrets1
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.SyncClusterWideSecrets`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncClusterWideSecretsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **secrets1** | [**Secrets1**](Secrets1.md) |  | 

### Return type

[**Secrets1**](Secrets1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDepartmentWideSecrets

> Secrets1 SyncDepartmentWideSecrets(ctx, clusterUuid, departmentId).Secrets1(secrets1).Execute()

sync the details of department-wide secrets to the control plane

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
    secrets1 := *openapiclient.NewSecrets1([]openapiclient.Secret3{*openapiclient.NewSecret3("secret-1", openapiclient.SecretType("access-key"))}) // Secrets1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.SyncDepartmentWideSecrets(context.Background(), clusterUuid, departmentId).Secrets1(secrets1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.SyncDepartmentWideSecrets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncDepartmentWideSecrets`: Secrets1
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.SyncDepartmentWideSecrets`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDepartmentWideSecretsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **secrets1** | [**Secrets1**](Secrets1.md) |  | 

### Return type

[**Secrets1**](Secrets1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDepartmentsWideSecrets

> Secrets1 SyncDepartmentsWideSecrets(ctx, clusterUuid, departmentId).Secrets1(secrets1).Execute()

sync the details of department-wide secrets to the control plane

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
    secrets1 := *openapiclient.NewSecrets1([]openapiclient.Secret3{*openapiclient.NewSecret3("secret-1", openapiclient.SecretType("access-key"))}) // Secrets1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.SyncDepartmentsWideSecrets(context.Background(), clusterUuid, departmentId).Secrets1(secrets1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.SyncDepartmentsWideSecrets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncDepartmentsWideSecrets`: Secrets1
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.SyncDepartmentsWideSecrets`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDepartmentsWideSecretsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **secrets1** | [**Secrets1**](Secrets1.md) |  | 

### Return type

[**Secrets1**](Secrets1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectSecrets

> Secrets1 SyncProjectSecrets(ctx, clusterUuid, namespace).Secrets1(secrets1).Execute()

sync the details of project secrets to the control plane

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
    secrets1 := *openapiclient.NewSecrets1([]openapiclient.Secret3{*openapiclient.NewSecret3("secret-1", openapiclient.SecretType("access-key"))}) // Secrets1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.SyncProjectSecrets(context.Background(), clusterUuid, namespace).Secrets1(secrets1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.SyncProjectSecrets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncProjectSecrets`: Secrets1
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.SyncProjectSecrets`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectSecretsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **secrets1** | [**Secrets1**](Secrets1.md) |  | 

### Return type

[**Secrets1**](Secrets1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectsSecrets

> Secrets1 SyncProjectsSecrets(ctx, clusterUuid, namespace).Secrets1(secrets1).Execute()

sync the details of project secrets to the control plane

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
    secrets1 := *openapiclient.NewSecrets1([]openapiclient.Secret3{*openapiclient.NewSecret3("secret-1", openapiclient.SecretType("access-key"))}) // Secrets1 | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.SecretsApi.SyncProjectsSecrets(context.Background(), clusterUuid, namespace).Secrets1(secrets1).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `SecretsApi.SyncProjectsSecrets``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncProjectsSecrets`: Secrets1
    fmt.Fprintf(os.Stdout, "Response from `SecretsApi.SyncProjectsSecrets`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectsSecretsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **secrets1** | [**Secrets1**](Secrets1.md) |  | 

### Return type

[**Secrets1**](Secrets1.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

