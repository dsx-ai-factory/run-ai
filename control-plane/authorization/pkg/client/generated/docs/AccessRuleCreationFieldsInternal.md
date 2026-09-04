# AccessRuleCreationFieldsInternal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SubjectId** | **string** | The unique identifier of the subject (e.g., user, group, or app). | 
**RoleId** | **int32** | The numeric ID of the assigned role. | 
**ScopeId** | **string** | The unique identifier of the scope. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 

## Methods

### NewAccessRuleCreationFieldsInternal

`func NewAccessRuleCreationFieldsInternal(subjectId string, roleId int32, scopeId string, scopeType ScopeType, ) *AccessRuleCreationFieldsInternal`

NewAccessRuleCreationFieldsInternal instantiates a new AccessRuleCreationFieldsInternal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleCreationFieldsInternalWithDefaults

`func NewAccessRuleCreationFieldsInternalWithDefaults() *AccessRuleCreationFieldsInternal`

NewAccessRuleCreationFieldsInternalWithDefaults instantiates a new AccessRuleCreationFieldsInternal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSubjectId

`func (o *AccessRuleCreationFieldsInternal) GetSubjectId() string`

GetSubjectId returns the SubjectId field if non-nil, zero value otherwise.

### GetSubjectIdOk

`func (o *AccessRuleCreationFieldsInternal) GetSubjectIdOk() (*string, bool)`

GetSubjectIdOk returns a tuple with the SubjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectId

`func (o *AccessRuleCreationFieldsInternal) SetSubjectId(v string)`

SetSubjectId sets SubjectId field to given value.


### GetRoleId

`func (o *AccessRuleCreationFieldsInternal) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *AccessRuleCreationFieldsInternal) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *AccessRuleCreationFieldsInternal) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.


### GetScopeId

`func (o *AccessRuleCreationFieldsInternal) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AccessRuleCreationFieldsInternal) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AccessRuleCreationFieldsInternal) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *AccessRuleCreationFieldsInternal) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AccessRuleCreationFieldsInternal) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AccessRuleCreationFieldsInternal) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetClusterId

`func (o *AccessRuleCreationFieldsInternal) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AccessRuleCreationFieldsInternal) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AccessRuleCreationFieldsInternal) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AccessRuleCreationFieldsInternal) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


