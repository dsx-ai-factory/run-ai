# UidGidSourceOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Options** | Pointer to [**[]UidGidSourceOptionsOptionsInner**](UidGidSourceOptionsOptionsInner.md) | Limit the permitted values for uidGidSource to a given set of values. | [optional] 

## Methods

### NewUidGidSourceOptions

`func NewUidGidSourceOptions() *UidGidSourceOptions`

NewUidGidSourceOptions instantiates a new UidGidSourceOptions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUidGidSourceOptionsWithDefaults

`func NewUidGidSourceOptionsWithDefaults() *UidGidSourceOptions`

NewUidGidSourceOptionsWithDefaults instantiates a new UidGidSourceOptions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOptions

`func (o *UidGidSourceOptions) GetOptions() []UidGidSourceOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *UidGidSourceOptions) GetOptionsOk() (*[]UidGidSourceOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *UidGidSourceOptions) SetOptions(v []UidGidSourceOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *UidGidSourceOptions) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *UidGidSourceOptions) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *UidGidSourceOptions) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


