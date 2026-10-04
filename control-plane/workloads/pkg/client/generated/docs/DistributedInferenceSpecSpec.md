# DistributedInferenceSpecSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**RestartPolicy** | Pointer to [**NullableDistributedInferenceRestartPolicy**](DistributedInferenceRestartPolicy.md) |  | [optional] [default to DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART]
**ServingPort** | Pointer to [**NullableDistributedInferenceServingPort**](DistributedInferenceServingPort.md) |  | [optional] 
**StartupPolicy** | Pointer to [**NullableDistributedInferenceStartupPolicy**](DistributedInferenceStartupPolicy.md) |  | [optional] [default to DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED]
**Workers** | Pointer to **NullableInt32** | Specifies the number of worker nodes to run. If set to 0, only the leader node will run, and no worker pods will be created. In this case, worker spec is not required. | [optional] [default to 0]
**Replicas** | Pointer to **NullableInt32** | Specifies the number of leader-worker sets to deploy. Each replica represents a group consisting of one leader pod and multiple worker pods.  For example, setting replicas: 3 will create 3 independent groups, each with its own leader and corresponding set of workers.  | [optional] [default to 1]
**Leader** | Pointer to [**NullableDistributedInferenceLeaderWorkerSpec**](DistributedInferenceLeaderWorkerSpec.md) | Defines the pod specification for the leader. Must always be provided, regardless of the number of workers. | [optional] 
**Worker** | Pointer to [**NullableDistributedInferenceLeaderWorkerSpec**](DistributedInferenceLeaderWorkerSpec.md) | Defines the pod specification for the workers. Required only if the number of workers is greater than 0. | [optional] 

## Methods

### NewDistributedInferenceSpecSpec

`func NewDistributedInferenceSpecSpec() *DistributedInferenceSpecSpec`

NewDistributedInferenceSpecSpec instantiates a new DistributedInferenceSpecSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceSpecSpecWithDefaults

`func NewDistributedInferenceSpecSpecWithDefaults() *DistributedInferenceSpecSpec`

NewDistributedInferenceSpecSpecWithDefaults instantiates a new DistributedInferenceSpecSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategory

`func (o *DistributedInferenceSpecSpec) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *DistributedInferenceSpecSpec) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *DistributedInferenceSpecSpec) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *DistributedInferenceSpecSpec) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *DistributedInferenceSpecSpec) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *DistributedInferenceSpecSpec) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetNodePools

`func (o *DistributedInferenceSpecSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *DistributedInferenceSpecSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *DistributedInferenceSpecSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *DistributedInferenceSpecSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *DistributedInferenceSpecSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *DistributedInferenceSpecSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *DistributedInferenceSpecSpec) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *DistributedInferenceSpecSpec) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *DistributedInferenceSpecSpec) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *DistributedInferenceSpecSpec) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *DistributedInferenceSpecSpec) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *DistributedInferenceSpecSpec) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *DistributedInferenceSpecSpec) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *DistributedInferenceSpecSpec) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *DistributedInferenceSpecSpec) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *DistributedInferenceSpecSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *DistributedInferenceSpecSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *DistributedInferenceSpecSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetRestartPolicy

`func (o *DistributedInferenceSpecSpec) GetRestartPolicy() DistributedInferenceRestartPolicy`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *DistributedInferenceSpecSpec) GetRestartPolicyOk() (*DistributedInferenceRestartPolicy, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *DistributedInferenceSpecSpec) SetRestartPolicy(v DistributedInferenceRestartPolicy)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *DistributedInferenceSpecSpec) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *DistributedInferenceSpecSpec) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *DistributedInferenceSpecSpec) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetServingPort

`func (o *DistributedInferenceSpecSpec) GetServingPort() DistributedInferenceServingPort`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *DistributedInferenceSpecSpec) GetServingPortOk() (*DistributedInferenceServingPort, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *DistributedInferenceSpecSpec) SetServingPort(v DistributedInferenceServingPort)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *DistributedInferenceSpecSpec) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *DistributedInferenceSpecSpec) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *DistributedInferenceSpecSpec) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetStartupPolicy

`func (o *DistributedInferenceSpecSpec) GetStartupPolicy() DistributedInferenceStartupPolicy`

