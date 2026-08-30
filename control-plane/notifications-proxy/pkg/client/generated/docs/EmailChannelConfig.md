# EmailChannelConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Username** | Pointer to **string** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**SmtpHost** | **string** |  | 
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

### NewEmailChannelConfig

`func NewEmailChannelConfig(smtpHost string, ) *EmailChannelConfig`

NewEmailChannelConfig instantiates a new EmailChannelConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEmailChannelConfigWithDefaults

`func NewEmailChannelConfigWithDefaults() *EmailChannelConfig`

NewEmailChannelConfigWithDefaults instantiates a new EmailChannelConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsername

`func (o *EmailChannelConfig) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *EmailChannelConfig) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *EmailChannelConfig) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *EmailChannelConfig) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetPassword

`func (o *EmailChannelConfig) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *EmailChannelConfig) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *EmailChannelConfig) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *EmailChannelConfig) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetSmtpHost

`func (o *EmailChannelConfig) GetSmtpHost() string`

GetSmtpHost returns the SmtpHost field if non-nil, zero value otherwise.

### GetSmtpHostOk

`func (o *EmailChannelConfig) GetSmtpHostOk() (*string, bool)`

GetSmtpHostOk returns a tuple with the SmtpHost field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpHost

`func (o *EmailChannelConfig) SetSmtpHost(v string)`

SetSmtpHost sets SmtpHost field to given value.


### GetFrom

`func (o *EmailChannelConfig) GetFrom() string`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *EmailChannelConfig) GetFromOk() (*string, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *EmailChannelConfig) SetFrom(v string)`

SetFrom sets From field to given value.

### HasFrom

`func (o *EmailChannelConfig) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetSmtpPort

`func (o *EmailChannelConfig) GetSmtpPort() int32`

GetSmtpPort returns the SmtpPort field if non-nil, zero value otherwise.

### GetSmtpPortOk

`func (o *EmailChannelConfig) GetSmtpPortOk() (*int32, bool)`

GetSmtpPortOk returns a tuple with the SmtpPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpPort

`func (o *EmailChannelConfig) SetSmtpPort(v int32)`

SetSmtpPort sets SmtpPort field to given value.

### HasSmtpPort

`func (o *EmailChannelConfig) HasSmtpPort() bool`

HasSmtpPort returns a boolean if a field has been set.

### GetAuthType

`func (o *EmailChannelConfig) GetAuthType() EmailAuthType`

GetAuthType returns the AuthType field if non-nil, zero value otherwise.

### GetAuthTypeOk

`func (o *EmailChannelConfig) GetAuthTypeOk() (*EmailAuthType, bool)`

GetAuthTypeOk returns a tuple with the AuthType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthType

`func (o *EmailChannelConfig) SetAuthType(v EmailAuthType)`

SetAuthType sets AuthType field to given value.

### HasAuthType

`func (o *EmailChannelConfig) HasAuthType() bool`

HasAuthType returns a boolean if a field has been set.

### GetSmtpTlsEnabled

`func (o *EmailChannelConfig) GetSmtpTlsEnabled() bool`

GetSmtpTlsEnabled returns the SmtpTlsEnabled field if non-nil, zero value otherwise.

### GetSmtpTlsEnabledOk

`func (o *EmailChannelConfig) GetSmtpTlsEnabledOk() (*bool, bool)`

GetSmtpTlsEnabledOk returns a tuple with the SmtpTlsEnabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSmtpTlsEnabled

`func (o *EmailChannelConfig) SetSmtpTlsEnabled(v bool)`

SetSmtpTlsEnabled sets SmtpTlsEnabled field to given value.

### HasSmtpTlsEnabled

`func (o *EmailChannelConfig) HasSmtpTlsEnabled() bool`

HasSmtpTlsEnabled returns a boolean if a field has been set.

### SetSmtpTlsEnabledNil

`func (o *EmailChannelConfig) SetSmtpTlsEnabledNil(b bool)`

 SetSmtpTlsEnabledNil sets the value for SmtpTlsEnabled to be an explicit nil

### UnsetSmtpTlsEnabled
`func (o *EmailChannelConfig) UnsetSmtpTlsEnabled()`

