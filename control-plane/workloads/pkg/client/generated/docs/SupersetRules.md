# SupersetRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Autoscaling** | Pointer to [**NullableAutoScalingRules**](AutoScalingRules.md) |  | [optional] 
**ServingConfiguration** | Pointer to [**NullableServingConfigurationRules**](ServingConfigurationRules.md) |  | [optional] 
**Leader** | Pointer to [**NullableDistributedInferenceLeaderWorkerRulesV2**](DistributedInferenceLeaderWorkerRulesV2.md) | Defines the pod specification for the leader. Must always be provided, regardless of the number of workers. | [optional] 
**Worker** | Pointer to [**NullableDistributedInferenceLeaderWorkerRulesV2**](DistributedInferenceLeaderWorkerRulesV2.md) | Defines the pod specification for the workers. Required only if the number of workers is greater than 0. | [optional] 
**Category** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibilityRules**](PreemptibilityRules.md) |  | [optional] 
**PriorityClass** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**RestartPolicy** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableSupersetServingPortRules**](SupersetServingPortRules.md) |  | [optional] 
**StartupPolicy** | Pointer to [**NullableDistributedInferenceStartupPolicyFieldRulesStartupPolicy**](DistributedInferenceStartupPolicyFieldRulesStartupPolicy.md) |  | [optional] 
**Workers** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Replicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Annotations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Args** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**BackoffLimit** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**CleanPodPolicy** | Pointer to [**NullableCleanPodPolicyRules**](CleanPodPolicyRules.md) |  | [optional] 
**Command** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Completions** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Compute** | Pointer to [**NullableSupersetRulesAllOfCompute**](SupersetRulesAllOfCompute.md) |  | [optional] 
**CreateHomeDir** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**DistributedFramework** | Pointer to [**NullableDistributedFrameworkRules**](DistributedFrameworkRules.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsRules**](ExposedUrlsRules.md) |  | [optional] 
**Image** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicyRules**](ImagePullPolicyRules.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**NullableImagePullSecretsRules**](ImagePullSecretsRules.md) |  | [optional] 
**Labels** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**MaxReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**MinReplicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**MpiLauncherCreationPolicy** | Pointer to [**NullableMpiLauncherCreationPolicyRules**](MpiLauncherCreationPolicyRules.md) |  | [optional] 
**NodeAffinityRequired** | Pointer to [**NullableArbitraryRules**](ArbitraryRules.md) |  | [optional] 
**NodeType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**NumWorkers** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Parallelism** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinityRules**](PodAffinityRules.md) |  | [optional] 
**Ports** | Pointer to [**NullablePortsRules**](PortsRules.md) |  | [optional] 
**Probes** | Pointer to [**NullableProbesRules**](ProbesRules.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsRules**](RelatedUrlsRules.md) |  | [optional] 
**Security** | Pointer to [**NullableSupersetRulesAllOfSecurity**](SupersetRulesAllOfSecurity.md) |  | [optional] 
**SlotsPerWorker** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**SshAuthMountPath** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Stdin** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**Storage** | Pointer to [**NullableSupersetRulesAllOfStorage**](SupersetRulesAllOfStorage.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**TerminationGracePeriodSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsRules**](TolerationsRules.md) |  | [optional] 
**Tty** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**WorkingDir** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewSupersetRules

`func NewSupersetRules() *SupersetRules`

NewSupersetRules instantiates a new SupersetRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetRulesWithDefaults

`func NewSupersetRulesWithDefaults() *SupersetRules`

NewSupersetRulesWithDefaults instantiates a new SupersetRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAutoscaling

`func (o *SupersetRules) GetAutoscaling() AutoScalingRules`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *SupersetRules) GetAutoscalingOk() (*AutoScalingRules, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *SupersetRules) SetAutoscaling(v AutoScalingRules)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *SupersetRules) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *SupersetRules) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *SupersetRules) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil
### GetServingConfiguration

