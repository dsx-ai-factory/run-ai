# EmailNotificationChannelConfig

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

## Methods

### NewEmailNotificationChannelConfig

`func NewEmailNotificationChannelConfig() *EmailNotificationChannelConfig`

NewEmailNotificationChannelConfig instantiates a new EmailNotificationChannelConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEmailNotificationChannelConfigWithDefaults

`func NewEmailNotificationChannelConfigWithDefaults() *EmailNotificationChannelConfig`

NewEmailNotificationChannelConfigWithDefaults instantiates a new EmailNotificationChannelConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsername

`func (o *EmailNotificationChannelConfig) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *EmailNotificationChannelConfig) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *EmailNotificationChannelConfig) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *EmailNotificationChannelConfig) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetPassword

`func (o *EmailNotificationChannelConfig) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *EmailNotificationChannelConfig) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *EmailNotificationChannelConfig) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *EmailNotificationChannelConfig) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetSmtpHost

`func (o *EmailNotificationChannelConfig) GetSmtpHost() string`

GetSmtpHost returns the SmtpHost field if non-nil, zero value otherwise.

### GetSmtpHostOk

`func (o *EmailNotificationChannelConfig) GetSmtpHostOk() (*string, bool)`

GetSmtpHostOk returns a tuple with the SmtpHost field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpHost

`func (o *EmailNotificationChannelConfig) SetSmtpHost(v string)`

SetSmtpHost sets SmtpHost field to given value.

### HasSmtpHost

`func (o *EmailNotificationChannelConfig) HasSmtpHost() bool`

HasSmtpHost returns a boolean if a field has been set.

### GetFrom

