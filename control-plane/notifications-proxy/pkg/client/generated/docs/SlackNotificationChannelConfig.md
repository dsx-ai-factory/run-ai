# SlackNotificationChannelConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Recipients** | Pointer to **[]string** |  | [optional] [default to []]
**DirectNotifications** | Pointer to **NullableBool** |  | [optional] 
**ApiToken** | Pointer to **string** |  | [optional] 

## Methods

### NewSlackNotificationChannelConfig

`func NewSlackNotificationChannelConfig() *SlackNotificationChannelConfig`

NewSlackNotificationChannelConfig instantiates a new SlackNotificationChannelConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSlackNotificationChannelConfigWithDefaults

`func NewSlackNotificationChannelConfigWithDefaults() *SlackNotificationChannelConfig`

NewSlackNotificationChannelConfigWithDefaults instantiates a new SlackNotificationChannelConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRecipients

`func (o *SlackNotificationChannelConfig) GetRecipients() []string`

GetRecipients returns the Recipients field if non-nil, zero value otherwise.

### GetRecipientsOk

`func (o *SlackNotificationChannelConfig) GetRecipientsOk() (*[]string, bool)`

GetRecipientsOk returns a tuple with the Recipients field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecipients

`func (o *SlackNotificationChannelConfig) SetRecipients(v []string)`

SetRecipients sets Recipients field to given value.

### HasRecipients

`func (o *SlackNotificationChannelConfig) HasRecipients() bool`

HasRecipients returns a boolean if a field has been set.

### GetDirectNotifications

`func (o *SlackNotificationChannelConfig) GetDirectNotifications() bool`

GetDirectNotifications returns the DirectNotifications field if non-nil, zero value otherwise.

### GetDirectNotificationsOk

`func (o *SlackNotificationChannelConfig) GetDirectNotificationsOk() (*bool, bool)`

GetDirectNotificationsOk returns a tuple with the DirectNotifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDirectNotifications

`func (o *SlackNotificationChannelConfig) SetDirectNotifications(v bool)`

SetDirectNotifications sets DirectNotifications field to given value.

### HasDirectNotifications

`func (o *SlackNotificationChannelConfig) HasDirectNotifications() bool`

HasDirectNotifications returns a boolean if a field has been set.

### SetDirectNotificationsNil

`func (o *SlackNotificationChannelConfig) SetDirectNotificationsNil(b bool)`

 SetDirectNotificationsNil sets the value for DirectNotifications to be an explicit nil

### UnsetDirectNotifications
`func (o *SlackNotificationChannelConfig) UnsetDirectNotifications()`

UnsetDirectNotifications ensures that no value is present for DirectNotifications, not even an explicit nil
### GetApiToken

`func (o *SlackNotificationChannelConfig) GetApiToken() string`

GetApiToken returns the ApiToken field if non-nil, zero value otherwise.

### GetApiTokenOk

`func (o *SlackNotificationChannelConfig) GetApiTokenOk() (*string, bool)`

GetApiTokenOk returns a tuple with the ApiToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiToken

`func (o *SlackNotificationChannelConfig) SetApiToken(v string)`

SetApiToken sets ApiToken field to given value.

### HasApiToken

`func (o *SlackNotificationChannelConfig) HasApiToken() bool`

HasApiToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


