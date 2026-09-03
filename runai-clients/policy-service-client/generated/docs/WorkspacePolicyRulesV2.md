# WorkspacePolicyRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Args** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**BackoffLimit** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Category** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Command** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Compute** | Pointer to [**NullableSupersetRulesAllOfCompute**](SupersetRulesAllOfCompute.md) |  | [optional] 
**CreateHomeDir** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsRules**](ExposedUrlsRules.md) |  | [optional] 
**Image** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicyRules**](ImagePullPolicyRules.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**NullableImagePullSecretsRules**](ImagePullSecretsRules.md) |  | [optional] 
**Labels** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**NodeAffinityRequired** | Pointer to [**NullableArbitraryRules**](ArbitraryRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**NodeType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinityRules**](PodAffinityRules.md) |  | [optional] 
**Ports** | Pointer to [**NullablePortsRules**](PortsRules.md) |  | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibilityRules**](PreemptibilityRules.md) |  | [optional] 
**PriorityClass** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Probes** | Pointer to [**NullableProbesRules**](ProbesRules.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsRules**](RelatedUrlsRules.md) |  | [optional] 
**RestartPolicy** | Pointer to [**NullableRestartPolicyRule**](RestartPolicyRule.md) |  | [optional] 
**Security** | Pointer to [**NullableSupersetRulesAllOfSecurity**](SupersetRulesAllOfSecurity.md) |  | [optional] 
**Stdin** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**Storage** | Pointer to [**NullableSupersetRulesAllOfStorage**](SupersetRulesAllOfStorage.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**TerminationGracePeriodSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsRules**](TolerationsRules.md) |  | [optional] 
**Tty** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**WorkingDir** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewWorkspacePolicyRulesV2

`func NewWorkspacePolicyRulesV2() *WorkspacePolicyRulesV2`

NewWorkspacePolicyRulesV2 instantiates a new WorkspacePolicyRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspacePolicyRulesV2WithDefaults

`func NewWorkspacePolicyRulesV2WithDefaults() *WorkspacePolicyRulesV2`

NewWorkspacePolicyRulesV2WithDefaults instantiates a new WorkspacePolicyRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *WorkspacePolicyRulesV2) GetAnnotations() InstancesRules`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *WorkspacePolicyRulesV2) GetAnnotationsOk() (*InstancesRules, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *WorkspacePolicyRulesV2) SetAnnotations(v InstancesRules)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *WorkspacePolicyRulesV2) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *WorkspacePolicyRulesV2) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *WorkspacePolicyRulesV2) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *WorkspacePolicyRulesV2) GetArgs() StringRules`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *WorkspacePolicyRulesV2) GetArgsOk() (*StringRules, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *WorkspacePolicyRulesV2) SetArgs(v StringRules)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *WorkspacePolicyRulesV2) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *WorkspacePolicyRulesV2) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *WorkspacePolicyRulesV2) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspacePolicyRulesV2) GetAutoDeletionTimeAfterCompletionSeconds() IntegerRules`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *WorkspacePolicyRulesV2) GetAutoDeletionTimeAfterCompletionSecondsOk() (*IntegerRules, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspacePolicyRulesV2) SetAutoDeletionTimeAfterCompletionSeconds(v IntegerRules)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspacePolicyRulesV2) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *WorkspacePolicyRulesV2) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *WorkspacePolicyRulesV2) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *WorkspacePolicyRulesV2) GetBackoffLimit() IntegerRules`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *WorkspacePolicyRulesV2) GetBackoffLimitOk() (*IntegerRules, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *WorkspacePolicyRulesV2) SetBackoffLimit(v IntegerRules)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *WorkspacePolicyRulesV2) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *WorkspacePolicyRulesV2) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *WorkspacePolicyRulesV2) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCategory

`func (o *WorkspacePolicyRulesV2) GetCategory() StringRules`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkspacePolicyRulesV2) GetCategoryOk() (*StringRules, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkspacePolicyRulesV2) SetCategory(v StringRules)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkspacePolicyRulesV2) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *WorkspacePolicyRulesV2) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *WorkspacePolicyRulesV2) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCommand

`func (o *WorkspacePolicyRulesV2) GetCommand() StringRules`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *WorkspacePolicyRulesV2) GetCommandOk() (*StringRules, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *WorkspacePolicyRulesV2) SetCommand(v StringRules)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *WorkspacePolicyRulesV2) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *WorkspacePolicyRulesV2) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *WorkspacePolicyRulesV2) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompute

`func (o *WorkspacePolicyRulesV2) GetCompute() SupersetRulesAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *WorkspacePolicyRulesV2) GetComputeOk() (*SupersetRulesAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *WorkspacePolicyRulesV2) SetCompute(v SupersetRulesAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *WorkspacePolicyRulesV2) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *WorkspacePolicyRulesV2) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *WorkspacePolicyRulesV2) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *WorkspacePolicyRulesV2) GetCreateHomeDir() BooleanRules`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *WorkspacePolicyRulesV2) GetCreateHomeDirOk() (*BooleanRules, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *WorkspacePolicyRulesV2) SetCreateHomeDir(v BooleanRules)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *WorkspacePolicyRulesV2) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *WorkspacePolicyRulesV2) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *WorkspacePolicyRulesV2) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetEnvironmentVariables

