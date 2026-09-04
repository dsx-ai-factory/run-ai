# AssetsRefOptional

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Environment** | [**AssetsRefEnvironment**](AssetsRefEnvironment.md) |  | 
**Compute** | Pointer to [**NullableAssetsRefCompute**](AssetsRefCompute.md) |  | [optional] 
**Datasources** | Pointer to [**[]DatasourceRef**](DatasourceRef.md) |  | [optional] 
**WorkloadVolumes** | Pointer to **[]string** |  | [optional] 

## Methods

### NewAssetsRefOptional

`func NewAssetsRefOptional(environment AssetsRefEnvironment, ) *AssetsRefOptional`

NewAssetsRefOptional instantiates a new AssetsRefOptional object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsRefOptionalWithDefaults

`func NewAssetsRefOptionalWithDefaults() *AssetsRefOptional`

NewAssetsRefOptionalWithDefaults instantiates a new AssetsRefOptional object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnvironment

`func (o *AssetsRefOptional) GetEnvironment() AssetsRefEnvironment`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *AssetsRefOptional) GetEnvironmentOk() (*AssetsRefEnvironment, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *AssetsRefOptional) SetEnvironment(v AssetsRefEnvironment)`

SetEnvironment sets Environment field to given value.


### GetCompute

`func (o *AssetsRefOptional) GetCompute() AssetsRefCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *AssetsRefOptional) GetComputeOk() (*AssetsRefCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *AssetsRefOptional) SetCompute(v AssetsRefCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *AssetsRefOptional) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *AssetsRefOptional) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *AssetsRefOptional) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetDatasources

`func (o *AssetsRefOptional) GetDatasources() []DatasourceRef`

GetDatasources returns the Datasources field if non-nil, zero value otherwise.

### GetDatasourcesOk

`func (o *AssetsRefOptional) GetDatasourcesOk() (*[]DatasourceRef, bool)`

GetDatasourcesOk returns a tuple with the Datasources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasources

`func (o *AssetsRefOptional) SetDatasources(v []DatasourceRef)`

SetDatasources sets Datasources field to given value.

### HasDatasources

`func (o *AssetsRefOptional) HasDatasources() bool`

HasDatasources returns a boolean if a field has been set.

### GetWorkloadVolumes

`func (o *AssetsRefOptional) GetWorkloadVolumes() []string`

GetWorkloadVolumes returns the WorkloadVolumes field if non-nil, zero value otherwise.

### GetWorkloadVolumesOk

`func (o *AssetsRefOptional) GetWorkloadVolumesOk() (*[]string, bool)`

GetWorkloadVolumesOk returns a tuple with the WorkloadVolumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadVolumes

`func (o *AssetsRefOptional) SetWorkloadVolumes(v []string)`

SetWorkloadVolumes sets WorkloadVolumes field to given value.

### HasWorkloadVolumes

`func (o *AssetsRefOptional) HasWorkloadVolumes() bool`

HasWorkloadVolumes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


