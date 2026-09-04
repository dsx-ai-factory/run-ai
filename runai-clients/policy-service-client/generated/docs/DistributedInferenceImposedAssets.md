# DistributedInferenceImposedAssets

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Leader** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 
**Worker** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 

## Methods

### NewDistributedInferenceImposedAssets

`func NewDistributedInferenceImposedAssets() *DistributedInferenceImposedAssets`

NewDistributedInferenceImposedAssets instantiates a new DistributedInferenceImposedAssets object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceImposedAssetsWithDefaults

`func NewDistributedInferenceImposedAssetsWithDefaults() *DistributedInferenceImposedAssets`

NewDistributedInferenceImposedAssetsWithDefaults instantiates a new DistributedInferenceImposedAssets object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLeader

`func (o *DistributedInferenceImposedAssets) GetLeader() []string`

GetLeader returns the Leader field if non-nil, zero value otherwise.

### GetLeaderOk

`func (o *DistributedInferenceImposedAssets) GetLeaderOk() (*[]string, bool)`

GetLeaderOk returns a tuple with the Leader field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLeader

`func (o *DistributedInferenceImposedAssets) SetLeader(v []string)`

SetLeader sets Leader field to given value.

### HasLeader

`func (o *DistributedInferenceImposedAssets) HasLeader() bool`

HasLeader returns a boolean if a field has been set.

### SetLeaderNil

`func (o *DistributedInferenceImposedAssets) SetLeaderNil(b bool)`

 SetLeaderNil sets the value for Leader to be an explicit nil

### UnsetLeader
`func (o *DistributedInferenceImposedAssets) UnsetLeader()`

UnsetLeader ensures that no value is present for Leader, not even an explicit nil
### GetWorker

`func (o *DistributedInferenceImposedAssets) GetWorker() []string`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedInferenceImposedAssets) GetWorkerOk() (*[]string, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedInferenceImposedAssets) SetWorker(v []string)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedInferenceImposedAssets) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedInferenceImposedAssets) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedInferenceImposedAssets) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


