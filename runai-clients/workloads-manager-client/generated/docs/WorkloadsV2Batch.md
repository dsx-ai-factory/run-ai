# WorkloadsV2Batch

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ids** | **[]string** | List of ids to perform the action on. | 
**Action** | [**WorkloadsV2BatchAction**](WorkloadsV2BatchAction.md) |  | 

## Methods

### NewWorkloadsV2Batch

`func NewWorkloadsV2Batch(ids []string, action WorkloadsV2BatchAction, ) *WorkloadsV2Batch`

NewWorkloadsV2Batch instantiates a new WorkloadsV2Batch object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadsV2BatchWithDefaults

`func NewWorkloadsV2BatchWithDefaults() *WorkloadsV2Batch`

NewWorkloadsV2BatchWithDefaults instantiates a new WorkloadsV2Batch object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIds

`func (o *WorkloadsV2Batch) GetIds() []string`

GetIds returns the Ids field if non-nil, zero value otherwise.

### GetIdsOk

`func (o *WorkloadsV2Batch) GetIdsOk() (*[]string, bool)`

GetIdsOk returns a tuple with the Ids field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIds

`func (o *WorkloadsV2Batch) SetIds(v []string)`

SetIds sets Ids field to given value.


### GetAction

`func (o *WorkloadsV2Batch) GetAction() WorkloadsV2BatchAction`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *WorkloadsV2Batch) GetActionOk() (*WorkloadsV2BatchAction, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *WorkloadsV2Batch) SetAction(v WorkloadsV2BatchAction)`

SetAction sets Action field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


