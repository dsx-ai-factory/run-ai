# ReplicaDefaultsV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workspace. | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workspace executes. | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workspace. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**WorkingDir** | Pointer to **NullableString** | Container&#39;s working directory. If not specified, the container runtime default will be used. This may be configured in the container image. | [optional] 
**CreateHomeDir** | Pointer to **NullableBool** | When set to &#x60;true&#x60;, creates a home directory for the container. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**NodeType** | Pointer to **NullableString** | Nodes (machines), or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes. For more information, see [Group Nodes](https://docs.run.ai/latest/admin/researcher-setup/limit-to-node-group). When using this flag with with Project-based affinity, it refines the list of allowable node groups set in the Project. For more information, see [Projects](https://docs.run.ai/admin/admin-ui-setup/project-setup). | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workspace on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinity**](PodAffinity.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableEnvironmentVariablesDefaults**](EnvironmentVariablesDefaults.md) |  | [optional] 
**Annotations** | Pointer to [**NullableAnnotationsDefaults**](AnnotationsDefaults.md) |  | [optional] 
**Labels** | Pointer to [**NullableLabelsDefaults**](LabelsDefaults.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsDefaults**](TolerationsDefaults.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates if the job should be terminated by the system after it has been preempted. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Ports** | Pointer to [**NullablePortsDefaults**](PortsDefaults.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsDefaults**](ExposedUrlsDefaults.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsDefaults**](RelatedUrlsDefaults.md) |  | [optional] 
**Security** | Pointer to [**NullableSecurityFlatFields**](SecurityFlatFields.md) |  | [optional] 
**Compute** | Pointer to [**NullableComputeFieldsDefaults**](ComputeFieldsDefaults.md) |  | [optional] 
**Storage** | Pointer to [**NullableStorageFieldsDefaults**](StorageFieldsDefaults.md) |  | [optional] 

## Methods

### NewReplicaDefaultsV2

`func NewReplicaDefaultsV2() *ReplicaDefaultsV2`

NewReplicaDefaultsV2 instantiates a new ReplicaDefaultsV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReplicaDefaultsV2WithDefaults

`func NewReplicaDefaultsV2WithDefaults() *ReplicaDefaultsV2`

NewReplicaDefaultsV2WithDefaults instantiates a new ReplicaDefaultsV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *ReplicaDefaultsV2) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *ReplicaDefaultsV2) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *ReplicaDefaultsV2) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *ReplicaDefaultsV2) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *ReplicaDefaultsV2) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *ReplicaDefaultsV2) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *ReplicaDefaultsV2) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *ReplicaDefaultsV2) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *ReplicaDefaultsV2) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *ReplicaDefaultsV2) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *ReplicaDefaultsV2) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *ReplicaDefaultsV2) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetImage

`func (o *ReplicaDefaultsV2) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *ReplicaDefaultsV2) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *ReplicaDefaultsV2) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *ReplicaDefaultsV2) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *ReplicaDefaultsV2) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *ReplicaDefaultsV2) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *ReplicaDefaultsV2) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *ReplicaDefaultsV2) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *ReplicaDefaultsV2) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *ReplicaDefaultsV2) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *ReplicaDefaultsV2) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *ReplicaDefaultsV2) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *ReplicaDefaultsV2) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *ReplicaDefaultsV2) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *ReplicaDefaultsV2) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *ReplicaDefaultsV2) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *ReplicaDefaultsV2) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *ReplicaDefaultsV2) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *ReplicaDefaultsV2) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *ReplicaDefaultsV2) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *ReplicaDefaultsV2) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *ReplicaDefaultsV2) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *ReplicaDefaultsV2) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *ReplicaDefaultsV2) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *ReplicaDefaultsV2) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *ReplicaDefaultsV2) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *ReplicaDefaultsV2) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *ReplicaDefaultsV2) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *ReplicaDefaultsV2) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *ReplicaDefaultsV2) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetNodeType

`func (o *ReplicaDefaultsV2) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *ReplicaDefaultsV2) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *ReplicaDefaultsV2) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *ReplicaDefaultsV2) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *ReplicaDefaultsV2) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *ReplicaDefaultsV2) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *ReplicaDefaultsV2) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *ReplicaDefaultsV2) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *ReplicaDefaultsV2) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *ReplicaDefaultsV2) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *ReplicaDefaultsV2) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *ReplicaDefaultsV2) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *ReplicaDefaultsV2) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *ReplicaDefaultsV2) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *ReplicaDefaultsV2) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *ReplicaDefaultsV2) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *ReplicaDefaultsV2) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *ReplicaDefaultsV2) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetEnvironmentVariables

