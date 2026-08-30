# UpdateSecuritySettingByKeyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** |  | [optional] 
**IdpAlias** | Pointer to **NullableString** |  | [optional] 
**IdleTimeout** | Pointer to **NullableFloat32** | Duration in minutes that a user can remain inactive before their browser session automatically expires and requires re-authentication | [optional] [default to 0]
**Uri** | Pointer to **NullableString** | The URI where users will be redirected after logging out of the application | [optional] 

## Methods

### NewUpdateSecuritySettingByKeyRequest

`func NewUpdateSecuritySettingByKeyRequest() *UpdateSecuritySettingByKeyRequest`

NewUpdateSecuritySettingByKeyRequest instantiates a new UpdateSecuritySettingByKeyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateSecuritySettingByKeyRequestWithDefaults

`func NewUpdateSecuritySettingByKeyRequestWithDefaults() *UpdateSecuritySettingByKeyRequest`

NewUpdateSecuritySettingByKeyRequestWithDefaults instantiates a new UpdateSecuritySettingByKeyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *UpdateSecuritySettingByKeyRequest) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *UpdateSecuritySettingByKeyRequest) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *UpdateSecuritySettingByKeyRequest) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *UpdateSecuritySettingByKeyRequest) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *UpdateSecuritySettingByKeyRequest) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *UpdateSecuritySettingByKeyRequest) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetIdpAlias

`func (o *UpdateSecuritySettingByKeyRequest) GetIdpAlias() string`

GetIdpAlias returns the IdpAlias field if non-nil, zero value otherwise.

### GetIdpAliasOk

`func (o *UpdateSecuritySettingByKeyRequest) GetIdpAliasOk() (*string, bool)`

GetIdpAliasOk returns a tuple with the IdpAlias field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdpAlias

`func (o *UpdateSecuritySettingByKeyRequest) SetIdpAlias(v string)`

SetIdpAlias sets IdpAlias field to given value.

### HasIdpAlias

`func (o *UpdateSecuritySettingByKeyRequest) HasIdpAlias() bool`

HasIdpAlias returns a boolean if a field has been set.

### SetIdpAliasNil

`func (o *UpdateSecuritySettingByKeyRequest) SetIdpAliasNil(b bool)`

 SetIdpAliasNil sets the value for IdpAlias to be an explicit nil

### UnsetIdpAlias
`func (o *UpdateSecuritySettingByKeyRequest) UnsetIdpAlias()`

UnsetIdpAlias ensures that no value is present for IdpAlias, not even an explicit nil
### GetIdleTimeout

`func (o *UpdateSecuritySettingByKeyRequest) GetIdleTimeout() float32`

GetIdleTimeout returns the IdleTimeout field if non-nil, zero value otherwise.

### GetIdleTimeoutOk

`func (o *UpdateSecuritySettingByKeyRequest) GetIdleTimeoutOk() (*float32, bool)`

GetIdleTimeoutOk returns a tuple with the IdleTimeout field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdleTimeout

`func (o *UpdateSecuritySettingByKeyRequest) SetIdleTimeout(v float32)`

SetIdleTimeout sets IdleTimeout field to given value.

### HasIdleTimeout

`func (o *UpdateSecuritySettingByKeyRequest) HasIdleTimeout() bool`

HasIdleTimeout returns a boolean if a field has been set.

### SetIdleTimeoutNil

`func (o *UpdateSecuritySettingByKeyRequest) SetIdleTimeoutNil(b bool)`

 SetIdleTimeoutNil sets the value for IdleTimeout to be an explicit nil

### UnsetIdleTimeout
`func (o *UpdateSecuritySettingByKeyRequest) UnsetIdleTimeout()`

UnsetIdleTimeout ensures that no value is present for IdleTimeout, not even an explicit nil
### GetUri

`func (o *UpdateSecuritySettingByKeyRequest) GetUri() string`

GetUri returns the Uri field if non-nil, zero value otherwise.

### GetUriOk

`func (o *UpdateSecuritySettingByKeyRequest) GetUriOk() (*string, bool)`

GetUriOk returns a tuple with the Uri field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUri

`func (o *UpdateSecuritySettingByKeyRequest) SetUri(v string)`

SetUri sets Uri field to given value.

### HasUri

`func (o *UpdateSecuritySettingByKeyRequest) HasUri() bool`

HasUri returns a boolean if a field has been set.

### SetUriNil

`func (o *UpdateSecuritySettingByKeyRequest) SetUriNil(b bool)`

 SetUriNil sets the value for Uri to be an explicit nil

### UnsetUri
`func (o *UpdateSecuritySettingByKeyRequest) UnsetUri()`

UnsetUri ensures that no value is present for Uri, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


