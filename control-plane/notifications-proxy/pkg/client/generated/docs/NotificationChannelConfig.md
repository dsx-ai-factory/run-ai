# NotificationChannelConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Username** | Pointer to **string** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**SmtpHost** | Pointer to **string** |  | [optional] 
**From** | Pointer to **string** |  | [optional] 
**SmtpPort** | Pointer to **int32** |  | [optional] [default to 587]
**AuthType** | Pointer to [**EmailAuthType**](EmailAuthType.md) |  | [optional] [default to EMAILAUTHTYPE_AUTH_LOGIN]
**SmtpTlsEnabled** | Pointer to **NullableBool** |  | [optional] 
**Recipients** | Pointer to **[]string** |  | [optional] [default to []]
**FromDisplayName** | Pointer to **string** |  | [optional] [default to "Run:ai"]
**DirectNotifications** | Pointer to **NullableBool** |  | [optional] 
**UseAttachments** | Pointer to **NullableBool** |  | [optional] 
**Subject** | Pointer to **string** |  | [optional] [default to "Run:ai notification - {{ .category }} {{ .reason }}"]
**ApiToken** | Pointer to **string** |  | [optional] 

## Methods

### NewNotificationChannelConfig

`func NewNotificationChannelConfig() *NotificationChannelConfig`

NewNotificationChannelConfig instantiates a new NotificationChannelConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNotificationChannelConfigWithDefaults

`func NewNotificationChannelConfigWithDefaults() *NotificationChannelConfig`

NewNotificationChannelConfigWithDefaults instantiates a new NotificationChannelConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsername

`func (o *NotificationChannelConfig) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *NotificationChannelConfig) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *NotificationChannelConfig) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *NotificationChannelConfig) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetPassword

`func (o *NotificationChannelConfig) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *NotificationChannelConfig) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *NotificationChannelConfig) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *NotificationChannelConfig) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetSmtpHost

`func (o *NotificationChannelConfig) GetSmtpHost() string`

GetSmtpHost returns the SmtpHost field if non-nil, zero value otherwise.

### GetSmtpHostOk

`func (o *NotificationChannelConfig) GetSmtpHostOk() (*string, bool)`

GetSmtpHostOk returns a tuple with the SmtpHost field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpHost

`func (o *NotificationChannelConfig) SetSmtpHost(v string)`

SetSmtpHost sets SmtpHost field to given value.

### HasSmtpHost

`func (o *NotificationChannelConfig) HasSmtpHost() bool`

HasSmtpHost returns a boolean if a field has been set.

### GetFrom

