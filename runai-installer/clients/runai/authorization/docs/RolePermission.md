# RolePermission

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceType** | [**ResourceType**](ResourceType.md) |  | 
**DisplayName** | Pointer to **string** |  | [optional] 
**Actions** | [**[]PermissionAction**](PermissionAction.md) | The list of allowed actions for this resource - create - read - update - delete  | 

## Methods

### NewRolePermission

`func NewRolePermission(resourceType ResourceType, actions []PermissionAction, ) *RolePermission`

NewRolePermission instantiates a new RolePermission object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRolePermissionWithDefaults

`func NewRolePermissionWithDefaults() *RolePermission`

NewRolePermissionWithDefaults instantiates a new RolePermission object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceType

`func (o *RolePermission) GetResourceType() ResourceType`

GetResourceType returns the ResourceType field if non-nil, zero value otherwise.

### GetResourceTypeOk

`func (o *RolePermission) GetResourceTypeOk() (*ResourceType, bool)`

GetResourceTypeOk returns a tuple with the ResourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceType

`func (o *RolePermission) SetResourceType(v ResourceType)`

SetResourceType sets ResourceType field to given value.


### GetDisplayName

`func (o *RolePermission) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *RolePermission) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *RolePermission) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *RolePermission) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetActions

`func (o *RolePermission) GetActions() []PermissionAction`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *RolePermission) GetActionsOk() (*[]PermissionAction, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *RolePermission) SetActions(v []PermissionAction)`

SetActions sets Actions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


