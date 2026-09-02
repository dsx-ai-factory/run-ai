# Permission

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceType** | [**ResourceType**](ResourceType.md) |  | 
**DisplayName** | **string** |  | 
**GroupId** | [**ResourceTypeGroupId**](ResourceTypeGroupId.md) |  | 
**Actions** | [**[]Action**](Action.md) |  | 

## Methods

### NewPermission

`func NewPermission(resourceType ResourceType, displayName string, groupId ResourceTypeGroupId, actions []Action, ) *Permission`

NewPermission instantiates a new Permission object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPermissionWithDefaults

`func NewPermissionWithDefaults() *Permission`

NewPermissionWithDefaults instantiates a new Permission object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceType

`func (o *Permission) GetResourceType() ResourceType`

GetResourceType returns the ResourceType field if non-nil, zero value otherwise.

### GetResourceTypeOk

`func (o *Permission) GetResourceTypeOk() (*ResourceType, bool)`

GetResourceTypeOk returns a tuple with the ResourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceType

`func (o *Permission) SetResourceType(v ResourceType)`

SetResourceType sets ResourceType field to given value.


### GetDisplayName

`func (o *Permission) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *Permission) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *Permission) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.


### GetGroupId

`func (o *Permission) GetGroupId() ResourceTypeGroupId`

GetGroupId returns the GroupId field if non-nil, zero value otherwise.

### GetGroupIdOk

`func (o *Permission) GetGroupIdOk() (*ResourceTypeGroupId, bool)`

GetGroupIdOk returns a tuple with the GroupId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupId

`func (o *Permission) SetGroupId(v ResourceTypeGroupId)`

SetGroupId sets GroupId field to given value.


### GetActions

`func (o *Permission) GetActions() []Action`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *Permission) GetActionsOk() (*[]Action, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *Permission) SetActions(v []Action)`

SetActions sets Actions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


