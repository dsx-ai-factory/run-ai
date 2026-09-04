# AssetsUsageRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Environment** | Pointer to [**NullableAssetsUsageRefEnvironment**](AssetsUsageRefEnvironment.md) |  | [optional] 
**Compute** | Pointer to [**NullableAssetsRefCompute**](AssetsRefCompute.md) |  | [optional] 
**Datasources** | Pointer to [**[]DatasourceRef**](DatasourceRef.md) |  | [optional] 

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

`func (o *AssetsUsageRef) GetEnvironment() AssetsUsageRefEnvironment`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *AssetsUsageRef) GetEnvironmentOk() (*AssetsUsageRefEnvironment, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *AssetsUsageRef) SetEnvironment(v AssetsUsageRefEnvironment)`

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
### GetCompute

`func (o *AssetsUsageRef) GetCompute() AssetsRefCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *AssetsUsageRef) GetComputeOk() (*AssetsRefCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *AssetsUsageRef) SetCompute(v AssetsRefCompute)`

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
### GetDatasources

`func (o *AssetsUsageRef) GetDatasources() []DatasourceRef`

GetDatasources returns the Datasources field if non-nil, zero value otherwise.

### GetDatasourcesOk

`func (o *AssetsUsageRef) GetDatasourcesOk() (*[]DatasourceRef, bool)`

GetDatasourcesOk returns a tuple with the Datasources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatasources

`func (o *AssetsUsageRef) SetDatasources(v []DatasourceRef)`

SetDatasources sets Datasources field to given value.

### HasDatasources

`func (o *AssetsUsageRef) HasDatasources() bool`

HasDatasources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


