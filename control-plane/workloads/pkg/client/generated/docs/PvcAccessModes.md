# PvcAccessModes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ReadWriteOnce** | Pointer to **NullableBool** | Mount the volume as read/write by a single node. | [optional] [default to true]
**ReadOnlyMany** | Pointer to **NullableBool** | Mount the volume as read-only by many nodes. | [optional] [default to false]
**ReadWriteMany** | Pointer to **NullableBool** | Mount the volume as read/write by many nodes. | [optional] [default to false]

## Methods

### NewPvcAccessModes

`func NewPvcAccessModes() *PvcAccessModes`

NewPvcAccessModes instantiates a new PvcAccessModes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPvcAccessModesWithDefaults

`func NewPvcAccessModesWithDefaults() *PvcAccessModes`

NewPvcAccessModesWithDefaults instantiates a new PvcAccessModes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReadWriteOnce

`func (o *PvcAccessModes) GetReadWriteOnce() bool`

GetReadWriteOnce returns the ReadWriteOnce field if non-nil, zero value otherwise.

### GetReadWriteOnceOk

`func (o *PvcAccessModes) GetReadWriteOnceOk() (*bool, bool)`

GetReadWriteOnceOk returns a tuple with the ReadWriteOnce field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadWriteOnce

`func (o *PvcAccessModes) SetReadWriteOnce(v bool)`

SetReadWriteOnce sets ReadWriteOnce field to given value.

### HasReadWriteOnce

`func (o *PvcAccessModes) HasReadWriteOnce() bool`

HasReadWriteOnce returns a boolean if a field has been set.

### SetReadWriteOnceNil

`func (o *PvcAccessModes) SetReadWriteOnceNil(b bool)`

 SetReadWriteOnceNil sets the value for ReadWriteOnce to be an explicit nil

### UnsetReadWriteOnce
`func (o *PvcAccessModes) UnsetReadWriteOnce()`

UnsetReadWriteOnce ensures that no value is present for ReadWriteOnce, not even an explicit nil
### GetReadOnlyMany

`func (o *PvcAccessModes) GetReadOnlyMany() bool`

GetReadOnlyMany returns the ReadOnlyMany field if non-nil, zero value otherwise.

### GetReadOnlyManyOk

`func (o *PvcAccessModes) GetReadOnlyManyOk() (*bool, bool)`

GetReadOnlyManyOk returns a tuple with the ReadOnlyMany field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyMany

`func (o *PvcAccessModes) SetReadOnlyMany(v bool)`

SetReadOnlyMany sets ReadOnlyMany field to given value.

### HasReadOnlyMany

`func (o *PvcAccessModes) HasReadOnlyMany() bool`

HasReadOnlyMany returns a boolean if a field has been set.

### SetReadOnlyManyNil

`func (o *PvcAccessModes) SetReadOnlyManyNil(b bool)`

 SetReadOnlyManyNil sets the value for ReadOnlyMany to be an explicit nil

### UnsetReadOnlyMany
`func (o *PvcAccessModes) UnsetReadOnlyMany()`

UnsetReadOnlyMany ensures that no value is present for ReadOnlyMany, not even an explicit nil
### GetReadWriteMany

`func (o *PvcAccessModes) GetReadWriteMany() bool`

GetReadWriteMany returns the ReadWriteMany field if non-nil, zero value otherwise.

### GetReadWriteManyOk

`func (o *PvcAccessModes) GetReadWriteManyOk() (*bool, bool)`

GetReadWriteManyOk returns a tuple with the ReadWriteMany field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadWriteMany

`func (o *PvcAccessModes) SetReadWriteMany(v bool)`

SetReadWriteMany sets ReadWriteMany field to given value.

### HasReadWriteMany

`func (o *PvcAccessModes) HasReadWriteMany() bool`

HasReadWriteMany returns a boolean if a field has been set.

### SetReadWriteManyNil

`func (o *PvcAccessModes) SetReadWriteManyNil(b bool)`

 SetReadWriteManyNil sets the value for ReadWriteMany to be an explicit nil

### UnsetReadWriteMany
`func (o *PvcAccessModes) UnsetReadWriteMany()`

UnsetReadWriteMany ensures that no value is present for ReadWriteMany, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


