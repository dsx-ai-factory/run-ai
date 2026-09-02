# AssetsIds

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Environment** | **string** | Unique identifier of the asset. | 
**Compute** | Pointer to **NullableString** | Unique identifier of the asset. | [optional] 
**Datasources** | Pointer to [**[]AssetIdAndKind**](AssetIdAndKind.md) |  | [optional] 
**WorkloadVolumes** | Pointer to **[]string** |  | [optional] 

## Methods

### NewAssetsIds

`func NewAssetsIds(environment string, ) *AssetsIds`

NewAssetsIds instantiates a new AssetsIds object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsIdsWithDefaults

`func NewAssetsIdsWithDefaults() *AssetsIds`

NewAssetsIdsWithDefaults instantiates a new AssetsIds object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnvironment

`func (o *AssetsIds) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *AssetsIds) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *AssetsIds) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.


### GetCompute

`func (o *AssetsIds) GetCompute() string`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *AssetsIds) GetComputeOk() (*string, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *AssetsIds) SetCompute(v string)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *AssetsIds) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *AssetsIds) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *AssetsIds) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetDatasources

`func (o *AssetsIds) GetDatasources() []AssetIdAndKind`

GetDatasources returns the Datasources field if non-nil, zero value otherwise.

### GetDatasourcesOk

`func (o *AssetsIds) GetDatasourcesOk() (*[]AssetIdAndKind, bool)`

GetDatasourcesOk returns a tuple with the Datasources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasources

`func (o *AssetsIds) SetDatasources(v []AssetIdAndKind)`

SetDatasources sets Datasources field to given value.

### HasDatasources

`func (o *AssetsIds) HasDatasources() bool`

HasDatasources returns a boolean if a field has been set.

### GetWorkloadVolumes

`func (o *AssetsIds) GetWorkloadVolumes() []string`

GetWorkloadVolumes returns the WorkloadVolumes field if non-nil, zero value otherwise.

### GetWorkloadVolumesOk

`func (o *AssetsIds) GetWorkloadVolumesOk() (*[]string, bool)`

GetWorkloadVolumesOk returns a tuple with the WorkloadVolumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadVolumes

`func (o *AssetsIds) SetWorkloadVolumes(v []string)`

SetWorkloadVolumes sets WorkloadVolumes field to given value.

### HasWorkloadVolumes

`func (o *AssetsIds) HasWorkloadVolumes() bool`

HasWorkloadVolumes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


