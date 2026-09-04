# MeAccessRule

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SubjectId** | **string** | The unique identifier of the subject (e.g., user, group, or app). | 
**SubjectType** | [**SubjectType**](SubjectType.md) |  | 
**RoleId** | **int32** | The numeric ID of the assigned role. | 
**ScopeId** | **string** | The unique identifier of the scope. | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**RoleName** | **string** | The name of the assigned role (e.g., admin, viewer). | 
**ScopeName** | **string** | Human-readable name of the scope (e.g., tenant-x). | 
**Id** | Pointer to **int32** |  | [optional] 
**TenantId** | **int32** | The id of the tenant. | 

## Methods

### NewMeAccessRule

`func NewMeAccessRule(subjectId string, subjectType SubjectType, roleId int32, scopeId string, scopeType ScopeType, roleName string, scopeName string, tenantId int32, ) *MeAccessRule`

NewMeAccessRule instantiates a new MeAccessRule object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMeAccessRuleWithDefaults

`func NewMeAccessRuleWithDefaults() *MeAccessRule`

NewMeAccessRuleWithDefaults instantiates a new MeAccessRule object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSubjectId

`func (o *MeAccessRule) GetSubjectId() string`

GetSubjectId returns the SubjectId field if non-nil, zero value otherwise.

### GetSubjectIdOk

`func (o *MeAccessRule) GetSubjectIdOk() (*string, bool)`

GetSubjectIdOk returns a tuple with the SubjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectId

`func (o *MeAccessRule) SetSubjectId(v string)`

SetSubjectId sets SubjectId field to given value.


### GetSubjectType

`func (o *MeAccessRule) GetSubjectType() SubjectType`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *MeAccessRule) GetSubjectTypeOk() (*SubjectType, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *MeAccessRule) SetSubjectType(v SubjectType)`

SetSubjectType sets SubjectType field to given value.


### GetRoleId

`func (o *MeAccessRule) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *MeAccessRule) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *MeAccessRule) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.


### GetScopeId

`func (o *MeAccessRule) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *MeAccessRule) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *MeAccessRule) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *MeAccessRule) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *MeAccessRule) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *MeAccessRule) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetClusterId

`func (o *MeAccessRule) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *MeAccessRule) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *MeAccessRule) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *MeAccessRule) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetRoleName

`func (o *MeAccessRule) GetRoleName() string`

GetRoleName returns the RoleName field if non-nil, zero value otherwise.

### GetRoleNameOk

`func (o *MeAccessRule) GetRoleNameOk() (*string, bool)`

GetRoleNameOk returns a tuple with the RoleName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleName

`func (o *MeAccessRule) SetRoleName(v string)`

SetRoleName sets RoleName field to given value.


### GetScopeName

`func (o *MeAccessRule) GetScopeName() string`

GetScopeName returns the ScopeName field if non-nil, zero value otherwise.

### GetScopeNameOk

`func (o *MeAccessRule) GetScopeNameOk() (*string, bool)`

GetScopeNameOk returns a tuple with the ScopeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeName

`func (o *MeAccessRule) SetScopeName(v string)`

SetScopeName sets ScopeName field to given value.


### GetId

`func (o *MeAccessRule) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *MeAccessRule) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *MeAccessRule) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *MeAccessRule) HasId() bool`

HasId returns a boolean if a field has been set.

### GetTenantId

`func (o *MeAccessRule) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *MeAccessRule) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *MeAccessRule) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


