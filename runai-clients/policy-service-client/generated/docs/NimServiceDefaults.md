# NimServiceDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**NullableAnnotationsDefaults**](AnnotationsDefaults.md) |  | [optional] 
**Autoscaling** | Pointer to [**NullableNimServiceDefaultsAllOfAutoscaling**](NimServiceDefaultsAllOfAutoscaling.md) |  | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Compute** | Pointer to [**NullableNimServiceDefaultsAllOfCompute**](NimServiceDefaultsAllOfCompute.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableEnvironmentVariablesDefaults**](EnvironmentVariablesDefaults.md) |  | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workload. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**NullableImagePullSecretsDefaults**](ImagePullSecretsDefaults.md) |  | [optional] 
**Labels** | Pointer to [**NullableLabelsDefaults**](LabelsDefaults.md) |  | [optional] 
**ModelStore** | Pointer to [**NullableNimServiceDefaultsAllOfModelStore**](NimServiceDefaultsAllOfModelStore.md) |  | [optional] 
**MultiNode** | Pointer to [**NullableNimServiceMultiNode**](NimServiceMultiNode.md) |  | [optional] 
**NgcAuthSecret** | Pointer to **NullableString** | The name of a Kubernetes secret containing the NGC access credentials. The secret must contain a key named NGC_API_KEY with the API key as the value. | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**Replicas** | Pointer to **NullableInt32** | The number of replicas to deploy. | [optional] [default to 1]
**Security** | Pointer to [**NullableNimServiceDefaultsAllOfSecurity**](NimServiceDefaultsAllOfSecurity.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableNimServiceServingPort**](NimServiceServingPort.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsDefaults**](TolerationsDefaults.md) |  | [optional] 

## Methods

### NewNimServiceDefaults

`func NewNimServiceDefaults() *NimServiceDefaults`

NewNimServiceDefaults instantiates a new NimServiceDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceDefaultsWithDefaults

`func NewNimServiceDefaultsWithDefaults() *NimServiceDefaults`

NewNimServiceDefaultsWithDefaults instantiates a new NimServiceDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *NimServiceDefaults) GetAnnotations() AnnotationsDefaults`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *NimServiceDefaults) GetAnnotationsOk() (*AnnotationsDefaults, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *NimServiceDefaults) SetAnnotations(v AnnotationsDefaults)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *NimServiceDefaults) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *NimServiceDefaults) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *NimServiceDefaults) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetAutoscaling

`func (o *NimServiceDefaults) GetAutoscaling() NimServiceDefaultsAllOfAutoscaling`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *NimServiceDefaults) GetAutoscalingOk() (*NimServiceDefaultsAllOfAutoscaling, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *NimServiceDefaults) SetAutoscaling(v NimServiceDefaultsAllOfAutoscaling)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *NimServiceDefaults) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *NimServiceDefaults) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *NimServiceDefaults) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil
### GetCategory

`func (o *NimServiceDefaults) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *NimServiceDefaults) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *NimServiceDefaults) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *NimServiceDefaults) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *NimServiceDefaults) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *NimServiceDefaults) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCompute

`func (o *NimServiceDefaults) GetCompute() NimServiceDefaultsAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *NimServiceDefaults) GetComputeOk() (*NimServiceDefaultsAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *NimServiceDefaults) SetCompute(v NimServiceDefaultsAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *NimServiceDefaults) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *NimServiceDefaults) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *NimServiceDefaults) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetEnvironmentVariables

`func (o *NimServiceDefaults) GetEnvironmentVariables() EnvironmentVariablesDefaults`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *NimServiceDefaults) GetEnvironmentVariablesOk() (*EnvironmentVariablesDefaults, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *NimServiceDefaults) SetEnvironmentVariables(v EnvironmentVariablesDefaults)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *NimServiceDefaults) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *NimServiceDefaults) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *NimServiceDefaults) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetImage

`func (o *NimServiceDefaults) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *NimServiceDefaults) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *NimServiceDefaults) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *NimServiceDefaults) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *NimServiceDefaults) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *NimServiceDefaults) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *NimServiceDefaults) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *NimServiceDefaults) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *NimServiceDefaults) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *NimServiceDefaults) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *NimServiceDefaults) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *NimServiceDefaults) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *NimServiceDefaults) GetImagePullSecrets() ImagePullSecretsDefaults`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *NimServiceDefaults) GetImagePullSecretsOk() (*ImagePullSecretsDefaults, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *NimServiceDefaults) SetImagePullSecrets(v ImagePullSecretsDefaults)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *NimServiceDefaults) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *NimServiceDefaults) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *NimServiceDefaults) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *NimServiceDefaults) GetLabels() LabelsDefaults`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *NimServiceDefaults) GetLabelsOk() (*LabelsDefaults, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *NimServiceDefaults) SetLabels(v LabelsDefaults)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *NimServiceDefaults) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *NimServiceDefaults) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *NimServiceDefaults) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetModelStore

`func (o *NimServiceDefaults) GetModelStore() NimServiceDefaultsAllOfModelStore`

GetModelStore returns the ModelStore field if non-nil, zero value otherwise.

### GetModelStoreOk

`func (o *NimServiceDefaults) GetModelStoreOk() (*NimServiceDefaultsAllOfModelStore, bool)`

