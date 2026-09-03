# App

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
**ClientId** | Pointer to **string** | The client ID of the application. | [optional] 
**Name** | **string** | The name of the application. | 
**Revoked** | Pointer to **bool** | Whether the application has been revoked. | [optional] 

## Methods

### NewApp

`func NewApp(name string, ) *App`

NewApp instantiates a new App object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAppWithDefaults

`func NewAppWithDefaults() *App`

NewAppWithDefaults instantiates a new App object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEntityType

`func (o *App) GetEntityType() AuthEntityType`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *App) GetEntityTypeOk() (*AuthEntityType, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *App) SetEntityType(v AuthEntityType)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *App) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetTenantId

`func (o *App) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *App) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *App) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *App) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetUserId

`func (o *App) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *App) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *App) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *App) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetPermitAllClusters

`func (o *App) GetPermitAllClusters() bool`

GetPermitAllClusters returns the PermitAllClusters field if non-nil, zero value otherwise.

### GetPermitAllClustersOk

`func (o *App) GetPermitAllClustersOk() (*bool, bool)`

GetPermitAllClustersOk returns a tuple with the PermitAllClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermitAllClusters

`func (o *App) SetPermitAllClusters(v bool)`

SetPermitAllClusters sets PermitAllClusters field to given value.

### HasPermitAllClusters

`func (o *App) HasPermitAllClusters() bool`

HasPermitAllClusters returns a boolean if a field has been set.

### GetPermittedClusters

`func (o *App) GetPermittedClusters() []string`

GetPermittedClusters returns the PermittedClusters field if non-nil, zero value otherwise.

### GetPermittedClustersOk

`func (o *App) GetPermittedClustersOk() (*[]string, bool)`

GetPermittedClustersOk returns a tuple with the PermittedClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermittedClusters

`func (o *App) SetPermittedClusters(v []string)`

SetPermittedClusters sets PermittedClusters field to given value.

### HasPermittedClusters

`func (o *App) HasPermittedClusters() bool`

HasPermittedClusters returns a boolean if a field has been set.

### GetRoles

`func (o *App) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *App) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *App) SetRoles(v []Role)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *App) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetCreatedAt

`func (o *App) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *App) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *App) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *App) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetClientId

`func (o *App) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *App) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *App) SetClientId(v string)`

SetClientId sets ClientId field to given value.

### HasClientId

`func (o *App) HasClientId() bool`

HasClientId returns a boolean if a field has been set.

### GetName

`func (o *App) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *App) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *App) SetName(v string)`

SetName sets Name field to given value.


### GetRevoked

`func (o *App) GetRevoked() bool`

GetRevoked returns the Revoked field if non-nil, zero value otherwise.

### GetRevokedOk

`func (o *App) GetRevokedOk() (*bool, bool)`

GetRevokedOk returns a tuple with the Revoked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevoked

`func (o *App) SetRevoked(v bool)`

SetRevoked sets Revoked field to given value.

### HasRevoked

`func (o *App) HasRevoked() bool`

HasRevoked returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


