# InfoDistributed

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumWorkers** | Pointer to **int32** |  | [optional] 
**NoMaster** | Pointer to **NullableBool** |  | [optional] 
**DistFramework** | Pointer to [**NullableDistributedFramework**](DistributedFramework.md) |  | [optional] 
**Master** | Pointer to [**NullableInfoDistMaster**](InfoDistMaster.md) |  | [optional] 

## Methods

### NewInfoDistributed

`func NewInfoDistributed() *InfoDistributed`

NewInfoDistributed instantiates a new InfoDistributed object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInfoDistributedWithDefaults

`func NewInfoDistributedWithDefaults() *InfoDistributed`

NewInfoDistributedWithDefaults instantiates a new InfoDistributed object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumWorkers

`func (o *InfoDistributed) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *InfoDistributed) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *InfoDistributed) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *InfoDistributed) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### GetNoMaster

`func (o *InfoDistributed) GetNoMaster() bool`

GetNoMaster returns the NoMaster field if non-nil, zero value otherwise.

### GetNoMasterOk

`func (o *InfoDistributed) GetNoMasterOk() (*bool, bool)`

GetNoMasterOk returns a tuple with the NoMaster field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNoMaster

`func (o *InfoDistributed) SetNoMaster(v bool)`

SetNoMaster sets NoMaster field to given value.

### HasNoMaster

`func (o *InfoDistributed) HasNoMaster() bool`

HasNoMaster returns a boolean if a field has been set.

### SetNoMasterNil

`func (o *InfoDistributed) SetNoMasterNil(b bool)`

 SetNoMasterNil sets the value for NoMaster to be an explicit nil

### UnsetNoMaster
`func (o *InfoDistributed) UnsetNoMaster()`

UnsetNoMaster ensures that no value is present for NoMaster, not even an explicit nil
### GetDistFramework

`func (o *InfoDistributed) GetDistFramework() DistributedFramework`

GetDistFramework returns the DistFramework field if non-nil, zero value otherwise.

### GetDistFrameworkOk

`func (o *InfoDistributed) GetDistFrameworkOk() (*DistributedFramework, bool)`

GetDistFrameworkOk returns a tuple with the DistFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistFramework

`func (o *InfoDistributed) SetDistFramework(v DistributedFramework)`

SetDistFramework sets DistFramework field to given value.

### HasDistFramework

`func (o *InfoDistributed) HasDistFramework() bool`

HasDistFramework returns a boolean if a field has been set.

### SetDistFrameworkNil

`func (o *InfoDistributed) SetDistFrameworkNil(b bool)`

 SetDistFrameworkNil sets the value for DistFramework to be an explicit nil

### UnsetDistFramework
`func (o *InfoDistributed) UnsetDistFramework()`

UnsetDistFramework ensures that no value is present for DistFramework, not even an explicit nil
### GetMaster

`func (o *InfoDistributed) GetMaster() InfoDistMaster`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *InfoDistributed) GetMasterOk() (*InfoDistMaster, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *InfoDistributed) SetMaster(v InfoDistMaster)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *InfoDistributed) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *InfoDistributed) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *InfoDistributed) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


