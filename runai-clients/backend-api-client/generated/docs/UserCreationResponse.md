# UserCreationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** | Email address of the user. | 
**Roles** | [**[]Role**](Role.md) |  | 
**EntityType** | Pointer to [**AuthEntityType**](AuthEntityType.md) |  | [optional] 
**TenantId** | Pointer to **int32** | The id of the tenant. | [optional] 
**Password** | **string** | The user&#39;s password. | 
**NeedToChangePassword** | Pointer to **bool** | True if the user is requested to change his password upon next login. | [optional] 
**PermitAllClusters** | Pointer to **bool** |  | [optional] 
**UserId** | Pointer to **string** | Unique identifier of the user | [optional] 
**PermittedClusters** | Pointer to **[]string** |  | [optional] 

## Methods

### NewUserCreationResponse

`func NewUserCreationResponse(email string, roles []Role, password string, ) *UserCreationResponse`

NewUserCreationResponse instantiates a new UserCreationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreationResponseWithDefaults

`func NewUserCreationResponseWithDefaults() *UserCreationResponse`

NewUserCreationResponseWithDefaults instantiates a new UserCreationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *UserCreationResponse) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserCreationResponse) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserCreationResponse) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetRoles

`func (o *UserCreationResponse) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *UserCreationResponse) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *UserCreationResponse) SetRoles(v []Role)`

SetRoles sets Roles field to given value.


### GetEntityType

`func (o *UserCreationResponse) GetEntityType() AuthEntityType`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *UserCreationResponse) GetEntityTypeOk() (*AuthEntityType, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *UserCreationResponse) SetEntityType(v AuthEntityType)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *UserCreationResponse) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetTenantId

`func (o *UserCreationResponse) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *UserCreationResponse) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *UserCreationResponse) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *UserCreationResponse) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetPassword

`func (o *UserCreationResponse) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *UserCreationResponse) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *UserCreationResponse) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetNeedToChangePassword

`func (o *UserCreationResponse) GetNeedToChangePassword() bool`

GetNeedToChangePassword returns the NeedToChangePassword field if non-nil, zero value otherwise.

### GetNeedToChangePasswordOk

`func (o *UserCreationResponse) GetNeedToChangePasswordOk() (*bool, bool)`

GetNeedToChangePasswordOk returns a tuple with the NeedToChangePassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNeedToChangePassword

`func (o *UserCreationResponse) SetNeedToChangePassword(v bool)`

SetNeedToChangePassword sets NeedToChangePassword field to given value.

### HasNeedToChangePassword

`func (o *UserCreationResponse) HasNeedToChangePassword() bool`

HasNeedToChangePassword returns a boolean if a field has been set.

### GetPermitAllClusters

`func (o *UserCreationResponse) GetPermitAllClusters() bool`

GetPermitAllClusters returns the PermitAllClusters field if non-nil, zero value otherwise.

### GetPermitAllClustersOk

`func (o *UserCreationResponse) GetPermitAllClustersOk() (*bool, bool)`

GetPermitAllClustersOk returns a tuple with the PermitAllClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermitAllClusters

`func (o *UserCreationResponse) SetPermitAllClusters(v bool)`

SetPermitAllClusters sets PermitAllClusters field to given value.

### HasPermitAllClusters

`func (o *UserCreationResponse) HasPermitAllClusters() bool`

HasPermitAllClusters returns a boolean if a field has been set.

### GetUserId

`func (o *UserCreationResponse) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *UserCreationResponse) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *UserCreationResponse) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *UserCreationResponse) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetPermittedClusters

`func (o *UserCreationResponse) GetPermittedClusters() []string`

GetPermittedClusters returns the PermittedClusters field if non-nil, zero value otherwise.

### GetPermittedClustersOk

`func (o *UserCreationResponse) GetPermittedClustersOk() (*[]string, bool)`

GetPermittedClustersOk returns a tuple with the PermittedClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermittedClusters

`func (o *UserCreationResponse) SetPermittedClusters(v []string)`

SetPermittedClusters sets PermittedClusters field to given value.

### HasPermittedClusters

`func (o *UserCreationResponse) HasPermittedClusters() bool`

HasPermittedClusters returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


