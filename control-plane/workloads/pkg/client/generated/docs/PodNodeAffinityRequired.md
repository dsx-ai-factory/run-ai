# PodNodeAffinityRequired

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodeSelectorTerms** | Pointer to [**[]NodeSelectorTerm**](NodeSelectorTerm.md) | A list of node selector terms. The terms are ORed. | [optional] 

## Methods

### NewPodNodeAffinityRequired

`func NewPodNodeAffinityRequired() *PodNodeAffinityRequired`

NewPodNodeAffinityRequired instantiates a new PodNodeAffinityRequired object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodNodeAffinityRequiredWithDefaults

`func NewPodNodeAffinityRequiredWithDefaults() *PodNodeAffinityRequired`

NewPodNodeAffinityRequiredWithDefaults instantiates a new PodNodeAffinityRequired object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodeSelectorTerms

`func (o *PodNodeAffinityRequired) GetNodeSelectorTerms() []NodeSelectorTerm`

GetNodeSelectorTerms returns the NodeSelectorTerms field if non-nil, zero value otherwise.

### GetNodeSelectorTermsOk

`func (o *PodNodeAffinityRequired) GetNodeSelectorTermsOk() (*[]NodeSelectorTerm, bool)`

GetNodeSelectorTermsOk returns a tuple with the NodeSelectorTerms field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeSelectorTerms

`func (o *PodNodeAffinityRequired) SetNodeSelectorTerms(v []NodeSelectorTerm)`

SetNodeSelectorTerms sets NodeSelectorTerms field to given value.

### HasNodeSelectorTerms

`func (o *PodNodeAffinityRequired) HasNodeSelectorTerms() bool`

HasNodeSelectorTerms returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


