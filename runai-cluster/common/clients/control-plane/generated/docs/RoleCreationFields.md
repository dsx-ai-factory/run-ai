# RoleCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Description** | **string** |  | 
**Permissions** | [**[]Permission**](Permission.md) |  | 

## Methods

### NewRoleCreationFields

`func NewRoleCreationFields(name string, description string, permissions []Permission, ) *RoleCreationFields`

NewRoleCreationFields instantiates a new RoleCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleCreationFieldsWithDefaults

`func NewRoleCreationFieldsWithDefaults() *RoleCreationFields`

NewRoleCreationFieldsWithDefaults instantiates a new RoleCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *RoleCreationFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RoleCreationFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RoleCreationFields) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *RoleCreationFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *RoleCreationFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *RoleCreationFields) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetPermissions

`func (o *RoleCreationFields) GetPermissions() []Permission`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *RoleCreationFields) GetPermissionsOk() (*[]Permission, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *RoleCreationFields) SetPermissions(v []Permission)`

SetPermissions sets Permissions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


