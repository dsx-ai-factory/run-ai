# AccessRulesBatchCreateFieldsPayload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RoleId** | **int32** | The numeric ID of the role to assign (e.g., admin, viewer). | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Subjects** | [**[]AccessRuleSubjectsInner**](AccessRuleSubjectsInner.md) | subjects for the access rules. must have at least one value | 
**Scopes** | [**[]AccessRuleScopesInner**](AccessRuleScopesInner.md) | scopes for the access rules. must have at least one value | 

## Methods

### NewAccessRulesBatchCreateFieldsPayload

`func NewAccessRulesBatchCreateFieldsPayload(roleId int32, subjects []AccessRuleSubjectsInner, scopes []AccessRuleScopesInner, ) *AccessRulesBatchCreateFieldsPayload`

NewAccessRulesBatchCreateFieldsPayload instantiates a new AccessRulesBatchCreateFieldsPayload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRulesBatchCreateFieldsPayloadWithDefaults

`func NewAccessRulesBatchCreateFieldsPayloadWithDefaults() *AccessRulesBatchCreateFieldsPayload`

NewAccessRulesBatchCreateFieldsPayloadWithDefaults instantiates a new AccessRulesBatchCreateFieldsPayload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRoleId

`func (o *AccessRulesBatchCreateFieldsPayload) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *AccessRulesBatchCreateFieldsPayload) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *AccessRulesBatchCreateFieldsPayload) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.


### GetClusterId

`func (o *AccessRulesBatchCreateFieldsPayload) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AccessRulesBatchCreateFieldsPayload) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AccessRulesBatchCreateFieldsPayload) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AccessRulesBatchCreateFieldsPayload) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetSubjects

`func (o *AccessRulesBatchCreateFieldsPayload) GetSubjects() []AccessRuleSubjectsInner`

GetSubjects returns the Subjects field if non-nil, zero value otherwise.

### GetSubjectsOk

`func (o *AccessRulesBatchCreateFieldsPayload) GetSubjectsOk() (*[]AccessRuleSubjectsInner, bool)`

GetSubjectsOk returns a tuple with the Subjects field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjects

`func (o *AccessRulesBatchCreateFieldsPayload) SetSubjects(v []AccessRuleSubjectsInner)`

SetSubjects sets Subjects field to given value.


### GetScopes

`func (o *AccessRulesBatchCreateFieldsPayload) GetScopes() []AccessRuleScopesInner`

GetScopes returns the Scopes field if non-nil, zero value otherwise.

### GetScopesOk

`func (o *AccessRulesBatchCreateFieldsPayload) GetScopesOk() (*[]AccessRuleScopesInner, bool)`

GetScopesOk returns a tuple with the Scopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopes

`func (o *AccessRulesBatchCreateFieldsPayload) SetScopes(v []AccessRuleScopesInner)`

SetScopes sets Scopes field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


