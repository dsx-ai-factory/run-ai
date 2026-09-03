# SpecificRunParamsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Autoscaling** | Pointer to [**NullableAutoScalingRules**](AutoScalingRules.md) |  | [optional] 
**ServingConfiguration** | Pointer to [**NullableServingConfigurationRules**](ServingConfigurationRules.md) |  | [optional] 
**Annotations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Args** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**BackoffLimit** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Category** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Command** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Completions** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**NullableImagePullSecretsRules**](ImagePullSecretsRules.md) |  | [optional] 
**Labels** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**NodeAffinityRequired** | Pointer to [**NullableArbitraryRules**](ArbitraryRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**NodeType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Parallelism** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinityRules**](PodAffinityRules.md) |  | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibilityRules**](PreemptibilityRules.md) |  | [optional] 
**PriorityClass** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**RestartPolicy** | Pointer to [**NullableRestartPolicyRule**](RestartPolicyRule.md) |  | [optional] 
**RunAsGid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**RunAsUid** | Pointer to [**NullableIntegerRulesOptional**](IntegerRulesOptional.md) |  | [optional] 
**SupplementalGroups** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**TerminationGracePeriodSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsRules**](TolerationsRules.md) |  | [optional] 

## Methods

### NewSpecificRunParamsRules

`func NewSpecificRunParamsRules() *SpecificRunParamsRules`

NewSpecificRunParamsRules instantiates a new SpecificRunParamsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunParamsRulesWithDefaults

`func NewSpecificRunParamsRulesWithDefaults() *SpecificRunParamsRules`

NewSpecificRunParamsRulesWithDefaults instantiates a new SpecificRunParamsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAutoscaling

`func (o *SpecificRunParamsRules) GetAutoscaling() AutoScalingRules`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *SpecificRunParamsRules) GetAutoscalingOk() (*AutoScalingRules, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *SpecificRunParamsRules) SetAutoscaling(v AutoScalingRules)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *SpecificRunParamsRules) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *SpecificRunParamsRules) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *SpecificRunParamsRules) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil
### GetServingConfiguration

`func (o *SpecificRunParamsRules) GetServingConfiguration() ServingConfigurationRules`

GetServingConfiguration returns the ServingConfiguration field if non-nil, zero value otherwise.

### GetServingConfigurationOk

`func (o *SpecificRunParamsRules) GetServingConfigurationOk() (*ServingConfigurationRules, bool)`

GetServingConfigurationOk returns a tuple with the ServingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingConfiguration

`func (o *SpecificRunParamsRules) SetServingConfiguration(v ServingConfigurationRules)`

SetServingConfiguration sets ServingConfiguration field to given value.

### HasServingConfiguration

`func (o *SpecificRunParamsRules) HasServingConfiguration() bool`

HasServingConfiguration returns a boolean if a field has been set.

### SetServingConfigurationNil

`func (o *SpecificRunParamsRules) SetServingConfigurationNil(b bool)`

 SetServingConfigurationNil sets the value for ServingConfiguration to be an explicit nil

### UnsetServingConfiguration
`func (o *SpecificRunParamsRules) UnsetServingConfiguration()`

UnsetServingConfiguration ensures that no value is present for ServingConfiguration, not even an explicit nil
### GetAnnotations

