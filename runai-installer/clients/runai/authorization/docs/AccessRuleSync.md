# AccessRuleSync

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
**Id** | **int32** | The ID of the access rule entry. | 
**CreatedAt** | **time.Time** | The timestamp for when the access rule was created. | 
**UpdatedAt** | **time.Time** | The timestamp for the last time the access rule was updated. | 
**DeletedAt** | Pointer to **NullableTime** | The timestamp indicating when the access rule was deleted. | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**CreatedBy** | **string** | Identifier of the user who created the access rule. | 
**PredefinedRole** | Pointer to [**NullableAccessRuleSyncAllOfPredefinedRole**](AccessRuleSyncAllOfPredefinedRole.md) |  | [optional] 
**IsDeleting** | Pointer to **bool** | Indicates if the access rule is in &#39;Deleting&#39; phase and should be deleted from the cluster | [optional] [default to false]

## Methods

### NewAccessRuleSync

`func NewAccessRuleSync(subjectId string, subjectType SubjectType, roleId int32, scopeId string, scopeType ScopeType, roleName string, scopeName string, id int32, createdAt time.Time, updatedAt time.Time, createdBy string, ) *AccessRuleSync`

NewAccessRuleSync instantiates a new AccessRuleSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessRuleSyncWithDefaults

`func NewAccessRuleSyncWithDefaults() *AccessRuleSync`

NewAccessRuleSyncWithDefaults instantiates a new AccessRuleSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSubjectId

`func (o *AccessRuleSync) GetSubjectId() string`

GetSubjectId returns the SubjectId field if non-nil, zero value otherwise.

### GetSubjectIdOk

`func (o *AccessRuleSync) GetSubjectIdOk() (*string, bool)`

GetSubjectIdOk returns a tuple with the SubjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectId

`func (o *AccessRuleSync) SetSubjectId(v string)`

SetSubjectId sets SubjectId field to given value.


### GetSubjectType

`func (o *AccessRuleSync) GetSubjectType() SubjectType`

GetSubjectType returns the SubjectType field if non-nil, zero value otherwise.

### GetSubjectTypeOk

`func (o *AccessRuleSync) GetSubjectTypeOk() (*SubjectType, bool)`

GetSubjectTypeOk returns a tuple with the SubjectType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubjectType

`func (o *AccessRuleSync) SetSubjectType(v SubjectType)`

SetSubjectType sets SubjectType field to given value.


### GetRoleId

`func (o *AccessRuleSync) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *AccessRuleSync) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *AccessRuleSync) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.


### GetScopeId

`func (o *AccessRuleSync) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *AccessRuleSync) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *AccessRuleSync) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetScopeType

`func (o *AccessRuleSync) GetScopeType() ScopeType`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *AccessRuleSync) GetScopeTypeOk() (*ScopeType, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *AccessRuleSync) SetScopeType(v ScopeType)`

SetScopeType sets ScopeType field to given value.


### GetClusterId

`func (o *AccessRuleSync) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *AccessRuleSync) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *AccessRuleSync) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *AccessRuleSync) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetRoleName

`func (o *AccessRuleSync) GetRoleName() string`

GetRoleName returns the RoleName field if non-nil, zero value otherwise.

### GetRoleNameOk

`func (o *AccessRuleSync) GetRoleNameOk() (*string, bool)`

GetRoleNameOk returns a tuple with the RoleName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleName

`func (o *AccessRuleSync) SetRoleName(v string)`

SetRoleName sets RoleName field to given value.


### GetScopeName

`func (o *AccessRuleSync) GetScopeName() string`

GetScopeName returns the ScopeName field if non-nil, zero value otherwise.

### GetScopeNameOk

`func (o *AccessRuleSync) GetScopeNameOk() (*string, bool)`

GetScopeNameOk returns a tuple with the ScopeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeName

`func (o *AccessRuleSync) SetScopeName(v string)`

SetScopeName sets ScopeName field to given value.


### GetId

`func (o *AccessRuleSync) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AccessRuleSync) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AccessRuleSync) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *AccessRuleSync) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AccessRuleSync) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AccessRuleSync) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *AccessRuleSync) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *AccessRuleSync) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *AccessRuleSync) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *AccessRuleSync) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *AccessRuleSync) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *AccessRuleSync) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *AccessRuleSync) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *AccessRuleSync) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *AccessRuleSync) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetTenantId

`func (o *AccessRuleSync) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AccessRuleSync) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AccessRuleSync) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *AccessRuleSync) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetCreatedBy

`func (o *AccessRuleSync) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *AccessRuleSync) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *AccessRuleSync) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetPredefinedRole

`func (o *AccessRuleSync) GetPredefinedRole() AccessRuleSyncAllOfPredefinedRole`

GetPredefinedRole returns the PredefinedRole field if non-nil, zero value otherwise.

### GetPredefinedRoleOk

`func (o *AccessRuleSync) GetPredefinedRoleOk() (*AccessRuleSyncAllOfPredefinedRole, bool)`

GetPredefinedRoleOk returns a tuple with the PredefinedRole field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPredefinedRole

`func (o *AccessRuleSync) SetPredefinedRole(v AccessRuleSyncAllOfPredefinedRole)`

SetPredefinedRole sets PredefinedRole field to given value.

### HasPredefinedRole

`func (o *AccessRuleSync) HasPredefinedRole() bool`

HasPredefinedRole returns a boolean if a field has been set.

### SetPredefinedRoleNil

`func (o *AccessRuleSync) SetPredefinedRoleNil(b bool)`

 SetPredefinedRoleNil sets the value for PredefinedRole to be an explicit nil

### UnsetPredefinedRole
`func (o *AccessRuleSync) UnsetPredefinedRole()`

UnsetPredefinedRole ensures that no value is present for PredefinedRole, not even an explicit nil
### GetIsDeleting

`func (o *AccessRuleSync) GetIsDeleting() bool`

GetIsDeleting returns the IsDeleting field if non-nil, zero value otherwise.

### GetIsDeletingOk

`func (o *AccessRuleSync) GetIsDeletingOk() (*bool, bool)`

GetIsDeletingOk returns a tuple with the IsDeleting field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeleting

`func (o *AccessRuleSync) SetIsDeleting(v bool)`

SetIsDeleting sets IsDeleting field to given value.

### HasIsDeleting

`func (o *AccessRuleSync) HasIsDeleting() bool`

HasIsDeleting returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


