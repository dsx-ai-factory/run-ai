# RestartPolicyOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Options** | Pointer to [**[]RestartPolicyOptionsOptionsInner**](RestartPolicyOptionsOptionsInner.md) | Limit the permitted values for restartPolicy to a given set of values. | [optional] 

## Methods

### NewRestartPolicyOptions

`func NewRestartPolicyOptions() *RestartPolicyOptions`

NewRestartPolicyOptions instantiates a new RestartPolicyOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRestartPolicyOptionsWithDefaults

`func NewRestartPolicyOptionsWithDefaults() *RestartPolicyOptions`

NewRestartPolicyOptionsWithDefaults instantiates a new RestartPolicyOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOptions

`func (o *RestartPolicyOptions) GetOptions() []RestartPolicyOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *RestartPolicyOptions) GetOptionsOk() (*[]RestartPolicyOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *RestartPolicyOptions) SetOptions(v []RestartPolicyOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *RestartPolicyOptions) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *RestartPolicyOptions) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *RestartPolicyOptions) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