`func (o *SpecificRunParamsRules) GetAnnotations() InstancesRules`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *SpecificRunParamsRules) GetAnnotationsOk() (*InstancesRules, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *SpecificRunParamsRules) SetAnnotations(v InstancesRules)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *SpecificRunParamsRules) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *SpecificRunParamsRules) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *SpecificRunParamsRules) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *SpecificRunParamsRules) GetArgs() StringRules`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *SpecificRunParamsRules) GetArgsOk() (*StringRules, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *SpecificRunParamsRules) SetArgs(v StringRules)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *SpecificRunParamsRules) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *SpecificRunParamsRules) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *SpecificRunParamsRules) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsRules) GetAutoDeletionTimeAfterCompletionSeconds() IntegerRules`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *SpecificRunParamsRules) GetAutoDeletionTimeAfterCompletionSecondsOk() (*IntegerRules, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsRules) SetAutoDeletionTimeAfterCompletionSeconds(v IntegerRules)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsRules) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *SpecificRunParamsRules) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *SpecificRunParamsRules) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *SpecificRunParamsRules) GetBackoffLimit() IntegerRules`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *SpecificRunParamsRules) GetBackoffLimitOk() (*IntegerRules, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *SpecificRunParamsRules) SetBackoffLimit(v IntegerRules)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *SpecificRunParamsRules) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *SpecificRunParamsRules) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *SpecificRunParamsRules) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCategory

`func (o *SpecificRunParamsRules) GetCategory() StringRules`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *SpecificRunParamsRules) GetCategoryOk() (*StringRules, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *SpecificRunParamsRules) SetCategory(v StringRules)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *SpecificRunParamsRules) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *SpecificRunParamsRules) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *SpecificRunParamsRules) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCommand

`func (o *SpecificRunParamsRules) GetCommand() StringRules`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *SpecificRunParamsRules) GetCommandOk() (*StringRules, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *SpecificRunParamsRules) SetCommand(v StringRules)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *SpecificRunParamsRules) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *SpecificRunParamsRules) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *SpecificRunParamsRules) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompletions

`func (o *SpecificRunParamsRules) GetCompletions() IntegerRules`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *SpecificRunParamsRules) GetCompletionsOk() (*IntegerRules, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *SpecificRunParamsRules) SetCompletions(v IntegerRules)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *SpecificRunParamsRules) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### SetCompletionsNil

`func (o *SpecificRunParamsRules) SetCompletionsNil(b bool)`

 SetCompletionsNil sets the value for Completions to be an explicit nil

### UnsetCompletions
`func (o *SpecificRunParamsRules) UnsetCompletions()`

UnsetCompletions ensures that no value is present for Completions, not even an explicit nil
### GetEnvironmentVariables

`func (o *SpecificRunParamsRules) GetEnvironmentVariables() InstancesRules`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *SpecificRunParamsRules) GetEnvironmentVariablesOk() (*InstancesRules, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *SpecificRunParamsRules) SetEnvironmentVariables(v InstancesRules)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *SpecificRunParamsRules) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *SpecificRunParamsRules) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *SpecificRunParamsRules) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetImagePullSecrets

`func (o *SpecificRunParamsRules) GetImagePullSecrets() ImagePullSecretsRules`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *SpecificRunParamsRules) GetImagePullSecretsOk() (*ImagePullSecretsRules, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *SpecificRunParamsRules) SetImagePullSecrets(v ImagePullSecretsRules)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *SpecificRunParamsRules) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *SpecificRunParamsRules) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *SpecificRunParamsRules) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *SpecificRunParamsRules) GetLabels() InstancesRules`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *SpecificRunParamsRules) GetLabelsOk() (*InstancesRules, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *SpecificRunParamsRules) SetLabels(v InstancesRules)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *SpecificRunParamsRules) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *SpecificRunParamsRules) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *SpecificRunParamsRules) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetNodeAffinityRequired

