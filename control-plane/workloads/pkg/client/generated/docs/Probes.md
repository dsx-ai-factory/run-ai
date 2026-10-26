# Probes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Readiness** | Pointer to [**NullableProbe**](Probe.md) |  | [optional] 

## Methods

### NewProbes

`func NewProbes() *Probes`

NewProbes instantiates a new Probes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProbesWithDefaults

`func NewProbesWithDefaults() *Probes`

NewProbesWithDefaults instantiates a new Probes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReadiness

`func (o *Probes) GetReadiness() Probe`

GetReadiness returns the Readiness field if non-nil, zero value otherwise.

### GetReadinessOk

`func (o *Probes) GetReadinessOk() (*Probe, bool)`

GetReadinessOk returns a tuple with the Readiness field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadiness

`func (o *Probes) SetReadiness(v Probe)`

SetReadiness sets Readiness field to given value.

### HasReadiness

`func (o *Probes) HasReadiness() bool`

HasReadiness returns a boolean if a field has been set.

### SetReadinessNil

`func (o *Probes) SetReadinessNil(b bool)`

 SetReadinessNil sets the value for Readiness to be an explicit nil

### UnsetReadiness
`func (o *Probes) UnsetReadiness()`

UnsetReadiness ensures that no value is present for Readiness, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


