# V1StorageClass

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StorageClassName** | **string** | Identifier of the storage class | 
**IsDefault** | Pointer to **bool** | Indicator if the storage class is the default storage class | [optional] 

## Methods

### NewV1StorageClass

`func NewV1StorageClass(storageClassName string, ) *V1StorageClass`

NewV1StorageClass instantiates a new V1StorageClass object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewV1StorageClassWithDefaults

`func NewV1StorageClassWithDefaults() *V1StorageClass`

NewV1StorageClassWithDefaults instantiates a new V1StorageClass object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStorageClassName

`func (o *V1StorageClass) GetStorageClassName() string`

GetStorageClassName returns the StorageClassName field if non-nil, zero value otherwise.

### GetStorageClassNameOk

`func (o *V1StorageClass) GetStorageClassNameOk() (*string, bool)`

GetStorageClassNameOk returns a tuple with the StorageClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassName

`func (o *V1StorageClass) SetStorageClassName(v string)`

SetStorageClassName sets StorageClassName field to given value.


### GetIsDefault

`func (o *V1StorageClass) GetIsDefault() bool`

GetIsDefault returns the IsDefault field if non-nil, zero value otherwise.

### GetIsDefaultOk

`func (o *V1StorageClass) GetIsDefaultOk() (*bool, bool)`

GetIsDefaultOk returns a tuple with the IsDefault field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDefault

`func (o *V1StorageClass) SetIsDefault(v bool)`

SetIsDefault sets IsDefault field to given value.

### HasIsDefault

`func (o *V1StorageClass) HasIsDefault() bool`

HasIsDefault returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