UnsetSmtpTlsEnabled ensures that no value is present for SmtpTlsEnabled, not even an explicit nil
### GetRecipients

`func (o *EmailChannelConfig) GetRecipients() []string`

GetRecipients returns the Recipients field if non-nil, zero value otherwise.

### GetRecipientsOk

`func (o *EmailChannelConfig) GetRecipientsOk() (*[]string, bool)`

GetRecipientsOk returns a tuple with the Recipients field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecipients

`func (o *EmailChannelConfig) SetRecipients(v []string)`

SetRecipients sets Recipients field to given value.

### HasRecipients

`func (o *EmailChannelConfig) HasRecipients() bool`

HasRecipients returns a boolean if a field has been set.

### GetFromDisplayName

`func (o *EmailChannelConfig) GetFromDisplayName() string`

GetFromDisplayName returns the FromDisplayName field if non-nil, zero value otherwise.

### GetFromDisplayNameOk

`func (o *EmailChannelConfig) GetFromDisplayNameOk() (*string, bool)`

GetFromDisplayNameOk returns a tuple with the FromDisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFromDisplayName

`func (o *EmailChannelConfig) SetFromDisplayName(v string)`

SetFromDisplayName sets FromDisplayName field to given value.

### HasFromDisplayName

`func (o *EmailChannelConfig) HasFromDisplayName() bool`

HasFromDisplayName returns a boolean if a field has been set.

### GetDirectNotifications

`func (o *EmailChannelConfig) GetDirectNotifications() bool`

GetDirectNotifications returns the DirectNotifications field if non-nil, zero value otherwise.

### GetDirectNotificationsOk

`func (o *EmailChannelConfig) GetDirectNotificationsOk() (*bool, bool)`

GetDirectNotificationsOk returns a tuple with the DirectNotifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDirectNotifications

`func (o *EmailChannelConfig) SetDirectNotifications(v bool)`

SetDirectNotifications sets DirectNotifications field to given value.

### HasDirectNotifications

`func (o *EmailChannelConfig) HasDirectNotifications() bool`

HasDirectNotifications returns a boolean if a field has been set.

### SetDirectNotificationsNil

`func (o *EmailChannelConfig) SetDirectNotificationsNil(b bool)`

 SetDirectNotificationsNil sets the value for DirectNotifications to be an explicit nil

### UnsetDirectNotifications
`func (o *EmailChannelConfig) UnsetDirectNotifications()`

UnsetDirectNotifications ensures that no value is present for DirectNotifications, not even an explicit nil
### GetUseAttachments

`func (o *EmailChannelConfig) GetUseAttachments() bool`

GetUseAttachments returns the UseAttachments field if non-nil, zero value otherwise.

### GetUseAttachmentsOk

`func (o *EmailChannelConfig) GetUseAttachmentsOk() (*bool, bool)`

GetUseAttachmentsOk returns a tuple with the UseAttachments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseAttachments

`func (o *EmailChannelConfig) SetUseAttachments(v bool)`

SetUseAttachments sets UseAttachments field to given value.

### HasUseAttachments

`func (o *EmailChannelConfig) HasUseAttachments() bool`

HasUseAttachments returns a boolean if a field has been set.

### SetUseAttachmentsNil

`func (o *EmailChannelConfig) SetUseAttachmentsNil(b bool)`

 SetUseAttachmentsNil sets the value for UseAttachments to be an explicit nil

### UnsetUseAttachments
`func (o *EmailChannelConfig) UnsetUseAttachments()`

UnsetUseAttachments ensures that no value is present for UseAttachments, not even an explicit nil
### GetSubject

`func (o *EmailChannelConfig) GetSubject() string`

GetSubject returns the Subject field if non-nil, zero value otherwise.

### GetSubjectOk

`func (o *EmailChannelConfig) GetSubjectOk() (*string, bool)`

GetSubjectOk returns a tuple with the Subject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubject

`func (o *EmailChannelConfig) SetSubject(v string)`

SetSubject sets Subject field to given value.

### HasSubject

`func (o *EmailChannelConfig) HasSubject() bool`

HasSubject returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