GetStartupPolicy returns the StartupPolicy field if non-nil, zero value otherwise.

### GetStartupPolicyOk

`func (o *DistributedInferenceSpecSpec) GetStartupPolicyOk() (*DistributedInferenceStartupPolicy, bool)`

GetStartupPolicyOk returns a tuple with the StartupPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartupPolicy

`func (o *DistributedInferenceSpecSpec) SetStartupPolicy(v DistributedInferenceStartupPolicy)`

SetStartupPolicy sets StartupPolicy field to given value.

### HasStartupPolicy

`func (o *DistributedInferenceSpecSpec) HasStartupPolicy() bool`

HasStartupPolicy returns a boolean if a field has been set.

### SetStartupPolicyNil

`func (o *DistributedInferenceSpecSpec) SetStartupPolicyNil(b bool)`

 SetStartupPolicyNil sets the value for StartupPolicy to be an explicit nil

### UnsetStartupPolicy
`func (o *DistributedInferenceSpecSpec) UnsetStartupPolicy()`

UnsetStartupPolicy ensures that no value is present for StartupPolicy, not even an explicit nil
### GetWorkers

`func (o *DistributedInferenceSpecSpec) GetWorkers() int32`

GetWorkers returns the Workers field if non-nil, zero value otherwise.

### GetWorkersOk

`func (o *DistributedInferenceSpecSpec) GetWorkersOk() (*int32, bool)`

GetWorkersOk returns a tuple with the Workers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkers

`func (o *DistributedInferenceSpecSpec) SetWorkers(v int32)`

SetWorkers sets Workers field to given value.

### HasWorkers

`func (o *DistributedInferenceSpecSpec) HasWorkers() bool`

HasWorkers returns a boolean if a field has been set.

### SetWorkersNil

`func (o *DistributedInferenceSpecSpec) SetWorkersNil(b bool)`

 SetWorkersNil sets the value for Workers to be an explicit nil

### UnsetWorkers
`func (o *DistributedInferenceSpecSpec) UnsetWorkers()`

UnsetWorkers ensures that no value is present for Workers, not even an explicit nil
### GetReplicas

`func (o *DistributedInferenceSpecSpec) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *DistributedInferenceSpecSpec) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *DistributedInferenceSpecSpec) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *DistributedInferenceSpecSpec) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### SetReplicasNil

`func (o *DistributedInferenceSpecSpec) SetReplicasNil(b bool)`

 SetReplicasNil sets the value for Replicas to be an explicit nil

### UnsetReplicas
`func (o *DistributedInferenceSpecSpec) UnsetReplicas()`

UnsetReplicas ensures that no value is present for Replicas, not even an explicit nil
### GetLeader

`func (o *DistributedInferenceSpecSpec) GetLeader() DistributedInferenceLeaderWorkerSpec`

GetLeader returns the Leader field if non-nil, zero value otherwise.

### GetLeaderOk

`func (o *DistributedInferenceSpecSpec) GetLeaderOk() (*DistributedInferenceLeaderWorkerSpec, bool)`

GetLeaderOk returns a tuple with the Leader field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLeader

`func (o *DistributedInferenceSpecSpec) SetLeader(v DistributedInferenceLeaderWorkerSpec)`

SetLeader sets Leader field to given value.

### HasLeader

`func (o *DistributedInferenceSpecSpec) HasLeader() bool`

HasLeader returns a boolean if a field has been set.

### SetLeaderNil

`func (o *DistributedInferenceSpecSpec) SetLeaderNil(b bool)`

 SetLeaderNil sets the value for Leader to be an explicit nil

### UnsetLeader
`func (o *DistributedInferenceSpecSpec) UnsetLeader()`

UnsetLeader ensures that no value is present for Leader, not even an explicit nil
### GetWorker

`func (o *DistributedInferenceSpecSpec) GetWorker() DistributedInferenceLeaderWorkerSpec`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *DistributedInferenceSpecSpec) GetWorkerOk() (*DistributedInferenceLeaderWorkerSpec, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *DistributedInferenceSpecSpec) SetWorker(v DistributedInferenceLeaderWorkerSpec)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *DistributedInferenceSpecSpec) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *DistributedInferenceSpecSpec) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *DistributedInferenceSpecSpec) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