`func (o *SpecificRunParamsRules) GetNodeAffinityRequired() ArbitraryRules`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *SpecificRunParamsRules) GetNodeAffinityRequiredOk() (*ArbitraryRules, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *SpecificRunParamsRules) SetNodeAffinityRequired(v ArbitraryRules)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *SpecificRunParamsRules) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *SpecificRunParamsRules) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *SpecificRunParamsRules) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *SpecificRunParamsRules) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *SpecificRunParamsRules) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *SpecificRunParamsRules) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *SpecificRunParamsRules) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *SpecificRunParamsRules) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *SpecificRunParamsRules) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *SpecificRunParamsRules) GetNodeType() StringRules`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *SpecificRunParamsRules) GetNodeTypeOk() (*StringRules, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *SpecificRunParamsRules) SetNodeType(v StringRules)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *SpecificRunParamsRules) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *SpecificRunParamsRules) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *SpecificRunParamsRules) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetParallelism

`func (o *SpecificRunParamsRules) GetParallelism() IntegerRules`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *SpecificRunParamsRules) GetParallelismOk() (*IntegerRules, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *SpecificRunParamsRules) SetParallelism(v IntegerRules)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *SpecificRunParamsRules) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### SetParallelismNil

`func (o *SpecificRunParamsRules) SetParallelismNil(b bool)`

 SetParallelismNil sets the value for Parallelism to be an explicit nil

### UnsetParallelism
`func (o *SpecificRunParamsRules) UnsetParallelism()`

UnsetParallelism ensures that no value is present for Parallelism, not even an explicit nil
### GetPodAffinity

`func (o *SpecificRunParamsRules) GetPodAffinity() PodAffinityRules`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *SpecificRunParamsRules) GetPodAffinityOk() (*PodAffinityRules, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *SpecificRunParamsRules) SetPodAffinity(v PodAffinityRules)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *SpecificRunParamsRules) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *SpecificRunParamsRules) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *SpecificRunParamsRules) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPreemptibility

`func (o *SpecificRunParamsRules) GetPreemptibility() PreemptibilityRules`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *SpecificRunParamsRules) GetPreemptibilityOk() (*PreemptibilityRules, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *SpecificRunParamsRules) SetPreemptibility(v PreemptibilityRules)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *SpecificRunParamsRules) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *SpecificRunParamsRules) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *SpecificRunParamsRules) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *SpecificRunParamsRules) GetPriorityClass() StringRules`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *SpecificRunParamsRules) GetPriorityClassOk() (*StringRules, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *SpecificRunParamsRules) SetPriorityClass(v StringRules)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *SpecificRunParamsRules) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *SpecificRunParamsRules) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *SpecificRunParamsRules) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetRestartPolicy

`func (o *SpecificRunParamsRules) GetRestartPolicy() RestartPolicyRule`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *SpecificRunParamsRules) GetRestartPolicyOk() (*RestartPolicyRule, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *SpecificRunParamsRules) SetRestartPolicy(v RestartPolicyRule)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *SpecificRunParamsRules) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *SpecificRunParamsRules) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *SpecificRunParamsRules) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetRunAsGid

`func (o *SpecificRunParamsRules) GetRunAsGid() IntegerRulesOptional`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *SpecificRunParamsRules) GetRunAsGidOk() (*IntegerRulesOptional, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *SpecificRunParamsRules) SetRunAsGid(v IntegerRulesOptional)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *SpecificRunParamsRules) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *SpecificRunParamsRules) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *SpecificRunParamsRules) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetRunAsUid

`func (o *SpecificRunParamsRules) GetRunAsUid() IntegerRulesOptional`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *SpecificRunParamsRules) GetRunAsUidOk() (*IntegerRulesOptional, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *SpecificRunParamsRules) SetRunAsUid(v IntegerRulesOptional)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *SpecificRunParamsRules) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *SpecificRunParamsRules) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *SpecificRunParamsRules) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetSupplementalGroups

`func (o *SpecificRunParamsRules) GetSupplementalGroups() StringRules`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *SpecificRunParamsRules) GetSupplementalGroupsOk() (*StringRules, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *SpecificRunParamsRules) SetSupplementalGroups(v StringRules)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *SpecificRunParamsRules) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *SpecificRunParamsRules) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *SpecificRunParamsRules) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *SpecificRunParamsRules) GetTerminateAfterPreemption() BooleanRules`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *SpecificRunParamsRules) GetTerminateAfterPreemptionOk() (*BooleanRules, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *SpecificRunParamsRules) SetTerminateAfterPreemption(v BooleanRules)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *SpecificRunParamsRules) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *SpecificRunParamsRules) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *SpecificRunParamsRules) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *SpecificRunParamsRules) GetTerminationGracePeriodSeconds() IntegerRules`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *SpecificRunParamsRules) GetTerminationGracePeriodSecondsOk() (*IntegerRules, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *SpecificRunParamsRules) SetTerminationGracePeriodSeconds(v IntegerRules)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *SpecificRunParamsRules) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *SpecificRunParamsRules) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *SpecificRunParamsRules) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *SpecificRunParamsRules) GetTolerations() TolerationsRules`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *SpecificRunParamsRules) GetTolerationsOk() (*TolerationsRules, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *SpecificRunParamsRules) SetTolerations(v TolerationsRules)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *SpecificRunParamsRules) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *SpecificRunParamsRules) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *SpecificRunParamsRules) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


