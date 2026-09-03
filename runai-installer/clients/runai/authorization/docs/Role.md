# Role

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the role to create. Must be unique within the given scope. | 
**Description** | **string** | A description of the role. | 
**ScopeType** | Pointer to **string** | The type of scope where the role applies. Valid values: system, tenant | [optional] [default to "tenant"]
**ScopeId** | Pointer to **NullableString** | The ID of the scope where this role is defined. | [optional] 
**Enabled** | Pointer to **NullableBool** | Specifies whether the role is active. Default: true | [optional] [default to true]
**PermissionSets** | Pointer to [**[]RolePermissionSet**](RolePermissionSet.md) | The permission sets assigned to this role. Each item includes the UUID and name of the permission set. | [optional] 
**KubernetesPermissions** | Pointer to [**KubernetesPermissions**](KubernetesPermissions.md) | Defines the Kubernetes cluster permissions associated with this role. | [optional] 
**Id** | **int32** | The ID of the role. | 
**CreatedAt** | **time.Time** | The timestamp for when the role was created. | 
**UpdatedAt** | **time.Time** | The timestamp for the last time the role was updated. | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**CreatedBy** | **string** | Identifier of the user who created the role. | 
**Custom** | **bool** |  | 
**EffectiveEnabled** | **bool** | Indicates whether the role is effectively enabled in the tenant. This reflects the role’s operational state after considering tenant-level configuration or restrictions. Default: true | [default to true]
**Deprecated** | Pointer to **bool** | Indicates whether the role is deprecated. When set to true, the role is considered deprecated and should not be used for new assignments. | [optional] 
**Permissions** | [**[]RolePermission**](RolePermission.md) |  | 

## Methods

### NewRole

`func NewRole(name string, description string, id int32, createdAt time.Time, updatedAt time.Time, createdBy string, custom bool, effectiveEnabled bool, permissions []RolePermission, ) *Role`

NewRole instantiates a new Role object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleWithDefaults

`func NewRoleWithDefaults() *Role`

NewRoleWithDefaults instantiates a new Role object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Role) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Role) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Role) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *Role) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Role) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Role) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetScopeType

`func (o *Role) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *Role) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *Role) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.

### HasScopeType

`func (o *Role) HasScopeType() bool`

HasScopeType returns a boolean if a field has been set.

### GetScopeId

`func (o *Role) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *Role) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *Role) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *Role) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.

### SetScopeIdNil

`func (o *Role) SetScopeIdNil(b bool)`

 SetScopeIdNil sets the value for ScopeId to be an explicit nil

### UnsetScopeId
`func (o *Role) UnsetScopeId()`

UnsetScopeId ensures that no value is present for ScopeId, not even an explicit nil
### GetEnabled

`func (o *Role) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *Role) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *Role) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *Role) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *Role) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *Role) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetPermissionSets

`func (o *Role) GetPermissionSets() []RolePermissionSet`

GetPermissionSets returns the PermissionSets field if non-nil, zero value otherwise.

### GetPermissionSetsOk

`func (o *Role) GetPermissionSetsOk() (*[]RolePermissionSet, bool)`

GetPermissionSetsOk returns a tuple with the PermissionSets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissionSets

`func (o *Role) SetPermissionSets(v []RolePermissionSet)`

SetPermissionSets sets PermissionSets field to given value.

### HasPermissionSets

`func (o *Role) HasPermissionSets() bool`

HasPermissionSets returns a boolean if a field has been set.

### GetKubernetesPermissions

`func (o *Role) GetKubernetesPermissions() KubernetesPermissions`

GetKubernetesPermissions returns the KubernetesPermissions field if non-nil, zero value otherwise.

### GetKubernetesPermissionsOk

`func (o *Role) GetKubernetesPermissionsOk() (*KubernetesPermissions, bool)`

GetKubernetesPermissionsOk returns a tuple with the KubernetesPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKubernetesPermissions

`func (o *Role) SetKubernetesPermissions(v KubernetesPermissions)`

SetKubernetesPermissions sets KubernetesPermissions field to given value.

### HasKubernetesPermissions

`func (o *Role) HasKubernetesPermissions() bool`

HasKubernetesPermissions returns a boolean if a field has been set.

### GetId

`func (o *Role) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Role) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Role) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *Role) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Role) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Role) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *Role) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Role) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Role) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *Role) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Role) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Role) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *Role) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *Role) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Role) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetCreatedBy

`func (o *Role) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Role) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Role) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCustom

`func (o *Role) GetCustom() bool`

GetCustom returns the Custom field if non-nil, zero value otherwise.

### GetCustomOk

`func (o *Role) GetCustomOk() (*bool, bool)`

GetCustomOk returns a tuple with the Custom field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustom

`func (o *Role) SetCustom(v bool)`

SetCustom sets Custom field to given value.


### GetEffectiveEnabled

`func (o *Role) GetEffectiveEnabled() bool`

GetEffectiveEnabled returns the EffectiveEnabled field if non-nil, zero value otherwise.

### GetEffectiveEnabledOk

`func (o *Role) GetEffectiveEnabledOk() (*bool, bool)`

GetEffectiveEnabledOk returns a tuple with the EffectiveEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveEnabled

`func (o *Role) SetEffectiveEnabled(v bool)`

SetEffectiveEnabled sets EffectiveEnabled field to given value.


### GetDeprecated

`func (o *Role) GetDeprecated() bool`

GetDeprecated returns the Deprecated field if non-nil, zero value otherwise.

### GetDeprecatedOk

`func (o *Role) GetDeprecatedOk() (*bool, bool)`

GetDeprecatedOk returns a tuple with the Deprecated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeprecated

`func (o *Role) SetDeprecated(v bool)`

SetDeprecated sets Deprecated field to given value.

### HasDeprecated

`func (o *Role) HasDeprecated() bool`

HasDeprecated returns a boolean if a field has been set.

### GetPermissions

`func (o *Role) GetPermissions() []RolePermission`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *Role) GetPermissionsOk() (*[]RolePermission, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *Role) SetPermissions(v []RolePermission)`

SetPermissions sets Permissions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


