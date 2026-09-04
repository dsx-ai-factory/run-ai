# StorageClass

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StorageClassName** | **string** | Identifier of the storage class | 
**IsDefault** | Pointer to **bool** | Indicator if the storage class is the default storage class | [optional] 

## Methods

### NewStorageClass

`func NewStorageClass(storageClassName string, ) *StorageClass`

NewStorageClass instantiates a new StorageClass object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassWithDefaults

`func NewStorageClassWithDefaults() *StorageClass`

NewStorageClassWithDefaults instantiates a new StorageClass object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStorageClassName

`func (o *StorageClass) GetStorageClassName() string`

GetStorageClassName returns the StorageClassName field if non-nil, zero value otherwise.

### GetStorageClassNameOk

`func (o *StorageClass) GetStorageClassNameOk() (*string, bool)`

GetStorageClassNameOk returns a tuple with the StorageClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassName

`func (o *StorageClass) SetStorageClassName(v string)`

SetStorageClassName sets StorageClassName field to given value.


### GetIsDefault

`func (o *StorageClass) GetIsDefault() bool`

GetIsDefault returns the IsDefault field if non-nil, zero value otherwise.

### GetIsDefaultOk

`func (o *StorageClass) GetIsDefaultOk() (*bool, bool)`

GetIsDefaultOk returns a tuple with the IsDefault field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDefault

`func (o *StorageClass) SetIsDefault(v bool)`

SetIsDefault sets IsDefault field to given value.

### HasIsDefault

`func (o *StorageClass) HasIsDefault() bool`

HasIsDefault returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


