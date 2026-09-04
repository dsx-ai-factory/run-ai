# StorageClassConfigurationPatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Permissions** | Pointer to [**NullableStorageClassPermissions**](StorageClassPermissions.md) |  | [optional] 
**Customization** | Pointer to [**NullableStorageClassConfigurationPatchRequestCustomization**](StorageClassConfigurationPatchRequestCustomization.md) |  | [optional] 

## Methods

### NewStorageClassConfigurationPatchRequest

`func NewStorageClassConfigurationPatchRequest() *StorageClassConfigurationPatchRequest`

NewStorageClassConfigurationPatchRequest instantiates a new StorageClassConfigurationPatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassConfigurationPatchRequestWithDefaults

`func NewStorageClassConfigurationPatchRequestWithDefaults() *StorageClassConfigurationPatchRequest`

NewStorageClassConfigurationPatchRequestWithDefaults instantiates a new StorageClassConfigurationPatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPermissions

`func (o *StorageClassConfigurationPatchRequest) GetPermissions() StorageClassPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *StorageClassConfigurationPatchRequest) GetPermissionsOk() (*StorageClassPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *StorageClassConfigurationPatchRequest) SetPermissions(v StorageClassPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *StorageClassConfigurationPatchRequest) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### SetPermissionsNil

`func (o *StorageClassConfigurationPatchRequest) SetPermissionsNil(b bool)`

 SetPermissionsNil sets the value for Permissions to be an explicit nil

### UnsetPermissions
`func (o *StorageClassConfigurationPatchRequest) UnsetPermissions()`

UnsetPermissions ensures that no value is present for Permissions, not even an explicit nil
### GetCustomization

`func (o *StorageClassConfigurationPatchRequest) GetCustomization() StorageClassConfigurationPatchRequestCustomization`

GetCustomization returns the Customization field if non-nil, zero value otherwise.

### GetCustomizationOk

`func (o *StorageClassConfigurationPatchRequest) GetCustomizationOk() (*StorageClassConfigurationPatchRequestCustomization, bool)`

GetCustomizationOk returns a tuple with the Customization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomization

`func (o *StorageClassConfigurationPatchRequest) SetCustomization(v StorageClassConfigurationPatchRequestCustomization)`

SetCustomization sets Customization field to given value.

### HasCustomization

`func (o *StorageClassConfigurationPatchRequest) HasCustomization() bool`

HasCustomization returns a boolean if a field has been set.

### SetCustomizationNil

`func (o *StorageClassConfigurationPatchRequest) SetCustomizationNil(b bool)`

 SetCustomizationNil sets the value for Customization to be an explicit nil

### UnsetCustomization
`func (o *StorageClassConfigurationPatchRequest) UnsetCustomization()`

UnsetCustomization ensures that no value is present for Customization, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


