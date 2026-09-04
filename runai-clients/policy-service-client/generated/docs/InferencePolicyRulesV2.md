# InferencePolicyRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Autoscaling** | Pointer to [**NullableAutoScalingRules**](AutoScalingRules.md) |  | [optional] 
**ServingConfiguration** | Pointer to [**NullableServingConfigurationRules**](ServingConfigurationRules.md) |  | [optional] 
**Annotations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Args** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
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
**Security** | Pointer to [**NullableInferencePolicyRulesV2AllOfSecurity**](InferencePolicyRulesV2AllOfSecurity.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableServingPortRules**](ServingPortRules.md) |  | [optional] 
**Storage** | Pointer to [**NullableInferencePolicyRulesV2AllOfStorage**](InferencePolicyRulesV2AllOfStorage.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsRules**](TolerationsRules.md) |  | [optional] 
**WorkingDir** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 

## Methods

### NewInferencePolicyRulesV2

`func NewInferencePolicyRulesV2() *InferencePolicyRulesV2`

NewInferencePolicyRulesV2 instantiates a new InferencePolicyRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyRulesV2WithDefaults

`func NewInferencePolicyRulesV2WithDefaults() *InferencePolicyRulesV2`

NewInferencePolicyRulesV2WithDefaults instantiates a new InferencePolicyRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAutoscaling

`func (o *InferencePolicyRulesV2) GetAutoscaling() AutoScalingRules`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *InferencePolicyRulesV2) GetAutoscalingOk() (*AutoScalingRules, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *InferencePolicyRulesV2) SetAutoscaling(v AutoScalingRules)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *InferencePolicyRulesV2) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *InferencePolicyRulesV2) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *InferencePolicyRulesV2) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil
### GetServingConfiguration

`func (o *InferencePolicyRulesV2) GetServingConfiguration() ServingConfigurationRules`

GetServingConfiguration returns the ServingConfiguration field if non-nil, zero value otherwise.

### GetServingConfigurationOk

`func (o *InferencePolicyRulesV2) GetServingConfigurationOk() (*ServingConfigurationRules, bool)`

GetServingConfigurationOk returns a tuple with the ServingConfiguration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingConfiguration

`func (o *InferencePolicyRulesV2) SetServingConfiguration(v ServingConfigurationRules)`

SetServingConfiguration sets ServingConfiguration field to given value.

### HasServingConfiguration

`func (o *InferencePolicyRulesV2) HasServingConfiguration() bool`

HasServingConfiguration returns a boolean if a field has been set.

### SetServingConfigurationNil

`func (o *InferencePolicyRulesV2) SetServingConfigurationNil(b bool)`

 SetServingConfigurationNil sets the value for ServingConfiguration to be an explicit nil

### UnsetServingConfiguration
`func (o *InferencePolicyRulesV2) UnsetServingConfiguration()`

UnsetServingConfiguration ensures that no value is present for ServingConfiguration, not even an explicit nil
### GetAnnotations

