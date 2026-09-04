# RoleCalculatedFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
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

### NewRoleCalculatedFields

`func NewRoleCalculatedFields(id int32, createdAt time.Time, updatedAt time.Time, createdBy string, custom bool, effectiveEnabled bool, permissions []RolePermission, ) *RoleCalculatedFields`

NewRoleCalculatedFields instantiates a new RoleCalculatedFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleCalculatedFieldsWithDefaults

`func NewRoleCalculatedFieldsWithDefaults() *RoleCalculatedFields`

NewRoleCalculatedFieldsWithDefaults instantiates a new RoleCalculatedFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *RoleCalculatedFields) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RoleCalculatedFields) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RoleCalculatedFields) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *RoleCalculatedFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *RoleCalculatedFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *RoleCalculatedFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *RoleCalculatedFields) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *RoleCalculatedFields) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *RoleCalculatedFields) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *RoleCalculatedFields) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *RoleCalculatedFields) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *RoleCalculatedFields) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *RoleCalculatedFields) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *RoleCalculatedFields) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *RoleCalculatedFields) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetCreatedBy

`func (o *RoleCalculatedFields) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *RoleCalculatedFields) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *RoleCalculatedFields) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCustom

`func (o *RoleCalculatedFields) GetCustom() bool`

GetCustom returns the Custom field if non-nil, zero value otherwise.

### GetCustomOk

`func (o *RoleCalculatedFields) GetCustomOk() (*bool, bool)`

GetCustomOk returns a tuple with the Custom field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustom

`func (o *RoleCalculatedFields) SetCustom(v bool)`

SetCustom sets Custom field to given value.


### GetEffectiveEnabled

`func (o *RoleCalculatedFields) GetEffectiveEnabled() bool`

GetEffectiveEnabled returns the EffectiveEnabled field if non-nil, zero value otherwise.

### GetEffectiveEnabledOk

`func (o *RoleCalculatedFields) GetEffectiveEnabledOk() (*bool, bool)`

GetEffectiveEnabledOk returns a tuple with the EffectiveEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveEnabled

`func (o *RoleCalculatedFields) SetEffectiveEnabled(v bool)`

SetEffectiveEnabled sets EffectiveEnabled field to given value.


### GetDeprecated

`func (o *RoleCalculatedFields) GetDeprecated() bool`

GetDeprecated returns the Deprecated field if non-nil, zero value otherwise.

### GetDeprecatedOk

`func (o *RoleCalculatedFields) GetDeprecatedOk() (*bool, bool)`

GetDeprecatedOk returns a tuple with the Deprecated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeprecated

`func (o *RoleCalculatedFields) SetDeprecated(v bool)`

SetDeprecated sets Deprecated field to given value.

### HasDeprecated

`func (o *RoleCalculatedFields) HasDeprecated() bool`

HasDeprecated returns a boolean if a field has been set.

### GetPermissions

`func (o *RoleCalculatedFields) GetPermissions() []RolePermission`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *RoleCalculatedFields) GetPermissionsOk() (*[]RolePermission, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *RoleCalculatedFields) SetPermissions(v []RolePermission)`

SetPermissions sets Permissions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


