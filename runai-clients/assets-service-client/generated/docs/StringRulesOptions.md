# StringRulesOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Options** | Pointer to [**[]StringOption**](StringOption.md) | Set of options that the value of the field must be chosen from. | [optional] 

## Methods

### NewStringRulesOptions

`func NewStringRulesOptions() *StringRulesOptions`

NewStringRulesOptions instantiates a new StringRulesOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStringRulesOptionsWithDefaults

`func NewStringRulesOptionsWithDefaults() *StringRulesOptions`

NewStringRulesOptionsWithDefaults instantiates a new StringRulesOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOptions

`func (o *StringRulesOptions) GetOptions() []StringOption`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *StringRulesOptions) GetOptionsOk() (*[]StringOption, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *StringRulesOptions) SetOptions(v []StringOption)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *StringRulesOptions) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *StringRulesOptions) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *StringRulesOptions) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


