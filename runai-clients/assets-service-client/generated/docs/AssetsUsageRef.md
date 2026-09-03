# AssetsUsageRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Environment** | Pointer to [**NullableEnvironmentAssetRef**](EnvironmentAssetRef.md) | environment asset. | [optional] 
**Environments** | Pointer to [**[]AssetRef**](AssetRef.md) |  | [optional] 
**Compute** | Pointer to [**NullableAssetsUsageRefCompute**](AssetsUsageRefCompute.md) |  | [optional] 
**Computes** | Pointer to [**[]AssetRef**](AssetRef.md) |  | [optional] 
**Datasources** | Pointer to [**[]AssetDatasourceRef**](AssetDatasourceRef.md) |  | [optional] 

## Methods

### NewAssetsUsageRef

`func NewAssetsUsageRef() *AssetsUsageRef`

NewAssetsUsageRef instantiates a new AssetsUsageRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsUsageRefWithDefaults

`func NewAssetsUsageRefWithDefaults() *AssetsUsageRef`

NewAssetsUsageRefWithDefaults instantiates a new AssetsUsageRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnvironment

`func (o *AssetsUsageRef) GetEnvironment() EnvironmentAssetRef`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *AssetsUsageRef) GetEnvironmentOk() (*EnvironmentAssetRef, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *AssetsUsageRef) SetEnvironment(v EnvironmentAssetRef)`

SetEnvironment sets Environment field to given value.

### HasEnvironment

`func (o *AssetsUsageRef) HasEnvironment() bool`

HasEnvironment returns a boolean if a field has been set.

### SetEnvironmentNil

`func (o *AssetsUsageRef) SetEnvironmentNil(b bool)`

 SetEnvironmentNil sets the value for Environment to be an explicit nil

### UnsetEnvironment
`func (o *AssetsUsageRef) UnsetEnvironment()`

UnsetEnvironment ensures that no value is present for Environment, not even an explicit nil
### GetEnvironments

`func (o *AssetsUsageRef) GetEnvironments() []AssetRef`

GetEnvironments returns the Environments field if non-nil, zero value otherwise.

### GetEnvironmentsOk

`func (o *AssetsUsageRef) GetEnvironmentsOk() (*[]AssetRef, bool)`

GetEnvironmentsOk returns a tuple with the Environments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironments

`func (o *AssetsUsageRef) SetEnvironments(v []AssetRef)`

SetEnvironments sets Environments field to given value.

### HasEnvironments

`func (o *AssetsUsageRef) HasEnvironments() bool`

HasEnvironments returns a boolean if a field has been set.

### GetCompute

`func (o *AssetsUsageRef) GetCompute() AssetsUsageRefCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *AssetsUsageRef) GetComputeOk() (*AssetsUsageRefCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *AssetsUsageRef) SetCompute(v AssetsUsageRefCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *AssetsUsageRef) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *AssetsUsageRef) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *AssetsUsageRef) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetComputes

`func (o *AssetsUsageRef) GetComputes() []AssetRef`

GetComputes returns the Computes field if non-nil, zero value otherwise.

### GetComputesOk

`func (o *AssetsUsageRef) GetComputesOk() (*[]AssetRef, bool)`

GetComputesOk returns a tuple with the Computes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComputes

`func (o *AssetsUsageRef) SetComputes(v []AssetRef)`

SetComputes sets Computes field to given value.

### HasComputes

`func (o *AssetsUsageRef) HasComputes() bool`

HasComputes returns a boolean if a field has been set.

### GetDatasources

`func (o *AssetsUsageRef) GetDatasources() []AssetDatasourceRef`

GetDatasources returns the Datasources field if non-nil, zero value otherwise.

### GetDatasourcesOk

`func (o *AssetsUsageRef) GetDatasourcesOk() (*[]AssetDatasourceRef, bool)`

GetDatasourcesOk returns a tuple with the Datasources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasources

`func (o *AssetsUsageRef) SetDatasources(v []AssetDatasourceRef)`

SetDatasources sets Datasources field to given value.

### HasDatasources

`func (o *AssetsUsageRef) HasDatasources() bool`

HasDatasources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


