# UpdateLogo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Logo** | **string** | Logo image as a base64-encoded string. Accepts PNG or SVG up to 128 KB. Send an empty string to reset the logo | 

## Methods

### NewUpdateLogo

`func NewUpdateLogo(logo string, ) *UpdateLogo`

NewUpdateLogo instantiates a new UpdateLogo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateLogoWithDefaults

`func NewUpdateLogoWithDefaults() *UpdateLogo`

NewUpdateLogoWithDefaults instantiates a new UpdateLogo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLogo

`func (o *UpdateLogo) GetLogo() string`

GetLogo returns the Logo field if non-nil, zero value otherwise.

### GetLogoOk

`func (o *UpdateLogo) GetLogoOk() (*string, bool)`

GetLogoOk returns a tuple with the Logo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLogo

`func (o *UpdateLogo) SetLogo(v string)`

SetLogo sets Logo field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


