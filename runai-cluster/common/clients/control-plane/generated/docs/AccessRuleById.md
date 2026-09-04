# AccessRuleById

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SubjectId** | **string** |  | 
**SubjectType** | [**SubjectType**](SubjectType.md) |  | 
**RoleId** | **int32** |  | 
**ScopeId** | **string** |  | 
**ScopeType** | [**ScopeType**](ScopeType.md) |  | 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**RoleName** | **string** |  | 
**ScopeName** | **string** |  | 
**Id** | **int32** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**TenantId** | **int32** | The id of the tenant. | 
**CreatedBy** | **string** |  | 
**ScopePath** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewAccessRuleById

`func NewAccessRuleById(subjectId string, subjectType SubjectType, roleId int32, scopeId string, scopeType ScopeType, roleName string, scopeName string, id int32, createdAt time.Time, updatedAt time.Time, tenantId int32, createdBy string, ) *AccessRuleById`

NewAccessRuleById instantiates a new AccessRuleById object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleByIdWithDefaults

`func NewAccessRuleByIdWithDefaults() *AccessRuleById`

NewAccessRuleByIdWithDefaults instantiates a new AccessRuleById object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSubjectId

`func (o *AccessRuleById) GetSubjectId() string`

GetSubjectId returns the SubjectId field if non-nil, zero value otherwise.

### GetSubjectIdOk

`func (o *AccessRuleById) GetSubjectIdOk() (*string, bool)`

GetSubjectIdOk returns a tuple with the SubjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectId

`func (o *AccessRuleById) SetSubjectId(v string)`

SetSubjectId sets SubjectId field to given value.


### GetSubjectType

`func (o *AccessRuleById) GetSubjectType() SubjectType`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *AccessRuleById) GetSubjectTypeOk() (*SubjectType, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *AccessRuleById) SetSubjectType(v SubjectType)`

SetSubjectType sets SubjectType field to given value.


### GetRoleId

`func (o *AccessRuleById) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *AccessRuleById) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *AccessRuleById) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.


### GetScopeId

`func (o *AccessRuleById) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AccessRuleById) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AccessRuleById) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *AccessRuleById) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AccessRuleById) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AccessRuleById) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetClusterId

`func (o *AccessRuleById) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AccessRuleById) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AccessRuleById) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AccessRuleById) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetRoleName

`func (o *AccessRuleById) GetRoleName() string`

GetRoleName returns the RoleName field if non-nil, zero value otherwise.

### GetRoleNameOk

`func (o *AccessRuleById) GetRoleNameOk() (*string, bool)`

GetRoleNameOk returns a tuple with the RoleName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleName

`func (o *AccessRuleById) SetRoleName(v string)`

SetRoleName sets RoleName field to given value.


### GetScopeName

`func (o *AccessRuleById) GetScopeName() string`

GetScopeName returns the ScopeName field if non-nil, zero value otherwise.

### GetScopeNameOk

`func (o *AccessRuleById) GetScopeNameOk() (*string, bool)`

GetScopeNameOk returns a tuple with the ScopeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeName

`func (o *AccessRuleById) SetScopeName(v string)`

SetScopeName sets ScopeName field to given value.


### GetId

`func (o *AccessRuleById) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AccessRuleById) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AccessRuleById) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *AccessRuleById) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AccessRuleById) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AccessRuleById) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *AccessRuleById) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AccessRuleById) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AccessRuleById) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *AccessRuleById) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *AccessRuleById) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *AccessRuleById) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *AccessRuleById) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *AccessRuleById) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *AccessRuleById) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetTenantId

`func (o *AccessRuleById) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AccessRuleById) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AccessRuleById) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetCreatedBy

`func (o *AccessRuleById) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AccessRuleById) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AccessRuleById) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetScopePath

`func (o *AccessRuleById) GetScopePath() string`

GetScopePath returns the ScopePath field if non-nil, zero value otherwise.

### GetScopePathOk

`func (o *AccessRuleById) GetScopePathOk() (*string, bool)`

GetScopePathOk returns a tuple with the ScopePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopePath

`func (o *AccessRuleById) SetScopePath(v string)`

SetScopePath sets ScopePath field to given value.

### HasScopePath

`func (o *AccessRuleById) HasScopePath() bool`

HasScopePath returns a boolean if a field has been set.

### SetScopePathNil

`func (o *AccessRuleById) SetScopePathNil(b bool)`

 SetScopePathNil sets the value for ScopePath to be an explicit nil

### UnsetScopePath
`func (o *AccessRuleById) UnsetScopePath()`

UnsetScopePath ensures that no value is present for ScopePath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