`func (o *SupersetRules) GetServingConfiguration() ServingConfigurationRules`

GetServingConfiguration returns the ServingConfiguration field if non-nil, zero value otherwise.

### GetServingConfigurationOk

`func (o *SupersetRules) GetServingConfigurationOk() (*ServingConfigurationRules, bool)`

GetServingConfigurationOk returns a tuple with the ServingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingConfiguration

`func (o *SupersetRules) SetServingConfiguration(v ServingConfigurationRules)`

SetServingConfiguration sets ServingConfiguration field to given value.

### HasServingConfiguration

`func (o *SupersetRules) HasServingConfiguration() bool`

HasServingConfiguration returns a boolean if a field has been set.

### SetServingConfigurationNil

`func (o *SupersetRules) SetServingConfigurationNil(b bool)`

 SetServingConfigurationNil sets the value for ServingConfiguration to be an explicit nil

### UnsetServingConfiguration
`func (o *SupersetRules) UnsetServingConfiguration()`

UnsetServingConfiguration ensures that no value is present for ServingConfiguration, not even an explicit nil
### GetLeader

`func (o *SupersetRules) GetLeader() DistributedInferenceLeaderWorkerRulesV2`

GetLeader returns the Leader field if non-nil, zero value otherwise.

### GetLeaderOk

`func (o *SupersetRules) GetLeaderOk() (*DistributedInferenceLeaderWorkerRulesV2, bool)`

GetLeaderOk returns a tuple with the Leader field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLeader

`func (o *SupersetRules) SetLeader(v DistributedInferenceLeaderWorkerRulesV2)`

SetLeader sets Leader field to given value.

### HasLeader

`func (o *SupersetRules) HasLeader() bool`

HasLeader returns a boolean if a field has been set.

### SetLeaderNil

`func (o *SupersetRules) SetLeaderNil(b bool)`

 SetLeaderNil sets the value for Leader to be an explicit nil

### UnsetLeader
`func (o *SupersetRules) UnsetLeader()`

UnsetLeader ensures that no value is present for Leader, not even an explicit nil
### GetWorker

`func (o *SupersetRules) GetWorker() DistributedInferenceLeaderWorkerRulesV2`

GetWorker returns the Worker field if non-nil, zero value otherwise.

### GetWorkerOk

`func (o *SupersetRules) GetWorkerOk() (*DistributedInferenceLeaderWorkerRulesV2, bool)`

GetWorkerOk returns a tuple with the Worker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorker

`func (o *SupersetRules) SetWorker(v DistributedInferenceLeaderWorkerRulesV2)`

SetWorker sets Worker field to given value.

### HasWorker

`func (o *SupersetRules) HasWorker() bool`

HasWorker returns a boolean if a field has been set.

### SetWorkerNil

`func (o *SupersetRules) SetWorkerNil(b bool)`

 SetWorkerNil sets the value for Worker to be an explicit nil

### UnsetWorker
`func (o *SupersetRules) UnsetWorker()`

UnsetWorker ensures that no value is present for Worker, not even an explicit nil
### GetCategory

`func (o *SupersetRules) GetCategory() StringRules`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *SupersetRules) GetCategoryOk() (*StringRules, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *SupersetRules) SetCategory(v StringRules)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *SupersetRules) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *SupersetRules) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *SupersetRules) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetNodePools

