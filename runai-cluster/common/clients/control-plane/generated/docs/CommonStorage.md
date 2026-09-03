# CommonStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Storage** | Pointer to [**NullableCommonStorageFields**](CommonStorageFields.md) |  | [optional] 

## Methods

### NewCommonStorage

`func NewCommonStorage() *CommonStorage`

NewCommonStorage instantiates a new CommonStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonStorageWithDefaults

`func NewCommonStorageWithDefaults() *CommonStorage`

NewCommonStorageWithDefaults instantiates a new CommonStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStorage

`func (o *CommonStorage) GetStorage() CommonStorageFields`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *CommonStorage) GetStorageOk() (*CommonStorageFields, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *CommonStorage) SetStorage(v CommonStorageFields)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *CommonStorage) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *CommonStorage) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *CommonStorage) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


