# \ConfigMapsAPI

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateClusterWideConfigmapV1**](ConfigMapsAPI.md#CreateClusterWideConfigmapV1) | **Post** /api/v1/config-maps/{clusterUuid}/{configMapName} | upload the details of cluster-wide ConfigMap to the control plane
[**CreateDepartmentsWideConfigmapV1**](ConfigMapsAPI.md#CreateDepartmentsWideConfigmapV1) | **Post** /api/v1/config-maps/{clusterUuid}/departments/{departmentId}/{configMapName} | upload the details of department-wide ConfigMap to the control plane
[**CreateProjectsConfigmapV1**](ConfigMapsAPI.md#CreateProjectsConfigmapV1) | **Post** /api/v1/config-maps/{clusterUuid}/namespaces/{namespace}/{configMapName} | upload the details of project ConfigMap to the control plane
[**DeleteClusterWideConfigmapV1**](ConfigMapsAPI.md#DeleteClusterWideConfigmapV1) | **Delete** /api/v1/config-maps/{clusterUuid}/{configMapName} | delete cluster-wide ConfigMap
[**DeleteDepartmentsWideConfigmapV1**](ConfigMapsAPI.md#DeleteDepartmentsWideConfigmapV1) | **Delete** /api/v1/config-maps/{clusterUuid}/departments/{departmentId}/{configMapName} | delete department-wide ConfigMap
[**DeleteProjectsConfigmapV1**](ConfigMapsAPI.md#DeleteProjectsConfigmapV1) | **Delete** /api/v1/config-maps/{clusterUuid}/namespaces/{namespace}/{configMapName} | delete project ConfigMap
[**ListConfigMaps**](ConfigMapsAPI.md#ListConfigMaps) | **Get** /api/v1/config-maps | list details of ConfigMaps
[**SyncClusterWideConfigmapsV1**](ConfigMapsAPI.md#SyncClusterWideConfigmapsV1) | **Post** /api/v1/config-maps/{clusterUuid} | sync the details of cluster-wide ConfigMaps to the control plane
[**SyncDepartmentsWideConfigmapV1**](ConfigMapsAPI.md#SyncDepartmentsWideConfigmapV1) | **Post** /api/v1/config-maps/{clusterUuid}/departments/{departmentId} | sync the details of department-wide ConfigMaps to the control plane
[**SyncProjectsConfigmapsV1**](ConfigMapsAPI.md#SyncProjectsConfigmapsV1) | **Post** /api/v1/config-maps/{clusterUuid}/namespaces/{namespace} | sync the details of project ConfigMaps to the control plane



## CreateClusterWideConfigmapV1

> ConfigMap CreateClusterWideConfigmapV1(ctx, clusterUuid, configMapName).ConfigMap(configMap).Execute()

upload the details of cluster-wide ConfigMap to the control plane

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
	configMapName := "configMapName_example" // string | The Config Map name.
	configMap := *openapiclient.NewConfigMap("cm-setting") // ConfigMap | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapsAPI.CreateClusterWideConfigmapV1(context.Background(), clusterUuid, configMapName).ConfigMap(configMap).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.CreateClusterWideConfigmapV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateClusterWideConfigmapV1`: ConfigMap
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapsAPI.CreateClusterWideConfigmapV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**configMapName** | **string** | The Config Map name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateClusterWideConfigmapV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **configMap** | [**ConfigMap**](ConfigMap.md) |  | 

### Return type

[**ConfigMap**](ConfigMap.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateDepartmentsWideConfigmapV1

> ConfigMap CreateDepartmentsWideConfigmapV1(ctx, clusterUuid, departmentId, configMapName).ConfigMap(configMap).Execute()

upload the details of department-wide ConfigMap to the control plane

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
	configMapName := "configMapName_example" // string | The Config Map name.
	configMap := *openapiclient.NewConfigMap("cm-setting") // ConfigMap | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapsAPI.CreateDepartmentsWideConfigmapV1(context.Background(), clusterUuid, departmentId, configMapName).ConfigMap(configMap).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.CreateDepartmentsWideConfigmapV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateDepartmentsWideConfigmapV1`: ConfigMap
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapsAPI.CreateDepartmentsWideConfigmapV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 
**configMapName** | **string** | The Config Map name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateDepartmentsWideConfigmapV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **configMap** | [**ConfigMap**](ConfigMap.md) |  | 

### Return type

[**ConfigMap**](ConfigMap.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateProjectsConfigmapV1

> ConfigMap CreateProjectsConfigmapV1(ctx, clusterUuid, namespace, configMapName).ConfigMap(configMap).Execute()

upload the details of project ConfigMap to the control plane

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
	configMapName := "configMapName_example" // string | The Config Map name.
	configMap := *openapiclient.NewConfigMap("cm-setting") // ConfigMap | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapsAPI.CreateProjectsConfigmapV1(context.Background(), clusterUuid, namespace, configMapName).ConfigMap(configMap).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.CreateProjectsConfigmapV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `CreateProjectsConfigmapV1`: ConfigMap
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapsAPI.CreateProjectsConfigmapV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 
**configMapName** | **string** | The Config Map name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateProjectsConfigmapV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **configMap** | [**ConfigMap**](ConfigMap.md) |  | 

### Return type

[**ConfigMap**](ConfigMap.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteClusterWideConfigmapV1

> DeleteClusterWideConfigmapV1(ctx, clusterUuid, configMapName).Execute()

delete cluster-wide ConfigMap

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
	configMapName := "configMapName_example" // string | The Config Map name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ConfigMapsAPI.DeleteClusterWideConfigmapV1(context.Background(), clusterUuid, configMapName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.DeleteClusterWideConfigmapV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**configMapName** | **string** | The Config Map name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteClusterWideConfigmapV1Request struct via the builder pattern


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


## DeleteDepartmentsWideConfigmapV1

> DeleteDepartmentsWideConfigmapV1(ctx, clusterUuid, departmentId, configMapName).Execute()

delete department-wide ConfigMap

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
	configMapName := "configMapName_example" // string | The Config Map name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ConfigMapsAPI.DeleteDepartmentsWideConfigmapV1(context.Background(), clusterUuid, departmentId, configMapName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.DeleteDepartmentsWideConfigmapV1``: %v\n", err)
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
**configMapName** | **string** | The Config Map name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteDepartmentsWideConfigmapV1Request struct via the builder pattern


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


## DeleteProjectsConfigmapV1

> DeleteProjectsConfigmapV1(ctx, clusterUuid, namespace, configMapName).Execute()

delete project ConfigMap

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
	configMapName := "configMapName_example" // string | The Config Map name.

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.ConfigMapsAPI.DeleteProjectsConfigmapV1(context.Background(), clusterUuid, namespace, configMapName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.DeleteProjectsConfigmapV1``: %v\n", err)
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
**configMapName** | **string** | The Config Map name. | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteProjectsConfigmapV1Request struct via the builder pattern


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


## ListConfigMaps

> ConfigMapsListResponse ListConfigMaps(ctx).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Execute()

list details of ConfigMaps

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapsAPI.ListConfigMaps(context.Background()).ScopeType(scopeType).ScopeId(scopeId).Name(name).ClusterId(clusterId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.ListConfigMaps``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `ListConfigMaps`: ConfigMapsListResponse
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapsAPI.ListConfigMaps`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiListConfigMapsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **scopeType** | **string** | The type of resource we want to filter by. | 
 **scopeId** | **string** | filter by scope id | 
 **name** | **string** | filter by name | 
 **clusterId** | **string** | Filter using the Universally Unique Identifier (UUID) of the cluster. | 

### Return type

[**ConfigMapsListResponse**](ConfigMapsListResponse.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncClusterWideConfigmapsV1

> ConfigMaps SyncClusterWideConfigmapsV1(ctx, clusterUuid).ConfigMaps(configMaps).Execute()

sync the details of cluster-wide ConfigMaps to the control plane

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
	configMaps := *openapiclient.NewConfigMaps([]openapiclient.ConfigMap{*openapiclient.NewConfigMap("cm-setting")}) // ConfigMaps | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapsAPI.SyncClusterWideConfigmapsV1(context.Background(), clusterUuid).ConfigMaps(configMaps).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.SyncClusterWideConfigmapsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncClusterWideConfigmapsV1`: ConfigMaps
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapsAPI.SyncClusterWideConfigmapsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncClusterWideConfigmapsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **configMaps** | [**ConfigMaps**](ConfigMaps.md) |  | 

### Return type

[**ConfigMaps**](ConfigMaps.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncDepartmentsWideConfigmapV1

> ConfigMaps SyncDepartmentsWideConfigmapV1(ctx, clusterUuid, departmentId).ConfigMaps(configMaps).Execute()

sync the details of department-wide ConfigMaps to the control plane

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
	configMaps := *openapiclient.NewConfigMaps([]openapiclient.ConfigMap{*openapiclient.NewConfigMap("cm-setting")}) // ConfigMaps | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapsAPI.SyncDepartmentsWideConfigmapV1(context.Background(), clusterUuid, departmentId).ConfigMaps(configMaps).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.SyncDepartmentsWideConfigmapV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncDepartmentsWideConfigmapV1`: ConfigMaps
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapsAPI.SyncDepartmentsWideConfigmapV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**departmentId** | **string** | The id of the department. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncDepartmentsWideConfigmapV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **configMaps** | [**ConfigMaps**](ConfigMaps.md) |  | 

### Return type

[**ConfigMaps**](ConfigMaps.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncProjectsConfigmapsV1

> ConfigMaps SyncProjectsConfigmapsV1(ctx, clusterUuid, namespace).ConfigMaps(configMaps).Execute()

sync the details of project ConfigMaps to the control plane

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
	configMaps := *openapiclient.NewConfigMaps([]openapiclient.ConfigMap{*openapiclient.NewConfigMap("cm-setting")}) // ConfigMaps | 

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.ConfigMapsAPI.SyncProjectsConfigmapsV1(context.Background(), clusterUuid, namespace).ConfigMaps(configMaps).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `ConfigMapsAPI.SyncProjectsConfigmapsV1``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SyncProjectsConfigmapsV1`: ConfigMaps
	fmt.Fprintf(os.Stdout, "Response from `ConfigMapsAPI.SyncProjectsConfigmapsV1`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**clusterUuid** | **string** | The Universally Unique Identifier (UUID) of the cluster. | 
**namespace** | **string** | The namespace of the project. | 

### Other Parameters

Other parameters are passed through a pointer to a apiSyncProjectsConfigmapsV1Request struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **configMaps** | [**ConfigMaps**](ConfigMaps.md) |  | 

### Return type

[**ConfigMaps**](ConfigMaps.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

