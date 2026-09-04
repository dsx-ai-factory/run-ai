# UpdateBrowserSessionTimeoutSettingValue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IdleTimeout** | Pointer to **NullableFloat32** | Duration in minutes that a user can remain inactive before their browser session automatically expires and requires re-authentication | [optional] [default to 0]

## Methods

### NewUpdateBrowserSessionTimeoutSettingValue

`func NewUpdateBrowserSessionTimeoutSettingValue() *UpdateBrowserSessionTimeoutSettingValue`

NewUpdateBrowserSessionTimeoutSettingValue instantiates a new UpdateBrowserSessionTimeoutSettingValue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateBrowserSessionTimeoutSettingValueWithDefaults

`func NewUpdateBrowserSessionTimeoutSettingValueWithDefaults() *UpdateBrowserSessionTimeoutSettingValue`

NewUpdateBrowserSessionTimeoutSettingValueWithDefaults instantiates a new UpdateBrowserSessionTimeoutSettingValue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIdleTimeout

`func (o *UpdateBrowserSessionTimeoutSettingValue) GetIdleTimeout() float32`

GetIdleTimeout returns the IdleTimeout field if non-nil, zero value otherwise.

### GetIdleTimeoutOk

`func (o *UpdateBrowserSessionTimeoutSettingValue) GetIdleTimeoutOk() (*float32, bool)`

GetIdleTimeoutOk returns a tuple with the IdleTimeout field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIdleTimeout

`func (o *UpdateBrowserSessionTimeoutSettingValue) SetIdleTimeout(v float32)`

SetIdleTimeout sets IdleTimeout field to given value.

### HasIdleTimeout

`func (o *UpdateBrowserSessionTimeoutSettingValue) HasIdleTimeout() bool`

HasIdleTimeout returns a boolean if a field has been set.

### SetIdleTimeoutNil

`func (o *UpdateBrowserSessionTimeoutSettingValue) SetIdleTimeoutNil(b bool)`

 SetIdleTimeoutNil sets the value for IdleTimeout to be an explicit nil

### UnsetIdleTimeout
`func (o *UpdateBrowserSessionTimeoutSettingValue) UnsetIdleTimeout()`

UnsetIdleTimeout ensures that no value is present for IdleTimeout, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


