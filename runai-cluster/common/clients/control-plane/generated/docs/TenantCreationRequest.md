# TenantCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Email** | **string** |  | 
**Password** | Pointer to **string** |  | [optional] 
**ContractType** | Pointer to **string** |  | [optional] 
**Role** | Pointer to [**NullableRoles1**](Roles1.md) |  | [optional] 
**MetricStore** | Pointer to [**NullableTenantCreationRequestMetricStore**](TenantCreationRequestMetricStore.md) |  | [optional] 

## Methods

### NewTenantCreationRequest

`func NewTenantCreationRequest(name string, email string, ) *TenantCreationRequest`

NewTenantCreationRequest instantiates a new TenantCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantCreationRequestWithDefaults

`func NewTenantCreationRequestWithDefaults() *TenantCreationRequest`

NewTenantCreationRequestWithDefaults instantiates a new TenantCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *TenantCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TenantCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TenantCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetEmail

`func (o *TenantCreationRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *TenantCreationRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *TenantCreationRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetPassword

`func (o *TenantCreationRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *TenantCreationRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *TenantCreationRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *TenantCreationRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetContractType

`func (o *TenantCreationRequest) GetContractType() string`

GetContractType returns the ContractType field if non-nil, zero value otherwise.

### GetContractTypeOk

`func (o *TenantCreationRequest) GetContractTypeOk() (*string, bool)`

GetContractTypeOk returns a tuple with the ContractType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractType

`func (o *TenantCreationRequest) SetContractType(v string)`

SetContractType sets ContractType field to given value.

### HasContractType

`func (o *TenantCreationRequest) HasContractType() bool`

HasContractType returns a boolean if a field has been set.

### GetRole

`func (o *TenantCreationRequest) GetRole() Roles1`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *TenantCreationRequest) GetRoleOk() (*Roles1, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *TenantCreationRequest) SetRole(v Roles1)`

SetRole sets Role field to given value.

### HasRole

`func (o *TenantCreationRequest) HasRole() bool`

HasRole returns a boolean if a field has been set.

### SetRoleNil

`func (o *TenantCreationRequest) SetRoleNil(b bool)`

 SetRoleNil sets the value for Role to be an explicit nil

### UnsetRole
`func (o *TenantCreationRequest) UnsetRole()`

UnsetRole ensures that no value is present for Role, not even an explicit nil
### GetMetricStore

`func (o *TenantCreationRequest) GetMetricStore() TenantCreationRequestMetricStore`

GetMetricStore returns the MetricStore field if non-nil, zero value otherwise.

### GetMetricStoreOk

`func (o *TenantCreationRequest) GetMetricStoreOk() (*TenantCreationRequestMetricStore, bool)`

GetMetricStoreOk returns a tuple with the MetricStore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetricStore

`func (o *TenantCreationRequest) SetMetricStore(v TenantCreationRequestMetricStore)`

SetMetricStore sets MetricStore field to given value.

### HasMetricStore

`func (o *TenantCreationRequest) HasMetricStore() bool`

HasMetricStore returns a boolean if a field has been set.

### SetMetricStoreNil

`func (o *TenantCreationRequest) SetMetricStoreNil(b bool)`

 SetMetricStoreNil sets the value for MetricStore to be an explicit nil

### UnsetMetricStore
`func (o *TenantCreationRequest) UnsetMetricStore()`

UnsetMetricStore ensures that no value is present for MetricStore, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


