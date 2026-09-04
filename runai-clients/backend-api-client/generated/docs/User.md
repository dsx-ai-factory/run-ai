# User

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EntityType** | Pointer to [**AuthEntityType**](AuthEntityType.md) |  | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**UserId** | Pointer to **string** | Unique identifier of the user | [optional] 
**PermitAllClusters** | Pointer to **bool** |  | [optional] 
**PermittedClusters** | Pointer to **[]string** | A list of clusters that the user or application can access. | [optional] 
**Roles** | Pointer to [**[]Role**](Role.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** | The creation date of the application. | [optional] 
**Email** | **string** | Email address of the user. | 
**LastLogin** | Pointer to **time.Time** | The last time the user logged into the system. | [optional] 

## Methods

### NewUser

`func NewUser(email string, ) *User`

NewUser instantiates a new User object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserWithDefaults

`func NewUserWithDefaults() *User`

NewUserWithDefaults instantiates a new User object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntityType

`func (o *User) GetEntityType() AuthEntityType`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *User) GetEntityTypeOk() (*AuthEntityType, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *User) SetEntityType(v AuthEntityType)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *User) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetTenantId

`func (o *User) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *User) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *User) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *User) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetUserId

`func (o *User) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *User) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *User) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *User) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetPermitAllClusters

`func (o *User) GetPermitAllClusters() bool`

GetPermitAllClusters returns the PermitAllClusters field if non-nil, zero value otherwise.

### GetPermitAllClustersOk

`func (o *User) GetPermitAllClustersOk() (*bool, bool)`

GetPermitAllClustersOk returns a tuple with the PermitAllClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermitAllClusters

`func (o *User) SetPermitAllClusters(v bool)`

SetPermitAllClusters sets PermitAllClusters field to given value.

### HasPermitAllClusters

`func (o *User) HasPermitAllClusters() bool`

HasPermitAllClusters returns a boolean if a field has been set.

### GetPermittedClusters

`func (o *User) GetPermittedClusters() []string`

GetPermittedClusters returns the PermittedClusters field if non-nil, zero value otherwise.

### GetPermittedClustersOk

`func (o *User) GetPermittedClustersOk() (*[]string, bool)`

GetPermittedClustersOk returns a tuple with the PermittedClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermittedClusters

`func (o *User) SetPermittedClusters(v []string)`

SetPermittedClusters sets PermittedClusters field to given value.

### HasPermittedClusters

`func (o *User) HasPermittedClusters() bool`

HasPermittedClusters returns a boolean if a field has been set.

### GetRoles

`func (o *User) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *User) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *User) SetRoles(v []Role)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *User) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetCreatedAt

`func (o *User) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *User) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *User) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *User) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetEmail

`func (o *User) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *User) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *User) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetLastLogin

`func (o *User) GetLastLogin() time.Time`

GetLastLogin returns the LastLogin field if non-nil, zero value otherwise.

### GetLastLoginOk

`func (o *User) GetLastLoginOk() (*time.Time, bool)`

GetLastLoginOk returns a tuple with the LastLogin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastLogin

`func (o *User) SetLastLogin(v time.Time)`

SetLastLogin sets LastLogin field to given value.

### HasLastLogin

`func (o *User) HasLastLogin() bool`

HasLastLogin returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


