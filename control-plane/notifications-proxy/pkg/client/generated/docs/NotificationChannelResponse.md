# NotificationChannelResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | [**NotificationChannelType**](NotificationChannelType.md) |  | [default to NOTIFICATIONCHANNELTYPE_EMAIL]
**Config** | [**NotificationChannelConfig**](NotificationChannelConfig.md) |  | 
**Name** | **string** |  | 

## Methods

### NewNotificationChannelResponse

`func NewNotificationChannelResponse(type_ NotificationChannelType, config NotificationChannelConfig, name string, ) *NotificationChannelResponse`

NewNotificationChannelResponse instantiates a new NotificationChannelResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNotificationChannelResponseWithDefaults

`func NewNotificationChannelResponseWithDefaults() *NotificationChannelResponse`

NewNotificationChannelResponseWithDefaults instantiates a new NotificationChannelResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *NotificationChannelResponse) GetType() NotificationChannelType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *NotificationChannelResponse) GetTypeOk() (*NotificationChannelType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *NotificationChannelResponse) SetType(v NotificationChannelType)`

SetType sets Type field to given value.


### GetConfig

`func (o *NotificationChannelResponse) GetConfig() NotificationChannelConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *NotificationChannelResponse) GetConfigOk() (*NotificationChannelConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *NotificationChannelResponse) SetConfig(v NotificationChannelConfig)`

SetConfig sets Config field to given value.


### GetName

`func (o *NotificationChannelResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NotificationChannelResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NotificationChannelResponse) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


