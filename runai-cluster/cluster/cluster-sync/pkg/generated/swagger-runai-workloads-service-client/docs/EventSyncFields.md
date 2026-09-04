# EventSyncFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**ClusterId** | **string** |  | 
**Type** | **string** |  | 
**Message** | **string** |  | 
**Reason** | **string** |  | 
**CreatedAt** | **time.Time** |  | 
**Source** | **string** |  | 
**InvolvedObject** | [**InvolvedObject**](InvolvedObject.md) |  | 

## Methods

### NewEventSyncFields

`func NewEventSyncFields(id string, clusterId string, type_ string, message string, reason string, createdAt time.Time, source string, involvedObject InvolvedObject, ) *EventSyncFields`

NewEventSyncFields instantiates a new EventSyncFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEventSyncFieldsWithDefaults

`func NewEventSyncFieldsWithDefaults() *EventSyncFields`

NewEventSyncFieldsWithDefaults instantiates a new EventSyncFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *EventSyncFields) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *EventSyncFields) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *EventSyncFields) SetId(v string)`

SetId sets Id field to given value.


### GetClusterId

`func (o *EventSyncFields) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *EventSyncFields) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *EventSyncFields) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetType

`func (o *EventSyncFields) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *EventSyncFields) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *EventSyncFields) SetType(v string)`

SetType sets Type field to given value.


### GetMessage

`func (o *EventSyncFields) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *EventSyncFields) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *EventSyncFields) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetReason

`func (o *EventSyncFields) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *EventSyncFields) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *EventSyncFields) SetReason(v string)`

SetReason sets Reason field to given value.


### GetCreatedAt

`func (o *EventSyncFields) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *EventSyncFields) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *EventSyncFields) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetSource

`func (o *EventSyncFields) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *EventSyncFields) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *EventSyncFields) SetSource(v string)`

SetSource sets Source field to given value.


### GetInvolvedObject

`func (o *EventSyncFields) GetInvolvedObject() InvolvedObject`

GetInvolvedObject returns the InvolvedObject field if non-nil, zero value otherwise.

### GetInvolvedObjectOk

`func (o *EventSyncFields) GetInvolvedObjectOk() (*InvolvedObject, bool)`

GetInvolvedObjectOk returns a tuple with the InvolvedObject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObject

`func (o *EventSyncFields) SetInvolvedObject(v InvolvedObject)`

SetInvolvedObject sets InvolvedObject field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


