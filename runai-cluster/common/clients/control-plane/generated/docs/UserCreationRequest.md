# UserCreationRequest

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

### NewUserCreationRequest

`func NewUserCreationRequest(email string, roles []Role, password string, ) *UserCreationRequest`

NewUserCreationRequest instantiates a new UserCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreationRequestWithDefaults

`func NewUserCreationRequestWithDefaults() *UserCreationRequest`

NewUserCreationRequestWithDefaults instantiates a new UserCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *UserCreationRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserCreationRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserCreationRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetRoles

`func (o *UserCreationRequest) GetRoles() []Role`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *UserCreationRequest) GetRolesOk() (*[]Role, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *UserCreationRequest) SetRoles(v []Role)`

SetRoles sets Roles field to given value.


### GetEntityType

`func (o *UserCreationRequest) GetEntityType() AuthEntityType`

GetEntityType returns the EntityType field if non-nil, zero value otherwise.

### GetEntityTypeOk

`func (o *UserCreationRequest) GetEntityTypeOk() (*AuthEntityType, bool)`

GetEntityTypeOk returns a tuple with the EntityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEntityType

`func (o *UserCreationRequest) SetEntityType(v AuthEntityType)`

SetEntityType sets EntityType field to given value.

### HasEntityType

`func (o *UserCreationRequest) HasEntityType() bool`

HasEntityType returns a boolean if a field has been set.

### GetTenantId

`func (o *UserCreationRequest) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *UserCreationRequest) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *UserCreationRequest) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *UserCreationRequest) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetPassword

`func (o *UserCreationRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *UserCreationRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *UserCreationRequest) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetNeedToChangePassword

`func (o *UserCreationRequest) GetNeedToChangePassword() bool`

GetNeedToChangePassword returns the NeedToChangePassword field if non-nil, zero value otherwise.

### GetNeedToChangePasswordOk

`func (o *UserCreationRequest) GetNeedToChangePasswordOk() (*bool, bool)`

GetNeedToChangePasswordOk returns a tuple with the NeedToChangePassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNeedToChangePassword

`func (o *UserCreationRequest) SetNeedToChangePassword(v bool)`

SetNeedToChangePassword sets NeedToChangePassword field to given value.

### HasNeedToChangePassword

`func (o *UserCreationRequest) HasNeedToChangePassword() bool`

HasNeedToChangePassword returns a boolean if a field has been set.

### GetPermitAllClusters

`func (o *UserCreationRequest) GetPermitAllClusters() bool`

GetPermitAllClusters returns the PermitAllClusters field if non-nil, zero value otherwise.

### GetPermitAllClustersOk

`func (o *UserCreationRequest) GetPermitAllClustersOk() (*bool, bool)`

GetPermitAllClustersOk returns a tuple with the PermitAllClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermitAllClusters

`func (o *UserCreationRequest) SetPermitAllClusters(v bool)`

SetPermitAllClusters sets PermitAllClusters field to given value.

### HasPermitAllClusters

`func (o *UserCreationRequest) HasPermitAllClusters() bool`

HasPermitAllClusters returns a boolean if a field has been set.

### GetUserId

`func (o *UserCreationRequest) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *UserCreationRequest) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *UserCreationRequest) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *UserCreationRequest) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetPermittedClusters

`func (o *UserCreationRequest) GetPermittedClusters() []string`

GetPermittedClusters returns the PermittedClusters field if non-nil, zero value otherwise.

### GetPermittedClustersOk

`func (o *UserCreationRequest) GetPermittedClustersOk() (*[]string, bool)`

GetPermittedClustersOk returns a tuple with the PermittedClusters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermittedClusters

`func (o *UserCreationRequest) SetPermittedClusters(v []string)`

SetPermittedClusters sets PermittedClusters field to given value.

### HasPermittedClusters

`func (o *UserCreationRequest) HasPermittedClusters() bool`

HasPermittedClusters returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


