# NotificationChannelForPatch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to [**NotificationChannelType**](NotificationChannelType.md) |  | [optional] [default to NOTIFICATIONCHANNELTYPE_EMAIL]
**Config** | Pointer to [**NotificationChannelConfig**](NotificationChannelConfig.md) |  | [optional] 

## Methods

### NewNotificationChannelForPatch

`func NewNotificationChannelForPatch() *NotificationChannelForPatch`

NewNotificationChannelForPatch instantiates a new NotificationChannelForPatch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNotificationChannelForPatchWithDefaults

`func NewNotificationChannelForPatchWithDefaults() *NotificationChannelForPatch`

NewNotificationChannelForPatchWithDefaults instantiates a new NotificationChannelForPatch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *NotificationChannelForPatch) GetType() NotificationChannelType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *NotificationChannelForPatch) GetTypeOk() (*NotificationChannelType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *NotificationChannelForPatch) SetType(v NotificationChannelType)`

SetType sets Type field to given value.

### HasType

`func (o *NotificationChannelForPatch) HasType() bool`

HasType returns a boolean if a field has been set.

### GetConfig

`func (o *NotificationChannelForPatch) GetConfig() NotificationChannelConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *NotificationChannelForPatch) GetConfigOk() (*NotificationChannelConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *NotificationChannelForPatch) SetConfig(v NotificationChannelConfig)`

SetConfig sets Config field to given value.

### HasConfig

`func (o *NotificationChannelForPatch) HasConfig() bool`

HasConfig returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


