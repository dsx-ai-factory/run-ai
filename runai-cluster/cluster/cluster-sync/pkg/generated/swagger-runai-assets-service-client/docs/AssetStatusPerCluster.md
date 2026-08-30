# AssetStatusPerCluster

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Clusters** | Pointer to [**[]AssetClusterStatus**](AssetClusterStatus.md) |  | [optional] 

## Methods

### NewAssetStatusPerCluster

`func NewAssetStatusPerCluster() *AssetStatusPerCluster`

NewAssetStatusPerCluster instantiates a new AssetStatusPerCluster object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetStatusPerClusterWithDefaults

`func NewAssetStatusPerClusterWithDefaults() *AssetStatusPerCluster`

NewAssetStatusPerClusterWithDefaults instantiates a new AssetStatusPerCluster object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusters

`func (o *AssetStatusPerCluster) GetClusters() []AssetClusterStatus`

GetClusters returns the Clusters field if non-nil, zero value otherwise.

### GetClustersOk

`func (o *AssetStatusPerCluster) GetClustersOk() (*[]AssetClusterStatus, bool)`

GetClustersOk returns a tuple with the Clusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusters

`func (o *AssetStatusPerCluster) SetClusters(v []AssetClusterStatus)`

SetClusters sets Clusters field to given value.

### HasClusters

`func (o *AssetStatusPerCluster) HasClusters() bool`

HasClusters returns a boolean if a field has been set.

### SetClustersNil

`func (o *AssetStatusPerCluster) SetClustersNil(b bool)`

 SetClustersNil sets the value for Clusters to be an explicit nil

### UnsetClusters
`func (o *AssetStatusPerCluster) UnsetClusters()`

UnsetClusters ensures that no value is present for Clusters, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


