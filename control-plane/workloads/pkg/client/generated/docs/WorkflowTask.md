# WorkflowTask

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workload. | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workload executes. | [optional] 
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workload. | [optional] 
**ComputeRef** | Pointer to **NullableString** | The name of the profile, which tasks use to reference it. It must be unique among the profiles of its kind. (mandatory) | [optional] 
**CreateHomeDir** | Pointer to **NullableBool** | When set to &#x60;true&#x60;, creates a home directory for the container. | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workload. | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workload. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**[]ImagePullSecret**](ImagePullSecret.md) | A list of references to Kubernetes secrets in the same namespace used for pulling container images. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workload. | [optional] 
**Name** | Pointer to **NullableString** | The name of the task. It must be unique in its group, and begin with a letter. It can be omitted in a group with a single task, which then takes the group&#39;s name. | [optional] 
**PlacementRef** | Pointer to **NullableString** | The name of the profile, which tasks use to reference it. It must be unique among the profiles of its kind. (mandatory) | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**Retry** | Pointer to [**NullableTaskRetry**](TaskRetry.md) |  | [optional] 
**Security** | Pointer to [**NullableWorkspaceSpecSpecAllOfSecurity**](WorkspaceSpecSpecAllOfSecurity.md) |  | [optional] 
**Stdin** | Pointer to **NullableBool** | Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF | [optional] 
**Storage** | Pointer to [**NullableWorkflowTaskStorage**](WorkflowTaskStorage.md) |  | [optional] 
**StorageRef** | Pointer to [**[]WorkflowS3StorageRef**](WorkflowS3StorageRef.md) | The storages of the workflow, in the workflow&#39;s &#x60;storage&#x60;, that the task uses. | [optional] 
**SuccessExitCodes** | Pointer to [**[]TaskExitCodeRange**](TaskExitCodeRange.md) | Ranges of exit codes that end the task as succeeded, in addition to 0. A non-zero exit code that isn&#39;t in this list or in &#x60;retry.onExitCodes&#x60; fails the task. | [optional] 
**TerminationGracePeriodSeconds** | Pointer to **NullableInt64** | Duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). | [optional] 
**Tty** | Pointer to **NullableBool** | Whether this container should allocate a TTY for itself, also requires &#39;stdin&#39; to be true. | [optional] 
**WorkingDir** | Pointer to **NullableString** | Container&#39;s working directory. If not specified, the container runtime default will be used. This may be configured in the container image. | [optional] 

## Methods

### NewWorkflowTask

`func NewWorkflowTask() *WorkflowTask`

NewWorkflowTask instantiates a new WorkflowTask object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowTaskWithDefaults

`func NewWorkflowTaskWithDefaults() *WorkflowTask`

NewWorkflowTaskWithDefaults instantiates a new WorkflowTask object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *WorkflowTask) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *WorkflowTask) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *WorkflowTask) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *WorkflowTask) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *WorkflowTask) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *WorkflowTask) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *WorkflowTask) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *WorkflowTask) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *WorkflowTask) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *WorkflowTask) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *WorkflowTask) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *WorkflowTask) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetCommand

`func (o *WorkflowTask) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *WorkflowTask) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *WorkflowTask) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *WorkflowTask) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *WorkflowTask) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *WorkflowTask) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetComputeRef

`func (o *WorkflowTask) GetComputeRef() string`

GetComputeRef returns the ComputeRef field if non-nil, zero value otherwise.

### GetComputeRefOk

`func (o *WorkflowTask) GetComputeRefOk() (*string, bool)`

GetComputeRefOk returns a tuple with the ComputeRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComputeRef

`func (o *WorkflowTask) SetComputeRef(v string)`

SetComputeRef sets ComputeRef field to given value.

### HasComputeRef

`func (o *WorkflowTask) HasComputeRef() bool`

HasComputeRef returns a boolean if a field has been set.

### SetComputeRefNil

`func (o *WorkflowTask) SetComputeRefNil(b bool)`

 SetComputeRefNil sets the value for ComputeRef to be an explicit nil

### UnsetComputeRef
`func (o *WorkflowTask) UnsetComputeRef()`

UnsetComputeRef ensures that no value is present for ComputeRef, not even an explicit nil
### GetCreateHomeDir

`func (o *WorkflowTask) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *WorkflowTask) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *WorkflowTask) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *WorkflowTask) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *WorkflowTask) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *WorkflowTask) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetEnvironmentVariables

`func (o *WorkflowTask) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *WorkflowTask) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *WorkflowTask) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *WorkflowTask) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *WorkflowTask) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *WorkflowTask) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetImage

`func (o *WorkflowTask) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *WorkflowTask) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *WorkflowTask) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *WorkflowTask) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *WorkflowTask) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *WorkflowTask) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *WorkflowTask) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *WorkflowTask) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *WorkflowTask) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *WorkflowTask) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *WorkflowTask) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *WorkflowTask) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *WorkflowTask) GetImagePullSecrets() []ImagePullSecret`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *WorkflowTask) GetImagePullSecretsOk() (*[]ImagePullSecret, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *WorkflowTask) SetImagePullSecrets(v []ImagePullSecret)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *WorkflowTask) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *WorkflowTask) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *WorkflowTask) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *WorkflowTask) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkflowTask) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkflowTask) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *WorkflowTask) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *WorkflowTask) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *WorkflowTask) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetName

`func (o *WorkflowTask) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkflowTask) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkflowTask) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkflowTask) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *WorkflowTask) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *WorkflowTask) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetPlacementRef

`func (o *WorkflowTask) GetPlacementRef() string`

GetPlacementRef returns the PlacementRef field if non-nil, zero value otherwise.

### GetPlacementRefOk

