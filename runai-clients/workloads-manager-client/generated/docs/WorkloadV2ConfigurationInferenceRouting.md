# WorkloadV2ConfigurationInferenceRouting

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LoadAwareRouting** | Pointer to **NullableBool** | Route requests to the serving port through the endpoint picker, which picks a Pod by load. Setting objectivePriority or modelRewrites also enables it. | [optional] [default to false]
**ObjectivePriority** | Pointer to **NullableInt32** | The priority of this workload&#39;s inference objective. The project policy sets the maximum value. | [optional] 
**ModelRewrites** | Pointer to [**[]ModelRewrite**](ModelRewrite.md) | Model name rewrites that the inference endpoint applies at runtime. Each rule maps a model name that clients request to one or more served model names. One target makes an alias; several targets split traffic by weight. | [optional] 
**ServingPortAuthorizationType** | Pointer to [**NullableServingPortAccessAuthorizationTypeEnum**](ServingPortAccessAuthorizationTypeEnum.md) |  | [optional] 

## Methods

### NewWorkloadV2ConfigurationInferenceRouting

`func NewWorkloadV2ConfigurationInferenceRouting() *WorkloadV2ConfigurationInferenceRouting`

NewWorkloadV2ConfigurationInferenceRouting instantiates a new WorkloadV2ConfigurationInferenceRouting object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2ConfigurationInferenceRoutingWithDefaults

`func NewWorkloadV2ConfigurationInferenceRoutingWithDefaults() *WorkloadV2ConfigurationInferenceRouting`

NewWorkloadV2ConfigurationInferenceRoutingWithDefaults instantiates a new WorkloadV2ConfigurationInferenceRouting object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLoadAwareRouting

`func (o *WorkloadV2ConfigurationInferenceRouting) GetLoadAwareRouting() bool`

GetLoadAwareRouting returns the LoadAwareRouting field if non-nil, zero value otherwise.

### GetLoadAwareRoutingOk

`func (o *WorkloadV2ConfigurationInferenceRouting) GetLoadAwareRoutingOk() (*bool, bool)`

GetLoadAwareRoutingOk returns a tuple with the LoadAwareRouting field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLoadAwareRouting

`func (o *WorkloadV2ConfigurationInferenceRouting) SetLoadAwareRouting(v bool)`

SetLoadAwareRouting sets LoadAwareRouting field to given value.

### HasLoadAwareRouting

`func (o *WorkloadV2ConfigurationInferenceRouting) HasLoadAwareRouting() bool`

HasLoadAwareRouting returns a boolean if a field has been set.

### SetLoadAwareRoutingNil

`func (o *WorkloadV2ConfigurationInferenceRouting) SetLoadAwareRoutingNil(b bool)`

 SetLoadAwareRoutingNil sets the value for LoadAwareRouting to be an explicit nil

### UnsetLoadAwareRouting
`func (o *WorkloadV2ConfigurationInferenceRouting) UnsetLoadAwareRouting()`

UnsetLoadAwareRouting ensures that no value is present for LoadAwareRouting, not even an explicit nil
### GetObjectivePriority

`func (o *WorkloadV2ConfigurationInferenceRouting) GetObjectivePriority() int32`

GetObjectivePriority returns the ObjectivePriority field if non-nil, zero value otherwise.

### GetObjectivePriorityOk

`func (o *WorkloadV2ConfigurationInferenceRouting) GetObjectivePriorityOk() (*int32, bool)`

GetObjectivePriorityOk returns a tuple with the ObjectivePriority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectivePriority

`func (o *WorkloadV2ConfigurationInferenceRouting) SetObjectivePriority(v int32)`

SetObjectivePriority sets ObjectivePriority field to given value.

### HasObjectivePriority

`func (o *WorkloadV2ConfigurationInferenceRouting) HasObjectivePriority() bool`

HasObjectivePriority returns a boolean if a field has been set.

### SetObjectivePriorityNil

`func (o *WorkloadV2ConfigurationInferenceRouting) SetObjectivePriorityNil(b bool)`

 SetObjectivePriorityNil sets the value for ObjectivePriority to be an explicit nil

### UnsetObjectivePriority
`func (o *WorkloadV2ConfigurationInferenceRouting) UnsetObjectivePriority()`

UnsetObjectivePriority ensures that no value is present for ObjectivePriority, not even an explicit nil
### GetModelRewrites

`func (o *WorkloadV2ConfigurationInferenceRouting) GetModelRewrites() []ModelRewrite`

GetModelRewrites returns the ModelRewrites field if non-nil, zero value otherwise.

### GetModelRewritesOk

`func (o *WorkloadV2ConfigurationInferenceRouting) GetModelRewritesOk() (*[]ModelRewrite, bool)`

GetModelRewritesOk returns a tuple with the ModelRewrites field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelRewrites

`func (o *WorkloadV2ConfigurationInferenceRouting) SetModelRewrites(v []ModelRewrite)`

SetModelRewrites sets ModelRewrites field to given value.

### HasModelRewrites

`func (o *WorkloadV2ConfigurationInferenceRouting) HasModelRewrites() bool`

HasModelRewrites returns a boolean if a field has been set.

### SetModelRewritesNil

`func (o *WorkloadV2ConfigurationInferenceRouting) SetModelRewritesNil(b bool)`

 SetModelRewritesNil sets the value for ModelRewrites to be an explicit nil

### UnsetModelRewrites
`func (o *WorkloadV2ConfigurationInferenceRouting) UnsetModelRewrites()`

UnsetModelRewrites ensures that no value is present for ModelRewrites, not even an explicit nil
### GetServingPortAuthorizationType

`func (o *WorkloadV2ConfigurationInferenceRouting) GetServingPortAuthorizationType() ServingPortAccessAuthorizationTypeEnum`

GetServingPortAuthorizationType returns the ServingPortAuthorizationType field if non-nil, zero value otherwise.

### GetServingPortAuthorizationTypeOk

`func (o *WorkloadV2ConfigurationInferenceRouting) GetServingPortAuthorizationTypeOk() (*ServingPortAccessAuthorizationTypeEnum, bool)`

GetServingPortAuthorizationTypeOk returns a tuple with the ServingPortAuthorizationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPortAuthorizationType

`func (o *WorkloadV2ConfigurationInferenceRouting) SetServingPortAuthorizationType(v ServingPortAccessAuthorizationTypeEnum)`

SetServingPortAuthorizationType sets ServingPortAuthorizationType field to given value.

### HasServingPortAuthorizationType

`func (o *WorkloadV2ConfigurationInferenceRouting) HasServingPortAuthorizationType() bool`

HasServingPortAuthorizationType returns a boolean if a field has been set.

### SetServingPortAuthorizationTypeNil

`func (o *WorkloadV2ConfigurationInferenceRouting) SetServingPortAuthorizationTypeNil(b bool)`

 SetServingPortAuthorizationTypeNil sets the value for ServingPortAuthorizationType to be an explicit nil

### UnsetServingPortAuthorizationType
`func (o *WorkloadV2ConfigurationInferenceRouting) UnsetServingPortAuthorizationType()`

UnsetServingPortAuthorizationType ensures that no value is present for ServingPortAuthorizationType, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