`func (o *SupersetRules) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *SupersetRules) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *SupersetRules) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *SupersetRules) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *SupersetRules) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *SupersetRules) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *SupersetRules) GetPreemptibility() PreemptibilityRules`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *SupersetRules) GetPreemptibilityOk() (*PreemptibilityRules, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *SupersetRules) SetPreemptibility(v PreemptibilityRules)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *SupersetRules) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *SupersetRules) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *SupersetRules) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *SupersetRules) GetPriorityClass() StringRules`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *SupersetRules) GetPriorityClassOk() (*StringRules, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *SupersetRules) SetPriorityClass(v StringRules)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *SupersetRules) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *SupersetRules) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *SupersetRules) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetRestartPolicy

`func (o *SupersetRules) GetRestartPolicy() StringRules`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *SupersetRules) GetRestartPolicyOk() (*StringRules, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *SupersetRules) SetRestartPolicy(v StringRules)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *SupersetRules) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *SupersetRules) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *SupersetRules) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetServingPort

`func (o *SupersetRules) GetServingPort() SupersetServingPortRules`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *SupersetRules) GetServingPortOk() (*SupersetServingPortRules, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *SupersetRules) SetServingPort(v SupersetServingPortRules)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *SupersetRules) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *SupersetRules) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *SupersetRules) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetStartupPolicy

`func (o *SupersetRules) GetStartupPolicy() DistributedInferenceStartupPolicyFieldRulesStartupPolicy`

GetStartupPolicy returns the StartupPolicy field if non-nil, zero value otherwise.

### GetStartupPolicyOk

`func (o *SupersetRules) GetStartupPolicyOk() (*DistributedInferenceStartupPolicyFieldRulesStartupPolicy, bool)`

GetStartupPolicyOk returns a tuple with the StartupPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartupPolicy

`func (o *SupersetRules) SetStartupPolicy(v DistributedInferenceStartupPolicyFieldRulesStartupPolicy)`

SetStartupPolicy sets StartupPolicy field to given value.

### HasStartupPolicy

`func (o *SupersetRules) HasStartupPolicy() bool`

HasStartupPolicy returns a boolean if a field has been set.

### SetStartupPolicyNil

`func (o *SupersetRules) SetStartupPolicyNil(b bool)`

 SetStartupPolicyNil sets the value for StartupPolicy to be an explicit nil

### UnsetStartupPolicy
`func (o *SupersetRules) UnsetStartupPolicy()`

UnsetStartupPolicy ensures that no value is present for StartupPolicy, not even an explicit nil
### GetWorkers

`func (o *SupersetRules) GetWorkers() IntegerRules`

GetWorkers returns the Workers field if non-nil, zero value otherwise.

### GetWorkersOk

`func (o *SupersetRules) GetWorkersOk() (*IntegerRules, bool)`

GetWorkersOk returns a tuple with the Workers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkers

`func (o *SupersetRules) SetWorkers(v IntegerRules)`

SetWorkers sets Workers field to given value.

### HasWorkers

`func (o *SupersetRules) HasWorkers() bool`

HasWorkers returns a boolean if a field has been set.

### SetWorkersNil

`func (o *SupersetRules) SetWorkersNil(b bool)`

 SetWorkersNil sets the value for Workers to be an explicit nil

### UnsetWorkers
`func (o *SupersetRules) UnsetWorkers()`

UnsetWorkers ensures that no value is present for Workers, not even an explicit nil
### GetReplicas

`func (o *SupersetRules) GetReplicas() IntegerRules`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *SupersetRules) GetReplicasOk() (*IntegerRules, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *SupersetRules) SetReplicas(v IntegerRules)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *SupersetRules) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### SetReplicasNil

`func (o *SupersetRules) SetReplicasNil(b bool)`

 SetReplicasNil sets the value for Replicas to be an explicit nil

### UnsetReplicas
`func (o *SupersetRules) UnsetReplicas()`

UnsetReplicas ensures that no value is present for Replicas, not even an explicit nil
### GetAnnotations

`func (o *SupersetRules) GetAnnotations() InstancesRules`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *SupersetRules) GetAnnotationsOk() (*InstancesRules, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *SupersetRules) SetAnnotations(v InstancesRules)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *SupersetRules) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *SupersetRules) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *SupersetRules) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *SupersetRules) GetArgs() StringRules`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *SupersetRules) GetArgsOk() (*StringRules, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *SupersetRules) SetArgs(v StringRules)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *SupersetRules) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *SupersetRules) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *SupersetRules) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *SupersetRules) GetAutoDeletionTimeAfterCompletionSeconds() IntegerRules`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *SupersetRules) GetAutoDeletionTimeAfterCompletionSecondsOk() (*IntegerRules, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *SupersetRules) SetAutoDeletionTimeAfterCompletionSeconds(v IntegerRules)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *SupersetRules) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *SupersetRules) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *SupersetRules) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *SupersetRules) GetBackoffLimit() IntegerRules`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *SupersetRules) GetBackoffLimitOk() (*IntegerRules, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *SupersetRules) SetBackoffLimit(v IntegerRules)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *SupersetRules) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *SupersetRules) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *SupersetRules) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCleanPodPolicy

