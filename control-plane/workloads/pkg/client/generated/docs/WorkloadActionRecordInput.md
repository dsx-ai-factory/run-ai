# WorkloadActionRecordInput

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | A unique ID of the workload action-history record. | 
**ClusterId** | **string** | The id of the cluster. | 
**WorkloadId** | **string** | A unique ID of the workload. | 
**WorkloadName** | **string** | The name of the workload. | 
**Action** | [**WorkloadActionType**](WorkloadActionType.md) |  | 
**AcceptedAt** | **time.Time** | The time when the workload action was accepted. | 

## Methods

### NewWorkloadActionRecordInput

`func NewWorkloadActionRecordInput(id string, clusterId string, workloadId string, workloadName string, action WorkloadActionType, acceptedAt time.Time, ) *WorkloadActionRecordInput`

NewWorkloadActionRecordInput instantiates a new WorkloadActionRecordInput object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadActionRecordInputWithDefaults

`func NewWorkloadActionRecordInputWithDefaults() *WorkloadActionRecordInput`

NewWorkloadActionRecordInputWithDefaults instantiates a new WorkloadActionRecordInput object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadActionRecordInput) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadActionRecordInput) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadActionRecordInput) SetId(v string)`

SetId sets Id field to given value.


### GetClusterId

`func (o *WorkloadActionRecordInput) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadActionRecordInput) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadActionRecordInput) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetWorkloadId

`func (o *WorkloadActionRecordInput) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *WorkloadActionRecordInput) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *WorkloadActionRecordInput) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetWorkloadName

`func (o *WorkloadActionRecordInput) GetWorkloadName() string`

GetWorkloadName returns the WorkloadName field if non-nil, zero value otherwise.

### GetWorkloadNameOk

`func (o *WorkloadActionRecordInput) GetWorkloadNameOk() (*string, bool)`

GetWorkloadNameOk returns a tuple with the WorkloadName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadName

`func (o *WorkloadActionRecordInput) SetWorkloadName(v string)`

SetWorkloadName sets WorkloadName field to given value.


### GetAction

`func (o *WorkloadActionRecordInput) GetAction() WorkloadActionType`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *WorkloadActionRecordInput) GetActionOk() (*WorkloadActionType, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *WorkloadActionRecordInput) SetAction(v WorkloadActionType)`

SetAction sets Action field to given value.


### GetAcceptedAt

`func (o *WorkloadActionRecordInput) GetAcceptedAt() time.Time`

GetAcceptedAt returns the AcceptedAt field if non-nil, zero value otherwise.

### GetAcceptedAtOk

`func (o *WorkloadActionRecordInput) GetAcceptedAtOk() (*time.Time, bool)`

GetAcceptedAtOk returns a tuple with the AcceptedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAcceptedAt

`func (o *WorkloadActionRecordInput) SetAcceptedAt(v time.Time)`

SetAcceptedAt sets AcceptedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


