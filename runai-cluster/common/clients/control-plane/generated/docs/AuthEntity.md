# AuthEntity

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

## Methods

### NewAuthEntity

`func NewAuthEntity() *AuthEntity`

NewAuthEntity instantiates a new AuthEntity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthEntityWithDefaults

`func NewAuthEntityWithDefaults() *AuthEntity`

NewAuthEntityWithDefaults instantiates a new AuthEntity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntityType

`func (o *AuthEntity) GetEntityType() AuthEntityType`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *AuthEntity) GetEntityTypeOk() (*AuthEntityType, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *AuthEntity) SetEntityType(v AuthEntityType)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *AuthEntity) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetTenantId

`func (o *AuthEntity) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *AuthEntity) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *AuthEntity) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *AuthEntity) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetUserId

`func (o *AuthEntity) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *AuthEntity) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *AuthEntity) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *AuthEntity) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetPermitAllClusters

`func (o *AuthEntity) GetPermitAllClusters() bool`

GetPermitAllClusters returns the PermitAllClusters field if non-nil, zero value otherwise.

### GetPermitAllClustersOk

`func (o *AuthEntity) GetPermitAllClustersOk() (*bool, bool)`

GetPermitAllClustersOk returns a tuple with the PermitAllClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermitAllClusters

`func (o *AuthEntity) SetPermitAllClusters(v bool)`

SetPermitAllClusters sets PermitAllClusters field to given value.

### HasPermitAllClusters

`func (o *AuthEntity) HasPermitAllClusters() bool`

HasPermitAllClusters returns a boolean if a field has been set.

### GetPermittedClusters

`func (o *AuthEntity) GetPermittedClusters() []string`

GetPermittedClusters returns the PermittedClusters field if non-nil, zero value otherwise.

### GetPermittedClustersOk

`func (o *AuthEntity) GetPermittedClustersOk() (*[]string, bool)`

GetPermittedClustersOk returns a tuple with the PermittedClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermittedClusters

`func (o *AuthEntity) SetPermittedClusters(v []string)`

SetPermittedClusters sets PermittedClusters field to given value.

### HasPermittedClusters

`func (o *AuthEntity) HasPermittedClusters() bool`

HasPermittedClusters returns a boolean if a field has been set.

### GetRoles

`func (o *AuthEntity) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *AuthEntity) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *AuthEntity) SetRoles(v []Role)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *AuthEntity) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetCreatedAt

`func (o *AuthEntity) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *AuthEntity) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *AuthEntity) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *AuthEntity) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


