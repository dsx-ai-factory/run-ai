# MasterTemplateSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workload. | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workload executes. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. Log retention is managed separately. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workload. | [optional] 
**Compute** | Pointer to [**NullableWorkspaceSpecSpecAllOfCompute**](WorkspaceSpecSpecAllOfCompute.md) |  | [optional] 
**CreateHomeDir** | Pointer to **NullableBool** | When set to &#x60;true&#x60;, creates a home directory for the container. | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workload. | [optional] 
**ExposedUrls** | Pointer to [**[]ExposedUrl**](ExposedUrl.md) | Set of container ports that the workload exposes via URLs. | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workload. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workload. | [optional] 
**NodeAffinityRequired** | Pointer to [**NullableNodeAffinityRequired**](NodeAffinityRequired.md) |  | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**NodeType** | Pointer to **NullableString** | Deprecated: use a workload policy node selector to target node types instead. Nodes (machines), or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes. For more information, see [Group Nodes](https://docs.run.ai/latest/admin/researcher-setup/limit-to-node-group). When using this flag with with Project-based affinity, it refines the list of allowable node groups set in the Project. For more information, see [Projects](https://docshub.run.ai/guides/platform-management/aiinitiatives/organization/projects). | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinity**](PodAffinity.md) |  | [optional] 
**Ports** | Pointer to [**[]Port**](Port.md) | Set of container ports that the workload exposes. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**RelatedUrls** | Pointer to [**[]RelatedUrl**](RelatedUrl.md) | Set of URLs that are related to the workload. | [optional] 
**RestartPolicy** | Pointer to [**NullableRestartPolicy**](RestartPolicy.md) |  | [optional] 
**Security** | Pointer to [**NullableWorkspaceSpecSpecAllOfSecurity**](WorkspaceSpecSpecAllOfSecurity.md) |  | [optional] 
**Stdin** | Pointer to **NullableBool** | Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF | [optional] 
**Storage** | Pointer to [**NullableWorkspaceSpecSpecAllOfStorage**](WorkspaceSpecSpecAllOfStorage.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates if the job should be terminated by the system after it has been preempted. | [optional] 
**TerminationGracePeriodSeconds** | Pointer to **NullableInt64** | Duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). | [optional] 
**Tolerations** | Pointer to [**[]Toleration**](Toleration.md) | Set of tolerations to apply to the workload. | [optional] 
**Tty** | Pointer to **NullableBool** | Whether this container should allocate a TTY for itself, also requires &#39;stdin&#39; to be true. | [optional] 
**WorkingDir** | Pointer to **NullableString** | Container&#39;s working directory. If not specified, the container runtime default will be used. This may be configured in the container image. | [optional] 

## Methods

### NewMasterTemplateSpec

`func NewMasterTemplateSpec() *MasterTemplateSpec`

NewMasterTemplateSpec instantiates a new MasterTemplateSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMasterTemplateSpecWithDefaults

`func NewMasterTemplateSpecWithDefaults() *MasterTemplateSpec`

NewMasterTemplateSpecWithDefaults instantiates a new MasterTemplateSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *MasterTemplateSpec) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *MasterTemplateSpec) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *MasterTemplateSpec) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *MasterTemplateSpec) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *MasterTemplateSpec) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *MasterTemplateSpec) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *MasterTemplateSpec) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *MasterTemplateSpec) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *MasterTemplateSpec) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *MasterTemplateSpec) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *MasterTemplateSpec) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *MasterTemplateSpec) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *MasterTemplateSpec) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *MasterTemplateSpec) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *MasterTemplateSpec) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *MasterTemplateSpec) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *MasterTemplateSpec) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *MasterTemplateSpec) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *MasterTemplateSpec) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *MasterTemplateSpec) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *MasterTemplateSpec) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *MasterTemplateSpec) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *MasterTemplateSpec) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *MasterTemplateSpec) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCategory

`func (o *MasterTemplateSpec) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *MasterTemplateSpec) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *MasterTemplateSpec) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *MasterTemplateSpec) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *MasterTemplateSpec) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *MasterTemplateSpec) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCommand

`func (o *MasterTemplateSpec) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *MasterTemplateSpec) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *MasterTemplateSpec) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *MasterTemplateSpec) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *MasterTemplateSpec) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *MasterTemplateSpec) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompute

`func (o *MasterTemplateSpec) GetCompute() WorkspaceSpecSpecAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *MasterTemplateSpec) GetComputeOk() (*WorkspaceSpecSpecAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *MasterTemplateSpec) SetCompute(v WorkspaceSpecSpecAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *MasterTemplateSpec) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *MasterTemplateSpec) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *MasterTemplateSpec) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *MasterTemplateSpec) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *MasterTemplateSpec) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *MasterTemplateSpec) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *MasterTemplateSpec) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *MasterTemplateSpec) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *MasterTemplateSpec) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetEnvironmentVariables

`func (o *MasterTemplateSpec) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *MasterTemplateSpec) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *MasterTemplateSpec) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *MasterTemplateSpec) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *MasterTemplateSpec) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *MasterTemplateSpec) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *MasterTemplateSpec) GetExposedUrls() []ExposedUrl`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *MasterTemplateSpec) GetExposedUrlsOk() (*[]ExposedUrl, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *MasterTemplateSpec) SetExposedUrls(v []ExposedUrl)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *MasterTemplateSpec) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *MasterTemplateSpec) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *MasterTemplateSpec) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *MasterTemplateSpec) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *MasterTemplateSpec) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *MasterTemplateSpec) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *MasterTemplateSpec) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *MasterTemplateSpec) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *MasterTemplateSpec) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *MasterTemplateSpec) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *MasterTemplateSpec) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *MasterTemplateSpec) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *MasterTemplateSpec) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *MasterTemplateSpec) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *MasterTemplateSpec) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetLabels

`func (o *MasterTemplateSpec) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *MasterTemplateSpec) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *MasterTemplateSpec) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *MasterTemplateSpec) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *MasterTemplateSpec) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *MasterTemplateSpec) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetNodeAffinityRequired

