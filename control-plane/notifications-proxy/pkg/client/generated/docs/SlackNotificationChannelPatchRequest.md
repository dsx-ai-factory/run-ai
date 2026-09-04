# SlackNotificationChannelPatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] [readonly] 
**Config** | [**SlackNotificationChannelPatchRequestConfig**](SlackNotificationChannelPatchRequestConfig.md) |  | 

## Methods

### NewSlackNotificationChannelPatchRequest

`func NewSlackNotificationChannelPatchRequest(config SlackNotificationChannelPatchRequestConfig, ) *SlackNotificationChannelPatchRequest`

NewSlackNotificationChannelPatchRequest instantiates a new SlackNotificationChannelPatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSlackNotificationChannelPatchRequestWithDefaults

`func NewSlackNotificationChannelPatchRequestWithDefaults() *SlackNotificationChannelPatchRequest`

NewSlackNotificationChannelPatchRequestWithDefaults instantiates a new SlackNotificationChannelPatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *SlackNotificationChannelPatchRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SlackNotificationChannelPatchRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SlackNotificationChannelPatchRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *SlackNotificationChannelPatchRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetConfig

`func (o *SlackNotificationChannelPatchRequest) GetConfig() SlackNotificationChannelPatchRequestConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *SlackNotificationChannelPatchRequest) GetConfigOk() (*SlackNotificationChannelPatchRequestConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *SlackNotificationChannelPatchRequest) SetConfig(v SlackNotificationChannelPatchRequestConfig)`

SetConfig sets Config field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


