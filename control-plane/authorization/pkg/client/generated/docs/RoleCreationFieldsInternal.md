# RoleCreationFieldsInternal

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
**Permissions** | Pointer to [**[]PostPermission**](PostPermission.md) |  | [optional] 

## Methods

### NewRoleCreationFieldsInternal

`func NewRoleCreationFieldsInternal(name string, description string, ) *RoleCreationFieldsInternal`

NewRoleCreationFieldsInternal instantiates a new RoleCreationFieldsInternal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleCreationFieldsInternalWithDefaults

`func NewRoleCreationFieldsInternalWithDefaults() *RoleCreationFieldsInternal`

NewRoleCreationFieldsInternalWithDefaults instantiates a new RoleCreationFieldsInternal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *RoleCreationFieldsInternal) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RoleCreationFieldsInternal) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RoleCreationFieldsInternal) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *RoleCreationFieldsInternal) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RoleCreationFieldsInternal) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RoleCreationFieldsInternal) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetScopeType

`func (o *RoleCreationFieldsInternal) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *RoleCreationFieldsInternal) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *RoleCreationFieldsInternal) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.

### HasScopeType

`func (o *RoleCreationFieldsInternal) HasScopeType() bool`

HasScopeType returns a boolean if a field has been set.

### GetScopeId

`func (o *RoleCreationFieldsInternal) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *RoleCreationFieldsInternal) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *RoleCreationFieldsInternal) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *RoleCreationFieldsInternal) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.

### SetScopeIdNil

`func (o *RoleCreationFieldsInternal) SetScopeIdNil(b bool)`

 SetScopeIdNil sets the value for ScopeId to be an explicit nil

### UnsetScopeId
`func (o *RoleCreationFieldsInternal) UnsetScopeId()`

UnsetScopeId ensures that no value is present for ScopeId, not even an explicit nil
### GetEnabled

`func (o *RoleCreationFieldsInternal) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *RoleCreationFieldsInternal) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *RoleCreationFieldsInternal) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *RoleCreationFieldsInternal) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *RoleCreationFieldsInternal) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *RoleCreationFieldsInternal) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetPermissionSets

`func (o *RoleCreationFieldsInternal) GetPermissionSets() []RolePermissionSet`

GetPermissionSets returns the PermissionSets field if non-nil, zero value otherwise.

### GetPermissionSetsOk

`func (o *RoleCreationFieldsInternal) GetPermissionSetsOk() (*[]RolePermissionSet, bool)`

GetPermissionSetsOk returns a tuple with the PermissionSets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissionSets

`func (o *RoleCreationFieldsInternal) SetPermissionSets(v []RolePermissionSet)`

SetPermissionSets sets PermissionSets field to given value.

### HasPermissionSets

`func (o *RoleCreationFieldsInternal) HasPermissionSets() bool`

HasPermissionSets returns a boolean if a field has been set.

### GetKubernetesPermissions

`func (o *RoleCreationFieldsInternal) GetKubernetesPermissions() KubernetesPermissions`

GetKubernetesPermissions returns the KubernetesPermissions field if non-nil, zero value otherwise.

### GetKubernetesPermissionsOk

`func (o *RoleCreationFieldsInternal) GetKubernetesPermissionsOk() (*KubernetesPermissions, bool)`

GetKubernetesPermissionsOk returns a tuple with the KubernetesPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKubernetesPermissions

`func (o *RoleCreationFieldsInternal) SetKubernetesPermissions(v KubernetesPermissions)`

SetKubernetesPermissions sets KubernetesPermissions field to given value.

### HasKubernetesPermissions

`func (o *RoleCreationFieldsInternal) HasKubernetesPermissions() bool`

HasKubernetesPermissions returns a boolean if a field has been set.

### GetPermissions

`func (o *RoleCreationFieldsInternal) GetPermissions() []PostPermission`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *RoleCreationFieldsInternal) GetPermissionsOk() (*[]PostPermission, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *RoleCreationFieldsInternal) SetPermissions(v []PostPermission)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *RoleCreationFieldsInternal) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


