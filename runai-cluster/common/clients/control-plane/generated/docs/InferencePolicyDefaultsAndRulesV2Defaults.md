# InferencePolicyDefaultsAndRulesV2Defaults

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
**Ports** | Pointer to [**NullablePortsDefaults**](PortsDefaults.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsDefaults**](ExposedUrlsDefaults.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsDefaults**](RelatedUrlsDefaults.md) |  | [optional] 
**Compute** | Pointer to [**NullableComputeFieldsDefaults**](ComputeFieldsDefaults.md) |  | [optional] 
**Storage** | Pointer to [**NullableCommonStorageFieldsDefaults**](CommonStorageFieldsDefaults.md) |  | [optional] 
**Security** | Pointer to [**NullableCommonSecurityFlatFields**](CommonSecurityFlatFields.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableServingPort**](ServingPort.md) |  | [optional] 
**Autoscaling** | Pointer to [**NullableAutoScaling**](AutoScaling.md) |  | [optional] 

## Methods

### NewInferencePolicyDefaultsAndRulesV2Defaults

`func NewInferencePolicyDefaultsAndRulesV2Defaults() *InferencePolicyDefaultsAndRulesV2Defaults`

NewInferencePolicyDefaultsAndRulesV2Defaults instantiates a new InferencePolicyDefaultsAndRulesV2Defaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyDefaultsAndRulesV2DefaultsWithDefaults

`func NewInferencePolicyDefaultsAndRulesV2DefaultsWithDefaults() *InferencePolicyDefaultsAndRulesV2Defaults`

NewInferencePolicyDefaultsAndRulesV2DefaultsWithDefaults instantiates a new InferencePolicyDefaultsAndRulesV2Defaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetImage

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetNodeType

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetEnvironmentVariables

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetEnvironmentVariables() EnvironmentVariablesDefaults`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetEnvironmentVariablesOk() (*EnvironmentVariablesDefaults, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetEnvironmentVariables(v EnvironmentVariablesDefaults)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetAnnotations

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetAnnotations() AnnotationsDefaults`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetAnnotationsOk() (*AnnotationsDefaults, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetAnnotations(v AnnotationsDefaults)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetLabels() LabelsDefaults`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetLabelsOk() (*LabelsDefaults, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetLabels(v LabelsDefaults)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetTolerations

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetTolerations() TolerationsDefaults`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetTolerationsOk() (*TolerationsDefaults, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetTolerations(v TolerationsDefaults)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetPorts

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetPorts() PortsDefaults`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetPortsOk() (*PortsDefaults, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetPorts(v PortsDefaults)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetExposedUrls() ExposedUrlsDefaults`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetExposedUrlsOk() (*ExposedUrlsDefaults, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetExposedUrls(v ExposedUrlsDefaults)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetRelatedUrls() RelatedUrlsDefaults`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetRelatedUrlsOk() (*RelatedUrlsDefaults, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetRelatedUrls(v RelatedUrlsDefaults)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetCompute

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetCompute() ComputeFieldsDefaults`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetComputeOk() (*ComputeFieldsDefaults, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetCompute(v ComputeFieldsDefaults)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetStorage

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetStorage() CommonStorageFieldsDefaults`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetStorageOk() (*CommonStorageFieldsDefaults, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetStorage(v CommonStorageFieldsDefaults)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetSecurity

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetSecurity() CommonSecurityFlatFields`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetSecurityOk() (*CommonSecurityFlatFields, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetSecurity(v CommonSecurityFlatFields)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetServingPort

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetServingPort() ServingPort`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetServingPortOk() (*ServingPort, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetServingPort(v ServingPort)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetAutoscaling

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetAutoscaling() AutoScaling`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) GetAutoscalingOk() (*AutoScaling, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetAutoscaling(v AutoScaling)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *InferencePolicyDefaultsAndRulesV2Defaults) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *InferencePolicyDefaultsAndRulesV2Defaults) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


