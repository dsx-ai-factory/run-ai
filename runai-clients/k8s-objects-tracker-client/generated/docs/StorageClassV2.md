# StorageClassV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | Cluster ID which has sent the resource info | 
**StorageClassName** | **string** | Storage class name | 
**Spec** | [**StorageClassSpec**](StorageClassSpec.md) |  | 
**Permissions** | Pointer to [**NullableStorageClassPermissions**](StorageClassPermissions.md) |  | [optional] 
**Customization** | Pointer to [**NullableStorageClassCustomization**](StorageClassCustomization.md) |  | [optional] 

## Methods

### NewStorageClassV2

`func NewStorageClassV2(clusterId string, storageClassName string, spec StorageClassSpec, ) *StorageClassV2`

NewStorageClassV2 instantiates a new StorageClassV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassV2WithDefaults

`func NewStorageClassV2WithDefaults() *StorageClassV2`

NewStorageClassV2WithDefaults instantiates a new StorageClassV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *StorageClassV2) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *StorageClassV2) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *StorageClassV2) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetStorageClassName

`func (o *StorageClassV2) GetStorageClassName() string`

GetStorageClassName returns the StorageClassName field if non-nil, zero value otherwise.

### GetStorageClassNameOk

`func (o *StorageClassV2) GetStorageClassNameOk() (*string, bool)`

GetStorageClassNameOk returns a tuple with the StorageClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassName

`func (o *StorageClassV2) SetStorageClassName(v string)`

SetStorageClassName sets StorageClassName field to given value.


### GetSpec

`func (o *StorageClassV2) GetSpec() StorageClassSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *StorageClassV2) GetSpecOk() (*StorageClassSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *StorageClassV2) SetSpec(v StorageClassSpec)`

SetSpec sets Spec field to given value.


### GetPermissions

`func (o *StorageClassV2) GetPermissions() StorageClassPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *StorageClassV2) GetPermissionsOk() (*StorageClassPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *StorageClassV2) SetPermissions(v StorageClassPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *StorageClassV2) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### SetPermissionsNil

`func (o *StorageClassV2) SetPermissionsNil(b bool)`

 SetPermissionsNil sets the value for Permissions to be an explicit nil

### UnsetPermissions
`func (o *StorageClassV2) UnsetPermissions()`

UnsetPermissions ensures that no value is present for Permissions, not even an explicit nil
### GetCustomization

`func (o *StorageClassV2) GetCustomization() StorageClassCustomization`

GetCustomization returns the Customization field if non-nil, zero value otherwise.

### GetCustomizationOk

`func (o *StorageClassV2) GetCustomizationOk() (*StorageClassCustomization, bool)`

GetCustomizationOk returns a tuple with the Customization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomization

`func (o *StorageClassV2) SetCustomization(v StorageClassCustomization)`

SetCustomization sets Customization field to given value.

### HasCustomization

`func (o *StorageClassV2) HasCustomization() bool`

HasCustomization returns a boolean if a field has been set.

### SetCustomizationNil

`func (o *StorageClassV2) SetCustomizationNil(b bool)`

 SetCustomizationNil sets the value for Customization to be an explicit nil

### UnsetCustomization
`func (o *StorageClassV2) UnsetCustomization()`

UnsetCustomization ensures that no value is present for Customization, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