`func (o *SupersetRules) GetCleanPodPolicy() CleanPodPolicyRules`

GetCleanPodPolicy returns the CleanPodPolicy field if non-nil, zero value otherwise.

### GetCleanPodPolicyOk

`func (o *SupersetRules) GetCleanPodPolicyOk() (*CleanPodPolicyRules, bool)`

GetCleanPodPolicyOk returns a tuple with the CleanPodPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCleanPodPolicy

`func (o *SupersetRules) SetCleanPodPolicy(v CleanPodPolicyRules)`

SetCleanPodPolicy sets CleanPodPolicy field to given value.

### HasCleanPodPolicy

`func (o *SupersetRules) HasCleanPodPolicy() bool`

HasCleanPodPolicy returns a boolean if a field has been set.

### SetCleanPodPolicyNil

`func (o *SupersetRules) SetCleanPodPolicyNil(b bool)`

 SetCleanPodPolicyNil sets the value for CleanPodPolicy to be an explicit nil

### UnsetCleanPodPolicy
`func (o *SupersetRules) UnsetCleanPodPolicy()`

UnsetCleanPodPolicy ensures that no value is present for CleanPodPolicy, not even an explicit nil
### GetCommand

`func (o *SupersetRules) GetCommand() StringRules`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *SupersetRules) GetCommandOk() (*StringRules, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *SupersetRules) SetCommand(v StringRules)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *SupersetRules) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *SupersetRules) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *SupersetRules) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompletions

`func (o *SupersetRules) GetCompletions() IntegerRules`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *SupersetRules) GetCompletionsOk() (*IntegerRules, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *SupersetRules) SetCompletions(v IntegerRules)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *SupersetRules) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### SetCompletionsNil

`func (o *SupersetRules) SetCompletionsNil(b bool)`

 SetCompletionsNil sets the value for Completions to be an explicit nil

### UnsetCompletions
`func (o *SupersetRules) UnsetCompletions()`

UnsetCompletions ensures that no value is present for Completions, not even an explicit nil
### GetCompute

`func (o *SupersetRules) GetCompute() SupersetRulesAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *SupersetRules) GetComputeOk() (*SupersetRulesAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *SupersetRules) SetCompute(v SupersetRulesAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *SupersetRules) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *SupersetRules) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *SupersetRules) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *SupersetRules) GetCreateHomeDir() BooleanRules`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *SupersetRules) GetCreateHomeDirOk() (*BooleanRules, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *SupersetRules) SetCreateHomeDir(v BooleanRules)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *SupersetRules) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *SupersetRules) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *SupersetRules) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetDistributedFramework

`func (o *SupersetRules) GetDistributedFramework() DistributedFrameworkRules`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *SupersetRules) GetDistributedFrameworkOk() (*DistributedFrameworkRules, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *SupersetRules) SetDistributedFramework(v DistributedFrameworkRules)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *SupersetRules) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### SetDistributedFrameworkNil

`func (o *SupersetRules) SetDistributedFrameworkNil(b bool)`

 SetDistributedFrameworkNil sets the value for DistributedFramework to be an explicit nil

### UnsetDistributedFramework
`func (o *SupersetRules) UnsetDistributedFramework()`

