# SpecificRunParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **NullableString** | A command to server as the entry point of the container running the workspace. | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workspace executes. | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workspace. | [optional] 
**RunAsUid** | Pointer to **NullableInt64** | The user id to run the entrypoint of the container which executes the workspace. Default to the (optional) value specified in the environment asset runAsUid field. Can be provided only when the source uid/gid of the environment asset is not fromTheImage, and overrideUidGidInWorkspace is enabled. | [optional] 
**RunAsGid** | Pointer to **NullableInt64** | The group id to run the entrypoint of the container which executes the workspace. Default to the (optional) value specified in the environment asset runAsGid field. Can be provided only when the source uid/gid of the environment asset is not fromTheImage, and overrideUidGidInWorkspace is enabled. | [optional] 
**SupplementalGroups** | Pointer to **NullableString** | Comma seperated list of groups that the user running the container belongs to, in addition to the group indicated by runAsGid. Can be provided only when the source uid/gid of the environment asset is not fromTheImage, and overrideUidGidInWorkspace is enabled. Empty string implies reverting to the supplementary groups of the image. | [optional] 
**NodeType** | Pointer to **NullableString** | Nodes (machines) or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes as explained in the Group Nodes guide at https://docs.run.ai/admin/researcher-setup/limit-to-node-group. This flag can be used in conjunction with Project-based affinity. In this case, the flag is used to refine the list of allowable node groups set in the Project. For more information consult the Projects guide at https://docs.run.ai/admin/admin-ui-setup/project-setup. | [optional] 
**NodePools** | Pointer to **[]string** | A prioritize list of node pools for the scheduler to run the workspace on. The scheduler will always try to use the first node pool before moving to the next one when the fist is not available. | [optional] 
**AllowOverQuota** | Pointer to **NullableBool** | Whether to allow the workspace to exceed the quota of the project. | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workspace. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workspace. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (Completed or Failed) will be automatically deleted. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed. | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates whether the job should be terminated, by the system, after it has been preempted. | [optional] 
**Connections** | Pointer to [**[]SpecificRunConnectionInfo**](SpecificRunConnectionInfo.md) |  | [optional] 
**AutoScaling** | Pointer to [**NullableSpecificRunAutoScalingAutoScaling**](SpecificRunAutoScalingAutoScaling.md) |  | [optional] 

## Methods

### NewSpecificRunParams

`func NewSpecificRunParams() *SpecificRunParams`

NewSpecificRunParams instantiates a new SpecificRunParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunParamsWithDefaults

`func NewSpecificRunParamsWithDefaults() *SpecificRunParams`

NewSpecificRunParamsWithDefaults instantiates a new SpecificRunParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *SpecificRunParams) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *SpecificRunParams) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *SpecificRunParams) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *SpecificRunParams) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *SpecificRunParams) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *SpecificRunParams) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *SpecificRunParams) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *SpecificRunParams) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *SpecificRunParams) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *SpecificRunParams) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *SpecificRunParams) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *SpecificRunParams) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetEnvironmentVariables

`func (o *SpecificRunParams) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *SpecificRunParams) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *SpecificRunParams) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *SpecificRunParams) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *SpecificRunParams) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *SpecificRunParams) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetRunAsUid

`func (o *SpecificRunParams) GetRunAsUid() int64`

GetRunAsUid returns the RunAsUid field if non-nil, zero value otherwise.

### GetRunAsUidOk

`func (o *SpecificRunParams) GetRunAsUidOk() (*int64, bool)`

GetRunAsUidOk returns a tuple with the RunAsUid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsUid

`func (o *SpecificRunParams) SetRunAsUid(v int64)`

SetRunAsUid sets RunAsUid field to given value.

### HasRunAsUid

`func (o *SpecificRunParams) HasRunAsUid() bool`

HasRunAsUid returns a boolean if a field has been set.

### SetRunAsUidNil

`func (o *SpecificRunParams) SetRunAsUidNil(b bool)`

 SetRunAsUidNil sets the value for RunAsUid to be an explicit nil

### UnsetRunAsUid
`func (o *SpecificRunParams) UnsetRunAsUid()`

UnsetRunAsUid ensures that no value is present for RunAsUid, not even an explicit nil
### GetRunAsGid

`func (o *SpecificRunParams) GetRunAsGid() int64`

GetRunAsGid returns the RunAsGid field if non-nil, zero value otherwise.

### GetRunAsGidOk

`func (o *SpecificRunParams) GetRunAsGidOk() (*int64, bool)`

GetRunAsGidOk returns a tuple with the RunAsGid field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunAsGid

`func (o *SpecificRunParams) SetRunAsGid(v int64)`

SetRunAsGid sets RunAsGid field to given value.

### HasRunAsGid

`func (o *SpecificRunParams) HasRunAsGid() bool`

HasRunAsGid returns a boolean if a field has been set.

### SetRunAsGidNil

`func (o *SpecificRunParams) SetRunAsGidNil(b bool)`

 SetRunAsGidNil sets the value for RunAsGid to be an explicit nil

### UnsetRunAsGid
`func (o *SpecificRunParams) UnsetRunAsGid()`

UnsetRunAsGid ensures that no value is present for RunAsGid, not even an explicit nil
### GetSupplementalGroups

`func (o *SpecificRunParams) GetSupplementalGroups() string`

GetSupplementalGroups returns the SupplementalGroups field if non-nil, zero value otherwise.

### GetSupplementalGroupsOk

`func (o *SpecificRunParams) GetSupplementalGroupsOk() (*string, bool)`

GetSupplementalGroupsOk returns a tuple with the SupplementalGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupplementalGroups

`func (o *SpecificRunParams) SetSupplementalGroups(v string)`

SetSupplementalGroups sets SupplementalGroups field to given value.

### HasSupplementalGroups

`func (o *SpecificRunParams) HasSupplementalGroups() bool`

HasSupplementalGroups returns a boolean if a field has been set.

### SetSupplementalGroupsNil

`func (o *SpecificRunParams) SetSupplementalGroupsNil(b bool)`

 SetSupplementalGroupsNil sets the value for SupplementalGroups to be an explicit nil

### UnsetSupplementalGroups
`func (o *SpecificRunParams) UnsetSupplementalGroups()`

UnsetSupplementalGroups ensures that no value is present for SupplementalGroups, not even an explicit nil
### GetNodeType

`func (o *SpecificRunParams) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *SpecificRunParams) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *SpecificRunParams) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *SpecificRunParams) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *SpecificRunParams) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *SpecificRunParams) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *SpecificRunParams) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *SpecificRunParams) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *SpecificRunParams) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *SpecificRunParams) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *SpecificRunParams) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *SpecificRunParams) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetAllowOverQuota

