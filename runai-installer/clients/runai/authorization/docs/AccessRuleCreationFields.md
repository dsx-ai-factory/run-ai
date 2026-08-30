# AccessRuleCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SubjectId** | **string** | The unique identifier of the subject (e.g., user, group, or app). | 
**SubjectType** | [**SubjectType**](SubjectType.md) |  | 
**RoleId** | **int32** | The numeric ID of the assigned role. | 
**ScopeId** | **string** | The unique identifier of the scope. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 

## Methods

### NewAccessRuleCreationFields

`func NewAccessRuleCreationFields(subjectId string, subjectType SubjectType, roleId int32, scopeId string, scopeType ScopeType, ) *AccessRuleCreationFields`

NewAccessRuleCreationFields instantiates a new AccessRuleCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleCreationFieldsWithDefaults

`func NewAccessRuleCreationFieldsWithDefaults() *AccessRuleCreationFields`

NewAccessRuleCreationFieldsWithDefaults instantiates a new AccessRuleCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSubjectId

`func (o *AccessRuleCreationFields) GetSubjectId() string`

GetSubjectId returns the SubjectId field if non-nil, zero value otherwise.

### GetSubjectIdOk

`func (o *AccessRuleCreationFields) GetSubjectIdOk() (*string, bool)`

GetSubjectIdOk returns a tuple with the SubjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectId

`func (o *AccessRuleCreationFields) SetSubjectId(v string)`

SetSubjectId sets SubjectId field to given value.


### GetSubjectType

`func (o *AccessRuleCreationFields) GetSubjectType() SubjectType`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *AccessRuleCreationFields) GetSubjectTypeOk() (*SubjectType, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *AccessRuleCreationFields) SetSubjectType(v SubjectType)`

SetSubjectType sets SubjectType field to given value.


### GetRoleId

`func (o *AccessRuleCreationFields) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *AccessRuleCreationFields) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *AccessRuleCreationFields) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.


### GetScopeId

`func (o *AccessRuleCreationFields) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AccessRuleCreationFields) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AccessRuleCreationFields) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *AccessRuleCreationFields) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AccessRuleCreationFields) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AccessRuleCreationFields) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetClusterId

`func (o *AccessRuleCreationFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AccessRuleCreationFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AccessRuleCreationFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AccessRuleCreationFields) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


