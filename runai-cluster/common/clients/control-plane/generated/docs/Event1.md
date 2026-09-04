# Event1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | **time.Time** |  | 
**Id** | Pointer to **string** |  | [optional] 
**Type** | **string** |  | 
**ClusterId** | **string** | The id of the cluster. | 
**Message** | **string** |  | 
**Reason** | Pointer to **string** |  | [optional] 
**Source** | Pointer to **string** |  | [optional] 
**InvolvedObject** | Pointer to [**InvolvedObject**](InvolvedObject.md) |  | [optional] 

## Methods

### NewEvent1

`func NewEvent1(createdAt time.Time, type_ string, clusterId string, message string, ) *Event1`

NewEvent1 instantiates a new Event1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEvent1WithDefaults

`func NewEvent1WithDefaults() *Event1`

NewEvent1WithDefaults instantiates a new Event1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *Event1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Event1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Event1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetId

`func (o *Event1) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Event1) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Event1) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Event1) HasId() bool`

HasId returns a boolean if a field has been set.

### GetType

`func (o *Event1) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Event1) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Event1) SetType(v string)`

SetType sets Type field to given value.


### GetClusterId

`func (o *Event1) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Event1) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Event1) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetMessage

`func (o *Event1) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *Event1) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *Event1) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetReason

`func (o *Event1) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *Event1) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *Event1) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *Event1) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetSource

`func (o *Event1) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *Event1) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *Event1) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *Event1) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetInvolvedObject

`func (o *Event1) GetInvolvedObject() InvolvedObject`

GetInvolvedObject returns the InvolvedObject field if non-nil, zero value otherwise.

### GetInvolvedObjectOk

`func (o *Event1) GetInvolvedObjectOk() (*InvolvedObject, bool)`

GetInvolvedObjectOk returns a tuple with the InvolvedObject field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvolvedObject

`func (o *Event1) SetInvolvedObject(v InvolvedObject)`

SetInvolvedObject sets InvolvedObject field to given value.

### HasInvolvedObject

`func (o *Event1) HasInvolvedObject() bool`

HasInvolvedObject returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


