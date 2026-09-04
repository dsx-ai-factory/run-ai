# GroupCreationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Username** | **string** | Group&#39;s name | 
**Email** | **string** | Empty email address | 
**Roles** | [**[]Role**](Role.md) |  | 
**EntityType** | Pointer to **string** |  | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**PermitAllClusters** | Pointer to **bool** |  | [optional] 
**UserId** | Pointer to **string** | Group&#39;s name | [optional] 
**PermittedClusters** | Pointer to **[]string** |  | [optional] 
**CreatedBy** | Pointer to **string** | uid of the user | [optional] 

## Methods

### NewGroupCreationResponse

`func NewGroupCreationResponse(username string, email string, roles []Role, ) *GroupCreationResponse`

NewGroupCreationResponse instantiates a new GroupCreationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGroupCreationResponseWithDefaults

`func NewGroupCreationResponseWithDefaults() *GroupCreationResponse`

NewGroupCreationResponseWithDefaults instantiates a new GroupCreationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsername

`func (o *GroupCreationResponse) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *GroupCreationResponse) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *GroupCreationResponse) SetUsername(v string)`

SetUsername sets Username field to given value.


### GetEmail

`func (o *GroupCreationResponse) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *GroupCreationResponse) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *GroupCreationResponse) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetRoles

`func (o *GroupCreationResponse) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *GroupCreationResponse) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *GroupCreationResponse) SetRoles(v []Role)`

SetRoles sets Roles field to given value.


### GetEntityType

`func (o *GroupCreationResponse) GetEntityType() string`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *GroupCreationResponse) GetEntityTypeOk() (*string, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *GroupCreationResponse) SetEntityType(v string)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *GroupCreationResponse) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetTenantId

`func (o *GroupCreationResponse) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *GroupCreationResponse) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *GroupCreationResponse) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *GroupCreationResponse) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetPermitAllClusters

`func (o *GroupCreationResponse) GetPermitAllClusters() bool`

GetPermitAllClusters returns the PermitAllClusters field if non-nil, zero value otherwise.

### GetPermitAllClustersOk

`func (o *GroupCreationResponse) GetPermitAllClustersOk() (*bool, bool)`

GetPermitAllClustersOk returns a tuple with the PermitAllClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermitAllClusters

`func (o *GroupCreationResponse) SetPermitAllClusters(v bool)`

SetPermitAllClusters sets PermitAllClusters field to given value.

### HasPermitAllClusters

`func (o *GroupCreationResponse) HasPermitAllClusters() bool`

HasPermitAllClusters returns a boolean if a field has been set.

### GetUserId

`func (o *GroupCreationResponse) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *GroupCreationResponse) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *GroupCreationResponse) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *GroupCreationResponse) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetPermittedClusters

`func (o *GroupCreationResponse) GetPermittedClusters() []string`

GetPermittedClusters returns the PermittedClusters field if non-nil, zero value otherwise.

### GetPermittedClustersOk

`func (o *GroupCreationResponse) GetPermittedClustersOk() (*[]string, bool)`

GetPermittedClustersOk returns a tuple with the PermittedClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermittedClusters

`func (o *GroupCreationResponse) SetPermittedClusters(v []string)`

SetPermittedClusters sets PermittedClusters field to given value.

### HasPermittedClusters

`func (o *GroupCreationResponse) HasPermittedClusters() bool`

HasPermittedClusters returns a boolean if a field has been set.

### GetCreatedBy

`func (o *GroupCreationResponse) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *GroupCreationResponse) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *GroupCreationResponse) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *GroupCreationResponse) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


