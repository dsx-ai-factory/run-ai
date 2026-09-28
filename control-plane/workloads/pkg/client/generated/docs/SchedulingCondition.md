# SchedulingCondition

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | **string** |  | 
**Reason** | [**PhaseReason**](PhaseReason.md) |  | 
**QueueDetails** | Pointer to [**NullableSchedulingConditionQueueDetails**](SchedulingConditionQueueDetails.md) |  | [optional] 

## Methods

### NewSchedulingCondition

`func NewSchedulingCondition(nodePool string, reason PhaseReason, ) *SchedulingCondition`

NewSchedulingCondition instantiates a new SchedulingCondition object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSchedulingConditionWithDefaults

`func NewSchedulingConditionWithDefaults() *SchedulingCondition`

NewSchedulingConditionWithDefaults instantiates a new SchedulingCondition object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *SchedulingCondition) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *SchedulingCondition) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *SchedulingCondition) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.


### GetReason

`func (o *SchedulingCondition) GetReason() PhaseReason`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *SchedulingCondition) GetReasonOk() (*PhaseReason, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *SchedulingCondition) SetReason(v PhaseReason)`

SetReason sets Reason field to given value.


### GetQueueDetails

`func (o *SchedulingCondition) GetQueueDetails() SchedulingConditionQueueDetails`

GetQueueDetails returns the QueueDetails field if non-nil, zero value otherwise.

### GetQueueDetailsOk

`func (o *SchedulingCondition) GetQueueDetailsOk() (*SchedulingConditionQueueDetails, bool)`

GetQueueDetailsOk returns a tuple with the QueueDetails field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueueDetails

`func (o *SchedulingCondition) SetQueueDetails(v SchedulingConditionQueueDetails)`

SetQueueDetails sets QueueDetails field to given value.

### HasQueueDetails

`func (o *SchedulingCondition) HasQueueDetails() bool`

HasQueueDetails returns a boolean if a field has been set.

### SetQueueDetailsNil

`func (o *SchedulingCondition) SetQueueDetailsNil(b bool)`

 SetQueueDetailsNil sets the value for QueueDetails to be an explicit nil

### UnsetQueueDetails
`func (o *SchedulingCondition) UnsetQueueDetails()`

UnsetQueueDetails ensures that no value is present for QueueDetails, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