`func (o *NotificationChannelConfig) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *NotificationChannelConfig) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *NotificationChannelConfig) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *NotificationChannelConfig) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetSmtpPort

`func (o *NotificationChannelConfig) GetSmtpPort() int32`

GetSmtpPort returns the SmtpPort field if non-nil, zero value otherwise.

### GetSmtpPortOk

`func (o *NotificationChannelConfig) GetSmtpPortOk() (*int32, bool)`

GetSmtpPortOk returns a tuple with the SmtpPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpPort

`func (o *NotificationChannelConfig) SetSmtpPort(v int32)`

SetSmtpPort sets SmtpPort field to given value.

### HasSmtpPort

`func (o *NotificationChannelConfig) HasSmtpPort() bool`

HasSmtpPort returns a boolean if a field has been set.

### GetAuthType

`func (o *NotificationChannelConfig) GetAuthType() EmailAuthType`

GetAuthType returns the AuthType field if non-nil, zero value otherwise.

### GetAuthTypeOk

`func (o *NotificationChannelConfig) GetAuthTypeOk() (*EmailAuthType, bool)`

GetAuthTypeOk returns a tuple with the AuthType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthType

`func (o *NotificationChannelConfig) SetAuthType(v EmailAuthType)`

SetAuthType sets AuthType field to given value.

### HasAuthType

`func (o *NotificationChannelConfig) HasAuthType() bool`

HasAuthType returns a boolean if a field has been set.

### GetSmtpTlsEnabled

`func (o *NotificationChannelConfig) GetSmtpTlsEnabled() bool`

GetSmtpTlsEnabled returns the SmtpTlsEnabled field if non-nil, zero value otherwise.

### GetSmtpTlsEnabledOk

`func (o *NotificationChannelConfig) GetSmtpTlsEnabledOk() (*bool, bool)`

GetSmtpTlsEnabledOk returns a tuple with the SmtpTlsEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpTlsEnabled

`func (o *NotificationChannelConfig) SetSmtpTlsEnabled(v bool)`

SetSmtpTlsEnabled sets SmtpTlsEnabled field to given value.

### HasSmtpTlsEnabled

`func (o *NotificationChannelConfig) HasSmtpTlsEnabled() bool`

HasSmtpTlsEnabled returns a boolean if a field has been set.

### SetSmtpTlsEnabledNil

`func (o *NotificationChannelConfig) SetSmtpTlsEnabledNil(b bool)`

 SetSmtpTlsEnabledNil sets the value for SmtpTlsEnabled to be an explicit nil

### UnsetSmtpTlsEnabled
`func (o *NotificationChannelConfig) UnsetSmtpTlsEnabled()`

UnsetSmtpTlsEnabled ensures that no value is present for SmtpTlsEnabled, not even an explicit nil
### GetRecipients

`func (o *NotificationChannelConfig) GetRecipients() []string`

GetRecipients returns the Recipients field if non-nil, zero value otherwise.

### GetRecipientsOk

`func (o *NotificationChannelConfig) GetRecipientsOk() (*[]string, bool)`

GetRecipientsOk returns a tuple with the Recipients field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecipients

`func (o *NotificationChannelConfig) SetRecipients(v []string)`

SetRecipients sets Recipients field to given value.

### HasRecipients

`func (o *NotificationChannelConfig) HasRecipients() bool`

HasRecipients returns a boolean if a field has been set.

### GetFromDisplayName

`func (o *NotificationChannelConfig) GetFromDisplayName() string`

GetFromDisplayName returns the FromDisplayName field if non-nil, zero value otherwise.

### GetFromDisplayNameOk

`func (o *NotificationChannelConfig) GetFromDisplayNameOk() (*string, bool)`

GetFromDisplayNameOk returns a tuple with the FromDisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFromDisplayName

`func (o *NotificationChannelConfig) SetFromDisplayName(v string)`

SetFromDisplayName sets FromDisplayName field to given value.

### HasFromDisplayName

`func (o *NotificationChannelConfig) HasFromDisplayName() bool`

HasFromDisplayName returns a boolean if a field has been set.

### GetDirectNotifications

`func (o *NotificationChannelConfig) GetDirectNotifications() bool`

GetDirectNotifications returns the DirectNotifications field if non-nil, zero value otherwise.

### GetDirectNotificationsOk

`func (o *NotificationChannelConfig) GetDirectNotificationsOk() (*bool, bool)`

GetDirectNotificationsOk returns a tuple with the DirectNotifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDirectNotifications

`func (o *NotificationChannelConfig) SetDirectNotifications(v bool)`

SetDirectNotifications sets DirectNotifications field to given value.

### HasDirectNotifications

`func (o *NotificationChannelConfig) HasDirectNotifications() bool`

HasDirectNotifications returns a boolean if a field has been set.

### SetDirectNotificationsNil

`func (o *NotificationChannelConfig) SetDirectNotificationsNil(b bool)`

 SetDirectNotificationsNil sets the value for DirectNotifications to be an explicit nil

### UnsetDirectNotifications
`func (o *NotificationChannelConfig) UnsetDirectNotifications()`

UnsetDirectNotifications ensures that no value is present for DirectNotifications, not even an explicit nil
### GetUseAttachments

`func (o *NotificationChannelConfig) GetUseAttachments() bool`

GetUseAttachments returns the UseAttachments field if non-nil, zero value otherwise.

### GetUseAttachmentsOk

`func (o *NotificationChannelConfig) GetUseAttachmentsOk() (*bool, bool)`

GetUseAttachmentsOk returns a tuple with the UseAttachments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseAttachments

`func (o *NotificationChannelConfig) SetUseAttachments(v bool)`

SetUseAttachments sets UseAttachments field to given value.

### HasUseAttachments

`func (o *NotificationChannelConfig) HasUseAttachments() bool`

HasUseAttachments returns a boolean if a field has been set.

### SetUseAttachmentsNil

`func (o *NotificationChannelConfig) SetUseAttachmentsNil(b bool)`

 SetUseAttachmentsNil sets the value for UseAttachments to be an explicit nil

### UnsetUseAttachments
`func (o *NotificationChannelConfig) UnsetUseAttachments()`

UnsetUseAttachments ensures that no value is present for UseAttachments, not even an explicit nil
### GetSubject

`func (o *NotificationChannelConfig) GetSubject() string`

GetSubject returns the Subject field if non-nil, zero value otherwise.

### GetSubjectOk

`func (o *NotificationChannelConfig) GetSubjectOk() (*string, bool)`

GetSubjectOk returns a tuple with the Subject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubject

`func (o *NotificationChannelConfig) SetSubject(v string)`

SetSubject sets Subject field to given value.

### HasSubject

`func (o *NotificationChannelConfig) HasSubject() bool`

HasSubject returns a boolean if a field has been set.

### GetApiToken

`func (o *NotificationChannelConfig) GetApiToken() string`

GetApiToken returns the ApiToken field if non-nil, zero value otherwise.

### GetApiTokenOk

`func (o *NotificationChannelConfig) GetApiTokenOk() (*string, bool)`

GetApiTokenOk returns a tuple with the ApiToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiToken

`func (o *NotificationChannelConfig) SetApiToken(v string)`

SetApiToken sets ApiToken field to given value.

### HasApiToken

`func (o *NotificationChannelConfig) HasApiToken() bool`

HasApiToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


