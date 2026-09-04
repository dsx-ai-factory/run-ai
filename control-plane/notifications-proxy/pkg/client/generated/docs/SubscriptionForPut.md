# SubscriptionForPut

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Events** | [**[]SubscriptionEvent**](SubscriptionEvent.md) |  | 
**Channels** | **[]string** |  | 

## Methods

### NewSubscriptionForPut

`func NewSubscriptionForPut(events []SubscriptionEvent, channels []string, ) *SubscriptionForPut`

NewSubscriptionForPut instantiates a new SubscriptionForPut object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubscriptionForPutWithDefaults

`func NewSubscriptionForPutWithDefaults() *SubscriptionForPut`

NewSubscriptionForPutWithDefaults instantiates a new SubscriptionForPut object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEvents

`func (o *SubscriptionForPut) GetEvents() []SubscriptionEvent`

GetEvents returns the Events field if non-nil, zero value otherwise.

### GetEventsOk

`func (o *SubscriptionForPut) GetEventsOk() (*[]SubscriptionEvent, bool)`

GetEventsOk returns a tuple with the Events field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEvents

`func (o *SubscriptionForPut) SetEvents(v []SubscriptionEvent)`

SetEvents sets Events field to given value.


### GetChannels

`func (o *SubscriptionForPut) GetChannels() []string`

GetChannels returns the Channels field if non-nil, zero value otherwise.

### GetChannelsOk

`func (o *SubscriptionForPut) GetChannelsOk() (*[]string, bool)`

GetChannelsOk returns a tuple with the Channels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannels

`func (o *SubscriptionForPut) SetChannels(v []string)`

SetChannels sets Channels field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


