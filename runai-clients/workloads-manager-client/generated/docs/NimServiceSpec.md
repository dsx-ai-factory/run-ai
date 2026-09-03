# NimServiceSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workload. | [optional] 
**Autoscaling** | Pointer to [**NullableNimServiceSpecAllOfAutoscaling**](NimServiceSpecAllOfAutoscaling.md) |  | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Compute** | Pointer to [**NullableNimServiceSpecAllOfCompute**](NimServiceSpecAllOfCompute.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workload. | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workload. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**[]ImagePullSecret**](ImagePullSecret.md) | A list of references to Kubernetes secrets in the same namespace used for pulling container images. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workload. | [optional] 
**ModelStore** | Pointer to [**NullableNimServiceSpecAllOfModelStore**](NimServiceSpecAllOfModelStore.md) |  | [optional] 
**MultiNode** | Pointer to [**NullableNimServiceMultiNode**](NimServiceMultiNode.md) |  | [optional] 
**NgcAuthSecret** | Pointer to **NullableString** | The name of a Kubernetes secret containing the NGC access credentials. The secret must contain a key named NGC_API_KEY with the API key as the value. | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**Replicas** | Pointer to **NullableInt32** | The number of replicas to deploy. | [optional] [default to 1]
**Security** | Pointer to [**NullableNimServiceSpecAllOfSecurity**](NimServiceSpecAllOfSecurity.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableNimServiceServingPort**](NimServiceServingPort.md) |  | [optional] 
**Tolerations** | Pointer to [**[]Toleration**](Toleration.md) | Set of tolerations to apply to the workload. | [optional] 

## Methods

### NewNimServiceSpec

`func NewNimServiceSpec() *NimServiceSpec`

NewNimServiceSpec instantiates a new NimServiceSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceSpecWithDefaults

`func NewNimServiceSpecWithDefaults() *NimServiceSpec`

NewNimServiceSpecWithDefaults instantiates a new NimServiceSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *NimServiceSpec) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *NimServiceSpec) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *NimServiceSpec) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *NimServiceSpec) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *NimServiceSpec) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *NimServiceSpec) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetAutoscaling

`func (o *NimServiceSpec) GetAutoscaling() NimServiceSpecAllOfAutoscaling`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *NimServiceSpec) GetAutoscalingOk() (*NimServiceSpecAllOfAutoscaling, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *NimServiceSpec) SetAutoscaling(v NimServiceSpecAllOfAutoscaling)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *NimServiceSpec) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *NimServiceSpec) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *NimServiceSpec) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil
### GetCategory

`func (o *NimServiceSpec) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *NimServiceSpec) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *NimServiceSpec) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *NimServiceSpec) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *NimServiceSpec) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *NimServiceSpec) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCompute

`func (o *NimServiceSpec) GetCompute() NimServiceSpecAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *NimServiceSpec) GetComputeOk() (*NimServiceSpecAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *NimServiceSpec) SetCompute(v NimServiceSpecAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *NimServiceSpec) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *NimServiceSpec) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *NimServiceSpec) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetEnvironmentVariables

