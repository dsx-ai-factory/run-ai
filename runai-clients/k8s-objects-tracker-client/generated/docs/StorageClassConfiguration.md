# StorageClassConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier of the storage class configuration. | 
**TenantId** | **int32** | The id of the tenant. | 
**ScopeType** | **string** | Currently supported value: &#39;cluster&#39;, which is also the default. | [default to "cluster"]
**ScopeId** | **string** | The ID of the cluster. | 
**Name** | **string** | The name of the storage class to which the configuration applies. | 
**Permissions** | Pointer to [**NullableStorageClassPermissions**](StorageClassPermissions.md) |  | [optional] 
**Customization** | Pointer to [**NullableStorageClassCustomization**](StorageClassCustomization.md) |  | [optional] 

## Methods

### NewStorageClassConfiguration

`func NewStorageClassConfiguration(id string, tenantId int32, scopeType string, scopeId string, name string, ) *StorageClassConfiguration`

NewStorageClassConfiguration instantiates a new StorageClassConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassConfigurationWithDefaults

`func NewStorageClassConfigurationWithDefaults() *StorageClassConfiguration`

NewStorageClassConfigurationWithDefaults instantiates a new StorageClassConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *StorageClassConfiguration) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *StorageClassConfiguration) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *StorageClassConfiguration) SetId(v string)`

SetId sets Id field to given value.


### GetTenantId

`func (o *StorageClassConfiguration) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *StorageClassConfiguration) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *StorageClassConfiguration) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetScopeType

`func (o *StorageClassConfiguration) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *StorageClassConfiguration) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *StorageClassConfiguration) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *StorageClassConfiguration) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *StorageClassConfiguration) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *StorageClassConfiguration) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetName

`func (o *StorageClassConfiguration) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *StorageClassConfiguration) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *StorageClassConfiguration) SetName(v string)`

SetName sets Name field to given value.


### GetPermissions

`func (o *StorageClassConfiguration) GetPermissions() StorageClassPermissions`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *StorageClassConfiguration) GetPermissionsOk() (*StorageClassPermissions, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *StorageClassConfiguration) SetPermissions(v StorageClassPermissions)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *StorageClassConfiguration) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### SetPermissionsNil

`func (o *StorageClassConfiguration) SetPermissionsNil(b bool)`

 SetPermissionsNil sets the value for Permissions to be an explicit nil

### UnsetPermissions
`func (o *StorageClassConfiguration) UnsetPermissions()`

UnsetPermissions ensures that no value is present for Permissions, not even an explicit nil
### GetCustomization

`func (o *StorageClassConfiguration) GetCustomization() StorageClassCustomization`

GetCustomization returns the Customization field if non-nil, zero value otherwise.

### GetCustomizationOk

`func (o *StorageClassConfiguration) GetCustomizationOk() (*StorageClassCustomization, bool)`

GetCustomizationOk returns a tuple with the Customization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomization

`func (o *StorageClassConfiguration) SetCustomization(v StorageClassCustomization)`

SetCustomization sets Customization field to given value.

### HasCustomization

`func (o *StorageClassConfiguration) HasCustomization() bool`

HasCustomization returns a boolean if a field has been set.

### SetCustomizationNil

`func (o *StorageClassConfiguration) SetCustomizationNil(b bool)`

 SetCustomizationNil sets the value for Customization to be an explicit nil

### UnsetCustomization
`func (o *StorageClassConfiguration) UnsetCustomization()`

UnsetCustomization ensures that no value is present for Customization, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