UnsetDistributedFramework ensures that no value is present for DistributedFramework, not even an explicit nil
### GetEnvironmentVariables

`func (o *SupersetRules) GetEnvironmentVariables() InstancesRules`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *SupersetRules) GetEnvironmentVariablesOk() (*InstancesRules, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *SupersetRules) SetEnvironmentVariables(v InstancesRules)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *SupersetRules) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *SupersetRules) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *SupersetRules) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *SupersetRules) GetExposedUrls() ExposedUrlsRules`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *SupersetRules) GetExposedUrlsOk() (*ExposedUrlsRules, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *SupersetRules) SetExposedUrls(v ExposedUrlsRules)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *SupersetRules) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *SupersetRules) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *SupersetRules) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *SupersetRules) GetImage() StringRules`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *SupersetRules) GetImageOk() (*StringRules, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *SupersetRules) SetImage(v StringRules)`

SetImage sets Image field to given value.

### HasImage

`func (o *SupersetRules) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *SupersetRules) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *SupersetRules) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *SupersetRules) GetImagePullPolicy() ImagePullPolicyRules`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *SupersetRules) GetImagePullPolicyOk() (*ImagePullPolicyRules, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *SupersetRules) SetImagePullPolicy(v ImagePullPolicyRules)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *SupersetRules) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *SupersetRules) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *SupersetRules) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *SupersetRules) GetImagePullSecrets() ImagePullSecretsRules`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *SupersetRules) GetImagePullSecretsOk() (*ImagePullSecretsRules, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *SupersetRules) SetImagePullSecrets(v ImagePullSecretsRules)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *SupersetRules) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *SupersetRules) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *SupersetRules) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *SupersetRules) GetLabels() InstancesRules`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *SupersetRules) GetLabelsOk() (*InstancesRules, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *SupersetRules) SetLabels(v InstancesRules)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *SupersetRules) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *SupersetRules) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *SupersetRules) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetMaxReplicas

`func (o *SupersetRules) GetMaxReplicas() IntegerRules`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *SupersetRules) GetMaxReplicasOk() (*IntegerRules, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *SupersetRules) SetMaxReplicas(v IntegerRules)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *SupersetRules) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *SupersetRules) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *SupersetRules) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetMinReplicas

`func (o *SupersetRules) GetMinReplicas() IntegerRules`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *SupersetRules) GetMinReplicasOk() (*IntegerRules, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *SupersetRules) SetMinReplicas(v IntegerRules)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *SupersetRules) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *SupersetRules) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *SupersetRules) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMpiLauncherCreationPolicy

`func (o *SupersetRules) GetMpiLauncherCreationPolicy() MpiLauncherCreationPolicyRules`

GetMpiLauncherCreationPolicy returns the MpiLauncherCreationPolicy field if non-nil, zero value otherwise.

### GetMpiLauncherCreationPolicyOk

`func (o *SupersetRules) GetMpiLauncherCreationPolicyOk() (*MpiLauncherCreationPolicyRules, bool)`

GetMpiLauncherCreationPolicyOk returns a tuple with the MpiLauncherCreationPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMpiLauncherCreationPolicy

`func (o *SupersetRules) SetMpiLauncherCreationPolicy(v MpiLauncherCreationPolicyRules)`

SetMpiLauncherCreationPolicy sets MpiLauncherCreationPolicy field to given value.

### HasMpiLauncherCreationPolicy

`func (o *SupersetRules) HasMpiLauncherCreationPolicy() bool`

HasMpiLauncherCreationPolicy returns a boolean if a field has been set.

### SetMpiLauncherCreationPolicyNil

`func (o *SupersetRules) SetMpiLauncherCreationPolicyNil(b bool)`

 SetMpiLauncherCreationPolicyNil sets the value for MpiLauncherCreationPolicy to be an explicit nil

### UnsetMpiLauncherCreationPolicy
`func (o *SupersetRules) UnsetMpiLauncherCreationPolicy()`

UnsetMpiLauncherCreationPolicy ensures that no value is present for MpiLauncherCreationPolicy, not even an explicit nil
### GetNodeAffinityRequired

`func (o *SupersetRules) GetNodeAffinityRequired() ArbitraryRules`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *SupersetRules) GetNodeAffinityRequiredOk() (*ArbitraryRules, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *SupersetRules) SetNodeAffinityRequired(v ArbitraryRules)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *SupersetRules) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *SupersetRules) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *SupersetRules) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodeType

`func (o *SupersetRules) GetNodeType() StringRules`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *SupersetRules) GetNodeTypeOk() (*StringRules, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *SupersetRules) SetNodeType(v StringRules)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *SupersetRules) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *SupersetRules) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *SupersetRules) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNumWorkers

`func (o *SupersetRules) GetNumWorkers() IntegerRules`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *SupersetRules) GetNumWorkersOk() (*IntegerRules, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *SupersetRules) SetNumWorkers(v IntegerRules)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *SupersetRules) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### SetNumWorkersNil

`func (o *SupersetRules) SetNumWorkersNil(b bool)`

 SetNumWorkersNil sets the value for NumWorkers to be an explicit nil

### UnsetNumWorkers
`func (o *SupersetRules) UnsetNumWorkers()`

UnsetNumWorkers ensures that no value is present for NumWorkers, not even an explicit nil
### GetParallelism

`func (o *SupersetRules) GetParallelism() IntegerRules`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *SupersetRules) GetParallelismOk() (*IntegerRules, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *SupersetRules) SetParallelism(v IntegerRules)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *SupersetRules) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### SetParallelismNil

`func (o *SupersetRules) SetParallelismNil(b bool)`

 SetParallelismNil sets the value for Parallelism to be an explicit nil

### UnsetParallelism
`func (o *SupersetRules) UnsetParallelism()`

UnsetParallelism ensures that no value is present for Parallelism, not even an explicit nil
### GetPodAffinity

`func (o *SupersetRules) GetPodAffinity() PodAffinityRules`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *SupersetRules) GetPodAffinityOk() (*PodAffinityRules, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *SupersetRules) SetPodAffinity(v PodAffinityRules)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *SupersetRules) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *SupersetRules) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *SupersetRules) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *SupersetRules) GetPorts() PortsRules`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *SupersetRules) GetPortsOk() (*PortsRules, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *SupersetRules) SetPorts(v PortsRules)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *SupersetRules) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *SupersetRules) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *SupersetRules) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetProbes

