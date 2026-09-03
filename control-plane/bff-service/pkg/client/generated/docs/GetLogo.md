# GetLogo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Value** | Pointer to [**GetLogoValue**](GetLogoValue.md) |  | [optional] 
**DisplayName** | Pointer to **string** | Human-readable name for this setting | [optional] [readonly] 
**Description** | Pointer to **string** | Short explanation of what this setting controls, shown alongside the setting | [optional] [readonly] 

## Methods

### NewGetLogo

`func NewGetLogo() *GetLogo`

NewGetLogo instantiates a new GetLogo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetLogoWithDefaults

`func NewGetLogoWithDefaults() *GetLogo`

NewGetLogoWithDefaults instantiates a new GetLogo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValue

`func (o *GetLogo) GetValue() GetLogoValue`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *GetLogo) GetValueOk() (*GetLogoValue, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *GetLogo) SetValue(v GetLogoValue)`

SetValue sets Value field to given value.

### HasValue

`func (o *GetLogo) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetDisplayName

`func (o *GetLogo) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *GetLogo) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *GetLogo) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *GetLogo) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *GetLogo) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GetLogo) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GetLogo) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GetLogo) HasDescription() bool`

HasDescription returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


