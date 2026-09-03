# StorageClass

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | Pointer to **string** | Cluster ID which has sent the resource info | [optional] 
**StorageClassName** | **string** | Storage class name | 
**Spec** | Pointer to [**StorageClassSpec**](StorageClassSpec.md) |  | [optional] 

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

### GetClusterId

`func (o *StorageClass) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *StorageClass) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *StorageClass) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *StorageClass) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

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


### GetSpec

`func (o *StorageClass) GetSpec() StorageClassSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *StorageClass) GetSpecOk() (*StorageClassSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *StorageClass) SetSpec(v StorageClassSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *StorageClass) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


