# PodNodeAffinityPreferred

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodeSelectorTerms** | Pointer to [**[]PodPreferredNodeSelectorTerm**](PodPreferredNodeSelectorTerm.md) | A list of node selector terms. | [optional] 

## Methods

### NewPodNodeAffinityPreferred

`func NewPodNodeAffinityPreferred() *PodNodeAffinityPreferred`

NewPodNodeAffinityPreferred instantiates a new PodNodeAffinityPreferred object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodNodeAffinityPreferredWithDefaults

`func NewPodNodeAffinityPreferredWithDefaults() *PodNodeAffinityPreferred`

NewPodNodeAffinityPreferredWithDefaults instantiates a new PodNodeAffinityPreferred object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodeSelectorTerms

`func (o *PodNodeAffinityPreferred) GetNodeSelectorTerms() []PodPreferredNodeSelectorTerm`

GetNodeSelectorTerms returns the NodeSelectorTerms field if non-nil, zero value otherwise.

### GetNodeSelectorTermsOk

`func (o *PodNodeAffinityPreferred) GetNodeSelectorTermsOk() (*[]PodPreferredNodeSelectorTerm, bool)`

GetNodeSelectorTermsOk returns a tuple with the NodeSelectorTerms field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeSelectorTerms

`func (o *PodNodeAffinityPreferred) SetNodeSelectorTerms(v []PodPreferredNodeSelectorTerm)`

SetNodeSelectorTerms sets NodeSelectorTerms field to given value.

### HasNodeSelectorTerms

`func (o *PodNodeAffinityPreferred) HasNodeSelectorTerms() bool`

HasNodeSelectorTerms returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


