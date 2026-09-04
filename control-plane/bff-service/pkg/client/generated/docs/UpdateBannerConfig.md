# UpdateBannerConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **NullableBool** | Whether the banner is displayed at the top of the UI for all authenticated users. | [optional] 
**Text** | Pointer to **NullableString** | Message displayed in the banner. Required when &#x60;enabled&#x60; is &#x60;true&#x60;. Maximum 500 characters. | [optional] 
**BackgroundColor** | Pointer to **NullableString** | Background color of the banner. Accepts any valid CSS color value (hex, rgb, named). Maximum 50 characters. | [optional] 
**FontColor** | Pointer to **NullableString** | Text color of the banner. Accepts any valid CSS color value (hex, rgb, named). Maximum 50 characters. | [optional] 
**FontSize** | Pointer to **NullableInt32** | Font size in pixels. Maximum 20. Defaults to 14 if omitted or null. | [optional] 

## Methods

### NewUpdateBannerConfig

`func NewUpdateBannerConfig() *UpdateBannerConfig`

NewUpdateBannerConfig instantiates a new UpdateBannerConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateBannerConfigWithDefaults

`func NewUpdateBannerConfigWithDefaults() *UpdateBannerConfig`

NewUpdateBannerConfigWithDefaults instantiates a new UpdateBannerConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *UpdateBannerConfig) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *UpdateBannerConfig) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *UpdateBannerConfig) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *UpdateBannerConfig) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### SetEnabledNil

`func (o *UpdateBannerConfig) SetEnabledNil(b bool)`

 SetEnabledNil sets the value for Enabled to be an explicit nil

### UnsetEnabled
`func (o *UpdateBannerConfig) UnsetEnabled()`

UnsetEnabled ensures that no value is present for Enabled, not even an explicit nil
### GetText

`func (o *UpdateBannerConfig) GetText() string`

GetText returns the Text field if non-nil, zero value otherwise.

### GetTextOk

`func (o *UpdateBannerConfig) GetTextOk() (*string, bool)`

GetTextOk returns a tuple with the Text field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetText

`func (o *UpdateBannerConfig) SetText(v string)`

SetText sets Text field to given value.

### HasText

`func (o *UpdateBannerConfig) HasText() bool`

HasText returns a boolean if a field has been set.

### SetTextNil

`func (o *UpdateBannerConfig) SetTextNil(b bool)`

 SetTextNil sets the value for Text to be an explicit nil

### UnsetText
`func (o *UpdateBannerConfig) UnsetText()`

UnsetText ensures that no value is present for Text, not even an explicit nil
### GetBackgroundColor

`func (o *UpdateBannerConfig) GetBackgroundColor() string`

GetBackgroundColor returns the BackgroundColor field if non-nil, zero value otherwise.

### GetBackgroundColorOk

`func (o *UpdateBannerConfig) GetBackgroundColorOk() (*string, bool)`

GetBackgroundColorOk returns a tuple with the BackgroundColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackgroundColor

`func (o *UpdateBannerConfig) SetBackgroundColor(v string)`

SetBackgroundColor sets BackgroundColor field to given value.

### HasBackgroundColor

`func (o *UpdateBannerConfig) HasBackgroundColor() bool`

HasBackgroundColor returns a boolean if a field has been set.

### SetBackgroundColorNil

`func (o *UpdateBannerConfig) SetBackgroundColorNil(b bool)`

 SetBackgroundColorNil sets the value for BackgroundColor to be an explicit nil

### UnsetBackgroundColor
`func (o *UpdateBannerConfig) UnsetBackgroundColor()`

UnsetBackgroundColor ensures that no value is present for BackgroundColor, not even an explicit nil
### GetFontColor

`func (o *UpdateBannerConfig) GetFontColor() string`

GetFontColor returns the FontColor field if non-nil, zero value otherwise.

### GetFontColorOk

`func (o *UpdateBannerConfig) GetFontColorOk() (*string, bool)`

GetFontColorOk returns a tuple with the FontColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFontColor

`func (o *UpdateBannerConfig) SetFontColor(v string)`

SetFontColor sets FontColor field to given value.

### HasFontColor

`func (o *UpdateBannerConfig) HasFontColor() bool`

HasFontColor returns a boolean if a field has been set.

### SetFontColorNil

`func (o *UpdateBannerConfig) SetFontColorNil(b bool)`

 SetFontColorNil sets the value for FontColor to be an explicit nil

### UnsetFontColor
`func (o *UpdateBannerConfig) UnsetFontColor()`

UnsetFontColor ensures that no value is present for FontColor, not even an explicit nil
### GetFontSize

`func (o *UpdateBannerConfig) GetFontSize() int32`

GetFontSize returns the FontSize field if non-nil, zero value otherwise.

### GetFontSizeOk

`func (o *UpdateBannerConfig) GetFontSizeOk() (*int32, bool)`

GetFontSizeOk returns a tuple with the FontSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFontSize

`func (o *UpdateBannerConfig) SetFontSize(v int32)`

SetFontSize sets FontSize field to given value.

### HasFontSize

`func (o *UpdateBannerConfig) HasFontSize() bool`

HasFontSize returns a boolean if a field has been set.

### SetFontSizeNil

`func (o *UpdateBannerConfig) SetFontSizeNil(b bool)`

 SetFontSizeNil sets the value for FontSize to be an explicit nil

### UnsetFontSize
`func (o *UpdateBannerConfig) UnsetFontSize()`

UnsetFontSize ensures that no value is present for FontSize, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


