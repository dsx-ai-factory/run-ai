# PvcSupportedAccessModes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ReadWriteOnce** | Pointer to **NullableBool** | Indicates whether the readWriteOnce access mode is supported. This mode allows a volume to be mounted as read/write by a single node. | [optional] [default to true]
**ReadOnlyMany** | Pointer to **NullableBool** | Indicates whether the readOnlyMany access mode is supported. This mode allows a volume to be mounted as read-only by many nodes. | [optional] [default to true]
**ReadWriteMany** | Pointer to **NullableBool** | Indicates whether the readWriteMany access mode is supported. This mode allows a volume to be mounted as read/write by many nodes. | [optional] [default to true]

## Methods

### NewPvcSupportedAccessModes

`func NewPvcSupportedAccessModes() *PvcSupportedAccessModes`

NewPvcSupportedAccessModes instantiates a new PvcSupportedAccessModes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPvcSupportedAccessModesWithDefaults

`func NewPvcSupportedAccessModesWithDefaults() *PvcSupportedAccessModes`

NewPvcSupportedAccessModesWithDefaults instantiates a new PvcSupportedAccessModes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReadWriteOnce

`func (o *PvcSupportedAccessModes) GetReadWriteOnce() bool`

GetReadWriteOnce returns the ReadWriteOnce field if non-nil, zero value otherwise.

### GetReadWriteOnceOk

`func (o *PvcSupportedAccessModes) GetReadWriteOnceOk() (*bool, bool)`

GetReadWriteOnceOk returns a tuple with the ReadWriteOnce field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadWriteOnce

`func (o *PvcSupportedAccessModes) SetReadWriteOnce(v bool)`

SetReadWriteOnce sets ReadWriteOnce field to given value.

### HasReadWriteOnce

`func (o *PvcSupportedAccessModes) HasReadWriteOnce() bool`

HasReadWriteOnce returns a boolean if a field has been set.

### SetReadWriteOnceNil

`func (o *PvcSupportedAccessModes) SetReadWriteOnceNil(b bool)`

 SetReadWriteOnceNil sets the value for ReadWriteOnce to be an explicit nil

### UnsetReadWriteOnce
`func (o *PvcSupportedAccessModes) UnsetReadWriteOnce()`

UnsetReadWriteOnce ensures that no value is present for ReadWriteOnce, not even an explicit nil
### GetReadOnlyMany

`func (o *PvcSupportedAccessModes) GetReadOnlyMany() bool`

GetReadOnlyMany returns the ReadOnlyMany field if non-nil, zero value otherwise.

### GetReadOnlyManyOk

`func (o *PvcSupportedAccessModes) GetReadOnlyManyOk() (*bool, bool)`

GetReadOnlyManyOk returns a tuple with the ReadOnlyMany field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnlyMany

`func (o *PvcSupportedAccessModes) SetReadOnlyMany(v bool)`

SetReadOnlyMany sets ReadOnlyMany field to given value.

### HasReadOnlyMany

`func (o *PvcSupportedAccessModes) HasReadOnlyMany() bool`

HasReadOnlyMany returns a boolean if a field has been set.

### SetReadOnlyManyNil

`func (o *PvcSupportedAccessModes) SetReadOnlyManyNil(b bool)`

 SetReadOnlyManyNil sets the value for ReadOnlyMany to be an explicit nil

### UnsetReadOnlyMany
`func (o *PvcSupportedAccessModes) UnsetReadOnlyMany()`

UnsetReadOnlyMany ensures that no value is present for ReadOnlyMany, not even an explicit nil
### GetReadWriteMany

`func (o *PvcSupportedAccessModes) GetReadWriteMany() bool`

GetReadWriteMany returns the ReadWriteMany field if non-nil, zero value otherwise.

### GetReadWriteManyOk

`func (o *PvcSupportedAccessModes) GetReadWriteManyOk() (*bool, bool)`

GetReadWriteManyOk returns a tuple with the ReadWriteMany field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadWriteMany

`func (o *PvcSupportedAccessModes) SetReadWriteMany(v bool)`

SetReadWriteMany sets ReadWriteMany field to given value.

### HasReadWriteMany

`func (o *PvcSupportedAccessModes) HasReadWriteMany() bool`

HasReadWriteMany returns a boolean if a field has been set.

### SetReadWriteManyNil

`func (o *PvcSupportedAccessModes) SetReadWriteManyNil(b bool)`

 SetReadWriteManyNil sets the value for ReadWriteMany to be an explicit nil

### UnsetReadWriteMany
`func (o *PvcSupportedAccessModes) UnsetReadWriteMany()`

UnsetReadWriteMany ensures that no value is present for ReadWriteMany, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


