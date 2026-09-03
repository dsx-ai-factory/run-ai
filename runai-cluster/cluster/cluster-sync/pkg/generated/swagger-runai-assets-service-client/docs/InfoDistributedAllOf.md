# InfoDistributedAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NumWorkers** | Pointer to **int32** |  | [optional] 
**NoMaster** | Pointer to **NullableBool** |  | [optional] 
**DistFramework** | Pointer to [**DistFramework**](DistFramework.md) |  | [optional] 
**Master** | Pointer to [**NullableInfoDistMaster**](InfoDistMaster.md) |  | [optional] 

## Methods

### NewInfoDistributedAllOf

`func NewInfoDistributedAllOf() *InfoDistributedAllOf`

NewInfoDistributedAllOf instantiates a new InfoDistributedAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInfoDistributedAllOfWithDefaults

`func NewInfoDistributedAllOfWithDefaults() *InfoDistributedAllOf`

NewInfoDistributedAllOfWithDefaults instantiates a new InfoDistributedAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumWorkers

`func (o *InfoDistributedAllOf) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *InfoDistributedAllOf) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *InfoDistributedAllOf) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *InfoDistributedAllOf) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### GetNoMaster

`func (o *InfoDistributedAllOf) GetNoMaster() bool`

GetNoMaster returns the NoMaster field if non-nil, zero value otherwise.

### GetNoMasterOk

`func (o *InfoDistributedAllOf) GetNoMasterOk() (*bool, bool)`

GetNoMasterOk returns a tuple with the NoMaster field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNoMaster

`func (o *InfoDistributedAllOf) SetNoMaster(v bool)`

SetNoMaster sets NoMaster field to given value.

### HasNoMaster

`func (o *InfoDistributedAllOf) HasNoMaster() bool`

HasNoMaster returns a boolean if a field has been set.

### SetNoMasterNil

`func (o *InfoDistributedAllOf) SetNoMasterNil(b bool)`

 SetNoMasterNil sets the value for NoMaster to be an explicit nil

### UnsetNoMaster
`func (o *InfoDistributedAllOf) UnsetNoMaster()`

UnsetNoMaster ensures that no value is present for NoMaster, not even an explicit nil
### GetDistFramework

`func (o *InfoDistributedAllOf) GetDistFramework() DistFramework`

GetDistFramework returns the DistFramework field if non-nil, zero value otherwise.

### GetDistFrameworkOk

`func (o *InfoDistributedAllOf) GetDistFrameworkOk() (*DistFramework, bool)`

GetDistFrameworkOk returns a tuple with the DistFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistFramework

`func (o *InfoDistributedAllOf) SetDistFramework(v DistFramework)`

SetDistFramework sets DistFramework field to given value.

### HasDistFramework

`func (o *InfoDistributedAllOf) HasDistFramework() bool`

HasDistFramework returns a boolean if a field has been set.

### GetMaster

`func (o *InfoDistributedAllOf) GetMaster() InfoDistMaster`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *InfoDistributedAllOf) GetMasterOk() (*InfoDistMaster, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *InfoDistributedAllOf) SetMaster(v InfoDistMaster)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *InfoDistributedAllOf) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *InfoDistributedAllOf) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *InfoDistributedAllOf) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