`func (o *SupersetRules) GetProbes() ProbesRules`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *SupersetRules) GetProbesOk() (*ProbesRules, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *SupersetRules) SetProbes(v ProbesRules)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *SupersetRules) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *SupersetRules) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *SupersetRules) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *SupersetRules) GetRelatedUrls() RelatedUrlsRules`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *SupersetRules) GetRelatedUrlsOk() (*RelatedUrlsRules, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *SupersetRules) SetRelatedUrls(v RelatedUrlsRules)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *SupersetRules) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *SupersetRules) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *SupersetRules) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetSecurity

`func (o *SupersetRules) GetSecurity() SupersetRulesAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *SupersetRules) GetSecurityOk() (*SupersetRulesAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *SupersetRules) SetSecurity(v SupersetRulesAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *SupersetRules) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *SupersetRules) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *SupersetRules) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetSlotsPerWorker

`func (o *SupersetRules) GetSlotsPerWorker() IntegerRules`

GetSlotsPerWorker returns the SlotsPerWorker field if non-nil, zero value otherwise.

### GetSlotsPerWorkerOk

`func (o *SupersetRules) GetSlotsPerWorkerOk() (*IntegerRules, bool)`

GetSlotsPerWorkerOk returns a tuple with the SlotsPerWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSlotsPerWorker

`func (o *SupersetRules) SetSlotsPerWorker(v IntegerRules)`

