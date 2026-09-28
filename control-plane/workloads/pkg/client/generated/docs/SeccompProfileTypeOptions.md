# SeccompProfileTypeOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Options** | Pointer to [**[]SeccompProfileTypeOptionsOptionsInner**](SeccompProfileTypeOptionsOptionsInner.md) | Limit the permitted values for secompProfileType to a given set of values. | [optional] 

## Methods

### NewSeccompProfileTypeOptions

`func NewSeccompProfileTypeOptions() *SeccompProfileTypeOptions`

NewSeccompProfileTypeOptions instantiates a new SeccompProfileTypeOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSeccompProfileTypeOptionsWithDefaults

`func NewSeccompProfileTypeOptionsWithDefaults() *SeccompProfileTypeOptions`

NewSeccompProfileTypeOptionsWithDefaults instantiates a new SeccompProfileTypeOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOptions

`func (o *SeccompProfileTypeOptions) GetOptions() []SeccompProfileTypeOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *SeccompProfileTypeOptions) GetOptionsOk() (*[]SeccompProfileTypeOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *SeccompProfileTypeOptions) SetOptions(v []SeccompProfileTypeOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *SeccompProfileTypeOptions) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *SeccompProfileTypeOptions) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *SeccompProfileTypeOptions) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