`func (o *WorkspacePolicyRulesV2) GetEnvironmentVariables() InstancesRules`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *WorkspacePolicyRulesV2) GetEnvironmentVariablesOk() (*InstancesRules, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *WorkspacePolicyRulesV2) SetEnvironmentVariables(v InstancesRules)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *WorkspacePolicyRulesV2) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *WorkspacePolicyRulesV2) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *WorkspacePolicyRulesV2) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *WorkspacePolicyRulesV2) GetExposedUrls() ExposedUrlsRules`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *WorkspacePolicyRulesV2) GetExposedUrlsOk() (*ExposedUrlsRules, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *WorkspacePolicyRulesV2) SetExposedUrls(v ExposedUrlsRules)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *WorkspacePolicyRulesV2) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *WorkspacePolicyRulesV2) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *WorkspacePolicyRulesV2) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *WorkspacePolicyRulesV2) GetImage() StringRules`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *WorkspacePolicyRulesV2) GetImageOk() (*StringRules, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *WorkspacePolicyRulesV2) SetImage(v StringRules)`

SetImage sets Image field to given value.

### HasImage

`func (o *WorkspacePolicyRulesV2) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *WorkspacePolicyRulesV2) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *WorkspacePolicyRulesV2) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *WorkspacePolicyRulesV2) GetImagePullPolicy() ImagePullPolicyRules`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *WorkspacePolicyRulesV2) GetImagePullPolicyOk() (*ImagePullPolicyRules, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *WorkspacePolicyRulesV2) SetImagePullPolicy(v ImagePullPolicyRules)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *WorkspacePolicyRulesV2) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *WorkspacePolicyRulesV2) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *WorkspacePolicyRulesV2) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *WorkspacePolicyRulesV2) GetImagePullSecrets() ImagePullSecretsRules`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *WorkspacePolicyRulesV2) GetImagePullSecretsOk() (*ImagePullSecretsRules, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *WorkspacePolicyRulesV2) SetImagePullSecrets(v ImagePullSecretsRules)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *WorkspacePolicyRulesV2) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *WorkspacePolicyRulesV2) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *WorkspacePolicyRulesV2) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *WorkspacePolicyRulesV2) GetLabels() InstancesRules`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkspacePolicyRulesV2) GetLabelsOk() (*InstancesRules, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkspacePolicyRulesV2) SetLabels(v InstancesRules)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *WorkspacePolicyRulesV2) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *WorkspacePolicyRulesV2) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *WorkspacePolicyRulesV2) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetNodeAffinityRequired

`func (o *WorkspacePolicyRulesV2) GetNodeAffinityRequired() ArbitraryRules`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *WorkspacePolicyRulesV2) GetNodeAffinityRequiredOk() (*ArbitraryRules, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *WorkspacePolicyRulesV2) SetNodeAffinityRequired(v ArbitraryRules)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *WorkspacePolicyRulesV2) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *WorkspacePolicyRulesV2) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *WorkspacePolicyRulesV2) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *WorkspacePolicyRulesV2) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *WorkspacePolicyRulesV2) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *WorkspacePolicyRulesV2) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *WorkspacePolicyRulesV2) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *WorkspacePolicyRulesV2) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *WorkspacePolicyRulesV2) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *WorkspacePolicyRulesV2) GetNodeType() StringRules`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *WorkspacePolicyRulesV2) GetNodeTypeOk() (*StringRules, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *WorkspacePolicyRulesV2) SetNodeType(v StringRules)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *WorkspacePolicyRulesV2) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *WorkspacePolicyRulesV2) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *WorkspacePolicyRulesV2) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetPodAffinity

`func (o *WorkspacePolicyRulesV2) GetPodAffinity() PodAffinityRules`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *WorkspacePolicyRulesV2) GetPodAffinityOk() (*PodAffinityRules, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *WorkspacePolicyRulesV2) SetPodAffinity(v PodAffinityRules)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *WorkspacePolicyRulesV2) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *WorkspacePolicyRulesV2) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *WorkspacePolicyRulesV2) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *WorkspacePolicyRulesV2) GetPorts() PortsRules`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *WorkspacePolicyRulesV2) GetPortsOk() (*PortsRules, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *WorkspacePolicyRulesV2) SetPorts(v PortsRules)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *WorkspacePolicyRulesV2) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *WorkspacePolicyRulesV2) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *WorkspacePolicyRulesV2) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetPreemptibility

`func (o *WorkspacePolicyRulesV2) GetPreemptibility() PreemptibilityRules`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkspacePolicyRulesV2) GetPreemptibilityOk() (*PreemptibilityRules, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkspacePolicyRulesV2) SetPreemptibility(v PreemptibilityRules)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkspacePolicyRulesV2) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkspacePolicyRulesV2) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkspacePolicyRulesV2) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *WorkspacePolicyRulesV2) GetPriorityClass() StringRules`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *WorkspacePolicyRulesV2) GetPriorityClassOk() (*StringRules, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *WorkspacePolicyRulesV2) SetPriorityClass(v StringRules)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *WorkspacePolicyRulesV2) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *WorkspacePolicyRulesV2) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *WorkspacePolicyRulesV2) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *WorkspacePolicyRulesV2) GetProbes() ProbesRules`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *WorkspacePolicyRulesV2) GetProbesOk() (*ProbesRules, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *WorkspacePolicyRulesV2) SetProbes(v ProbesRules)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *WorkspacePolicyRulesV2) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *WorkspacePolicyRulesV2) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *WorkspacePolicyRulesV2) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *WorkspacePolicyRulesV2) GetRelatedUrls() RelatedUrlsRules`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *WorkspacePolicyRulesV2) GetRelatedUrlsOk() (*RelatedUrlsRules, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *WorkspacePolicyRulesV2) SetRelatedUrls(v RelatedUrlsRules)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *WorkspacePolicyRulesV2) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *WorkspacePolicyRulesV2) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *WorkspacePolicyRulesV2) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetRestartPolicy

`func (o *WorkspacePolicyRulesV2) GetRestartPolicy() RestartPolicyRule`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *WorkspacePolicyRulesV2) GetRestartPolicyOk() (*RestartPolicyRule, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *WorkspacePolicyRulesV2) SetRestartPolicy(v RestartPolicyRule)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *WorkspacePolicyRulesV2) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *WorkspacePolicyRulesV2) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *WorkspacePolicyRulesV2) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetSecurity

`func (o *WorkspacePolicyRulesV2) GetSecurity() SupersetRulesAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *WorkspacePolicyRulesV2) GetSecurityOk() (*SupersetRulesAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *WorkspacePolicyRulesV2) SetSecurity(v SupersetRulesAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *WorkspacePolicyRulesV2) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *WorkspacePolicyRulesV2) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *WorkspacePolicyRulesV2) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetStdin

`func (o *WorkspacePolicyRulesV2) GetStdin() BooleanRules`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *WorkspacePolicyRulesV2) GetStdinOk() (*BooleanRules, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *WorkspacePolicyRulesV2) SetStdin(v BooleanRules)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *WorkspacePolicyRulesV2) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *WorkspacePolicyRulesV2) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *WorkspacePolicyRulesV2) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *WorkspacePolicyRulesV2) GetStorage() SupersetRulesAllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkspacePolicyRulesV2) GetStorageOk() (*SupersetRulesAllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkspacePolicyRulesV2) SetStorage(v SupersetRulesAllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *WorkspacePolicyRulesV2) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *WorkspacePolicyRulesV2) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *WorkspacePolicyRulesV2) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *WorkspacePolicyRulesV2) GetTerminateAfterPreemption() BooleanRules`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *WorkspacePolicyRulesV2) GetTerminateAfterPreemptionOk() (*BooleanRules, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *WorkspacePolicyRulesV2) SetTerminateAfterPreemption(v BooleanRules)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *WorkspacePolicyRulesV2) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *WorkspacePolicyRulesV2) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *WorkspacePolicyRulesV2) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *WorkspacePolicyRulesV2) GetTerminationGracePeriodSeconds() IntegerRules`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *WorkspacePolicyRulesV2) GetTerminationGracePeriodSecondsOk() (*IntegerRules, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *WorkspacePolicyRulesV2) SetTerminationGracePeriodSeconds(v IntegerRules)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *WorkspacePolicyRulesV2) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *WorkspacePolicyRulesV2) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *WorkspacePolicyRulesV2) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *WorkspacePolicyRulesV2) GetTolerations() TolerationsRules`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *WorkspacePolicyRulesV2) GetTolerationsOk() (*TolerationsRules, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *WorkspacePolicyRulesV2) SetTolerations(v TolerationsRules)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *WorkspacePolicyRulesV2) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *WorkspacePolicyRulesV2) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *WorkspacePolicyRulesV2) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTty

`func (o *WorkspacePolicyRulesV2) GetTty() BooleanRules`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *WorkspacePolicyRulesV2) GetTtyOk() (*BooleanRules, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *WorkspacePolicyRulesV2) SetTty(v BooleanRules)`

SetTty sets Tty field to given value.

### HasTty

`func (o *WorkspacePolicyRulesV2) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *WorkspacePolicyRulesV2) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *WorkspacePolicyRulesV2) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *WorkspacePolicyRulesV2) GetWorkingDir() StringRules`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *WorkspacePolicyRulesV2) GetWorkingDirOk() (*StringRules, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *WorkspacePolicyRulesV2) SetWorkingDir(v StringRules)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *WorkspacePolicyRulesV2) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *WorkspacePolicyRulesV2) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *WorkspacePolicyRulesV2) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


