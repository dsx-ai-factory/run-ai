# StringOption

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Value** | **string** | The value that the field should hold | 
**Displayed** | Pointer to **NullableString** | Textual description of the value. to be used by user interface applications. | [optional] 

## Methods

### NewStringOption

`func NewStringOption(value string, ) *StringOption`

NewStringOption instantiates a new StringOption object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStringOptionWithDefaults

`func NewStringOptionWithDefaults() *StringOption`

NewStringOptionWithDefaults instantiates a new StringOption object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValue

`func (o *StringOption) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *StringOption) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *StringOption) SetValue(v string)`

SetValue sets Value field to given value.


### GetDisplayed

`func (o *StringOption) GetDisplayed() string`

GetDisplayed returns the Displayed field if non-nil, zero value otherwise.

### GetDisplayedOk

`func (o *StringOption) GetDisplayedOk() (*string, bool)`

GetDisplayedOk returns a tuple with the Displayed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayed

`func (o *StringOption) SetDisplayed(v string)`

SetDisplayed sets Displayed field to given value.

### HasDisplayed

`func (o *StringOption) HasDisplayed() bool`

HasDisplayed returns a boolean if a field has been set.

### SetDisplayedNil

`func (o *StringOption) SetDisplayedNil(b bool)`

 SetDisplayedNil sets the value for Displayed to be an explicit nil

### UnsetDisplayed
`func (o *StringOption) UnsetDisplayed()`

UnsetDisplayed ensures that no value is present for Displayed, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


