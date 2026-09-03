# SpecificRunParamsCommon

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

## Methods

### NewSpecificRunParamsCommon

`func NewSpecificRunParamsCommon() *SpecificRunParamsCommon`

NewSpecificRunParamsCommon instantiates a new SpecificRunParamsCommon object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunParamsCommonWithDefaults

`func NewSpecificRunParamsCommonWithDefaults() *SpecificRunParamsCommon`

NewSpecificRunParamsCommonWithDefaults instantiates a new SpecificRunParamsCommon object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *SpecificRunParamsCommon) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *SpecificRunParamsCommon) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *SpecificRunParamsCommon) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *SpecificRunParamsCommon) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *SpecificRunParamsCommon) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *SpecificRunParamsCommon) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *SpecificRunParamsCommon) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *SpecificRunParamsCommon) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *SpecificRunParamsCommon) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *SpecificRunParamsCommon) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *SpecificRunParamsCommon) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *SpecificRunParamsCommon) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetRunAsUid

`func (o *SpecificRunParamsCommon) GetRunAsUid() int64`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *SpecificRunParamsCommon) GetRunAsUidOk() (*int64, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *SpecificRunParamsCommon) SetRunAsUid(v int64)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *SpecificRunParamsCommon) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *SpecificRunParamsCommon) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *SpecificRunParamsCommon) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *SpecificRunParamsCommon) GetRunAsGid() int64`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *SpecificRunParamsCommon) GetRunAsGidOk() (*int64, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *SpecificRunParamsCommon) SetRunAsGid(v int64)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *SpecificRunParamsCommon) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *SpecificRunParamsCommon) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *SpecificRunParamsCommon) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *SpecificRunParamsCommon) GetSupplementalGroups() string`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *SpecificRunParamsCommon) GetSupplementalGroupsOk() (*string, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *SpecificRunParamsCommon) SetSupplementalGroups(v string)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *SpecificRunParamsCommon) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *SpecificRunParamsCommon) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *SpecificRunParamsCommon) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetEnvironmentVariables

`func (o *SpecificRunParamsCommon) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *SpecificRunParamsCommon) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *SpecificRunParamsCommon) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *SpecificRunParamsCommon) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *SpecificRunParamsCommon) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *SpecificRunParamsCommon) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetNodeType

`func (o *SpecificRunParamsCommon) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *SpecificRunParamsCommon) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *SpecificRunParamsCommon) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *SpecificRunParamsCommon) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *SpecificRunParamsCommon) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *SpecificRunParamsCommon) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *SpecificRunParamsCommon) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *SpecificRunParamsCommon) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *SpecificRunParamsCommon) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *SpecificRunParamsCommon) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *SpecificRunParamsCommon) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *SpecificRunParamsCommon) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *SpecificRunParamsCommon) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *SpecificRunParamsCommon) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *SpecificRunParamsCommon) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *SpecificRunParamsCommon) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *SpecificRunParamsCommon) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *SpecificRunParamsCommon) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *SpecificRunParamsCommon) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *SpecificRunParamsCommon) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *SpecificRunParamsCommon) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *SpecificRunParamsCommon) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *SpecificRunParamsCommon) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *SpecificRunParamsCommon) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsCommon) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *SpecificRunParamsCommon) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsCommon) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsCommon) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *SpecificRunParamsCommon) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *SpecificRunParamsCommon) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *SpecificRunParamsCommon) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *SpecificRunParamsCommon) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *SpecificRunParamsCommon) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *SpecificRunParamsCommon) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *SpecificRunParamsCommon) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *SpecificRunParamsCommon) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetAnnotations

`func (o *SpecificRunParamsCommon) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *SpecificRunParamsCommon) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *SpecificRunParamsCommon) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *SpecificRunParamsCommon) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *SpecificRunParamsCommon) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *SpecificRunParamsCommon) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *SpecificRunParamsCommon) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *SpecificRunParamsCommon) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *SpecificRunParamsCommon) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *SpecificRunParamsCommon) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *SpecificRunParamsCommon) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *SpecificRunParamsCommon) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetConnections

`func (o *SpecificRunParamsCommon) GetConnections() []SpecificRunConnectionInfo`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *SpecificRunParamsCommon) GetConnectionsOk() (*[]SpecificRunConnectionInfo, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *SpecificRunParamsCommon) SetConnections(v []SpecificRunConnectionInfo)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *SpecificRunParamsCommon) HasConnections() bool`

HasConnections returns a boolean if a field has been set.

### SetConnectionsNil

`func (o *SpecificRunParamsCommon) SetConnectionsNil(b bool)`

 SetConnectionsNil sets the value for Connections to be an explicit nil

### UnsetConnections
`func (o *SpecificRunParamsCommon) UnsetConnections()`

UnsetConnections ensures that no value is present for Connections, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


