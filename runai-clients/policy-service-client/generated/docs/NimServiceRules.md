# NimServiceRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Autoscaling** | Pointer to [**NullableNimServiceRulesAllOfAutoscaling**](NimServiceRulesAllOfAutoscaling.md) |  | [optional] 
**Category** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Compute** | Pointer to [**NullableNimServiceRulesAllOfCompute**](NimServiceRulesAllOfCompute.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Image** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicyRules**](ImagePullPolicyRules.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**NullableImagePullSecretsRules**](ImagePullSecretsRules.md) |  | [optional] 
**Labels** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**ModelStore** | Pointer to [**NullableNimServiceRulesAllOfModelStore**](NimServiceRulesAllOfModelStore.md) |  | [optional] 
**MultiNode** | Pointer to [**NullableNimServiceMultiNodeRules**](NimServiceMultiNodeRules.md) |  | [optional] 
**NgcAuthSecret** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibilityRules**](PreemptibilityRules.md) |  | [optional] 
**PriorityClass** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Probes** | Pointer to [**NullableProbesRules**](ProbesRules.md) |  | [optional] 
**Replicas** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Security** | Pointer to [**NullableNimServiceRulesAllOfSecurity**](NimServiceRulesAllOfSecurity.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableNimServiceServingPortRules**](NimServiceServingPortRules.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsRules**](TolerationsRules.md) |  | [optional] 

## Methods

### NewNimServiceRules

`func NewNimServiceRules() *NimServiceRules`

NewNimServiceRules instantiates a new NimServiceRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceRulesWithDefaults

`func NewNimServiceRulesWithDefaults() *NimServiceRules`

NewNimServiceRulesWithDefaults instantiates a new NimServiceRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *NimServiceRules) GetAnnotations() InstancesRules`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *NimServiceRules) GetAnnotationsOk() (*InstancesRules, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *NimServiceRules) SetAnnotations(v InstancesRules)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *NimServiceRules) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *NimServiceRules) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *NimServiceRules) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetAutoscaling

`func (o *NimServiceRules) GetAutoscaling() NimServiceRulesAllOfAutoscaling`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *NimServiceRules) GetAutoscalingOk() (*NimServiceRulesAllOfAutoscaling, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *NimServiceRules) SetAutoscaling(v NimServiceRulesAllOfAutoscaling)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *NimServiceRules) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *NimServiceRules) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *NimServiceRules) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil
### GetCategory

`func (o *NimServiceRules) GetCategory() StringRules`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *NimServiceRules) GetCategoryOk() (*StringRules, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *NimServiceRules) SetCategory(v StringRules)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *NimServiceRules) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *NimServiceRules) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *NimServiceRules) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCompute

`func (o *NimServiceRules) GetCompute() NimServiceRulesAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *NimServiceRules) GetComputeOk() (*NimServiceRulesAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *NimServiceRules) SetCompute(v NimServiceRulesAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *NimServiceRules) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *NimServiceRules) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *NimServiceRules) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetEnvironmentVariables

`func (o *NimServiceRules) GetEnvironmentVariables() InstancesRules`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *NimServiceRules) GetEnvironmentVariablesOk() (*InstancesRules, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *NimServiceRules) SetEnvironmentVariables(v InstancesRules)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *NimServiceRules) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *NimServiceRules) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *NimServiceRules) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetImage

`func (o *NimServiceRules) GetImage() StringRules`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *NimServiceRules) GetImageOk() (*StringRules, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *NimServiceRules) SetImage(v StringRules)`

SetImage sets Image field to given value.

### HasImage

`func (o *NimServiceRules) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *NimServiceRules) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *NimServiceRules) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *NimServiceRules) GetImagePullPolicy() ImagePullPolicyRules`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *NimServiceRules) GetImagePullPolicyOk() (*ImagePullPolicyRules, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *NimServiceRules) SetImagePullPolicy(v ImagePullPolicyRules)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *NimServiceRules) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *NimServiceRules) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *NimServiceRules) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *NimServiceRules) GetImagePullSecrets() ImagePullSecretsRules`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *NimServiceRules) GetImagePullSecretsOk() (*ImagePullSecretsRules, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *NimServiceRules) SetImagePullSecrets(v ImagePullSecretsRules)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *NimServiceRules) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *NimServiceRules) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *NimServiceRules) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *NimServiceRules) GetLabels() InstancesRules`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *NimServiceRules) GetLabelsOk() (*InstancesRules, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *NimServiceRules) SetLabels(v InstancesRules)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *NimServiceRules) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *NimServiceRules) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *NimServiceRules) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetModelStore

`func (o *NimServiceRules) GetModelStore() NimServiceRulesAllOfModelStore`

GetModelStore returns the ModelStore field if non-nil, zero value otherwise.

### GetModelStoreOk

`func (o *NimServiceRules) GetModelStoreOk() (*NimServiceRulesAllOfModelStore, bool)`

GetModelStoreOk returns a tuple with the ModelStore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelStore

`func (o *NimServiceRules) SetModelStore(v NimServiceRulesAllOfModelStore)`

SetModelStore sets ModelStore field to given value.

### HasModelStore

`func (o *NimServiceRules) HasModelStore() bool`

HasModelStore returns a boolean if a field has been set.

### SetModelStoreNil

`func (o *NimServiceRules) SetModelStoreNil(b bool)`

 SetModelStoreNil sets the value for ModelStore to be an explicit nil

### UnsetModelStore
`func (o *NimServiceRules) UnsetModelStore()`

UnsetModelStore ensures that no value is present for ModelStore, not even an explicit nil
### GetMultiNode

`func (o *NimServiceRules) GetMultiNode() NimServiceMultiNodeRules`

GetMultiNode returns the MultiNode field if non-nil, zero value otherwise.

### GetMultiNodeOk

`func (o *NimServiceRules) GetMultiNodeOk() (*NimServiceMultiNodeRules, bool)`

GetMultiNodeOk returns a tuple with the MultiNode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMultiNode

`func (o *NimServiceRules) SetMultiNode(v NimServiceMultiNodeRules)`

SetMultiNode sets MultiNode field to given value.

### HasMultiNode

`func (o *NimServiceRules) HasMultiNode() bool`

HasMultiNode returns a boolean if a field has been set.

### SetMultiNodeNil

`func (o *NimServiceRules) SetMultiNodeNil(b bool)`

 SetMultiNodeNil sets the value for MultiNode to be an explicit nil

### UnsetMultiNode
`func (o *NimServiceRules) UnsetMultiNode()`

UnsetMultiNode ensures that no value is present for MultiNode, not even an explicit nil
### GetNgcAuthSecret

`func (o *NimServiceRules) GetNgcAuthSecret() StringRules`

GetNgcAuthSecret returns the NgcAuthSecret field if non-nil, zero value otherwise.

### GetNgcAuthSecretOk

`func (o *NimServiceRules) GetNgcAuthSecretOk() (*StringRules, bool)`

GetNgcAuthSecretOk returns a tuple with the NgcAuthSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNgcAuthSecret

`func (o *NimServiceRules) SetNgcAuthSecret(v StringRules)`

SetNgcAuthSecret sets NgcAuthSecret field to given value.

### HasNgcAuthSecret

`func (o *NimServiceRules) HasNgcAuthSecret() bool`

HasNgcAuthSecret returns a boolean if a field has been set.

### SetNgcAuthSecretNil

`func (o *NimServiceRules) SetNgcAuthSecretNil(b bool)`

 SetNgcAuthSecretNil sets the value for NgcAuthSecret to be an explicit nil

### UnsetNgcAuthSecret
`func (o *NimServiceRules) UnsetNgcAuthSecret()`

UnsetNgcAuthSecret ensures that no value is present for NgcAuthSecret, not even an explicit nil
### GetNodePools

`func (o *NimServiceRules) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *NimServiceRules) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *NimServiceRules) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *NimServiceRules) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *NimServiceRules) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *NimServiceRules) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *NimServiceRules) GetPreemptibility() PreemptibilityRules`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *NimServiceRules) GetPreemptibilityOk() (*PreemptibilityRules, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *NimServiceRules) SetPreemptibility(v PreemptibilityRules)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *NimServiceRules) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *NimServiceRules) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *NimServiceRules) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *NimServiceRules) GetPriorityClass() StringRules`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *NimServiceRules) GetPriorityClassOk() (*StringRules, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *NimServiceRules) SetPriorityClass(v StringRules)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *NimServiceRules) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *NimServiceRules) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *NimServiceRules) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *NimServiceRules) GetProbes() ProbesRules`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *NimServiceRules) GetProbesOk() (*ProbesRules, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *NimServiceRules) SetProbes(v ProbesRules)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *NimServiceRules) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *NimServiceRules) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *NimServiceRules) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetReplicas

`func (o *NimServiceRules) GetReplicas() IntegerRules`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *NimServiceRules) GetReplicasOk() (*IntegerRules, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *NimServiceRules) SetReplicas(v IntegerRules)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *NimServiceRules) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### SetReplicasNil

`func (o *NimServiceRules) SetReplicasNil(b bool)`

 SetReplicasNil sets the value for Replicas to be an explicit nil

### UnsetReplicas
`func (o *NimServiceRules) UnsetReplicas()`

UnsetReplicas ensures that no value is present for Replicas, not even an explicit nil
### GetSecurity

`func (o *NimServiceRules) GetSecurity() NimServiceRulesAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *NimServiceRules) GetSecurityOk() (*NimServiceRulesAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *NimServiceRules) SetSecurity(v NimServiceRulesAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *NimServiceRules) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *NimServiceRules) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *NimServiceRules) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetServingPort

`func (o *NimServiceRules) GetServingPort() NimServiceServingPortRules`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *NimServiceRules) GetServingPortOk() (*NimServiceServingPortRules, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *NimServiceRules) SetServingPort(v NimServiceServingPortRules)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *NimServiceRules) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *NimServiceRules) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *NimServiceRules) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetTolerations

`func (o *NimServiceRules) GetTolerations() TolerationsRules`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *NimServiceRules) GetTolerationsOk() (*TolerationsRules, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *NimServiceRules) SetTolerations(v TolerationsRules)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *NimServiceRules) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *NimServiceRules) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *NimServiceRules) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


