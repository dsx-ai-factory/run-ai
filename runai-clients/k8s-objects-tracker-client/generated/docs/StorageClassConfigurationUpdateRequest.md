# StorageClassConfigurationUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Permissions** | Pointer to [**NullableStorageClassPermissions**](StorageClassPermissions.md) |  | [optional] 
**Customization** | Pointer to [**NullableStorageClassCustomization**](StorageClassCustomization.md) |  | [optional] 

## Methods

### NewStorageClassConfigurationUpdateRequest

`func NewStorageClassConfigurationUpdateRequest() *StorageClassConfigurationUpdateRequest`

NewStorageClassConfigurationUpdateRequest instantiates a new StorageClassConfigurationUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassConfigurationUpdateRequestWithDefaults

`func NewStorageClassConfigurationUpdateRequestWithDefaults() *StorageClassConfigurationUpdateRequest`

NewStorageClassConfigurationUpdateRequestWithDefaults instantiates a new StorageClassConfigurationUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPermissions

`func (o *StorageClassConfigurationUpdateRequest) GetPermissions() StorageClassPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *StorageClassConfigurationUpdateRequest) GetPermissionsOk() (*StorageClassPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *StorageClassConfigurationUpdateRequest) SetPermissions(v StorageClassPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *StorageClassConfigurationUpdateRequest) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### SetPermissionsNil

`func (o *StorageClassConfigurationUpdateRequest) SetPermissionsNil(b bool)`

 SetPermissionsNil sets the value for Permissions to be an explicit nil

### UnsetPermissions
`func (o *StorageClassConfigurationUpdateRequest) UnsetPermissions()`

UnsetPermissions ensures that no value is present for Permissions, not even an explicit nil
### GetCustomization

`func (o *StorageClassConfigurationUpdateRequest) GetCustomization() StorageClassCustomization`

GetCustomization returns the Customization field if non-nil, zero value otherwise.

### GetCustomizationOk

`func (o *StorageClassConfigurationUpdateRequest) GetCustomizationOk() (*StorageClassCustomization, bool)`

GetCustomizationOk returns a tuple with the Customization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomization

`func (o *StorageClassConfigurationUpdateRequest) SetCustomization(v StorageClassCustomization)`

SetCustomization sets Customization field to given value.

### HasCustomization

`func (o *StorageClassConfigurationUpdateRequest) HasCustomization() bool`

HasCustomization returns a boolean if a field has been set.

### SetCustomizationNil

`func (o *StorageClassConfigurationUpdateRequest) SetCustomizationNil(b bool)`

 SetCustomizationNil sets the value for Customization to be an explicit nil

### UnsetCustomization
`func (o *StorageClassConfigurationUpdateRequest) UnsetCustomization()`

UnsetCustomization ensures that no value is present for Customization, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


