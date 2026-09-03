# AuthenticatedUsersAccessRuleCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RoleId** | **int32** | The ID of the role to assign. | 
**ScopeId** | **string** | The unique identifier of the scope. | 
**ScopeType** | **string** | The type of scope the access rule applies to. | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 

## Methods

### NewAuthenticatedUsersAccessRuleCreationFields

`func NewAuthenticatedUsersAccessRuleCreationFields(roleId int32, scopeId string, scopeType string, ) *AuthenticatedUsersAccessRuleCreationFields`

NewAuthenticatedUsersAccessRuleCreationFields instantiates a new AuthenticatedUsersAccessRuleCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthenticatedUsersAccessRuleCreationFieldsWithDefaults

`func NewAuthenticatedUsersAccessRuleCreationFieldsWithDefaults() *AuthenticatedUsersAccessRuleCreationFields`

NewAuthenticatedUsersAccessRuleCreationFieldsWithDefaults instantiates a new AuthenticatedUsersAccessRuleCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRoleId

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *AuthenticatedUsersAccessRuleCreationFields) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.


### GetScopeId

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AuthenticatedUsersAccessRuleCreationFields) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AuthenticatedUsersAccessRuleCreationFields) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetClusterId

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AuthenticatedUsersAccessRuleCreationFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AuthenticatedUsersAccessRuleCreationFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AuthenticatedUsersAccessRuleCreationFields) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


