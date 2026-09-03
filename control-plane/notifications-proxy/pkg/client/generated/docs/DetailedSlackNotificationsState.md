# DetailedSlackNotificationsState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** |  | [optional] 
**Settings** | Pointer to [**SlackDetailsSettings**](SlackDetailsSettings.md) |  | [optional] 

## Methods

### NewDetailedSlackNotificationsState

`func NewDetailedSlackNotificationsState() *DetailedSlackNotificationsState`

NewDetailedSlackNotificationsState instantiates a new DetailedSlackNotificationsState object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDetailedSlackNotificationsStateWithDefaults

`func NewDetailedSlackNotificationsStateWithDefaults() *DetailedSlackNotificationsState`

NewDetailedSlackNotificationsStateWithDefaults instantiates a new DetailedSlackNotificationsState object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *DetailedSlackNotificationsState) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *DetailedSlackNotificationsState) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *DetailedSlackNotificationsState) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *DetailedSlackNotificationsState) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *DetailedSlackNotificationsState) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *DetailedSlackNotificationsState) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetSettings

`func (o *DetailedSlackNotificationsState) GetSettings() SlackDetailsSettings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *DetailedSlackNotificationsState) GetSettingsOk() (*SlackDetailsSettings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *DetailedSlackNotificationsState) SetSettings(v SlackDetailsSettings)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *DetailedSlackNotificationsState) HasSettings() bool`

HasSettings returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


