# RoleV1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **int32** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**CreatedBy** | **string** |  | 
**Custom** | **bool** |  | 
**EffectiveEnabled** | **bool** |  | [default to true]
**Permissions** | [**[]Permission**](Permission.md) |  | 
**Name** | **string** |  | 
**Description** | **string** |  | 
**ScopeType** | Pointer to **string** |  | [optional] [default to "tenant"]
**ScopeId** | Pointer to **NullableString** |  | [optional] 
**Enabled** | Pointer to **NullableBool** |  | [optional] [default to true]
**KubernetesPermissions** | Pointer to [**KubernetesPermissions**](KubernetesPermissions.md) |  | [optional] 

## Methods

### NewRoleV1

`func NewRoleV1(id int32, createdAt time.Time, updatedAt time.Time, createdBy string, custom bool, effectiveEnabled bool, permissions []Permission, name string, description string, ) *RoleV1`

NewRoleV1 instantiates a new RoleV1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleV1WithDefaults

`func NewRoleV1WithDefaults() *RoleV1`

NewRoleV1WithDefaults instantiates a new RoleV1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *RoleV1) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RoleV1) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RoleV1) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *RoleV1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *RoleV1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *RoleV1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *RoleV1) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *RoleV1) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *RoleV1) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *RoleV1) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *RoleV1) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *RoleV1) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *RoleV1) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *RoleV1) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *RoleV1) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetCreatedBy

`func (o *RoleV1) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *RoleV1) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *RoleV1) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCustom

`func (o *RoleV1) GetCustom() bool`

GetCustom returns the Custom field if non-nil, zero value otherwise.

### GetCustomOk

`func (o *RoleV1) GetCustomOk() (*bool, bool)`

GetCustomOk returns a tuple with the Custom field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustom

`func (o *RoleV1) SetCustom(v bool)`

SetCustom sets Custom field to given value.


### GetEffectiveEnabled

`func (o *RoleV1) GetEffectiveEnabled() bool`

GetEffectiveEnabled returns the EffectiveEnabled field if non-nil, zero value otherwise.

### GetEffectiveEnabledOk

`func (o *RoleV1) GetEffectiveEnabledOk() (*bool, bool)`

GetEffectiveEnabledOk returns a tuple with the EffectiveEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveEnabled

`func (o *RoleV1) SetEffectiveEnabled(v bool)`

SetEffectiveEnabled sets EffectiveEnabled field to given value.


### GetPermissions

`func (o *RoleV1) GetPermissions() []Permission`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *RoleV1) GetPermissionsOk() (*[]Permission, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *RoleV1) SetPermissions(v []Permission)`

SetPermissions sets Permissions field to given value.


### GetName

`func (o *RoleV1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RoleV1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RoleV1) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *RoleV1) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RoleV1) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RoleV1) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetScopeType

`func (o *RoleV1) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *RoleV1) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *RoleV1) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.

### HasScopeType

`func (o *RoleV1) HasScopeType() bool`

HasScopeType returns a boolean if a field has been set.

### GetScopeId

`func (o *RoleV1) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *RoleV1) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *RoleV1) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *RoleV1) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.

### SetScopeIdNil

`func (o *RoleV1) SetScopeIdNil(b bool)`

 SetScopeIdNil sets the value for ScopeId to be an explicit nil

### UnsetScopeId
`func (o *RoleV1) UnsetScopeId()`

UnsetScopeId ensures that no value is present for ScopeId, not even an explicit nil
### GetEnabled

`func (o *RoleV1) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *RoleV1) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *RoleV1) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *RoleV1) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *RoleV1) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *RoleV1) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetKubernetesPermissions

`func (o *RoleV1) GetKubernetesPermissions() KubernetesPermissions`

GetKubernetesPermissions returns the KubernetesPermissions field if non-nil, zero value otherwise.

### GetKubernetesPermissionsOk

`func (o *RoleV1) GetKubernetesPermissionsOk() (*KubernetesPermissions, bool)`

GetKubernetesPermissionsOk returns a tuple with the KubernetesPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKubernetesPermissions

`func (o *RoleV1) SetKubernetesPermissions(v KubernetesPermissions)`

SetKubernetesPermissions sets KubernetesPermissions field to given value.

### HasKubernetesPermissions

`func (o *RoleV1) HasKubernetesPermissions() bool`

HasKubernetesPermissions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