`func (o *WorkflowTask) GetPlacementRefOk() (*string, bool)`

GetPlacementRefOk returns a tuple with the PlacementRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementRef

`func (o *WorkflowTask) SetPlacementRef(v string)`

SetPlacementRef sets PlacementRef field to given value.

### HasPlacementRef

`func (o *WorkflowTask) HasPlacementRef() bool`

HasPlacementRef returns a boolean if a field has been set.

### SetPlacementRefNil

`func (o *WorkflowTask) SetPlacementRefNil(b bool)`

 SetPlacementRefNil sets the value for PlacementRef to be an explicit nil

### UnsetPlacementRef
`func (o *WorkflowTask) UnsetPlacementRef()`

UnsetPlacementRef ensures that no value is present for PlacementRef, not even an explicit nil
### GetProbes

`func (o *WorkflowTask) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *WorkflowTask) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *WorkflowTask) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *WorkflowTask) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *WorkflowTask) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *WorkflowTask) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRetry

`func (o *WorkflowTask) GetRetry() TaskRetry`

GetRetry returns the Retry field if non-nil, zero value otherwise.

### GetRetryOk

`func (o *WorkflowTask) GetRetryOk() (*TaskRetry, bool)`

GetRetryOk returns a tuple with the Retry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetry

`func (o *WorkflowTask) SetRetry(v TaskRetry)`

SetRetry sets Retry field to given value.

### HasRetry

`func (o *WorkflowTask) HasRetry() bool`

HasRetry returns a boolean if a field has been set.

### SetRetryNil

`func (o *WorkflowTask) SetRetryNil(b bool)`

 SetRetryNil sets the value for Retry to be an explicit nil

### UnsetRetry
`func (o *WorkflowTask) UnsetRetry()`

UnsetRetry ensures that no value is present for Retry, not even an explicit nil
### GetSecurity

`func (o *WorkflowTask) GetSecurity() WorkspaceSpecSpecAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *WorkflowTask) GetSecurityOk() (*WorkspaceSpecSpecAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *WorkflowTask) SetSecurity(v WorkspaceSpecSpecAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *WorkflowTask) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *WorkflowTask) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *WorkflowTask) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetStdin

`func (o *WorkflowTask) GetStdin() bool`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *WorkflowTask) GetStdinOk() (*bool, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *WorkflowTask) SetStdin(v bool)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *WorkflowTask) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *WorkflowTask) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *WorkflowTask) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *WorkflowTask) GetStorage() WorkflowTaskStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkflowTask) GetStorageOk() (*WorkflowTaskStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkflowTask) SetStorage(v WorkflowTaskStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *WorkflowTask) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *WorkflowTask) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *WorkflowTask) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetStorageRef

`func (o *WorkflowTask) GetStorageRef() []WorkflowS3StorageRef`

GetStorageRef returns the StorageRef field if non-nil, zero value otherwise.

### GetStorageRefOk

`func (o *WorkflowTask) GetStorageRefOk() (*[]WorkflowS3StorageRef, bool)`

GetStorageRefOk returns a tuple with the StorageRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageRef

`func (o *WorkflowTask) SetStorageRef(v []WorkflowS3StorageRef)`

SetStorageRef sets StorageRef field to given value.

### HasStorageRef

`func (o *WorkflowTask) HasStorageRef() bool`

HasStorageRef returns a boolean if a field has been set.

### SetStorageRefNil

`func (o *WorkflowTask) SetStorageRefNil(b bool)`

 SetStorageRefNil sets the value for StorageRef to be an explicit nil

### UnsetStorageRef
`func (o *WorkflowTask) UnsetStorageRef()`

UnsetStorageRef ensures that no value is present for StorageRef, not even an explicit nil
### GetSuccessExitCodes

`func (o *WorkflowTask) GetSuccessExitCodes() []TaskExitCodeRange`

GetSuccessExitCodes returns the SuccessExitCodes field if non-nil, zero value otherwise.

### GetSuccessExitCodesOk

`func (o *WorkflowTask) GetSuccessExitCodesOk() (*[]TaskExitCodeRange, bool)`

GetSuccessExitCodesOk returns a tuple with the SuccessExitCodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuccessExitCodes

`func (o *WorkflowTask) SetSuccessExitCodes(v []TaskExitCodeRange)`

SetSuccessExitCodes sets SuccessExitCodes field to given value.

### HasSuccessExitCodes

`func (o *WorkflowTask) HasSuccessExitCodes() bool`

HasSuccessExitCodes returns a boolean if a field has been set.

### SetSuccessExitCodesNil

`func (o *WorkflowTask) SetSuccessExitCodesNil(b bool)`

 SetSuccessExitCodesNil sets the value for SuccessExitCodes to be an explicit nil

### UnsetSuccessExitCodes
`func (o *WorkflowTask) UnsetSuccessExitCodes()`

UnsetSuccessExitCodes ensures that no value is present for SuccessExitCodes, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *WorkflowTask) GetTerminationGracePeriodSeconds() int64`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *WorkflowTask) GetTerminationGracePeriodSecondsOk() (*int64, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *WorkflowTask) SetTerminationGracePeriodSeconds(v int64)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *WorkflowTask) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *WorkflowTask) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *WorkflowTask) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTty

`func (o *WorkflowTask) GetTty() bool`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *WorkflowTask) GetTtyOk() (*bool, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *WorkflowTask) SetTty(v bool)`

SetTty sets Tty field to given value.

### HasTty

`func (o *WorkflowTask) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *WorkflowTask) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *WorkflowTask) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *WorkflowTask) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *WorkflowTask) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *WorkflowTask) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *WorkflowTask) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *WorkflowTask) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *WorkflowTask) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


