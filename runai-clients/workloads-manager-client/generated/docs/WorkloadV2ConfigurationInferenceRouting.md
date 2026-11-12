# WorkloadV2ConfigurationInferenceRouting

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LoadAwareRouting** | Pointer to **NullableBool** | Route requests to the serving port through the endpoint picker, which picks a Pod by load. It can be set only when the workload is created. | [optional] [default to false]

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

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


