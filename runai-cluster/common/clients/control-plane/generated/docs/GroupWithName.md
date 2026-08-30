# GroupWithName

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EntityType** | Pointer to **string** |  | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**UserId** | Pointer to **string** | Group&#39;s name | [optional] 
**PermitAllClusters** | Pointer to **bool** |  | [optional] 
**PermittedClusters** | Pointer to **[]string** | A list of clusters that the user or application can access. | [optional] 
**Roles** | Pointer to [**[]Role**](Role.md) |  | [optional] 
**CreatedBy** | Pointer to **string** | uid of the user | [optional] 
**Username** | Pointer to **string** | group name | [optional] 

## Methods

### NewGroupWithName

`func NewGroupWithName() *GroupWithName`

NewGroupWithName instantiates a new GroupWithName object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGroupWithNameWithDefaults

`func NewGroupWithNameWithDefaults() *GroupWithName`

NewGroupWithNameWithDefaults instantiates a new GroupWithName object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntityType

`func (o *GroupWithName) GetEntityType() string`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *GroupWithName) GetEntityTypeOk() (*string, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *GroupWithName) SetEntityType(v string)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *GroupWithName) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetTenantId

`func (o *GroupWithName) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *GroupWithName) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *GroupWithName) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *GroupWithName) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetUserId

`func (o *GroupWithName) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *GroupWithName) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *GroupWithName) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *GroupWithName) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetPermitAllClusters

`func (o *GroupWithName) GetPermitAllClusters() bool`

GetPermitAllClusters returns the PermitAllClusters field if non-nil, zero value otherwise.

### GetPermitAllClustersOk

`func (o *GroupWithName) GetPermitAllClustersOk() (*bool, bool)`

GetPermitAllClustersOk returns a tuple with the PermitAllClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermitAllClusters

`func (o *GroupWithName) SetPermitAllClusters(v bool)`

SetPermitAllClusters sets PermitAllClusters field to given value.

### HasPermitAllClusters

`func (o *GroupWithName) HasPermitAllClusters() bool`

HasPermitAllClusters returns a boolean if a field has been set.

### GetPermittedClusters

`func (o *GroupWithName) GetPermittedClusters() []string`

GetPermittedClusters returns the PermittedClusters field if non-nil, zero value otherwise.

### GetPermittedClustersOk

`func (o *GroupWithName) GetPermittedClustersOk() (*[]string, bool)`

GetPermittedClustersOk returns a tuple with the PermittedClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermittedClusters

`func (o *GroupWithName) SetPermittedClusters(v []string)`

SetPermittedClusters sets PermittedClusters field to given value.

### HasPermittedClusters

`func (o *GroupWithName) HasPermittedClusters() bool`

HasPermittedClusters returns a boolean if a field has been set.

### GetRoles

`func (o *GroupWithName) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *GroupWithName) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *GroupWithName) SetRoles(v []Role)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *GroupWithName) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetCreatedBy

`func (o *GroupWithName) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *GroupWithName) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *GroupWithName) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *GroupWithName) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUsername

`func (o *GroupWithName) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *GroupWithName) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *GroupWithName) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *GroupWithName) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


