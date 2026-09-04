# WorkloadBatch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ids** | **[]string** | List of ids to perform the action on. | 
**Action** | [**Action**](Action.md) |  | 

## Methods

### NewWorkloadBatch

`func NewWorkloadBatch(ids []string, action Action, ) *WorkloadBatch`

NewWorkloadBatch instantiates a new WorkloadBatch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadBatchWithDefaults

`func NewWorkloadBatchWithDefaults() *WorkloadBatch`

NewWorkloadBatchWithDefaults instantiates a new WorkloadBatch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIds

`func (o *WorkloadBatch) GetIds() []string`

GetIds returns the Ids field if non-nil, zero value otherwise.

### GetIdsOk

`func (o *WorkloadBatch) GetIdsOk() (*[]string, bool)`

GetIdsOk returns a tuple with the Ids field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIds

`func (o *WorkloadBatch) SetIds(v []string)`

SetIds sets Ids field to given value.


### GetAction

`func (o *WorkloadBatch) GetAction() Action`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *WorkloadBatch) GetActionOk() (*Action, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *WorkloadBatch) SetAction(v Action)`

SetAction sets Action field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


