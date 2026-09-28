# WorkspaceTemplateSpec

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

### NewWorkspaceTemplateSpec

`func NewWorkspaceTemplateSpec() *WorkspaceTemplateSpec`

NewWorkspaceTemplateSpec instantiates a new WorkspaceTemplateSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceTemplateSpecWithDefaults

`func NewWorkspaceTemplateSpecWithDefaults() *WorkspaceTemplateSpec`

NewWorkspaceTemplateSpecWithDefaults instantiates a new WorkspaceTemplateSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *WorkspaceTemplateSpec) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *WorkspaceTemplateSpec) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *WorkspaceTemplateSpec) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *WorkspaceTemplateSpec) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *WorkspaceTemplateSpec) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *WorkspaceTemplateSpec) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *WorkspaceTemplateSpec) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *WorkspaceTemplateSpec) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *WorkspaceTemplateSpec) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *WorkspaceTemplateSpec) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *WorkspaceTemplateSpec) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *WorkspaceTemplateSpec) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspaceTemplateSpec) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *WorkspaceTemplateSpec) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspaceTemplateSpec) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspaceTemplateSpec) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *WorkspaceTemplateSpec) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *WorkspaceTemplateSpec) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *WorkspaceTemplateSpec) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *WorkspaceTemplateSpec) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *WorkspaceTemplateSpec) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *WorkspaceTemplateSpec) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *WorkspaceTemplateSpec) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *WorkspaceTemplateSpec) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCategory

`func (o *WorkspaceTemplateSpec) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkspaceTemplateSpec) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkspaceTemplateSpec) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkspaceTemplateSpec) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *WorkspaceTemplateSpec) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *WorkspaceTemplateSpec) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCommand

`func (o *WorkspaceTemplateSpec) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *WorkspaceTemplateSpec) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *WorkspaceTemplateSpec) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *WorkspaceTemplateSpec) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *WorkspaceTemplateSpec) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *WorkspaceTemplateSpec) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompute

`func (o *WorkspaceTemplateSpec) GetCompute() WorkspaceSpecSpecAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *WorkspaceTemplateSpec) GetComputeOk() (*WorkspaceSpecSpecAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *WorkspaceTemplateSpec) SetCompute(v WorkspaceSpecSpecAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *WorkspaceTemplateSpec) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *WorkspaceTemplateSpec) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *WorkspaceTemplateSpec) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *WorkspaceTemplateSpec) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *WorkspaceTemplateSpec) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *WorkspaceTemplateSpec) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *WorkspaceTemplateSpec) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *WorkspaceTemplateSpec) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *WorkspaceTemplateSpec) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetEnvironmentVariables

`func (o *WorkspaceTemplateSpec) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *WorkspaceTemplateSpec) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *WorkspaceTemplateSpec) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *WorkspaceTemplateSpec) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *WorkspaceTemplateSpec) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *WorkspaceTemplateSpec) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *WorkspaceTemplateSpec) GetExposedUrls() []ExposedUrl`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *WorkspaceTemplateSpec) GetExposedUrlsOk() (*[]ExposedUrl, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *WorkspaceTemplateSpec) SetExposedUrls(v []ExposedUrl)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *WorkspaceTemplateSpec) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *WorkspaceTemplateSpec) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *WorkspaceTemplateSpec) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *WorkspaceTemplateSpec) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *WorkspaceTemplateSpec) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *WorkspaceTemplateSpec) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *WorkspaceTemplateSpec) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *WorkspaceTemplateSpec) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *WorkspaceTemplateSpec) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *WorkspaceTemplateSpec) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *WorkspaceTemplateSpec) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *WorkspaceTemplateSpec) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *WorkspaceTemplateSpec) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *WorkspaceTemplateSpec) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *WorkspaceTemplateSpec) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetLabels

`func (o *WorkspaceTemplateSpec) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkspaceTemplateSpec) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkspaceTemplateSpec) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *WorkspaceTemplateSpec) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *WorkspaceTemplateSpec) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *WorkspaceTemplateSpec) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetNodeAffinityRequired

`func (o *WorkspaceTemplateSpec) GetNodeAffinityRequired() NodeAffinityRequired`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *WorkspaceTemplateSpec) GetNodeAffinityRequiredOk() (*NodeAffinityRequired, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *WorkspaceTemplateSpec) SetNodeAffinityRequired(v NodeAffinityRequired)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *WorkspaceTemplateSpec) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *WorkspaceTemplateSpec) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *WorkspaceTemplateSpec) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *WorkspaceTemplateSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *WorkspaceTemplateSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *WorkspaceTemplateSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *WorkspaceTemplateSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *WorkspaceTemplateSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *WorkspaceTemplateSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *WorkspaceTemplateSpec) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *WorkspaceTemplateSpec) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *WorkspaceTemplateSpec) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *WorkspaceTemplateSpec) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *WorkspaceTemplateSpec) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *WorkspaceTemplateSpec) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetPodAffinity

`func (o *WorkspaceTemplateSpec) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *WorkspaceTemplateSpec) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *WorkspaceTemplateSpec) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *WorkspaceTemplateSpec) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *WorkspaceTemplateSpec) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *WorkspaceTemplateSpec) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *WorkspaceTemplateSpec) GetPorts() []Port`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *WorkspaceTemplateSpec) GetPortsOk() (*[]Port, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *WorkspaceTemplateSpec) SetPorts(v []Port)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *WorkspaceTemplateSpec) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *WorkspaceTemplateSpec) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *WorkspaceTemplateSpec) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetPreemptibility

`func (o *WorkspaceTemplateSpec) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkspaceTemplateSpec) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkspaceTemplateSpec) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkspaceTemplateSpec) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkspaceTemplateSpec) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkspaceTemplateSpec) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *WorkspaceTemplateSpec) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *WorkspaceTemplateSpec) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *WorkspaceTemplateSpec) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *WorkspaceTemplateSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *WorkspaceTemplateSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *WorkspaceTemplateSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *WorkspaceTemplateSpec) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *WorkspaceTemplateSpec) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *WorkspaceTemplateSpec) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *WorkspaceTemplateSpec) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *WorkspaceTemplateSpec) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *WorkspaceTemplateSpec) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *WorkspaceTemplateSpec) GetRelatedUrls() []RelatedUrl`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *WorkspaceTemplateSpec) GetRelatedUrlsOk() (*[]RelatedUrl, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *WorkspaceTemplateSpec) SetRelatedUrls(v []RelatedUrl)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *WorkspaceTemplateSpec) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *WorkspaceTemplateSpec) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *WorkspaceTemplateSpec) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetRestartPolicy

`func (o *WorkspaceTemplateSpec) GetRestartPolicy() RestartPolicy`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *WorkspaceTemplateSpec) GetRestartPolicyOk() (*RestartPolicy, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *WorkspaceTemplateSpec) SetRestartPolicy(v RestartPolicy)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *WorkspaceTemplateSpec) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *WorkspaceTemplateSpec) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *WorkspaceTemplateSpec) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetSecurity

`func (o *WorkspaceTemplateSpec) GetSecurity() WorkspaceSpecSpecAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *WorkspaceTemplateSpec) GetSecurityOk() (*WorkspaceSpecSpecAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *WorkspaceTemplateSpec) SetSecurity(v WorkspaceSpecSpecAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *WorkspaceTemplateSpec) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *WorkspaceTemplateSpec) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *WorkspaceTemplateSpec) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetStdin

`func (o *WorkspaceTemplateSpec) GetStdin() bool`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *WorkspaceTemplateSpec) GetStdinOk() (*bool, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *WorkspaceTemplateSpec) SetStdin(v bool)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *WorkspaceTemplateSpec) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *WorkspaceTemplateSpec) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *WorkspaceTemplateSpec) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *WorkspaceTemplateSpec) GetStorage() WorkspaceSpecSpecAllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkspaceTemplateSpec) GetStorageOk() (*WorkspaceSpecSpecAllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkspaceTemplateSpec) SetStorage(v WorkspaceSpecSpecAllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *WorkspaceTemplateSpec) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *WorkspaceTemplateSpec) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *WorkspaceTemplateSpec) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *WorkspaceTemplateSpec) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *WorkspaceTemplateSpec) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *WorkspaceTemplateSpec) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *WorkspaceTemplateSpec) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *WorkspaceTemplateSpec) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *WorkspaceTemplateSpec) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *WorkspaceTemplateSpec) GetTerminationGracePeriodSeconds() int64`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *WorkspaceTemplateSpec) GetTerminationGracePeriodSecondsOk() (*int64, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *WorkspaceTemplateSpec) SetTerminationGracePeriodSeconds(v int64)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *WorkspaceTemplateSpec) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *WorkspaceTemplateSpec) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *WorkspaceTemplateSpec) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *WorkspaceTemplateSpec) GetTolerations() []Toleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *WorkspaceTemplateSpec) GetTolerationsOk() (*[]Toleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *WorkspaceTemplateSpec) SetTolerations(v []Toleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *WorkspaceTemplateSpec) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *WorkspaceTemplateSpec) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *WorkspaceTemplateSpec) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTty

`func (o *WorkspaceTemplateSpec) GetTty() bool`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *WorkspaceTemplateSpec) GetTtyOk() (*bool, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *WorkspaceTemplateSpec) SetTty(v bool)`

SetTty sets Tty field to given value.

### HasTty

`func (o *WorkspaceTemplateSpec) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *WorkspaceTemplateSpec) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *WorkspaceTemplateSpec) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *WorkspaceTemplateSpec) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *WorkspaceTemplateSpec) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *WorkspaceTemplateSpec) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *WorkspaceTemplateSpec) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *WorkspaceTemplateSpec) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *WorkspaceTemplateSpec) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


