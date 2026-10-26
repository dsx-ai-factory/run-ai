# ClaimInfoUpdatable

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | Pointer to **NullableString** | Requested size for the PVC. Mandatory when existingPvc is false. Recommended sizes: TB/GB/MB/TIB/GIB/MIB | [optional] 
**AddedAttrValues** | Pointer to [**[]PvcAddedAttrValue**](PvcAddedAttrValue.md) | an optional array of key-values pairs that are written as annotations on the created PVC. the allowed attributes are determined according to the storage class configuration (see k8s-objects-tracker for further info). | [optional] 

## Methods

### NewClaimInfoUpdatable

`func NewClaimInfoUpdatable() *ClaimInfoUpdatable`

NewClaimInfoUpdatable instantiates a new ClaimInfoUpdatable object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClaimInfoUpdatableWithDefaults

`func NewClaimInfoUpdatableWithDefaults() *ClaimInfoUpdatable`

NewClaimInfoUpdatableWithDefaults instantiates a new ClaimInfoUpdatable object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSize

`func (o *ClaimInfoUpdatable) GetSize() string`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *ClaimInfoUpdatable) GetSizeOk() (*string, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *ClaimInfoUpdatable) SetSize(v string)`

SetSize sets Size field to given value.

### HasSize

`func (o *ClaimInfoUpdatable) HasSize() bool`

HasSize returns a boolean if a field has been set.

### SetSizeNil

`func (o *ClaimInfoUpdatable) SetSizeNil(b bool)`

 SetSizeNil sets the value for Size to be an explicit nil

### UnsetSize
`func (o *ClaimInfoUpdatable) UnsetSize()`

UnsetSize ensures that no value is present for Size, not even an explicit nil
### GetAddedAttrValues

`func (o *ClaimInfoUpdatable) GetAddedAttrValues() []PvcAddedAttrValue`

GetAddedAttrValues returns the AddedAttrValues field if non-nil, zero value otherwise.

### GetAddedAttrValuesOk

`func (o *ClaimInfoUpdatable) GetAddedAttrValuesOk() (*[]PvcAddedAttrValue, bool)`

GetAddedAttrValuesOk returns a tuple with the AddedAttrValues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAddedAttrValues

`func (o *ClaimInfoUpdatable) SetAddedAttrValues(v []PvcAddedAttrValue)`

SetAddedAttrValues sets AddedAttrValues field to given value.

### HasAddedAttrValues

`func (o *ClaimInfoUpdatable) HasAddedAttrValues() bool`

HasAddedAttrValues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


