# DistributedInferenceCommonSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StartupPolicy** | Pointer to [**NullableDistributedInferenceStartupPolicy**](DistributedInferenceStartupPolicy.md) |  | [optional] [default to DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED]
**Workers** | Pointer to **NullableInt32** | Specifies the number of worker nodes to run. If set to 0, only the leader node will run, and no worker pods will be created. In this case, worker spec is not required. | [optional] [default to 0]
**Replicas** | Pointer to **NullableInt32** | Specifies the number of leader-worker sets to deploy. Each replica represents a group consisting of one leader pod and multiple worker pods.  For example, setting replicas: 3 will create 3 independent groups, each with its own leader and corresponding set of workers.  | [optional] [default to 1]
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**RestartPolicy** | Pointer to [**NullableDistributedInferenceRestartPolicy**](DistributedInferenceRestartPolicy.md) |  | [optional] [default to DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART]
**ServingPort** | Pointer to [**NullableDistributedInferenceServingPort**](DistributedInferenceServingPort.md) |  | [optional] 

## Methods

### NewDistributedInferenceCommonSpec

`func NewDistributedInferenceCommonSpec() *DistributedInferenceCommonSpec`

NewDistributedInferenceCommonSpec instantiates a new DistributedInferenceCommonSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceCommonSpecWithDefaults

`func NewDistributedInferenceCommonSpecWithDefaults() *DistributedInferenceCommonSpec`

NewDistributedInferenceCommonSpecWithDefaults instantiates a new DistributedInferenceCommonSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStartupPolicy

`func (o *DistributedInferenceCommonSpec) GetStartupPolicy() DistributedInferenceStartupPolicy`

GetStartupPolicy returns the StartupPolicy field if non-nil, zero value otherwise.

### GetStartupPolicyOk

`func (o *DistributedInferenceCommonSpec) GetStartupPolicyOk() (*DistributedInferenceStartupPolicy, bool)`

GetStartupPolicyOk returns a tuple with the StartupPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartupPolicy

`func (o *DistributedInferenceCommonSpec) SetStartupPolicy(v DistributedInferenceStartupPolicy)`

SetStartupPolicy sets StartupPolicy field to given value.

### HasStartupPolicy

`func (o *DistributedInferenceCommonSpec) HasStartupPolicy() bool`

HasStartupPolicy returns a boolean if a field has been set.

### SetStartupPolicyNil

`func (o *DistributedInferenceCommonSpec) SetStartupPolicyNil(b bool)`

 SetStartupPolicyNil sets the value for StartupPolicy to be an explicit nil

### UnsetStartupPolicy
`func (o *DistributedInferenceCommonSpec) UnsetStartupPolicy()`

UnsetStartupPolicy ensures that no value is present for StartupPolicy, not even an explicit nil
### GetWorkers

`func (o *DistributedInferenceCommonSpec) GetWorkers() int32`

GetWorkers returns the Workers field if non-nil, zero value otherwise.

### GetWorkersOk

`func (o *DistributedInferenceCommonSpec) GetWorkersOk() (*int32, bool)`

GetWorkersOk returns a tuple with the Workers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkers

`func (o *DistributedInferenceCommonSpec) SetWorkers(v int32)`

SetWorkers sets Workers field to given value.

### HasWorkers

`func (o *DistributedInferenceCommonSpec) HasWorkers() bool`

HasWorkers returns a boolean if a field has been set.

### SetWorkersNil

`func (o *DistributedInferenceCommonSpec) SetWorkersNil(b bool)`

 SetWorkersNil sets the value for Workers to be an explicit nil

### UnsetWorkers
`func (o *DistributedInferenceCommonSpec) UnsetWorkers()`

UnsetWorkers ensures that no value is present for Workers, not even an explicit nil
### GetReplicas

`func (o *DistributedInferenceCommonSpec) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *DistributedInferenceCommonSpec) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *DistributedInferenceCommonSpec) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *DistributedInferenceCommonSpec) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### SetReplicasNil

`func (o *DistributedInferenceCommonSpec) SetReplicasNil(b bool)`

 SetReplicasNil sets the value for Replicas to be an explicit nil

### UnsetReplicas
`func (o *DistributedInferenceCommonSpec) UnsetReplicas()`

UnsetReplicas ensures that no value is present for Replicas, not even an explicit nil
### GetCategory

`func (o *DistributedInferenceCommonSpec) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *DistributedInferenceCommonSpec) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *DistributedInferenceCommonSpec) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *DistributedInferenceCommonSpec) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *DistributedInferenceCommonSpec) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *DistributedInferenceCommonSpec) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetNodePools

`func (o *DistributedInferenceCommonSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *DistributedInferenceCommonSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *DistributedInferenceCommonSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *DistributedInferenceCommonSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *DistributedInferenceCommonSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *DistributedInferenceCommonSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *DistributedInferenceCommonSpec) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *DistributedInferenceCommonSpec) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *DistributedInferenceCommonSpec) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *DistributedInferenceCommonSpec) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *DistributedInferenceCommonSpec) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *DistributedInferenceCommonSpec) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *DistributedInferenceCommonSpec) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *DistributedInferenceCommonSpec) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *DistributedInferenceCommonSpec) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *DistributedInferenceCommonSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *DistributedInferenceCommonSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *DistributedInferenceCommonSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetRestartPolicy

`func (o *DistributedInferenceCommonSpec) GetRestartPolicy() DistributedInferenceRestartPolicy`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *DistributedInferenceCommonSpec) GetRestartPolicyOk() (*DistributedInferenceRestartPolicy, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *DistributedInferenceCommonSpec) SetRestartPolicy(v DistributedInferenceRestartPolicy)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *DistributedInferenceCommonSpec) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *DistributedInferenceCommonSpec) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *DistributedInferenceCommonSpec) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetServingPort

`func (o *DistributedInferenceCommonSpec) GetServingPort() DistributedInferenceServingPort`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *DistributedInferenceCommonSpec) GetServingPortOk() (*DistributedInferenceServingPort, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *DistributedInferenceCommonSpec) SetServingPort(v DistributedInferenceServingPort)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *DistributedInferenceCommonSpec) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *DistributedInferenceCommonSpec) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *DistributedInferenceCommonSpec) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


