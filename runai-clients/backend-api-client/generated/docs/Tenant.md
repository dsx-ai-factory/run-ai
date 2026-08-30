# Tenant

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AuthSettingsId** | Pointer to **int32** |  | [optional] 
**DisplayName** | Pointer to **string** |  | [optional] 
**SsoAuthSettingsId** | Pointer to **int32** |  | [optional] 
**AuthRealm** | Pointer to **string** |  | [optional] 
**Id** | **int32** | The id of the tenant. | 
**Name** | Pointer to **string** |  | [optional] 
**EulaSigningUser** | Pointer to **string** |  | [optional] 
**ContractType** | Pointer to **string** |  | [optional] 
**SmgEnabled** | Pointer to **bool** |  | [optional] 
**AuthClientID** | Pointer to **string** |  | [optional] 
**CliClientId** | Pointer to **string** |  | [optional] 

## Methods

### NewTenant

`func NewTenant(id int32, ) *Tenant`

NewTenant instantiates a new Tenant object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTenantWithDefaults

`func NewTenantWithDefaults() *Tenant`

NewTenantWithDefaults instantiates a new Tenant object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthSettingsId

`func (o *Tenant) GetAuthSettingsId() int32`

GetAuthSettingsId returns the AuthSettingsId field if non-nil, zero value otherwise.

### GetAuthSettingsIdOk

`func (o *Tenant) GetAuthSettingsIdOk() (*int32, bool)`

GetAuthSettingsIdOk returns a tuple with the AuthSettingsId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthSettingsId

`func (o *Tenant) SetAuthSettingsId(v int32)`

SetAuthSettingsId sets AuthSettingsId field to given value.

### HasAuthSettingsId

`func (o *Tenant) HasAuthSettingsId() bool`

HasAuthSettingsId returns a boolean if a field has been set.

### GetDisplayName

`func (o *Tenant) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *Tenant) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *Tenant) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *Tenant) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetSsoAuthSettingsId

`func (o *Tenant) GetSsoAuthSettingsId() int32`

GetSsoAuthSettingsId returns the SsoAuthSettingsId field if non-nil, zero value otherwise.

### GetSsoAuthSettingsIdOk

`func (o *Tenant) GetSsoAuthSettingsIdOk() (*int32, bool)`

GetSsoAuthSettingsIdOk returns a tuple with the SsoAuthSettingsId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSsoAuthSettingsId

`func (o *Tenant) SetSsoAuthSettingsId(v int32)`

SetSsoAuthSettingsId sets SsoAuthSettingsId field to given value.

### HasSsoAuthSettingsId

`func (o *Tenant) HasSsoAuthSettingsId() bool`

HasSsoAuthSettingsId returns a boolean if a field has been set.

### GetAuthRealm

`func (o *Tenant) GetAuthRealm() string`

GetAuthRealm returns the AuthRealm field if non-nil, zero value otherwise.

### GetAuthRealmOk

`func (o *Tenant) GetAuthRealmOk() (*string, bool)`

GetAuthRealmOk returns a tuple with the AuthRealm field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthRealm

`func (o *Tenant) SetAuthRealm(v string)`

SetAuthRealm sets AuthRealm field to given value.

### HasAuthRealm

`func (o *Tenant) HasAuthRealm() bool`

HasAuthRealm returns a boolean if a field has been set.

### GetId

`func (o *Tenant) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Tenant) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Tenant) SetId(v int32)`

SetId sets Id field to given value.


### GetName

`func (o *Tenant) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Tenant) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Tenant) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Tenant) HasName() bool`

HasName returns a boolean if a field has been set.

### GetEulaSigningUser

`func (o *Tenant) GetEulaSigningUser() string`

GetEulaSigningUser returns the EulaSigningUser field if non-nil, zero value otherwise.

### GetEulaSigningUserOk

`func (o *Tenant) GetEulaSigningUserOk() (*string, bool)`

GetEulaSigningUserOk returns a tuple with the EulaSigningUser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEulaSigningUser

`func (o *Tenant) SetEulaSigningUser(v string)`

SetEulaSigningUser sets EulaSigningUser field to given value.

### HasEulaSigningUser

`func (o *Tenant) HasEulaSigningUser() bool`

HasEulaSigningUser returns a boolean if a field has been set.

### GetContractType

`func (o *Tenant) GetContractType() string`

GetContractType returns the ContractType field if non-nil, zero value otherwise.

### GetContractTypeOk

`func (o *Tenant) GetContractTypeOk() (*string, bool)`

GetContractTypeOk returns a tuple with the ContractType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContractType

`func (o *Tenant) SetContractType(v string)`

SetContractType sets ContractType field to given value.

### HasContractType

`func (o *Tenant) HasContractType() bool`

HasContractType returns a boolean if a field has been set.

### GetSmgEnabled

`func (o *Tenant) GetSmgEnabled() bool`

GetSmgEnabled returns the SmgEnabled field if non-nil, zero value otherwise.

### GetSmgEnabledOk

`func (o *Tenant) GetSmgEnabledOk() (*bool, bool)`

GetSmgEnabledOk returns a tuple with the SmgEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmgEnabled

`func (o *Tenant) SetSmgEnabled(v bool)`

SetSmgEnabled sets SmgEnabled field to given value.

### HasSmgEnabled

`func (o *Tenant) HasSmgEnabled() bool`

HasSmgEnabled returns a boolean if a field has been set.

### GetAuthClientID

`func (o *Tenant) GetAuthClientID() string`

GetAuthClientID returns the AuthClientID field if non-nil, zero value otherwise.

### GetAuthClientIDOk

`func (o *Tenant) GetAuthClientIDOk() (*string, bool)`

GetAuthClientIDOk returns a tuple with the AuthClientID field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthClientID

`func (o *Tenant) SetAuthClientID(v string)`

SetAuthClientID sets AuthClientID field to given value.

### HasAuthClientID

`func (o *Tenant) HasAuthClientID() bool`

HasAuthClientID returns a boolean if a field has been set.

### GetCliClientId

`func (o *Tenant) GetCliClientId() string`

GetCliClientId returns the CliClientId field if non-nil, zero value otherwise.

### GetCliClientIdOk

`func (o *Tenant) GetCliClientIdOk() (*string, bool)`

GetCliClientIdOk returns a tuple with the CliClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCliClientId

`func (o *Tenant) SetCliClientId(v string)`

SetCliClientId sets CliClientId field to given value.

### HasCliClientId

`func (o *Tenant) HasCliClientId() bool`

HasCliClientId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


