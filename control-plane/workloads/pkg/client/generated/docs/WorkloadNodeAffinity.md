# WorkloadNodeAffinity

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SpecName** | **string** | The name of which spec this node affinity is referring to. | 
**NodeAffinity** | Pointer to [**PodNodeAffinity**](PodNodeAffinity.md) |  | [optional] 

## Methods

### NewWorkloadNodeAffinity

`func NewWorkloadNodeAffinity(specName string, ) *WorkloadNodeAffinity`

NewWorkloadNodeAffinity instantiates a new WorkloadNodeAffinity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadNodeAffinityWithDefaults

`func NewWorkloadNodeAffinityWithDefaults() *WorkloadNodeAffinity`

NewWorkloadNodeAffinityWithDefaults instantiates a new WorkloadNodeAffinity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpecName

`func (o *WorkloadNodeAffinity) GetSpecName() string`

GetSpecName returns the SpecName field if non-nil, zero value otherwise.

### GetSpecNameOk

`func (o *WorkloadNodeAffinity) GetSpecNameOk() (*string, bool)`

GetSpecNameOk returns a tuple with the SpecName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecName

`func (o *WorkloadNodeAffinity) SetSpecName(v string)`

SetSpecName sets SpecName field to given value.


### GetNodeAffinity

`func (o *WorkloadNodeAffinity) GetNodeAffinity() PodNodeAffinity`

GetNodeAffinity returns the NodeAffinity field if non-nil, zero value otherwise.

### GetNodeAffinityOk

`func (o *WorkloadNodeAffinity) GetNodeAffinityOk() (*PodNodeAffinity, bool)`

GetNodeAffinityOk returns a tuple with the NodeAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinity

`func (o *WorkloadNodeAffinity) SetNodeAffinity(v PodNodeAffinity)`

SetNodeAffinity sets NodeAffinity field to given value.

### HasNodeAffinity

`func (o *WorkloadNodeAffinity) HasNodeAffinity() bool`

HasNodeAffinity returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