`func (o *MasterTemplateSpec) GetNodeAffinityRequired() NodeAffinityRequired`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *MasterTemplateSpec) GetNodeAffinityRequiredOk() (*NodeAffinityRequired, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *MasterTemplateSpec) SetNodeAffinityRequired(v NodeAffinityRequired)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *MasterTemplateSpec) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *MasterTemplateSpec) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *MasterTemplateSpec) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *MasterTemplateSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *MasterTemplateSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *MasterTemplateSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *MasterTemplateSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *MasterTemplateSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *MasterTemplateSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *MasterTemplateSpec) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *MasterTemplateSpec) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *MasterTemplateSpec) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *MasterTemplateSpec) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *MasterTemplateSpec) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *MasterTemplateSpec) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetPodAffinity

`func (o *MasterTemplateSpec) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *MasterTemplateSpec) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *MasterTemplateSpec) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *MasterTemplateSpec) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *MasterTemplateSpec) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *MasterTemplateSpec) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *MasterTemplateSpec) GetPorts() []Port`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *MasterTemplateSpec) GetPortsOk() (*[]Port, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *MasterTemplateSpec) SetPorts(v []Port)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *MasterTemplateSpec) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *MasterTemplateSpec) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *MasterTemplateSpec) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetPreemptibility

`func (o *MasterTemplateSpec) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *MasterTemplateSpec) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *MasterTemplateSpec) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *MasterTemplateSpec) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *MasterTemplateSpec) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *MasterTemplateSpec) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *MasterTemplateSpec) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *MasterTemplateSpec) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *MasterTemplateSpec) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *MasterTemplateSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *MasterTemplateSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *MasterTemplateSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *MasterTemplateSpec) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *MasterTemplateSpec) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *MasterTemplateSpec) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *MasterTemplateSpec) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *MasterTemplateSpec) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *MasterTemplateSpec) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *MasterTemplateSpec) GetRelatedUrls() []RelatedUrl`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *MasterTemplateSpec) GetRelatedUrlsOk() (*[]RelatedUrl, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *MasterTemplateSpec) SetRelatedUrls(v []RelatedUrl)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *MasterTemplateSpec) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *MasterTemplateSpec) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *MasterTemplateSpec) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetRestartPolicy

`func (o *MasterTemplateSpec) GetRestartPolicy() RestartPolicy`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *MasterTemplateSpec) GetRestartPolicyOk() (*RestartPolicy, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *MasterTemplateSpec) SetRestartPolicy(v RestartPolicy)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *MasterTemplateSpec) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *MasterTemplateSpec) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *MasterTemplateSpec) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetSecurity

`func (o *MasterTemplateSpec) GetSecurity() WorkspaceSpecSpecAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *MasterTemplateSpec) GetSecurityOk() (*WorkspaceSpecSpecAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *MasterTemplateSpec) SetSecurity(v WorkspaceSpecSpecAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *MasterTemplateSpec) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *MasterTemplateSpec) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *MasterTemplateSpec) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetStdin

`func (o *MasterTemplateSpec) GetStdin() bool`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *MasterTemplateSpec) GetStdinOk() (*bool, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *MasterTemplateSpec) SetStdin(v bool)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *MasterTemplateSpec) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *MasterTemplateSpec) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *MasterTemplateSpec) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *MasterTemplateSpec) GetStorage() WorkspaceSpecSpecAllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *MasterTemplateSpec) GetStorageOk() (*WorkspaceSpecSpecAllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *MasterTemplateSpec) SetStorage(v WorkspaceSpecSpecAllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *MasterTemplateSpec) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *MasterTemplateSpec) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *MasterTemplateSpec) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *MasterTemplateSpec) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *MasterTemplateSpec) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *MasterTemplateSpec) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *MasterTemplateSpec) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *MasterTemplateSpec) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *MasterTemplateSpec) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *MasterTemplateSpec) GetTerminationGracePeriodSeconds() int64`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *MasterTemplateSpec) GetTerminationGracePeriodSecondsOk() (*int64, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *MasterTemplateSpec) SetTerminationGracePeriodSeconds(v int64)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *MasterTemplateSpec) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *MasterTemplateSpec) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *MasterTemplateSpec) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *MasterTemplateSpec) GetTolerations() []Toleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *MasterTemplateSpec) GetTolerationsOk() (*[]Toleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *MasterTemplateSpec) SetTolerations(v []Toleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *MasterTemplateSpec) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *MasterTemplateSpec) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *MasterTemplateSpec) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTty

`func (o *MasterTemplateSpec) GetTty() bool`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *MasterTemplateSpec) GetTtyOk() (*bool, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *MasterTemplateSpec) SetTty(v bool)`

SetTty sets Tty field to given value.

### HasTty

`func (o *MasterTemplateSpec) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *MasterTemplateSpec) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *MasterTemplateSpec) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *MasterTemplateSpec) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *MasterTemplateSpec) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *MasterTemplateSpec) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *MasterTemplateSpec) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *MasterTemplateSpec) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *MasterTemplateSpec) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


