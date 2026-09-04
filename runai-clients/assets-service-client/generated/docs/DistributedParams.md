# DistributedParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumWorkers** | **int32** | number of worker replicas to process the workload. | 
**Framework** | Pointer to [**NullableDistributedFramework**](DistributedFramework.md) |  | [optional] 
**Master** | Pointer to [**NullableDistributedMaster**](DistributedMaster.md) |  | [optional] 

## Methods

### NewDistributedParams

`func NewDistributedParams(numWorkers int32, ) *DistributedParams`

NewDistributedParams instantiates a new DistributedParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedParamsWithDefaults

`func NewDistributedParamsWithDefaults() *DistributedParams`

NewDistributedParamsWithDefaults instantiates a new DistributedParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumWorkers

`func (o *DistributedParams) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *DistributedParams) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *DistributedParams) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.


### GetFramework

`func (o *DistributedParams) GetFramework() DistributedFramework`

GetFramework returns the Framework field if non-nil, zero value otherwise.

### GetFrameworkOk

`func (o *DistributedParams) GetFrameworkOk() (*DistributedFramework, bool)`

GetFrameworkOk returns a tuple with the Framework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFramework

`func (o *DistributedParams) SetFramework(v DistributedFramework)`

SetFramework sets Framework field to given value.

### HasFramework

`func (o *DistributedParams) HasFramework() bool`

HasFramework returns a boolean if a field has been set.

### SetFrameworkNil

`func (o *DistributedParams) SetFrameworkNil(b bool)`

 SetFrameworkNil sets the value for Framework to be an explicit nil

### UnsetFramework
`func (o *DistributedParams) UnsetFramework()`

UnsetFramework ensures that no value is present for Framework, not even an explicit nil
### GetMaster

`func (o *DistributedParams) GetMaster() DistributedMaster`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *DistributedParams) GetMasterOk() (*DistributedMaster, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *DistributedParams) SetMaster(v DistributedMaster)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *DistributedParams) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *DistributedParams) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *DistributedParams) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


