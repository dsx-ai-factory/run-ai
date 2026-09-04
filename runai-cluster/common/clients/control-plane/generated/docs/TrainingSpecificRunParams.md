# TrainingSpecificRunParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workspace. | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workspace executes. | [optional] 
**RunAsUid** | Pointer to **NullableInt64** | The user id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsUid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**RunAsGid** | Pointer to **NullableInt64** | The group id to run the entrypoint of the container which executes the workspace. Default to the value specified in the environment asset &#x60;runAsGid&#x60; field (optional). Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. | [optional] 
**SupplementalGroups** | Pointer to **NullableString** | Comma separated list of groups that the user running the container belongs to, in addition to the group indicated by runAsGid. Use only when the source uid/gid of the environment asset is not &#x60;fromTheImage&#x60;, and &#x60;overrideUidGidInWorkspace&#x60; is enabled. Using an empty string implies reverting the supplementary groups of the image. | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workspace. | [optional] 
**NodeType** | Pointer to **NullableString** | Nodes (machines), or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes. For more information, see [Group Nodes](https://docs.run.ai/latest/admin/researcher-setup/limit-to-node-group). When using this flag with with Project-based affinity, it refines the list of allowable node groups set in the Project. For more information, see [Projects](https://docs.run.ai/admin/admin-ui-setup/project-setup). | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workspace on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinity**](PodAffinity.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates if the job should be terminated by the system after it has been preempted. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workspace. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workspace. | [optional] 
**Connections** | Pointer to [**[]SpecificRunConnectionInfo**](SpecificRunConnectionInfo.md) |  | [optional] 
**AllowOverQuota** | Pointer to **NullableBool** | Whether to allow the workload to exceed the quota of the project. | [optional] 
**NumWorkers** | Pointer to **int32** |  | [optional] 
**NoMaster** | Pointer to **NullableBool** |  | [optional] 
**DistFramework** | Pointer to [**NullableDistributedFramework**](DistributedFramework.md) |  | [optional] 
**Master** | Pointer to [**NullableInfoDistMaster**](InfoDistMaster.md) |  | [optional] 

## Methods

### NewTrainingSpecificRunParams

`func NewTrainingSpecificRunParams() *TrainingSpecificRunParams`

NewTrainingSpecificRunParams instantiates a new TrainingSpecificRunParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingSpecificRunParamsWithDefaults

`func NewTrainingSpecificRunParamsWithDefaults() *TrainingSpecificRunParams`

NewTrainingSpecificRunParamsWithDefaults instantiates a new TrainingSpecificRunParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *TrainingSpecificRunParams) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *TrainingSpecificRunParams) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *TrainingSpecificRunParams) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *TrainingSpecificRunParams) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *TrainingSpecificRunParams) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *TrainingSpecificRunParams) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *TrainingSpecificRunParams) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *TrainingSpecificRunParams) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *TrainingSpecificRunParams) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *TrainingSpecificRunParams) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *TrainingSpecificRunParams) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *TrainingSpecificRunParams) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetRunAsUid

`func (o *TrainingSpecificRunParams) GetRunAsUid() int64`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *TrainingSpecificRunParams) GetRunAsUidOk() (*int64, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *TrainingSpecificRunParams) SetRunAsUid(v int64)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *TrainingSpecificRunParams) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *TrainingSpecificRunParams) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *TrainingSpecificRunParams) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *TrainingSpecificRunParams) GetRunAsGid() int64`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *TrainingSpecificRunParams) GetRunAsGidOk() (*int64, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *TrainingSpecificRunParams) SetRunAsGid(v int64)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *TrainingSpecificRunParams) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *TrainingSpecificRunParams) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *TrainingSpecificRunParams) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *TrainingSpecificRunParams) GetSupplementalGroups() string`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *TrainingSpecificRunParams) GetSupplementalGroupsOk() (*string, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *TrainingSpecificRunParams) SetSupplementalGroups(v string)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *TrainingSpecificRunParams) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *TrainingSpecificRunParams) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *TrainingSpecificRunParams) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetEnvironmentVariables

`func (o *TrainingSpecificRunParams) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *TrainingSpecificRunParams) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *TrainingSpecificRunParams) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *TrainingSpecificRunParams) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *TrainingSpecificRunParams) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *TrainingSpecificRunParams) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetNodeType

`func (o *TrainingSpecificRunParams) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *TrainingSpecificRunParams) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *TrainingSpecificRunParams) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *TrainingSpecificRunParams) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *TrainingSpecificRunParams) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *TrainingSpecificRunParams) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *TrainingSpecificRunParams) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *TrainingSpecificRunParams) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *TrainingSpecificRunParams) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *TrainingSpecificRunParams) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *TrainingSpecificRunParams) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *TrainingSpecificRunParams) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *TrainingSpecificRunParams) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *TrainingSpecificRunParams) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *TrainingSpecificRunParams) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *TrainingSpecificRunParams) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *TrainingSpecificRunParams) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *TrainingSpecificRunParams) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *TrainingSpecificRunParams) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *TrainingSpecificRunParams) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *TrainingSpecificRunParams) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *TrainingSpecificRunParams) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *TrainingSpecificRunParams) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *TrainingSpecificRunParams) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingSpecificRunParams) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *TrainingSpecificRunParams) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingSpecificRunParams) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingSpecificRunParams) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *TrainingSpecificRunParams) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *TrainingSpecificRunParams) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *TrainingSpecificRunParams) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *TrainingSpecificRunParams) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *TrainingSpecificRunParams) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *TrainingSpecificRunParams) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *TrainingSpecificRunParams) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *TrainingSpecificRunParams) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetAnnotations

`func (o *TrainingSpecificRunParams) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *TrainingSpecificRunParams) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *TrainingSpecificRunParams) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *TrainingSpecificRunParams) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *TrainingSpecificRunParams) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *TrainingSpecificRunParams) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *TrainingSpecificRunParams) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *TrainingSpecificRunParams) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *TrainingSpecificRunParams) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *TrainingSpecificRunParams) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *TrainingSpecificRunParams) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *TrainingSpecificRunParams) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetConnections

`func (o *TrainingSpecificRunParams) GetConnections() []SpecificRunConnectionInfo`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *TrainingSpecificRunParams) GetConnectionsOk() (*[]SpecificRunConnectionInfo, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *TrainingSpecificRunParams) SetConnections(v []SpecificRunConnectionInfo)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *TrainingSpecificRunParams) HasConnections() bool`

HasConnections returns a boolean if a field has been set.

### SetConnectionsNil

`func (o *TrainingSpecificRunParams) SetConnectionsNil(b bool)`

 SetConnectionsNil sets the value for Connections to be an explicit nil

### UnsetConnections
`func (o *TrainingSpecificRunParams) UnsetConnections()`

UnsetConnections ensures that no value is present for Connections, not even an explicit nil
### GetAllowOverQuota

`func (o *TrainingSpecificRunParams) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *TrainingSpecificRunParams) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *TrainingSpecificRunParams) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *TrainingSpecificRunParams) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### SetAllowOverQuotaNil

`func (o *TrainingSpecificRunParams) SetAllowOverQuotaNil(b bool)`

 SetAllowOverQuotaNil sets the value for AllowOverQuota to be an explicit nil

### UnsetAllowOverQuota
`func (o *TrainingSpecificRunParams) UnsetAllowOverQuota()`

UnsetAllowOverQuota ensures that no value is present for AllowOverQuota, not even an explicit nil
### GetNumWorkers

`func (o *TrainingSpecificRunParams) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *TrainingSpecificRunParams) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *TrainingSpecificRunParams) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *TrainingSpecificRunParams) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### GetNoMaster

`func (o *TrainingSpecificRunParams) GetNoMaster() bool`

GetNoMaster returns the NoMaster field if non-nil, zero value otherwise.

### GetNoMasterOk

`func (o *TrainingSpecificRunParams) GetNoMasterOk() (*bool, bool)`

GetNoMasterOk returns a tuple with the NoMaster field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNoMaster

`func (o *TrainingSpecificRunParams) SetNoMaster(v bool)`

SetNoMaster sets NoMaster field to given value.

### HasNoMaster

`func (o *TrainingSpecificRunParams) HasNoMaster() bool`

HasNoMaster returns a boolean if a field has been set.

### SetNoMasterNil

`func (o *TrainingSpecificRunParams) SetNoMasterNil(b bool)`

 SetNoMasterNil sets the value for NoMaster to be an explicit nil

### UnsetNoMaster
`func (o *TrainingSpecificRunParams) UnsetNoMaster()`

UnsetNoMaster ensures that no value is present for NoMaster, not even an explicit nil
### GetDistFramework

`func (o *TrainingSpecificRunParams) GetDistFramework() DistributedFramework`

GetDistFramework returns the DistFramework field if non-nil, zero value otherwise.

### GetDistFrameworkOk

`func (o *TrainingSpecificRunParams) GetDistFrameworkOk() (*DistributedFramework, bool)`

GetDistFrameworkOk returns a tuple with the DistFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistFramework

`func (o *TrainingSpecificRunParams) SetDistFramework(v DistributedFramework)`

SetDistFramework sets DistFramework field to given value.

### HasDistFramework

`func (o *TrainingSpecificRunParams) HasDistFramework() bool`

HasDistFramework returns a boolean if a field has been set.

### SetDistFrameworkNil

`func (o *TrainingSpecificRunParams) SetDistFrameworkNil(b bool)`

 SetDistFrameworkNil sets the value for DistFramework to be an explicit nil

### UnsetDistFramework
`func (o *TrainingSpecificRunParams) UnsetDistFramework()`

UnsetDistFramework ensures that no value is present for DistFramework, not even an explicit nil
### GetMaster

`func (o *TrainingSpecificRunParams) GetMaster() InfoDistMaster`

GetMaster returns the Master field if non-nil, zero value otherwise.

### GetMasterOk

`func (o *TrainingSpecificRunParams) GetMasterOk() (*InfoDistMaster, bool)`

GetMasterOk returns a tuple with the Master field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaster

`func (o *TrainingSpecificRunParams) SetMaster(v InfoDistMaster)`

SetMaster sets Master field to given value.

### HasMaster

`func (o *TrainingSpecificRunParams) HasMaster() bool`

HasMaster returns a boolean if a field has been set.

### SetMasterNil

`func (o *TrainingSpecificRunParams) SetMasterNil(b bool)`

 SetMasterNil sets the value for Master to be an explicit nil

### UnsetMaster
`func (o *TrainingSpecificRunParams) UnsetMaster()`

UnsetMaster ensures that no value is present for Master, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


