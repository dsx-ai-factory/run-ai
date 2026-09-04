# StorageInstanceName

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | unique name to identify the instance. primarily used for policy locked rules. | [optional] 

## Methods

### NewStorageInstanceName

`func NewStorageInstanceName() *StorageInstanceName`

NewStorageInstanceName instantiates a new StorageInstanceName object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageInstanceNameWithDefaults

`func NewStorageInstanceNameWithDefaults() *StorageInstanceName`

NewStorageInstanceNameWithDefaults instantiates a new StorageInstanceName object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *StorageInstanceName) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *StorageInstanceName) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *StorageInstanceName) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *StorageInstanceName) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *StorageInstanceName) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *StorageInstanceName) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