`func (o *ReplicaDefaultsV2) GetEnvironmentVariables() EnvironmentVariablesDefaults`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *ReplicaDefaultsV2) GetEnvironmentVariablesOk() (*EnvironmentVariablesDefaults, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *ReplicaDefaultsV2) SetEnvironmentVariables(v EnvironmentVariablesDefaults)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *ReplicaDefaultsV2) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *ReplicaDefaultsV2) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *ReplicaDefaultsV2) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetAnnotations

`func (o *ReplicaDefaultsV2) GetAnnotations() AnnotationsDefaults`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *ReplicaDefaultsV2) GetAnnotationsOk() (*AnnotationsDefaults, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *ReplicaDefaultsV2) SetAnnotations(v AnnotationsDefaults)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *ReplicaDefaultsV2) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *ReplicaDefaultsV2) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *ReplicaDefaultsV2) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *ReplicaDefaultsV2) GetLabels() LabelsDefaults`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ReplicaDefaultsV2) GetLabelsOk() (*LabelsDefaults, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ReplicaDefaultsV2) SetLabels(v LabelsDefaults)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ReplicaDefaultsV2) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *ReplicaDefaultsV2) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *ReplicaDefaultsV2) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetTolerations

`func (o *ReplicaDefaultsV2) GetTolerations() TolerationsDefaults`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *ReplicaDefaultsV2) GetTolerationsOk() (*TolerationsDefaults, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *ReplicaDefaultsV2) SetTolerations(v TolerationsDefaults)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *ReplicaDefaultsV2) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *ReplicaDefaultsV2) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *ReplicaDefaultsV2) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *ReplicaDefaultsV2) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *ReplicaDefaultsV2) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *ReplicaDefaultsV2) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *ReplicaDefaultsV2) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *ReplicaDefaultsV2) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *ReplicaDefaultsV2) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *ReplicaDefaultsV2) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *ReplicaDefaultsV2) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *ReplicaDefaultsV2) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *ReplicaDefaultsV2) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *ReplicaDefaultsV2) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *ReplicaDefaultsV2) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *ReplicaDefaultsV2) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *ReplicaDefaultsV2) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *ReplicaDefaultsV2) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *ReplicaDefaultsV2) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *ReplicaDefaultsV2) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *ReplicaDefaultsV2) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetPorts

`func (o *ReplicaDefaultsV2) GetPorts() PortsDefaults`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *ReplicaDefaultsV2) GetPortsOk() (*PortsDefaults, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *ReplicaDefaultsV2) SetPorts(v PortsDefaults)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *ReplicaDefaultsV2) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *ReplicaDefaultsV2) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *ReplicaDefaultsV2) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *ReplicaDefaultsV2) GetExposedUrls() ExposedUrlsDefaults`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *ReplicaDefaultsV2) GetExposedUrlsOk() (*ExposedUrlsDefaults, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *ReplicaDefaultsV2) SetExposedUrls(v ExposedUrlsDefaults)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *ReplicaDefaultsV2) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *ReplicaDefaultsV2) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *ReplicaDefaultsV2) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *ReplicaDefaultsV2) GetRelatedUrls() RelatedUrlsDefaults`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *ReplicaDefaultsV2) GetRelatedUrlsOk() (*RelatedUrlsDefaults, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *ReplicaDefaultsV2) SetRelatedUrls(v RelatedUrlsDefaults)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *ReplicaDefaultsV2) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *ReplicaDefaultsV2) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *ReplicaDefaultsV2) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetSecurity

`func (o *ReplicaDefaultsV2) GetSecurity() SecurityFlatFields`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *ReplicaDefaultsV2) GetSecurityOk() (*SecurityFlatFields, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *ReplicaDefaultsV2) SetSecurity(v SecurityFlatFields)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *ReplicaDefaultsV2) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *ReplicaDefaultsV2) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *ReplicaDefaultsV2) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetCompute

`func (o *ReplicaDefaultsV2) GetCompute() ComputeFieldsDefaults`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *ReplicaDefaultsV2) GetComputeOk() (*ComputeFieldsDefaults, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *ReplicaDefaultsV2) SetCompute(v ComputeFieldsDefaults)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *ReplicaDefaultsV2) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *ReplicaDefaultsV2) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *ReplicaDefaultsV2) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetStorage

`func (o *ReplicaDefaultsV2) GetStorage() StorageFieldsDefaults`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *ReplicaDefaultsV2) GetStorageOk() (*StorageFieldsDefaults, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *ReplicaDefaultsV2) SetStorage(v StorageFieldsDefaults)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *ReplicaDefaultsV2) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *ReplicaDefaultsV2) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *ReplicaDefaultsV2) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


