# Role

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Description** | **string** |  | 
**Permissions** | [**[]Permission**](Permission.md) |  | 
**Id** | **int32** |  | 
**CreatedAt** | **time.Time** |  | 
**UpdatedAt** | **time.Time** |  | 
**DeletedAt** | Pointer to **NullableTime** |  | [optional] 
**TenantId** | Pointer to **int32** |  | [optional] 
**CreatedBy** | **string** |  | 
**Assignments** | Pointer to [**[]RoleAllOfAssignments**](RoleAllOfAssignments.md) |  | [optional] 

## Methods

### NewRole

`func NewRole(name string, description string, permissions []Permission, id int32, createdAt time.Time, updatedAt time.Time, createdBy string, ) *Role`

NewRole instantiates a new Role object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRoleWithDefaults

`func NewRoleWithDefaults() *Role`

NewRoleWithDefaults instantiates a new Role object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Role) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Role) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Role) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *Role) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Role) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Role) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetPermissions

`func (o *Role) GetPermissions() []Permission`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *Role) GetPermissionsOk() (*[]Permission, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *Role) SetPermissions(v []Permission)`

SetPermissions sets Permissions field to given value.


### GetId

`func (o *Role) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Role) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Role) SetId(v int32)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *Role) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Role) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Role) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *Role) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Role) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Role) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetDeletedAt

`func (o *Role) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Role) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Role) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *Role) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *Role) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Role) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetTenantId

`func (o *Role) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *Role) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *Role) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *Role) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetCreatedBy

`func (o *Role) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Role) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Role) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetAssignments

`func (o *Role) GetAssignments() []RoleAllOfAssignments`

GetAssignments returns the Assignments field if non-nil, zero value otherwise.

### GetAssignmentsOk

`func (o *Role) GetAssignmentsOk() (*[]RoleAllOfAssignments, bool)`

GetAssignmentsOk returns a tuple with the Assignments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssignments

`func (o *Role) SetAssignments(v []RoleAllOfAssignments)`

SetAssignments sets Assignments field to given value.

### HasAssignments

`func (o *Role) HasAssignments() bool`

HasAssignments returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


