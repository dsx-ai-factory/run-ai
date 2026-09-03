# FormatterArrayKey

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ArgName** | **string** | The name of the argument that supplies the unique key value per item. | 
**SpecRef** | Pointer to **NullableString** | An optional direct dotted path into the workload spec that locates the key value for each item. | [optional] 

## Methods

### NewFormatterArrayKey

`func NewFormatterArrayKey(argName string, ) *FormatterArrayKey`

NewFormatterArrayKey instantiates a new FormatterArrayKey object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFormatterArrayKeyWithDefaults

`func NewFormatterArrayKeyWithDefaults() *FormatterArrayKey`

NewFormatterArrayKeyWithDefaults instantiates a new FormatterArrayKey object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetArgName

`func (o *FormatterArrayKey) GetArgName() string`

GetArgName returns the ArgName field if non-nil, zero value otherwise.

### GetArgNameOk

`func (o *FormatterArrayKey) GetArgNameOk() (*string, bool)`

GetArgNameOk returns a tuple with the ArgName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgName

`func (o *FormatterArrayKey) SetArgName(v string)`

SetArgName sets ArgName field to given value.


### GetSpecRef

`func (o *FormatterArrayKey) GetSpecRef() string`

GetSpecRef returns the SpecRef field if non-nil, zero value otherwise.

### GetSpecRefOk

`func (o *FormatterArrayKey) GetSpecRefOk() (*string, bool)`

GetSpecRefOk returns a tuple with the SpecRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecRef

`func (o *FormatterArrayKey) SetSpecRef(v string)`

SetSpecRef sets SpecRef field to given value.

### HasSpecRef

`func (o *FormatterArrayKey) HasSpecRef() bool`

HasSpecRef returns a boolean if a field has been set.

### SetSpecRefNil

`func (o *FormatterArrayKey) SetSpecRefNil(b bool)`

 SetSpecRefNil sets the value for SpecRef to be an explicit nil

### UnsetSpecRef
`func (o *FormatterArrayKey) UnsetSpecRef()`

UnsetSpecRef ensures that no value is present for SpecRef, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


