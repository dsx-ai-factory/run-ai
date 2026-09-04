# PendingSchedulingMessage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePool** | **string** |  | 
**PhaseReason** | [**PhaseReason**](PhaseReason.md) |  | 
**Reason** | **string** |  | 
**OrgType** | Pointer to [**NullableOrgType**](OrgType.md) |  | [optional] 
**UserMessage** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewPendingSchedulingMessage

`func NewPendingSchedulingMessage(nodePool string, phaseReason PhaseReason, reason string, ) *PendingSchedulingMessage`

NewPendingSchedulingMessage instantiates a new PendingSchedulingMessage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPendingSchedulingMessageWithDefaults

`func NewPendingSchedulingMessageWithDefaults() *PendingSchedulingMessage`

NewPendingSchedulingMessageWithDefaults instantiates a new PendingSchedulingMessage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePool

`func (o *PendingSchedulingMessage) GetNodePool() string`

GetNodePool returns the NodePool field if non-nil, zero value otherwise.

### GetNodePoolOk

`func (o *PendingSchedulingMessage) GetNodePoolOk() (*string, bool)`

GetNodePoolOk returns a tuple with the NodePool field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePool

`func (o *PendingSchedulingMessage) SetNodePool(v string)`

SetNodePool sets NodePool field to given value.


### GetPhaseReason

`func (o *PendingSchedulingMessage) GetPhaseReason() PhaseReason`

GetPhaseReason returns the PhaseReason field if non-nil, zero value otherwise.

### GetPhaseReasonOk

`func (o *PendingSchedulingMessage) GetPhaseReasonOk() (*PhaseReason, bool)`

GetPhaseReasonOk returns a tuple with the PhaseReason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhaseReason

`func (o *PendingSchedulingMessage) SetPhaseReason(v PhaseReason)`

SetPhaseReason sets PhaseReason field to given value.


### GetReason

`func (o *PendingSchedulingMessage) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *PendingSchedulingMessage) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *PendingSchedulingMessage) SetReason(v string)`

SetReason sets Reason field to given value.


### GetOrgType

`func (o *PendingSchedulingMessage) GetOrgType() OrgType`

GetOrgType returns the OrgType field if non-nil, zero value otherwise.

### GetOrgTypeOk

`func (o *PendingSchedulingMessage) GetOrgTypeOk() (*OrgType, bool)`

GetOrgTypeOk returns a tuple with the OrgType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgType

`func (o *PendingSchedulingMessage) SetOrgType(v OrgType)`

SetOrgType sets OrgType field to given value.

### HasOrgType

`func (o *PendingSchedulingMessage) HasOrgType() bool`

HasOrgType returns a boolean if a field has been set.

### SetOrgTypeNil

`func (o *PendingSchedulingMessage) SetOrgTypeNil(b bool)`

 SetOrgTypeNil sets the value for OrgType to be an explicit nil

### UnsetOrgType
`func (o *PendingSchedulingMessage) UnsetOrgType()`

UnsetOrgType ensures that no value is present for OrgType, not even an explicit nil
### GetUserMessage

`func (o *PendingSchedulingMessage) GetUserMessage() string`

GetUserMessage returns the UserMessage field if non-nil, zero value otherwise.

### GetUserMessageOk

`func (o *PendingSchedulingMessage) GetUserMessageOk() (*string, bool)`

GetUserMessageOk returns a tuple with the UserMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserMessage

`func (o *PendingSchedulingMessage) SetUserMessage(v string)`

SetUserMessage sets UserMessage field to given value.

### HasUserMessage

`func (o *PendingSchedulingMessage) HasUserMessage() bool`

HasUserMessage returns a boolean if a field has been set.

### SetUserMessageNil

`func (o *PendingSchedulingMessage) SetUserMessageNil(b bool)`

 SetUserMessageNil sets the value for UserMessage to be an explicit nil

### UnsetUserMessage
`func (o *PendingSchedulingMessage) UnsetUserMessage()`

UnsetUserMessage ensures that no value is present for UserMessage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