`func (o *InferencePolicyRulesV2) GetAnnotations() InstancesRules`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *InferencePolicyRulesV2) GetAnnotationsOk() (*InstancesRules, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *InferencePolicyRulesV2) SetAnnotations(v InstancesRules)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *InferencePolicyRulesV2) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *InferencePolicyRulesV2) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *InferencePolicyRulesV2) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *InferencePolicyRulesV2) GetArgs() StringRules`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *InferencePolicyRulesV2) GetArgsOk() (*StringRules, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *InferencePolicyRulesV2) SetArgs(v StringRules)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *InferencePolicyRulesV2) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *InferencePolicyRulesV2) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *InferencePolicyRulesV2) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetCategory

`func (o *InferencePolicyRulesV2) GetCategory() StringRules`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *InferencePolicyRulesV2) GetCategoryOk() (*StringRules, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *InferencePolicyRulesV2) SetCategory(v StringRules)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *InferencePolicyRulesV2) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *InferencePolicyRulesV2) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *InferencePolicyRulesV2) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCommand

`func (o *InferencePolicyRulesV2) GetCommand() StringRules`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *InferencePolicyRulesV2) GetCommandOk() (*StringRules, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *InferencePolicyRulesV2) SetCommand(v StringRules)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *InferencePolicyRulesV2) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *InferencePolicyRulesV2) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *InferencePolicyRulesV2) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompute

`func (o *InferencePolicyRulesV2) GetCompute() SupersetRulesAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *InferencePolicyRulesV2) GetComputeOk() (*SupersetRulesAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *InferencePolicyRulesV2) SetCompute(v SupersetRulesAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *InferencePolicyRulesV2) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *InferencePolicyRulesV2) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *InferencePolicyRulesV2) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *InferencePolicyRulesV2) GetCreateHomeDir() BooleanRules`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *InferencePolicyRulesV2) GetCreateHomeDirOk() (*BooleanRules, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *InferencePolicyRulesV2) SetCreateHomeDir(v BooleanRules)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *InferencePolicyRulesV2) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *InferencePolicyRulesV2) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *InferencePolicyRulesV2) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetEnvironmentVariables

`func (o *InferencePolicyRulesV2) GetEnvironmentVariables() InstancesRules`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *InferencePolicyRulesV2) GetEnvironmentVariablesOk() (*InstancesRules, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *InferencePolicyRulesV2) SetEnvironmentVariables(v InstancesRules)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *InferencePolicyRulesV2) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *InferencePolicyRulesV2) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *InferencePolicyRulesV2) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *InferencePolicyRulesV2) GetExposedUrls() ExposedUrlsRules`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *InferencePolicyRulesV2) GetExposedUrlsOk() (*ExposedUrlsRules, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *InferencePolicyRulesV2) SetExposedUrls(v ExposedUrlsRules)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *InferencePolicyRulesV2) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *InferencePolicyRulesV2) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *InferencePolicyRulesV2) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *InferencePolicyRulesV2) GetImage() StringRules`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *InferencePolicyRulesV2) GetImageOk() (*StringRules, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *InferencePolicyRulesV2) SetImage(v StringRules)`

SetImage sets Image field to given value.

### HasImage

`func (o *InferencePolicyRulesV2) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *InferencePolicyRulesV2) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *InferencePolicyRulesV2) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *InferencePolicyRulesV2) GetImagePullPolicy() ImagePullPolicyRules`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *InferencePolicyRulesV2) GetImagePullPolicyOk() (*ImagePullPolicyRules, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *InferencePolicyRulesV2) SetImagePullPolicy(v ImagePullPolicyRules)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *InferencePolicyRulesV2) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *InferencePolicyRulesV2) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *InferencePolicyRulesV2) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *InferencePolicyRulesV2) GetImagePullSecrets() ImagePullSecretsRules`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *InferencePolicyRulesV2) GetImagePullSecretsOk() (*ImagePullSecretsRules, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *InferencePolicyRulesV2) SetImagePullSecrets(v ImagePullSecretsRules)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *InferencePolicyRulesV2) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *InferencePolicyRulesV2) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *InferencePolicyRulesV2) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *InferencePolicyRulesV2) GetLabels() InstancesRules`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *InferencePolicyRulesV2) GetLabelsOk() (*InstancesRules, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *InferencePolicyRulesV2) SetLabels(v InstancesRules)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *InferencePolicyRulesV2) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *InferencePolicyRulesV2) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *InferencePolicyRulesV2) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetNodeAffinityRequired

`func (o *InferencePolicyRulesV2) GetNodeAffinityRequired() ArbitraryRules`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *InferencePolicyRulesV2) GetNodeAffinityRequiredOk() (*ArbitraryRules, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *InferencePolicyRulesV2) SetNodeAffinityRequired(v ArbitraryRules)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *InferencePolicyRulesV2) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *InferencePolicyRulesV2) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *InferencePolicyRulesV2) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *InferencePolicyRulesV2) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *InferencePolicyRulesV2) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *InferencePolicyRulesV2) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *InferencePolicyRulesV2) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *InferencePolicyRulesV2) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *InferencePolicyRulesV2) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *InferencePolicyRulesV2) GetNodeType() StringRules`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *InferencePolicyRulesV2) GetNodeTypeOk() (*StringRules, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *InferencePolicyRulesV2) SetNodeType(v StringRules)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *InferencePolicyRulesV2) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *InferencePolicyRulesV2) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *InferencePolicyRulesV2) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetPodAffinity

`func (o *InferencePolicyRulesV2) GetPodAffinity() PodAffinityRules`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *InferencePolicyRulesV2) GetPodAffinityOk() (*PodAffinityRules, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *InferencePolicyRulesV2) SetPodAffinity(v PodAffinityRules)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *InferencePolicyRulesV2) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *InferencePolicyRulesV2) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *InferencePolicyRulesV2) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *InferencePolicyRulesV2) GetPorts() PortsRules`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *InferencePolicyRulesV2) GetPortsOk() (*PortsRules, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *InferencePolicyRulesV2) SetPorts(v PortsRules)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *InferencePolicyRulesV2) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *InferencePolicyRulesV2) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *InferencePolicyRulesV2) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetPreemptibility

