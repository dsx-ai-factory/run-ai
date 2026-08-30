# PostPermission

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceType** | [**ResourceType**](ResourceType.md) |  | 
**DisplayName** | **string** |  | 
**GroupId** | Pointer to [**NullableResourceTypeGroupId**](ResourceTypeGroupId.md) |  | [optional] 
**Actions** | [**[]Action**](Action.md) |  | 

## Methods

### NewPostPermission

`func NewPostPermission(resourceType ResourceType, displayName string, actions []Action, ) *PostPermission`

NewPostPermission instantiates a new PostPermission object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostPermissionWithDefaults

`func NewPostPermissionWithDefaults() *PostPermission`

NewPostPermissionWithDefaults instantiates a new PostPermission object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceType

`func (o *PostPermission) GetResourceType() ResourceType`

GetResourceType returns the ResourceType field if non-nil, zero value otherwise.

### GetResourceTypeOk

`func (o *PostPermission) GetResourceTypeOk() (*ResourceType, bool)`

GetResourceTypeOk returns a tuple with the ResourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceType

`func (o *PostPermission) SetResourceType(v ResourceType)`

SetResourceType sets ResourceType field to given value.


### GetDisplayName

`func (o *PostPermission) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *PostPermission) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *PostPermission) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.


### GetGroupId

`func (o *PostPermission) GetGroupId() ResourceTypeGroupId`

GetGroupId returns the GroupId field if non-nil, zero value otherwise.

### GetGroupIdOk

`func (o *PostPermission) GetGroupIdOk() (*ResourceTypeGroupId, bool)`

GetGroupIdOk returns a tuple with the GroupId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupId

`func (o *PostPermission) SetGroupId(v ResourceTypeGroupId)`

SetGroupId sets GroupId field to given value.

### HasGroupId

`func (o *PostPermission) HasGroupId() bool`

HasGroupId returns a boolean if a field has been set.

### SetGroupIdNil

`func (o *PostPermission) SetGroupIdNil(b bool)`

 SetGroupIdNil sets the value for GroupId to be an explicit nil

### UnsetGroupId
`func (o *PostPermission) UnsetGroupId()`

UnsetGroupId ensures that no value is present for GroupId, not even an explicit nil
### GetActions

`func (o *PostPermission) GetActions() []Action`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *PostPermission) GetActionsOk() (*[]Action, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *PostPermission) SetActions(v []Action)`

SetActions sets Actions field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


