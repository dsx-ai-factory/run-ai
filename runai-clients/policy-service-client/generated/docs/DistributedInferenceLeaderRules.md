# DistributedInferenceLeaderRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Leader** | Pointer to [**NullableDistributedInferenceLeaderWorkerRulesV2**](DistributedInferenceLeaderWorkerRulesV2.md) | Defines the pod specification for the leader. Must always be provided, regardless of the number of workers. | [optional] 

## Methods

### NewDistributedInferenceLeaderRules

`func NewDistributedInferenceLeaderRules() *DistributedInferenceLeaderRules`

NewDistributedInferenceLeaderRules instantiates a new DistributedInferenceLeaderRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceLeaderRulesWithDefaults

`func NewDistributedInferenceLeaderRulesWithDefaults() *DistributedInferenceLeaderRules`

NewDistributedInferenceLeaderRulesWithDefaults instantiates a new DistributedInferenceLeaderRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLeader

`func (o *DistributedInferenceLeaderRules) GetLeader() DistributedInferenceLeaderWorkerRulesV2`

GetLeader returns the Leader field if non-nil, zero value otherwise.

### GetLeaderOk

`func (o *DistributedInferenceLeaderRules) GetLeaderOk() (*DistributedInferenceLeaderWorkerRulesV2, bool)`

GetLeaderOk returns a tuple with the Leader field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLeader

`func (o *DistributedInferenceLeaderRules) SetLeader(v DistributedInferenceLeaderWorkerRulesV2)`

SetLeader sets Leader field to given value.

### HasLeader

`func (o *DistributedInferenceLeaderRules) HasLeader() bool`

HasLeader returns a boolean if a field has been set.

### SetLeaderNil

`func (o *DistributedInferenceLeaderRules) SetLeaderNil(b bool)`

 SetLeaderNil sets the value for Leader to be an explicit nil

### UnsetLeader
`func (o *DistributedInferenceLeaderRules) UnsetLeader()`

UnsetLeader ensures that no value is present for Leader, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


