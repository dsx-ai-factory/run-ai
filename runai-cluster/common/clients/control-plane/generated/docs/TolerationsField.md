# TolerationsField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Tolerations** | Pointer to [**[]Toleration**](Toleration.md) | Set of tolerations to apply to the workload. | [optional] 

## Methods

### NewTolerationsField

`func NewTolerationsField() *TolerationsField`

NewTolerationsField instantiates a new TolerationsField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTolerationsFieldWithDefaults

`func NewTolerationsFieldWithDefaults() *TolerationsField`

NewTolerationsFieldWithDefaults instantiates a new TolerationsField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTolerations

`func (o *TolerationsField) GetTolerations() []Toleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *TolerationsField) GetTolerationsOk() (*[]Toleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *TolerationsField) SetTolerations(v []Toleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *TolerationsField) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *TolerationsField) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *TolerationsField) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


