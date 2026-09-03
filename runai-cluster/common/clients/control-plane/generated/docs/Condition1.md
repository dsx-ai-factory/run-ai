# Condition1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | The type of the condition, such as Failed or Available. See Types of domain status conditions. | 
**Status** | **string** | The status of the condition, such as True, False or Unknown. | 
**Message** | Pointer to **string** | An optional, human-readable message providing more details about the condition. | [optional] 
**Reason** | Pointer to **string** | The reason for the Failed condition. Not applicable to other types of condition. | [optional] 
**LastTransitionTime** | Pointer to **NullableTime** | A timestamp of when the condition was created or the last time time the condition transitioned from one status to another. | [optional] 

## Methods

### NewCondition1

`func NewCondition1(type_ string, status string, ) *Condition1`

NewCondition1 instantiates a new Condition1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCondition1WithDefaults

`func NewCondition1WithDefaults() *Condition1`

NewCondition1WithDefaults instantiates a new Condition1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *Condition1) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Condition1) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Condition1) SetType(v string)`

SetType sets Type field to given value.


### GetStatus

`func (o *Condition1) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Condition1) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Condition1) SetStatus(v string)`

SetStatus sets Status field to given value.


### GetMessage

`func (o *Condition1) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *Condition1) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *Condition1) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *Condition1) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetReason

`func (o *Condition1) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *Condition1) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *Condition1) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *Condition1) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetLastTransitionTime

`func (o *Condition1) GetLastTransitionTime() time.Time`

GetLastTransitionTime returns the LastTransitionTime field if non-nil, zero value otherwise.

### GetLastTransitionTimeOk

`func (o *Condition1) GetLastTransitionTimeOk() (*time.Time, bool)`

GetLastTransitionTimeOk returns a tuple with the LastTransitionTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastTransitionTime

`func (o *Condition1) SetLastTransitionTime(v time.Time)`

SetLastTransitionTime sets LastTransitionTime field to given value.

### HasLastTransitionTime

`func (o *Condition1) HasLastTransitionTime() bool`

HasLastTransitionTime returns a boolean if a field has been set.

### SetLastTransitionTimeNil

`func (o *Condition1) SetLastTransitionTimeNil(b bool)`

 SetLastTransitionTimeNil sets the value for LastTransitionTime to be an explicit nil

### UnsetLastTransitionTime
`func (o *Condition1) UnsetLastTransitionTime()`

UnsetLastTransitionTime ensures that no value is present for LastTransitionTime, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


