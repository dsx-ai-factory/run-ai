# PodNodeAffinity

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Required** | Pointer to [**NullablePodNodeAffinityRequired**](PodNodeAffinityRequired.md) |  | [optional] 
**Preferred** | Pointer to [**NullablePodNodeAffinityPreferred**](PodNodeAffinityPreferred.md) |  | [optional] 

## Methods

### NewPodNodeAffinity

`func NewPodNodeAffinity() *PodNodeAffinity`

NewPodNodeAffinity instantiates a new PodNodeAffinity object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodNodeAffinityWithDefaults

`func NewPodNodeAffinityWithDefaults() *PodNodeAffinity`

NewPodNodeAffinityWithDefaults instantiates a new PodNodeAffinity object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequired

`func (o *PodNodeAffinity) GetRequired() PodNodeAffinityRequired`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *PodNodeAffinity) GetRequiredOk() (*PodNodeAffinityRequired, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *PodNodeAffinity) SetRequired(v PodNodeAffinityRequired)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *PodNodeAffinity) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *PodNodeAffinity) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *PodNodeAffinity) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetPreferred

`func (o *PodNodeAffinity) GetPreferred() PodNodeAffinityPreferred`

GetPreferred returns the Preferred field if non-nil, zero value otherwise.

### GetPreferredOk

`func (o *PodNodeAffinity) GetPreferredOk() (*PodNodeAffinityPreferred, bool)`

GetPreferredOk returns a tuple with the Preferred field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreferred

`func (o *PodNodeAffinity) SetPreferred(v PodNodeAffinityPreferred)`

SetPreferred sets Preferred field to given value.

### HasPreferred

`func (o *PodNodeAffinity) HasPreferred() bool`

HasPreferred returns a boolean if a field has been set.

### SetPreferredNil

`func (o *PodNodeAffinity) SetPreferredNil(b bool)`

 SetPreferredNil sets the value for Preferred to be an explicit nil

### UnsetPreferred
`func (o *PodNodeAffinity) UnsetPreferred()`

UnsetPreferred ensures that no value is present for Preferred, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


