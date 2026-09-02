# SlackNotificationChannelPatchRequestConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Recipients** | Pointer to **[]string** |  | [optional] [default to []]
**DirectNotifications** | Pointer to **NullableBool** |  | [optional] 

## Methods

### NewSlackNotificationChannelPatchRequestConfig

`func NewSlackNotificationChannelPatchRequestConfig() *SlackNotificationChannelPatchRequestConfig`

NewSlackNotificationChannelPatchRequestConfig instantiates a new SlackNotificationChannelPatchRequestConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSlackNotificationChannelPatchRequestConfigWithDefaults

`func NewSlackNotificationChannelPatchRequestConfigWithDefaults() *SlackNotificationChannelPatchRequestConfig`

NewSlackNotificationChannelPatchRequestConfigWithDefaults instantiates a new SlackNotificationChannelPatchRequestConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRecipients

`func (o *SlackNotificationChannelPatchRequestConfig) GetRecipients() []string`

GetRecipients returns the Recipients field if non-nil, zero value otherwise.

### GetRecipientsOk

`func (o *SlackNotificationChannelPatchRequestConfig) GetRecipientsOk() (*[]string, bool)`

GetRecipientsOk returns a tuple with the Recipients field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecipients

`func (o *SlackNotificationChannelPatchRequestConfig) SetRecipients(v []string)`

SetRecipients sets Recipients field to given value.

### HasRecipients

`func (o *SlackNotificationChannelPatchRequestConfig) HasRecipients() bool`

HasRecipients returns a boolean if a field has been set.

### GetDirectNotifications

`func (o *SlackNotificationChannelPatchRequestConfig) GetDirectNotifications() bool`

GetDirectNotifications returns the DirectNotifications field if non-nil, zero value otherwise.

### GetDirectNotificationsOk

`func (o *SlackNotificationChannelPatchRequestConfig) GetDirectNotificationsOk() (*bool, bool)`

GetDirectNotificationsOk returns a tuple with the DirectNotifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDirectNotifications

`func (o *SlackNotificationChannelPatchRequestConfig) SetDirectNotifications(v bool)`

SetDirectNotifications sets DirectNotifications field to given value.

### HasDirectNotifications

`func (o *SlackNotificationChannelPatchRequestConfig) HasDirectNotifications() bool`

HasDirectNotifications returns a boolean if a field has been set.

### SetDirectNotificationsNil

`func (o *SlackNotificationChannelPatchRequestConfig) SetDirectNotificationsNil(b bool)`

 SetDirectNotificationsNil sets the value for DirectNotifications to be an explicit nil

### UnsetDirectNotifications
`func (o *SlackNotificationChannelPatchRequestConfig) UnsetDirectNotifications()`

UnsetDirectNotifications ensures that no value is present for DirectNotifications, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


