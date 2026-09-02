# DatavolumeStatusInternalAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IsDeleted** | Pointer to **NullableBool** | If true, the datavolume is deleted in the cluster and &#39;phase&#39; field is irrelevant. | [optional] 

## Methods

### NewDatavolumeStatusInternalAllOf

`func NewDatavolumeStatusInternalAllOf() *DatavolumeStatusInternalAllOf`

NewDatavolumeStatusInternalAllOf instantiates a new DatavolumeStatusInternalAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeStatusInternalAllOfWithDefaults

`func NewDatavolumeStatusInternalAllOfWithDefaults() *DatavolumeStatusInternalAllOf`

NewDatavolumeStatusInternalAllOfWithDefaults instantiates a new DatavolumeStatusInternalAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIsDeleted

`func (o *DatavolumeStatusInternalAllOf) GetIsDeleted() bool`

GetIsDeleted returns the IsDeleted field if non-nil, zero value otherwise.

### GetIsDeletedOk

`func (o *DatavolumeStatusInternalAllOf) GetIsDeletedOk() (*bool, bool)`

GetIsDeletedOk returns a tuple with the IsDeleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeleted

`func (o *DatavolumeStatusInternalAllOf) SetIsDeleted(v bool)`

SetIsDeleted sets IsDeleted field to given value.

### HasIsDeleted

`func (o *DatavolumeStatusInternalAllOf) HasIsDeleted() bool`

HasIsDeleted returns a boolean if a field has been set.

### SetIsDeletedNil

`func (o *DatavolumeStatusInternalAllOf) SetIsDeletedNil(b bool)`

 SetIsDeletedNil sets the value for IsDeleted to be an explicit nil

### UnsetIsDeleted
`func (o *DatavolumeStatusInternalAllOf) UnsetIsDeleted()`

UnsetIsDeleted ensures that no value is present for IsDeleted, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