`func (o *NimServiceSpec) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *NimServiceSpec) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *NimServiceSpec) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *NimServiceSpec) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *NimServiceSpec) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *NimServiceSpec) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetImage

`func (o *NimServiceSpec) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *NimServiceSpec) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *NimServiceSpec) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *NimServiceSpec) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *NimServiceSpec) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *NimServiceSpec) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *NimServiceSpec) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *NimServiceSpec) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *NimServiceSpec) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *NimServiceSpec) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *NimServiceSpec) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *NimServiceSpec) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *NimServiceSpec) GetImagePullSecrets() []ImagePullSecret`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *NimServiceSpec) GetImagePullSecretsOk() (*[]ImagePullSecret, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *NimServiceSpec) SetImagePullSecrets(v []ImagePullSecret)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *NimServiceSpec) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *NimServiceSpec) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *NimServiceSpec) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *NimServiceSpec) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *NimServiceSpec) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *NimServiceSpec) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *NimServiceSpec) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *NimServiceSpec) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *NimServiceSpec) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetModelStore

`func (o *NimServiceSpec) GetModelStore() NimServiceSpecAllOfModelStore`

GetModelStore returns the ModelStore field if non-nil, zero value otherwise.

### GetModelStoreOk

`func (o *NimServiceSpec) GetModelStoreOk() (*NimServiceSpecAllOfModelStore, bool)`

GetModelStoreOk returns a tuple with the ModelStore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelStore

`func (o *NimServiceSpec) SetModelStore(v NimServiceSpecAllOfModelStore)`

SetModelStore sets ModelStore field to given value.

### HasModelStore

`func (o *NimServiceSpec) HasModelStore() bool`

HasModelStore returns a boolean if a field has been set.

### SetModelStoreNil

`func (o *NimServiceSpec) SetModelStoreNil(b bool)`

 SetModelStoreNil sets the value for ModelStore to be an explicit nil

### UnsetModelStore
`func (o *NimServiceSpec) UnsetModelStore()`

UnsetModelStore ensures that no value is present for ModelStore, not even an explicit nil
### GetMultiNode

`func (o *NimServiceSpec) GetMultiNode() NimServiceMultiNode`

GetMultiNode returns the MultiNode field if non-nil, zero value otherwise.

### GetMultiNodeOk

`func (o *NimServiceSpec) GetMultiNodeOk() (*NimServiceMultiNode, bool)`

GetMultiNodeOk returns a tuple with the MultiNode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMultiNode

`func (o *NimServiceSpec) SetMultiNode(v NimServiceMultiNode)`

SetMultiNode sets MultiNode field to given value.

### HasMultiNode

`func (o *NimServiceSpec) HasMultiNode() bool`

HasMultiNode returns a boolean if a field has been set.

### SetMultiNodeNil

`func (o *NimServiceSpec) SetMultiNodeNil(b bool)`

 SetMultiNodeNil sets the value for MultiNode to be an explicit nil

### UnsetMultiNode
`func (o *NimServiceSpec) UnsetMultiNode()`

UnsetMultiNode ensures that no value is present for MultiNode, not even an explicit nil
### GetNgcAuthSecret

`func (o *NimServiceSpec) GetNgcAuthSecret() string`

GetNgcAuthSecret returns the NgcAuthSecret field if non-nil, zero value otherwise.

### GetNgcAuthSecretOk

`func (o *NimServiceSpec) GetNgcAuthSecretOk() (*string, bool)`

GetNgcAuthSecretOk returns a tuple with the NgcAuthSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNgcAuthSecret

`func (o *NimServiceSpec) SetNgcAuthSecret(v string)`

SetNgcAuthSecret sets NgcAuthSecret field to given value.

### HasNgcAuthSecret

`func (o *NimServiceSpec) HasNgcAuthSecret() bool`

HasNgcAuthSecret returns a boolean if a field has been set.

### SetNgcAuthSecretNil

`func (o *NimServiceSpec) SetNgcAuthSecretNil(b bool)`

 SetNgcAuthSecretNil sets the value for NgcAuthSecret to be an explicit nil

### UnsetNgcAuthSecret
`func (o *NimServiceSpec) UnsetNgcAuthSecret()`

UnsetNgcAuthSecret ensures that no value is present for NgcAuthSecret, not even an explicit nil
### GetNodePools

`func (o *NimServiceSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *NimServiceSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *NimServiceSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *NimServiceSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *NimServiceSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *NimServiceSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *NimServiceSpec) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *NimServiceSpec) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *NimServiceSpec) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *NimServiceSpec) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *NimServiceSpec) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *NimServiceSpec) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *NimServiceSpec) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *NimServiceSpec) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *NimServiceSpec) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *NimServiceSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *NimServiceSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *NimServiceSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *NimServiceSpec) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *NimServiceSpec) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *NimServiceSpec) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *NimServiceSpec) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *NimServiceSpec) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *NimServiceSpec) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetReplicas

`func (o *NimServiceSpec) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *NimServiceSpec) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *NimServiceSpec) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *NimServiceSpec) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### SetReplicasNil

`func (o *NimServiceSpec) SetReplicasNil(b bool)`

 SetReplicasNil sets the value for Replicas to be an explicit nil

### UnsetReplicas
`func (o *NimServiceSpec) UnsetReplicas()`

UnsetReplicas ensures that no value is present for Replicas, not even an explicit nil
### GetSecurity

`func (o *NimServiceSpec) GetSecurity() NimServiceSpecAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *NimServiceSpec) GetSecurityOk() (*NimServiceSpecAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *NimServiceSpec) SetSecurity(v NimServiceSpecAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *NimServiceSpec) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *NimServiceSpec) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *NimServiceSpec) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetServingPort

`func (o *NimServiceSpec) GetServingPort() NimServiceServingPort`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *NimServiceSpec) GetServingPortOk() (*NimServiceServingPort, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *NimServiceSpec) SetServingPort(v NimServiceServingPort)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *NimServiceSpec) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *NimServiceSpec) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *NimServiceSpec) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetTolerations

`func (o *NimServiceSpec) GetTolerations() []Toleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *NimServiceSpec) GetTolerationsOk() (*[]Toleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *NimServiceSpec) SetTolerations(v []Toleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *NimServiceSpec) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *NimServiceSpec) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *NimServiceSpec) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


