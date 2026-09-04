# StorageClassConfigurationCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TenantId** | Pointer to **NullableInt32** | The id of the tenant. | [optional] 
**ScopeType** | **string** | the scope type of the configuration. | [default to "cluster"]
**ScopeId** | **string** | The ID of the cluster. | 
**Name** | **string** | The name of the storage class to which the configuration applies. | 
**Permissions** | Pointer to [**NullableStorageClassPermissions**](StorageClassPermissions.md) |  | [optional] 
**Customization** | Pointer to [**NullableStorageClassCustomization**](StorageClassCustomization.md) |  | [optional] 

## Methods

### NewStorageClassConfigurationCreationRequest

`func NewStorageClassConfigurationCreationRequest(scopeType string, scopeId string, name string, ) *StorageClassConfigurationCreationRequest`

NewStorageClassConfigurationCreationRequest instantiates a new StorageClassConfigurationCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassConfigurationCreationRequestWithDefaults

`func NewStorageClassConfigurationCreationRequestWithDefaults() *StorageClassConfigurationCreationRequest`

NewStorageClassConfigurationCreationRequestWithDefaults instantiates a new StorageClassConfigurationCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTenantId

`func (o *StorageClassConfigurationCreationRequest) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *StorageClassConfigurationCreationRequest) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *StorageClassConfigurationCreationRequest) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *StorageClassConfigurationCreationRequest) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### SetTenantIdNil

`func (o *StorageClassConfigurationCreationRequest) SetTenantIdNil(b bool)`

 SetTenantIdNil sets the value for TenantId to be an explicit nil

### UnsetTenantId
`func (o *StorageClassConfigurationCreationRequest) UnsetTenantId()`

UnsetTenantId ensures that no value is present for TenantId, not even an explicit nil
### GetScopeType

`func (o *StorageClassConfigurationCreationRequest) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *StorageClassConfigurationCreationRequest) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *StorageClassConfigurationCreationRequest) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *StorageClassConfigurationCreationRequest) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *StorageClassConfigurationCreationRequest) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *StorageClassConfigurationCreationRequest) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetName

`func (o *StorageClassConfigurationCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *StorageClassConfigurationCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *StorageClassConfigurationCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetPermissions

`func (o *StorageClassConfigurationCreationRequest) GetPermissions() StorageClassPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *StorageClassConfigurationCreationRequest) GetPermissionsOk() (*StorageClassPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *StorageClassConfigurationCreationRequest) SetPermissions(v StorageClassPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *StorageClassConfigurationCreationRequest) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### SetPermissionsNil

`func (o *StorageClassConfigurationCreationRequest) SetPermissionsNil(b bool)`

 SetPermissionsNil sets the value for Permissions to be an explicit nil

### UnsetPermissions
`func (o *StorageClassConfigurationCreationRequest) UnsetPermissions()`

UnsetPermissions ensures that no value is present for Permissions, not even an explicit nil
### GetCustomization

`func (o *StorageClassConfigurationCreationRequest) GetCustomization() StorageClassCustomization`

GetCustomization returns the Customization field if non-nil, zero value otherwise.

### GetCustomizationOk

`func (o *StorageClassConfigurationCreationRequest) GetCustomizationOk() (*StorageClassCustomization, bool)`

GetCustomizationOk returns a tuple with the Customization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomization

`func (o *StorageClassConfigurationCreationRequest) SetCustomization(v StorageClassCustomization)`

SetCustomization sets Customization field to given value.

### HasCustomization

`func (o *StorageClassConfigurationCreationRequest) HasCustomization() bool`

HasCustomization returns a boolean if a field has been set.

### SetCustomizationNil

`func (o *StorageClassConfigurationCreationRequest) SetCustomizationNil(b bool)`

 SetCustomizationNil sets the value for Customization to be an explicit nil

### UnsetCustomization
`func (o *StorageClassConfigurationCreationRequest) UnsetCustomization()`

UnsetCustomization ensures that no value is present for Customization, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