`func (o *SpecificRunParams) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *SpecificRunParams) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *SpecificRunParams) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *SpecificRunParams) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### SetAllowOverQuotaNil

`func (o *SpecificRunParams) SetAllowOverQuotaNil(b bool)`

 SetAllowOverQuotaNil sets the value for AllowOverQuota to be an explicit nil

### UnsetAllowOverQuota
`func (o *SpecificRunParams) UnsetAllowOverQuota()`

UnsetAllowOverQuota ensures that no value is present for AllowOverQuota, not even an explicit nil
### GetAnnotations

`func (o *SpecificRunParams) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *SpecificRunParams) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *SpecificRunParams) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *SpecificRunParams) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetLabels

`func (o *SpecificRunParams) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *SpecificRunParams) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *SpecificRunParams) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *SpecificRunParams) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParams) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *SpecificRunParams) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParams) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParams) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *SpecificRunParams) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *SpecificRunParams) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *SpecificRunParams) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *SpecificRunParams) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *SpecificRunParams) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *SpecificRunParams) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *SpecificRunParams) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *SpecificRunParams) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *SpecificRunParams) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *SpecificRunParams) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *SpecificRunParams) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *SpecificRunParams) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *SpecificRunParams) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *SpecificRunParams) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetConnections

`func (o *SpecificRunParams) GetConnections() []SpecificRunConnectionInfo`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *SpecificRunParams) GetConnectionsOk() (*[]SpecificRunConnectionInfo, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *SpecificRunParams) SetConnections(v []SpecificRunConnectionInfo)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *SpecificRunParams) HasConnections() bool`

HasConnections returns a boolean if a field has been set.

### SetConnectionsNil

`func (o *SpecificRunParams) SetConnectionsNil(b bool)`

 SetConnectionsNil sets the value for Connections to be an explicit nil

### UnsetConnections
`func (o *SpecificRunParams) UnsetConnections()`

UnsetConnections ensures that no value is present for Connections, not even an explicit nil
### GetAutoScaling

`func (o *SpecificRunParams) GetAutoScaling() SpecificRunAutoScalingAutoScaling`

GetAutoScaling returns the AutoScaling field if non-nil, zero value otherwise.

### GetAutoScalingOk

`func (o *SpecificRunParams) GetAutoScalingOk() (*SpecificRunAutoScalingAutoScaling, bool)`

GetAutoScalingOk returns a tuple with the AutoScaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoScaling

`func (o *SpecificRunParams) SetAutoScaling(v SpecificRunAutoScalingAutoScaling)`

SetAutoScaling sets AutoScaling field to given value.

### HasAutoScaling

`func (o *SpecificRunParams) HasAutoScaling() bool`

HasAutoScaling returns a boolean if a field has been set.

### SetAutoScalingNil

`func (o *SpecificRunParams) SetAutoScalingNil(b bool)`

 SetAutoScalingNil sets the value for AutoScaling to be an explicit nil

### UnsetAutoScaling
`func (o *SpecificRunParams) UnsetAutoScaling()`

UnsetAutoScaling ensures that no value is present for AutoScaling, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


