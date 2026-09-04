# Event

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique ID of the event | 
**SourceListener** | **string** | ID of the listener that received the event | 
**Properties** | **map[string]interface{}** | Event-specific dynamic properties | 
**CreatedAt** | **time.Time** | When the event was created | 
**Read** | **bool** | Whether the event has been read | 

## Methods

### NewEvent

`func NewEvent(id string, sourceListener string, properties map[string]interface{}, createdAt time.Time, read bool, ) *Event`

NewEvent instantiates a new Event object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEventWithDefaults

`func NewEventWithDefaults() *Event`

NewEventWithDefaults instantiates a new Event object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Event) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Event) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Event) SetId(v string)`

SetId sets Id field to given value.


### GetSourceListener

`func (o *Event) GetSourceListener() string`

GetSourceListener returns the SourceListener field if non-nil, zero value otherwise.

### GetSourceListenerOk

`func (o *Event) GetSourceListenerOk() (*string, bool)`

GetSourceListenerOk returns a tuple with the SourceListener field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceListener

`func (o *Event) SetSourceListener(v string)`

SetSourceListener sets SourceListener field to given value.


### GetProperties

`func (o *Event) GetProperties() map[string]interface{}`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *Event) GetPropertiesOk() (*map[string]interface{}, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *Event) SetProperties(v map[string]interface{})`

SetProperties sets Properties field to given value.


### GetCreatedAt

`func (o *Event) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Event) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Event) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetRead

`func (o *Event) GetRead() bool`

GetRead returns the Read field if non-nil, zero value otherwise.

### GetReadOk

`func (o *Event) GetReadOk() (*bool, bool)`

GetReadOk returns a tuple with the Read field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRead

`func (o *Event) SetRead(v bool)`

SetRead sets Read field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


