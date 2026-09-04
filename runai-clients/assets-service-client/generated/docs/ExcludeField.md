# ExcludeField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the item is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewExcludeField

`func NewExcludeField() *ExcludeField`

NewExcludeField instantiates a new ExcludeField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExcludeFieldWithDefaults

`func NewExcludeFieldWithDefaults() *ExcludeField`

NewExcludeFieldWithDefaults instantiates a new ExcludeField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExclude

`func (o *ExcludeField) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *ExcludeField) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *ExcludeField) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *ExcludeField) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *ExcludeField) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *ExcludeField) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


