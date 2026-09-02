# GetBannerConfigValue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** | Whether the banner is displayed at the top of the UI for all authenticated users. | [optional] 
**Text** | Pointer to **NullableString** | Message displayed in the banner. Required when &#x60;enabled&#x60; is &#x60;true&#x60;. Maximum 500 characters. | [optional] 
**BackgroundColor** | Pointer to **NullableString** | Background color of the banner. Accepts any valid CSS color value (hex, rgb, named). Maximum 50 characters. | [optional] 
**FontColor** | Pointer to **NullableString** | Text color of the banner. Accepts any valid CSS color value (hex, rgb, named). Maximum 50 characters. | [optional] 
**FontSize** | Pointer to **NullableInt32** | Font size in pixels. Maximum 20. Defaults to 14 if omitted or null. | [optional] 

## Methods

### NewGetBannerConfigValue

`func NewGetBannerConfigValue() *GetBannerConfigValue`

NewGetBannerConfigValue instantiates a new GetBannerConfigValue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetBannerConfigValueWithDefaults

`func NewGetBannerConfigValueWithDefaults() *GetBannerConfigValue`

NewGetBannerConfigValueWithDefaults instantiates a new GetBannerConfigValue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *GetBannerConfigValue) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *GetBannerConfigValue) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *GetBannerConfigValue) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *GetBannerConfigValue) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *GetBannerConfigValue) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *GetBannerConfigValue) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetText

`func (o *GetBannerConfigValue) GetText() string`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *GetBannerConfigValue) GetTextOk() (*string, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *GetBannerConfigValue) SetText(v string)`

SetText sets Text field to given value.

### HasText

`func (o *GetBannerConfigValue) HasText() bool`

HasText returns a boolean if a field has been set.

### SetTextNil

`func (o *GetBannerConfigValue) SetTextNil(b bool)`

 SetTextNil sets the value for Text to be an explicit nil

### UnsetText
`func (o *GetBannerConfigValue) UnsetText()`

UnsetText ensures that no value is present for Text, not even an explicit nil
### GetBackgroundColor

`func (o *GetBannerConfigValue) GetBackgroundColor() string`

GetBackgroundColor returns the BackgroundColor field if non-nil, zero value otherwise.

### GetBackgroundColorOk

`func (o *GetBannerConfigValue) GetBackgroundColorOk() (*string, bool)`

GetBackgroundColorOk returns a tuple with the BackgroundColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackgroundColor

`func (o *GetBannerConfigValue) SetBackgroundColor(v string)`

SetBackgroundColor sets BackgroundColor field to given value.

### HasBackgroundColor

`func (o *GetBannerConfigValue) HasBackgroundColor() bool`

HasBackgroundColor returns a boolean if a field has been set.

### SetBackgroundColorNil

`func (o *GetBannerConfigValue) SetBackgroundColorNil(b bool)`

 SetBackgroundColorNil sets the value for BackgroundColor to be an explicit nil

### UnsetBackgroundColor
`func (o *GetBannerConfigValue) UnsetBackgroundColor()`

UnsetBackgroundColor ensures that no value is present for BackgroundColor, not even an explicit nil
### GetFontColor

`func (o *GetBannerConfigValue) GetFontColor() string`

GetFontColor returns the FontColor field if non-nil, zero value otherwise.

### GetFontColorOk

`func (o *GetBannerConfigValue) GetFontColorOk() (*string, bool)`

GetFontColorOk returns a tuple with the FontColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFontColor

`func (o *GetBannerConfigValue) SetFontColor(v string)`

SetFontColor sets FontColor field to given value.

### HasFontColor

`func (o *GetBannerConfigValue) HasFontColor() bool`

HasFontColor returns a boolean if a field has been set.

### SetFontColorNil

`func (o *GetBannerConfigValue) SetFontColorNil(b bool)`

 SetFontColorNil sets the value for FontColor to be an explicit nil

### UnsetFontColor
`func (o *GetBannerConfigValue) UnsetFontColor()`

UnsetFontColor ensures that no value is present for FontColor, not even an explicit nil
### GetFontSize

`func (o *GetBannerConfigValue) GetFontSize() int32`

GetFontSize returns the FontSize field if non-nil, zero value otherwise.

### GetFontSizeOk

`func (o *GetBannerConfigValue) GetFontSizeOk() (*int32, bool)`

GetFontSizeOk returns a tuple with the FontSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFontSize

`func (o *GetBannerConfigValue) SetFontSize(v int32)`

SetFontSize sets FontSize field to given value.

### HasFontSize

`func (o *GetBannerConfigValue) HasFontSize() bool`

HasFontSize returns a boolean if a field has been set.

### SetFontSizeNil

`func (o *GetBannerConfigValue) SetFontSizeNil(b bool)`

 SetFontSizeNil sets the value for FontSize to be an explicit nil

### UnsetFontSize
`func (o *GetBannerConfigValue) UnsetFontSize()`

UnsetFontSize ensures that no value is present for FontSize, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


