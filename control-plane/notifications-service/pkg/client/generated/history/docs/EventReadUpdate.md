# EventReadUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Read** | Pointer to **NullableBool** | true to mark the event as read, false to mark it as unread | [optional] 

## Methods

### NewEventReadUpdate

`func NewEventReadUpdate() *EventReadUpdate`

NewEventReadUpdate instantiates a new EventReadUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEventReadUpdateWithDefaults

`func NewEventReadUpdateWithDefaults() *EventReadUpdate`

NewEventReadUpdateWithDefaults instantiates a new EventReadUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRead

`func (o *EventReadUpdate) GetRead() bool`

GetRead returns the Read field if non-nil, zero value otherwise.

### GetReadOk

`func (o *EventReadUpdate) GetReadOk() (*bool, bool)`

GetReadOk returns a tuple with the Read field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRead

`func (o *EventReadUpdate) SetRead(v bool)`

SetRead sets Read field to given value.

### HasRead

`func (o *EventReadUpdate) HasRead() bool`

HasRead returns a boolean if a field has been set.

### SetReadNil

`func (o *EventReadUpdate) SetReadNil(b bool)`

 SetReadNil sets the value for Read to be an explicit nil

### UnsetRead
`func (o *EventReadUpdate) UnsetRead()`

UnsetRead ensures that no value is present for Read, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


