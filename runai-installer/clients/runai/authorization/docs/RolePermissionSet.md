# RolePermissionSet

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The universally unique ID (UUID) of the permission set. | 
**Name** | Pointer to **string** | The name of the permission set. | [optional] 

## Methods

### NewRolePermissionSet

`func NewRolePermissionSet(id string, ) *RolePermissionSet`

NewRolePermissionSet instantiates a new RolePermissionSet object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRolePermissionSetWithDefaults

`func NewRolePermissionSetWithDefaults() *RolePermissionSet`

NewRolePermissionSetWithDefaults instantiates a new RolePermissionSet object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *RolePermissionSet) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RolePermissionSet) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RolePermissionSet) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *RolePermissionSet) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RolePermissionSet) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RolePermissionSet) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *RolePermissionSet) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


