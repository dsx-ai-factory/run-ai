# {{classname}}

All URIs are relative to *https://app.run.ai*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateCluster**](ClustersApi.md#CreateCluster) | **Post** /v1/k8s/clusters | Create a cluster
[**InternalClustersUuidSubResourcesDelete**](ClustersApi.md#InternalClustersUuidSubResourcesDelete) | **Delete** /internal/clusters/{uuid}/sub-resources | Delete all cluster sub-resources
[**InternalClustersUuidSubResourcesPost**](ClustersApi.md#InternalClustersUuidSubResourcesPost) | **Post** /internal/clusters/{uuid}/sub-resources | Create all cluster sub-resources
[**UpdateCluster**](ClustersApi.md#UpdateCluster) | **Put** /v1/k8s/clusters/{uuid} | Update a Cluster by Id
[**UpdateClusterInfo**](ClustersApi.md#UpdateClusterInfo) | **Post** /v1/k8s/clusters/{uuid}/cluster-info | Send cluster-info. This endpoint is reserved for internal run:ai purposes.
[**UpdateClusterLiveness**](ClustersApi.md#UpdateClusterLiveness) | **Post** /v1/k8s/clusters/{uuid}/liveness | Update liveness status for a specific cluster. This endpoint is reserved for internal run:ai purposes.
[**V1K8sClustersClusterUuidConfigControlPlaneGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigControlPlaneGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config/control-plane | Returns cluster&#x27;s control plane configuration
[**V1K8sClustersClusterUuidConfigCustomGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigCustomGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config/custom | Returns custom cluster configuration
[**V1K8sClustersClusterUuidConfigCustomPost**](ClustersApi.md#V1K8sClustersClusterUuidConfigCustomPost) | **Post** /v1/k8s/clusters/{cluster_uuid}/config/custom | Update custom cluster configuration
[**V1K8sClustersClusterUuidConfigGet**](ClustersApi.md#V1K8sClustersClusterUuidConfigGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/config | Returns cluster configuration
[**V1K8sClustersClusterUuidInstallfileGet**](ClustersApi.md#V1K8sClustersClusterUuidInstallfileGet) | **Get** /v1/k8s/clusters/{cluster_uuid}/installfile | Returns cluster installation file by Id
[**V1K8sClustersGet**](ClustersApi.md#V1K8sClustersGet) | **Get** /v1/k8s/clusters | List clusters.
[**V1K8sClustersUuidDelete**](ClustersApi.md#V1K8sClustersUuidDelete) | **Delete** /v1/k8s/clusters/{uuid} | Delete a Cluster
[**V1K8sClustersUuidGet**](ClustersApi.md#V1K8sClustersUuidGet) | **Get** /v1/k8s/clusters/{uuid} | Get cluster by id

# **CreateCluster**
> Cluster CreateCluster(ctx, body)
Create a cluster

Create a new cluster for the tenant. This endpoint requires ADMIN role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CreateCluster**](CreateCluster.md)|  | 

### Return type

[**Cluster**](Cluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **InternalClustersUuidSubResourcesDelete**
> InternalClustersUuidSubResourcesDelete(ctx, uuid)
Delete all cluster sub-resources

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **InternalClustersUuidSubResourcesPost**
> InternalClustersUuidSubResourcesPost(ctx, body, uuid)
Create all cluster sub-resources

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ClusterMetaData**](ClusterMetaData.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateCluster**
> Cluster UpdateCluster(ctx, body, uuid)
Update a Cluster by Id

Update the details of a given cluster. This endpoint requires ADMIN role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Cluster**](Cluster.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

[**Cluster**](Cluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateClusterInfo**
> UpdateClusterInfo(ctx, body, uuid)
Send cluster-info. This endpoint is reserved for internal run:ai purposes.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ClusterInfo**](ClusterInfo.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateClusterLiveness**
> UpdateClusterLiveness(ctx, body, uuid)
Update liveness status for a specific cluster. This endpoint is reserved for internal run:ai purposes.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**UuidLivenessBody**](UuidLivenessBody.md)|  | 
  **uuid** | [**string**](.md)| Unique identifier of the cluster | 

### Return type

 (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterUuidConfigControlPlaneGet**
> ClusterControlPlaneConfig V1K8sClustersClusterUuidConfigControlPlaneGet(ctx, clusterUuid)
Returns cluster's control plane configuration

Retrieve the control plane configuration of a given cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterUuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

[**ClusterControlPlaneConfig**](ClusterControlPlaneConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterUuidConfigCustomGet**
> ClusterCustomConfig V1K8sClustersClusterUuidConfigCustomGet(ctx, clusterUuid)
Returns custom cluster configuration

Retrieve custom configuration of a given cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterUuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

[**ClusterCustomConfig**](ClusterCustomConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterUuidConfigCustomPost**
> ClusterCustomConfig V1K8sClustersClusterUuidConfigCustomPost(ctx, body, clusterUuid)
Update custom cluster configuration

Update custom configuration of a given cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ClusterCustomConfig**](ClusterCustomConfig.md)|  | 
  **clusterUuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

[**ClusterCustomConfig**](ClusterCustomConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterUuidConfigGet**
> ClusterConfig V1K8sClustersClusterUuidConfigGet(ctx, clusterUuid, optional)
Returns cluster configuration

Retrieve the configuration of a given cluster

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterUuid** | [**string**](.md)| Unique identifier of the cluster. | 
 **optional** | ***ClustersApiV1K8sClustersClusterUuidConfigGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ClustersApiV1K8sClustersClusterUuidConfigGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **clusterDomain** | **optional.String**| The domain of the cluster. | 
 **clusterVersion** | **optional.String**| The version of the configuration. | 

### Return type

[**ClusterConfig**](ClusterConfig.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersClusterUuidInstallfileGet**
> string V1K8sClustersClusterUuidInstallfileGet(ctx, clusterUuid, optional)
Returns cluster installation file by Id

Retrieve the installation values file of a given cluster by ID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **clusterUuid** | [**string**](.md)| Unique identifier of the cluster. | 
 **optional** | ***ClustersApiV1K8sClustersClusterUuidInstallfileGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ClustersApiV1K8sClustersClusterUuidInstallfileGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **cloud** | **optional.String**| Cloud type identifier. | 
 **clusterip** | **optional.String**| Comma-separated list of IP addresses that provide access to the cluster. | 
 **format** | **optional.String**| Format of the output file. | [default to yaml]

### Return type

**string**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersGet**
> []Cluster V1K8sClustersGet(ctx, )
List clusters.

Get a list of all the clusters of the tenant.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]Cluster**](Cluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersUuidDelete**
> InlineResponse200 V1K8sClustersUuidDelete(ctx, uuid)
Delete a Cluster

Delete a Cluster by Cluster ID. This endpoint requires ADMIN role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

[**InlineResponse200**](inline_response_200.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **V1K8sClustersUuidGet**
> Cluster V1K8sClustersUuidGet(ctx, uuid)
Get cluster by id

Retrieve a cluster given a cluster id. This endpoint requires ADMIN role.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **uuid** | [**string**](.md)| Unique identifier of the cluster. | 

### Return type

[**Cluster**](Cluster.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

