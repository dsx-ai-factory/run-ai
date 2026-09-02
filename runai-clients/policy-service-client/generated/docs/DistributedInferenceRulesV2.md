# DistributedInferenceRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Category** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibilityRules**](PreemptibilityRules.md) |  | [optional] 
**PriorityClass** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**RestartPolicy** | Pointer to [**NullableDistributedInferenceRestartPolicyRules**](DistributedInferenceRestartPolicyRules.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableDistributedInferenceServingPortRules**](DistributedInferenceServingPortRules.md) |  | [optional] 
**StartupPolicy** | Pointer to [**NullableDistributedInferenceStartupPolicyFieldRulesStartupPolicy**](DistributedInferenceStartupPolicyFieldRulesStartupPolicy.md) |  | [optional] 
**Workers** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Replicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Leader** | Pointer to [**NullableDistributedInferenceLeaderWorkerRulesV2**](DistributedInferenceLeaderWorkerRulesV2.md) | Defines the pod specification for the leader. Must always be provided, regardless of the number of workers. | [optional] 
**Worker** | Pointer to [**NullableDistributedInferenceLeaderWorkerRulesV2**](DistributedInferenceLeaderWorkerRulesV2.md) | Defines the pod specification for the workers. Required only if the number of workers is greater than 0. | [optional] 

## Methods

### NewDistributedInferenceRulesV2

`func NewDistributedInferenceRulesV2() *DistributedInferenceRulesV2`

NewDistributedInferenceRulesV2 instantiates a new DistributedInferenceRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceRulesV2WithDefaults

`func NewDistributedInferenceRulesV2WithDefaults() *DistributedInferenceRulesV2`

NewDistributedInferenceRulesV2WithDefaults instantiates a new DistributedInferenceRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategory

`func (o *DistributedInferenceRulesV2) GetCategory() StringRules`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *DistributedInferenceRulesV2) GetCategoryOk() (*StringRules, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *DistributedInferenceRulesV2) SetCategory(v StringRules)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *DistributedInferenceRulesV2) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *DistributedInferenceRulesV2) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *DistributedInferenceRulesV2) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetNodePools

`func (o *DistributedInferenceRulesV2) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *DistributedInferenceRulesV2) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *DistributedInferenceRulesV2) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *DistributedInferenceRulesV2) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *DistributedInferenceRulesV2) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *DistributedInferenceRulesV2) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *DistributedInferenceRulesV2) GetPreemptibility() PreemptibilityRules`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *DistributedInferenceRulesV2) GetPreemptibilityOk() (*PreemptibilityRules, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *DistributedInferenceRulesV2) SetPreemptibility(v PreemptibilityRules)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *DistributedInferenceRulesV2) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *DistributedInferenceRulesV2) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *DistributedInferenceRulesV2) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *DistributedInferenceRulesV2) GetPriorityClass() StringRules`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *DistributedInferenceRulesV2) GetPriorityClassOk() (*StringRules, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *DistributedInferenceRulesV2) SetPriorityClass(v StringRules)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *DistributedInferenceRulesV2) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *DistributedInferenceRulesV2) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *DistributedInferenceRulesV2) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetRestartPolicy

`func (o *DistributedInferenceRulesV2) GetRestartPolicy() DistributedInferenceRestartPolicyRules`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *DistributedInferenceRulesV2) GetRestartPolicyOk() (*DistributedInferenceRestartPolicyRules, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *DistributedInferenceRulesV2) SetRestartPolicy(v DistributedInferenceRestartPolicyRules)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *DistributedInferenceRulesV2) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *DistributedInferenceRulesV2) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *DistributedInferenceRulesV2) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetServingPort

`func (o *DistributedInferenceRulesV2) GetServingPort() DistributedInferenceServingPortRules`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *DistributedInferenceRulesV2) GetServingPortOk() (*DistributedInferenceServingPortRules, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *DistributedInferenceRulesV2) SetServingPort(v DistributedInferenceServingPortRules)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *DistributedInferenceRulesV2) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *DistributedInferenceRulesV2) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *DistributedInferenceRulesV2) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetStartupPolicy

`func (o *DistributedInferenceRulesV2) GetStartupPolicy() DistributedInferenceStartupPolicyFieldRulesStartupPolicy`

GetStartupPolicy returns the StartupPolicy field if non-nil, zero value otherwise.

### GetStartupPolicyOk

`func (o *DistributedInferenceRulesV2) GetStartupPolicyOk() (*DistributedInferenceStartupPolicyFieldRulesStartupPolicy, bool)`

GetStartupPolicyOk returns a tuple with the StartupPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartupPolicy

`func (o *DistributedInferenceRulesV2) SetStartupPolicy(v DistributedInferenceStartupPolicyFieldRulesStartupPolicy)`

SetStartupPolicy sets StartupPolicy field to given value.

### HasStartupPolicy

`func (o *DistributedInferenceRulesV2) HasStartupPolicy() bool`

HasStartupPolicy returns a boolean if a field has been set.

### SetStartupPolicyNil

`func (o *DistributedInferenceRulesV2) SetStartupPolicyNil(b bool)`

 SetStartupPolicyNil sets the value for StartupPolicy to be an explicit nil

### UnsetStartupPolicy
`func (o *DistributedInferenceRulesV2) UnsetStartupPolicy()`

UnsetStartupPolicy ensures that no value is present for StartupPolicy, not even an explicit nil
### GetWorkers

`func (o *DistributedInferenceRulesV2) GetWorkers() IntegerRules`

GetWorkers returns the Workers field if non-nil, zero value otherwise.

### GetWorkersOk

`func (o *DistributedInferenceRulesV2) GetWorkersOk() (*IntegerRules, bool)`

GetWorkersOk returns a tuple with the Workers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkers

`func (o *DistributedInferenceRulesV2) SetWorkers(v IntegerRules)`

SetWorkers sets Workers field to given value.

### HasWorkers

`func (o *DistributedInferenceRulesV2) HasWorkers() bool`

HasWorkers returns a boolean if a field has been set.

### SetWorkersNil

`func (o *DistributedInferenceRulesV2) SetWorkersNil(b bool)`

 SetWorkersNil sets the value for Workers to be an explicit nil

### UnsetWorkers
`func (o *DistributedInferenceRulesV2) UnsetWorkers()`

UnsetWorkers ensures that no value is present for Workers, not even an explicit nil
### GetReplicas

`func (o *DistributedInferenceRulesV2) GetReplicas() IntegerRules`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *DistributedInferenceRulesV2) GetReplicasOk() (*IntegerRules, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *DistributedInferenceRulesV2) SetReplicas(v IntegerRules)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *DistributedInferenceRulesV2) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### SetReplicasNil

`func (o *DistributedInferenceRulesV2) SetReplicasNil(b bool)`

 SetReplicasNil sets the value for Replicas to be an explicit nil

### UnsetReplicas
`func (o *DistributedInferenceRulesV2) UnsetReplicas()`

UnsetReplicas ensures that no value is present for Replicas, not even an explicit nil
### GetLeader

`func (o *DistributedInferenceRulesV2) GetLeader() DistributedInferenceLeaderWorkerRulesV2`

GetLeader returns the Leader field if non-nil, zero value otherwise.

### GetLeaderOk

`func (o *DistributedInferenceRulesV2) GetLeaderOk() (*DistributedInferenceLeaderWorkerRulesV2, bool)`

GetLeaderOk returns a tuple with the Leader field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLeader

`func (o *DistributedInferenceRulesV2) SetLeader(v DistributedInferenceLeaderWorkerRulesV2)`

SetLeader sets Leader field to given value.

### HasLeader

`func (o *DistributedInferenceRulesV2) HasLeader() bool`

HasLeader returns a boolean if a field has been set.

### SetLeaderNil

`func (o *DistributedInferenceRulesV2) SetLeaderNil(b bool)`

 SetLeaderNil sets the value for Leader to be an explicit nil

### UnsetLeader
`func (o *DistributedInferenceRulesV2) UnsetLeader()`

UnsetLeader ensures that no value is present for Leader, not even an explicit nil
### GetWorker

`func (o *DistributedInferenceRulesV2) GetWorker() DistributedInferenceLeaderWorkerRulesV2`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedInferenceRulesV2) GetWorkerOk() (*DistributedInferenceLeaderWorkerRulesV2, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedInferenceRulesV2) SetWorker(v DistributedInferenceLeaderWorkerRulesV2)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedInferenceRulesV2) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedInferenceRulesV2) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedInferenceRulesV2) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


