# DetailedNotificationStateSlack

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** |  | [optional] 
**Settings** | Pointer to [**SlackDetailsSettings**](SlackDetailsSettings.md) |  | [optional] 

## Methods

### NewDetailedNotificationStateSlack

`func NewDetailedNotificationStateSlack() *DetailedNotificationStateSlack`

NewDetailedNotificationStateSlack instantiates a new DetailedNotificationStateSlack object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDetailedNotificationStateSlackWithDefaults

`func NewDetailedNotificationStateSlackWithDefaults() *DetailedNotificationStateSlack`

NewDetailedNotificationStateSlackWithDefaults instantiates a new DetailedNotificationStateSlack object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *DetailedNotificationStateSlack) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *DetailedNotificationStateSlack) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *DetailedNotificationStateSlack) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *DetailedNotificationStateSlack) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *DetailedNotificationStateSlack) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *DetailedNotificationStateSlack) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetSettings

`func (o *DetailedNotificationStateSlack) GetSettings() SlackDetailsSettings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *DetailedNotificationStateSlack) GetSettingsOk() (*SlackDetailsSettings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *DetailedNotificationStateSlack) SetSettings(v SlackDetailsSettings)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *DetailedNotificationStateSlack) HasSettings() bool`

HasSettings returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


