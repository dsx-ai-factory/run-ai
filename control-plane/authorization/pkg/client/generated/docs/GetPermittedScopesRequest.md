# GetPermittedScopesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ResourceType** | [**ResourceType**](ResourceType.md) |  | 
**Action** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewGetPermittedScopesRequest

`func NewGetPermittedScopesRequest(resourceType ResourceType, ) *GetPermittedScopesRequest`

NewGetPermittedScopesRequest instantiates a new GetPermittedScopesRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetPermittedScopesRequestWithDefaults

`func NewGetPermittedScopesRequestWithDefaults() *GetPermittedScopesRequest`

NewGetPermittedScopesRequestWithDefaults instantiates a new GetPermittedScopesRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResourceType

`func (o *GetPermittedScopesRequest) GetResourceType() ResourceType`

GetResourceType returns the ResourceType field if non-nil, zero value otherwise.

### GetResourceTypeOk

`func (o *GetPermittedScopesRequest) GetResourceTypeOk() (*ResourceType, bool)`

GetResourceTypeOk returns a tuple with the ResourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceType

`func (o *GetPermittedScopesRequest) SetResourceType(v ResourceType)`

SetResourceType sets ResourceType field to given value.


### GetAction

`func (o *GetPermittedScopesRequest) GetAction() string`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *GetPermittedScopesRequest) GetActionOk() (*string, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *GetPermittedScopesRequest) SetAction(v string)`

SetAction sets Action field to given value.

### HasAction

`func (o *GetPermittedScopesRequest) HasAction() bool`

HasAction returns a boolean if a field has been set.

### SetActionNil

`func (o *GetPermittedScopesRequest) SetActionNil(b bool)`

 SetActionNil sets the value for Action to be an explicit nil

### UnsetAction
`func (o *GetPermittedScopesRequest) UnsetAction()`

UnsetAction ensures that no value is present for Action, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