`func (o *InferencePolicyRulesV2) GetPreemptibility() PreemptibilityRules`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *InferencePolicyRulesV2) GetPreemptibilityOk() (*PreemptibilityRules, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *InferencePolicyRulesV2) SetPreemptibility(v PreemptibilityRules)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *InferencePolicyRulesV2) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *InferencePolicyRulesV2) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *InferencePolicyRulesV2) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *InferencePolicyRulesV2) GetPriorityClass() StringRules`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *InferencePolicyRulesV2) GetPriorityClassOk() (*StringRules, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *InferencePolicyRulesV2) SetPriorityClass(v StringRules)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *InferencePolicyRulesV2) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *InferencePolicyRulesV2) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *InferencePolicyRulesV2) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *InferencePolicyRulesV2) GetProbes() ProbesRules`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *InferencePolicyRulesV2) GetProbesOk() (*ProbesRules, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *InferencePolicyRulesV2) SetProbes(v ProbesRules)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *InferencePolicyRulesV2) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *InferencePolicyRulesV2) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *InferencePolicyRulesV2) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *InferencePolicyRulesV2) GetRelatedUrls() RelatedUrlsRules`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *InferencePolicyRulesV2) GetRelatedUrlsOk() (*RelatedUrlsRules, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *InferencePolicyRulesV2) SetRelatedUrls(v RelatedUrlsRules)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *InferencePolicyRulesV2) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *InferencePolicyRulesV2) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *InferencePolicyRulesV2) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetSecurity

`func (o *InferencePolicyRulesV2) GetSecurity() InferencePolicyRulesV2AllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *InferencePolicyRulesV2) GetSecurityOk() (*InferencePolicyRulesV2AllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *InferencePolicyRulesV2) SetSecurity(v InferencePolicyRulesV2AllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *InferencePolicyRulesV2) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *InferencePolicyRulesV2) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *InferencePolicyRulesV2) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetServingPort

`func (o *InferencePolicyRulesV2) GetServingPort() ServingPortRules`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *InferencePolicyRulesV2) GetServingPortOk() (*ServingPortRules, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *InferencePolicyRulesV2) SetServingPort(v ServingPortRules)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *InferencePolicyRulesV2) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *InferencePolicyRulesV2) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *InferencePolicyRulesV2) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetStorage

`func (o *InferencePolicyRulesV2) GetStorage() InferencePolicyRulesV2AllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *InferencePolicyRulesV2) GetStorageOk() (*InferencePolicyRulesV2AllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *InferencePolicyRulesV2) SetStorage(v InferencePolicyRulesV2AllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *InferencePolicyRulesV2) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *InferencePolicyRulesV2) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *InferencePolicyRulesV2) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTolerations

`func (o *InferencePolicyRulesV2) GetTolerations() TolerationsRules`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *InferencePolicyRulesV2) GetTolerationsOk() (*TolerationsRules, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *InferencePolicyRulesV2) SetTolerations(v TolerationsRules)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *InferencePolicyRulesV2) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *InferencePolicyRulesV2) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *InferencePolicyRulesV2) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetWorkingDir

`func (o *InferencePolicyRulesV2) GetWorkingDir() StringRules`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *InferencePolicyRulesV2) GetWorkingDirOk() (*StringRules, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *InferencePolicyRulesV2) SetWorkingDir(v StringRules)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *InferencePolicyRulesV2) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *InferencePolicyRulesV2) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *InferencePolicyRulesV2) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