SetSlotsPerWorker sets SlotsPerWorker field to given value.

### HasSlotsPerWorker

`func (o *SupersetRules) HasSlotsPerWorker() bool`

HasSlotsPerWorker returns a boolean if a field has been set.

### SetSlotsPerWorkerNil

`func (o *SupersetRules) SetSlotsPerWorkerNil(b bool)`

 SetSlotsPerWorkerNil sets the value for SlotsPerWorker to be an explicit nil

### UnsetSlotsPerWorker
`func (o *SupersetRules) UnsetSlotsPerWorker()`

UnsetSlotsPerWorker ensures that no value is present for SlotsPerWorker, not even an explicit nil
### GetSshAuthMountPath

`func (o *SupersetRules) GetSshAuthMountPath() StringRules`

GetSshAuthMountPath returns the SshAuthMountPath field if non-nil, zero value otherwise.

### GetSshAuthMountPathOk

`func (o *SupersetRules) GetSshAuthMountPathOk() (*StringRules, bool)`

GetSshAuthMountPathOk returns a tuple with the SshAuthMountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshAuthMountPath

`func (o *SupersetRules) SetSshAuthMountPath(v StringRules)`

SetSshAuthMountPath sets SshAuthMountPath field to given value.

### HasSshAuthMountPath

`func (o *SupersetRules) HasSshAuthMountPath() bool`

HasSshAuthMountPath returns a boolean if a field has been set.

### SetSshAuthMountPathNil

`func (o *SupersetRules) SetSshAuthMountPathNil(b bool)`

 SetSshAuthMountPathNil sets the value for SshAuthMountPath to be an explicit nil

### UnsetSshAuthMountPath
`func (o *SupersetRules) UnsetSshAuthMountPath()`

UnsetSshAuthMountPath ensures that no value is present for SshAuthMountPath, not even an explicit nil
### GetStdin

`func (o *SupersetRules) GetStdin() BooleanRules`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *SupersetRules) GetStdinOk() (*BooleanRules, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *SupersetRules) SetStdin(v BooleanRules)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *SupersetRules) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *SupersetRules) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *SupersetRules) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *SupersetRules) GetStorage() SupersetRulesAllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *SupersetRules) GetStorageOk() (*SupersetRulesAllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *SupersetRules) SetStorage(v SupersetRulesAllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *SupersetRules) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *SupersetRules) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *SupersetRules) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *SupersetRules) GetTerminateAfterPreemption() BooleanRules`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *SupersetRules) GetTerminateAfterPreemptionOk() (*BooleanRules, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *SupersetRules) SetTerminateAfterPreemption(v BooleanRules)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *SupersetRules) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *SupersetRules) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *SupersetRules) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *SupersetRules) GetTerminationGracePeriodSeconds() IntegerRules`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *SupersetRules) GetTerminationGracePeriodSecondsOk() (*IntegerRules, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *SupersetRules) SetTerminationGracePeriodSeconds(v IntegerRules)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *SupersetRules) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *SupersetRules) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *SupersetRules) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *SupersetRules) GetTolerations() TolerationsRules`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *SupersetRules) GetTolerationsOk() (*TolerationsRules, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *SupersetRules) SetTolerations(v TolerationsRules)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *SupersetRules) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *SupersetRules) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *SupersetRules) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTty

`func (o *SupersetRules) GetTty() BooleanRules`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *SupersetRules) GetTtyOk() (*BooleanRules, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *SupersetRules) SetTty(v BooleanRules)`

SetTty sets Tty field to given value.

### HasTty

`func (o *SupersetRules) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *SupersetRules) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *SupersetRules) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *SupersetRules) GetWorkingDir() StringRules`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *SupersetRules) GetWorkingDirOk() (*StringRules, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *SupersetRules) SetWorkingDir(v StringRules)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *SupersetRules) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *SupersetRules) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *SupersetRules) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