`func (o *EmailNotificationChannelConfig) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *EmailNotificationChannelConfig) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *EmailNotificationChannelConfig) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *EmailNotificationChannelConfig) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetSmtpPort

`func (o *EmailNotificationChannelConfig) GetSmtpPort() int32`

GetSmtpPort returns the SmtpPort field if non-nil, zero value otherwise.

### GetSmtpPortOk

`func (o *EmailNotificationChannelConfig) GetSmtpPortOk() (*int32, bool)`

GetSmtpPortOk returns a tuple with the SmtpPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpPort

`func (o *EmailNotificationChannelConfig) SetSmtpPort(v int32)`

SetSmtpPort sets SmtpPort field to given value.

### HasSmtpPort

`func (o *EmailNotificationChannelConfig) HasSmtpPort() bool`

HasSmtpPort returns a boolean if a field has been set.

### GetAuthType

`func (o *EmailNotificationChannelConfig) GetAuthType() EmailAuthType`

GetAuthType returns the AuthType field if non-nil, zero value otherwise.

### GetAuthTypeOk

`func (o *EmailNotificationChannelConfig) GetAuthTypeOk() (*EmailAuthType, bool)`

GetAuthTypeOk returns a tuple with the AuthType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthType

`func (o *EmailNotificationChannelConfig) SetAuthType(v EmailAuthType)`

SetAuthType sets AuthType field to given value.

### HasAuthType

`func (o *EmailNotificationChannelConfig) HasAuthType() bool`

HasAuthType returns a boolean if a field has been set.

### GetSmtpTlsEnabled

`func (o *EmailNotificationChannelConfig) GetSmtpTlsEnabled() bool`

GetSmtpTlsEnabled returns the SmtpTlsEnabled field if non-nil, zero value otherwise.

### GetSmtpTlsEnabledOk

`func (o *EmailNotificationChannelConfig) GetSmtpTlsEnabledOk() (*bool, bool)`

GetSmtpTlsEnabledOk returns a tuple with the SmtpTlsEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpTlsEnabled

`func (o *EmailNotificationChannelConfig) SetSmtpTlsEnabled(v bool)`

SetSmtpTlsEnabled sets SmtpTlsEnabled field to given value.

### HasSmtpTlsEnabled

`func (o *EmailNotificationChannelConfig) HasSmtpTlsEnabled() bool`

HasSmtpTlsEnabled returns a boolean if a field has been set.

### SetSmtpTlsEnabledNil

`func (o *EmailNotificationChannelConfig) SetSmtpTlsEnabledNil(b bool)`

 SetSmtpTlsEnabledNil sets the value for SmtpTlsEnabled to be an explicit nil

### UnsetSmtpTlsEnabled
`func (o *EmailNotificationChannelConfig) UnsetSmtpTlsEnabled()`

UnsetSmtpTlsEnabled ensures that no value is present for SmtpTlsEnabled, not even an explicit nil
### GetRecipients

`func (o *EmailNotificationChannelConfig) GetRecipients() []string`

GetRecipients returns the Recipients field if non-nil, zero value otherwise.

### GetRecipientsOk

`func (o *EmailNotificationChannelConfig) GetRecipientsOk() (*[]string, bool)`

GetRecipientsOk returns a tuple with the Recipients field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecipients

`func (o *EmailNotificationChannelConfig) SetRecipients(v []string)`

SetRecipients sets Recipients field to given value.

### HasRecipients

`func (o *EmailNotificationChannelConfig) HasRecipients() bool`

HasRecipients returns a boolean if a field has been set.

### GetFromDisplayName

`func (o *EmailNotificationChannelConfig) GetFromDisplayName() string`

GetFromDisplayName returns the FromDisplayName field if non-nil, zero value otherwise.

### GetFromDisplayNameOk

`func (o *EmailNotificationChannelConfig) GetFromDisplayNameOk() (*string, bool)`

GetFromDisplayNameOk returns a tuple with the FromDisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFromDisplayName

`func (o *EmailNotificationChannelConfig) SetFromDisplayName(v string)`

SetFromDisplayName sets FromDisplayName field to given value.

### HasFromDisplayName

`func (o *EmailNotificationChannelConfig) HasFromDisplayName() bool`

HasFromDisplayName returns a boolean if a field has been set.

### GetDirectNotifications

`func (o *EmailNotificationChannelConfig) GetDirectNotifications() bool`

GetDirectNotifications returns the DirectNotifications field if non-nil, zero value otherwise.

### GetDirectNotificationsOk

`func (o *EmailNotificationChannelConfig) GetDirectNotificationsOk() (*bool, bool)`

GetDirectNotificationsOk returns a tuple with the DirectNotifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDirectNotifications

`func (o *EmailNotificationChannelConfig) SetDirectNotifications(v bool)`

SetDirectNotifications sets DirectNotifications field to given value.

### HasDirectNotifications

`func (o *EmailNotificationChannelConfig) HasDirectNotifications() bool`

HasDirectNotifications returns a boolean if a field has been set.

### SetDirectNotificationsNil

`func (o *EmailNotificationChannelConfig) SetDirectNotificationsNil(b bool)`

 SetDirectNotificationsNil sets the value for DirectNotifications to be an explicit nil

### UnsetDirectNotifications
`func (o *EmailNotificationChannelConfig) UnsetDirectNotifications()`

UnsetDirectNotifications ensures that no value is present for DirectNotifications, not even an explicit nil
### GetUseAttachments

`func (o *EmailNotificationChannelConfig) GetUseAttachments() bool`

GetUseAttachments returns the UseAttachments field if non-nil, zero value otherwise.

### GetUseAttachmentsOk

`func (o *EmailNotificationChannelConfig) GetUseAttachmentsOk() (*bool, bool)`

GetUseAttachmentsOk returns a tuple with the UseAttachments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseAttachments

`func (o *EmailNotificationChannelConfig) SetUseAttachments(v bool)`

SetUseAttachments sets UseAttachments field to given value.

### HasUseAttachments

`func (o *EmailNotificationChannelConfig) HasUseAttachments() bool`

HasUseAttachments returns a boolean if a field has been set.

### SetUseAttachmentsNil

`func (o *EmailNotificationChannelConfig) SetUseAttachmentsNil(b bool)`

 SetUseAttachmentsNil sets the value for UseAttachments to be an explicit nil

### UnsetUseAttachments
`func (o *EmailNotificationChannelConfig) UnsetUseAttachments()`

UnsetUseAttachments ensures that no value is present for UseAttachments, not even an explicit nil
### GetSubject

`func (o *EmailNotificationChannelConfig) GetSubject() string`

GetSubject returns the Subject field if non-nil, zero value otherwise.

### GetSubjectOk

`func (o *EmailNotificationChannelConfig) GetSubjectOk() (*string, bool)`

GetSubjectOk returns a tuple with the Subject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubject

`func (o *EmailNotificationChannelConfig) SetSubject(v string)`

SetSubject sets Subject field to given value.

### HasSubject

`func (o *EmailNotificationChannelConfig) HasSubject() bool`

HasSubject returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


