# PodPreferredNodeSelectorTerm

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Weight** | Pointer to **float32** | Weight associated with matching the corresponding nodeSelectorTerm, in the range 1-100. | [optional] 
**Preference** | Pointer to [**NullableNodeSelectorTerm**](NodeSelectorTerm.md) |  | [optional] 

## Methods

### NewPodPreferredNodeSelectorTerm

`func NewPodPreferredNodeSelectorTerm() *PodPreferredNodeSelectorTerm`

NewPodPreferredNodeSelectorTerm instantiates a new PodPreferredNodeSelectorTerm object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodPreferredNodeSelectorTermWithDefaults

`func NewPodPreferredNodeSelectorTermWithDefaults() *PodPreferredNodeSelectorTerm`

NewPodPreferredNodeSelectorTermWithDefaults instantiates a new PodPreferredNodeSelectorTerm object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWeight

`func (o *PodPreferredNodeSelectorTerm) GetWeight() float32`

GetWeight returns the Weight field if non-nil, zero value otherwise.

### GetWeightOk

`func (o *PodPreferredNodeSelectorTerm) GetWeightOk() (*float32, bool)`

GetWeightOk returns a tuple with the Weight field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWeight

`func (o *PodPreferredNodeSelectorTerm) SetWeight(v float32)`

SetWeight sets Weight field to given value.

### HasWeight

`func (o *PodPreferredNodeSelectorTerm) HasWeight() bool`

HasWeight returns a boolean if a field has been set.

### GetPreference

`func (o *PodPreferredNodeSelectorTerm) GetPreference() NodeSelectorTerm`

GetPreference returns the Preference field if non-nil, zero value otherwise.

### GetPreferenceOk

`func (o *PodPreferredNodeSelectorTerm) GetPreferenceOk() (*NodeSelectorTerm, bool)`

GetPreferenceOk returns a tuple with the Preference field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreference

`func (o *PodPreferredNodeSelectorTerm) SetPreference(v NodeSelectorTerm)`

SetPreference sets Preference field to given value.

### HasPreference

`func (o *PodPreferredNodeSelectorTerm) HasPreference() bool`

HasPreference returns a boolean if a field has been set.

### SetPreferenceNil

`func (o *PodPreferredNodeSelectorTerm) SetPreferenceNil(b bool)`

 SetPreferenceNil sets the value for Preference to be an explicit nil

### UnsetPreference
`func (o *PodPreferredNodeSelectorTerm) UnsetPreference()`

UnsetPreference ensures that no value is present for Preference, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


