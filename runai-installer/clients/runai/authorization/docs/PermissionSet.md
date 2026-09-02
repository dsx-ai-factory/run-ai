# PermissionSet

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The universally unique ID (UUID) of the permission set. | 
**Name** | **string** | The name of the permission set. | 
**Description** | Pointer to **string** | A description of the permission set. | [optional] 
**Permissions** | [**[]RolePermission**](RolePermission.md) |  | 

## Methods

### NewPermissionSet

`func NewPermissionSet(id string, name string, permissions []RolePermission, ) *PermissionSet`

NewPermissionSet instantiates a new PermissionSet object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPermissionSetWithDefaults

`func NewPermissionSetWithDefaults() *PermissionSet`

NewPermissionSetWithDefaults instantiates a new PermissionSet object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *PermissionSet) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PermissionSet) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PermissionSet) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *PermissionSet) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PermissionSet) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PermissionSet) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *PermissionSet) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *PermissionSet) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *PermissionSet) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *PermissionSet) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetPermissions

`func (o *PermissionSet) GetPermissions() []RolePermission`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *PermissionSet) GetPermissionsOk() (*[]RolePermission, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *PermissionSet) SetPermissions(v []RolePermission)`

SetPermissions sets Permissions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