GetModelStoreOk returns a tuple with the ModelStore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelStore

`func (o *NimServiceDefaults) SetModelStore(v NimServiceDefaultsAllOfModelStore)`

SetModelStore sets ModelStore field to given value.

### HasModelStore

`func (o *NimServiceDefaults) HasModelStore() bool`

HasModelStore returns a boolean if a field has been set.

### SetModelStoreNil

`func (o *NimServiceDefaults) SetModelStoreNil(b bool)`

 SetModelStoreNil sets the value for ModelStore to be an explicit nil

### UnsetModelStore
`func (o *NimServiceDefaults) UnsetModelStore()`

UnsetModelStore ensures that no value is present for ModelStore, not even an explicit nil
### GetMultiNode

`func (o *NimServiceDefaults) GetMultiNode() NimServiceMultiNode`

GetMultiNode returns the MultiNode field if non-nil, zero value otherwise.

### GetMultiNodeOk

`func (o *NimServiceDefaults) GetMultiNodeOk() (*NimServiceMultiNode, bool)`

GetMultiNodeOk returns a tuple with the MultiNode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMultiNode

`func (o *NimServiceDefaults) SetMultiNode(v NimServiceMultiNode)`

SetMultiNode sets MultiNode field to given value.

### HasMultiNode

`func (o *NimServiceDefaults) HasMultiNode() bool`

HasMultiNode returns a boolean if a field has been set.

### SetMultiNodeNil

`func (o *NimServiceDefaults) SetMultiNodeNil(b bool)`

 SetMultiNodeNil sets the value for MultiNode to be an explicit nil

### UnsetMultiNode
`func (o *NimServiceDefaults) UnsetMultiNode()`

UnsetMultiNode ensures that no value is present for MultiNode, not even an explicit nil
### GetNgcAuthSecret

`func (o *NimServiceDefaults) GetNgcAuthSecret() string`

GetNgcAuthSecret returns the NgcAuthSecret field if non-nil, zero value otherwise.

### GetNgcAuthSecretOk

`func (o *NimServiceDefaults) GetNgcAuthSecretOk() (*string, bool)`

GetNgcAuthSecretOk returns a tuple with the NgcAuthSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNgcAuthSecret

`func (o *NimServiceDefaults) SetNgcAuthSecret(v string)`

SetNgcAuthSecret sets NgcAuthSecret field to given value.

### HasNgcAuthSecret

`func (o *NimServiceDefaults) HasNgcAuthSecret() bool`

HasNgcAuthSecret returns a boolean if a field has been set.

### SetNgcAuthSecretNil

`func (o *NimServiceDefaults) SetNgcAuthSecretNil(b bool)`

 SetNgcAuthSecretNil sets the value for NgcAuthSecret to be an explicit nil

### UnsetNgcAuthSecret
`func (o *NimServiceDefaults) UnsetNgcAuthSecret()`

UnsetNgcAuthSecret ensures that no value is present for NgcAuthSecret, not even an explicit nil
### GetNodePools

`func (o *NimServiceDefaults) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *NimServiceDefaults) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *NimServiceDefaults) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *NimServiceDefaults) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *NimServiceDefaults) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *NimServiceDefaults) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *NimServiceDefaults) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *NimServiceDefaults) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *NimServiceDefaults) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *NimServiceDefaults) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *NimServiceDefaults) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *NimServiceDefaults) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *NimServiceDefaults) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *NimServiceDefaults) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *NimServiceDefaults) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *NimServiceDefaults) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *NimServiceDefaults) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *NimServiceDefaults) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *NimServiceDefaults) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *NimServiceDefaults) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *NimServiceDefaults) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *NimServiceDefaults) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *NimServiceDefaults) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *NimServiceDefaults) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetReplicas

`func (o *NimServiceDefaults) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *NimServiceDefaults) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *NimServiceDefaults) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *NimServiceDefaults) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### SetReplicasNil

`func (o *NimServiceDefaults) SetReplicasNil(b bool)`

 SetReplicasNil sets the value for Replicas to be an explicit nil

### UnsetReplicas
`func (o *NimServiceDefaults) UnsetReplicas()`

UnsetReplicas ensures that no value is present for Replicas, not even an explicit nil
### GetSecurity

`func (o *NimServiceDefaults) GetSecurity() NimServiceDefaultsAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *NimServiceDefaults) GetSecurityOk() (*NimServiceDefaultsAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *NimServiceDefaults) SetSecurity(v NimServiceDefaultsAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *NimServiceDefaults) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *NimServiceDefaults) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *NimServiceDefaults) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetServingPort

`func (o *NimServiceDefaults) GetServingPort() NimServiceServingPort`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *NimServiceDefaults) GetServingPortOk() (*NimServiceServingPort, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *NimServiceDefaults) SetServingPort(v NimServiceServingPort)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *NimServiceDefaults) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *NimServiceDefaults) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *NimServiceDefaults) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetTolerations

`func (o *NimServiceDefaults) GetTolerations() TolerationsDefaults`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *NimServiceDefaults) GetTolerationsOk() (*TolerationsDefaults, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *NimServiceDefaults) SetTolerations(v TolerationsDefaults)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *NimServiceDefaults) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *NimServiceDefaults) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *NimServiceDefaults) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


