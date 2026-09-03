# GetBannerConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Value** | Pointer to [**GetBannerConfigValue**](GetBannerConfigValue.md) |  | [optional] 
**DisplayName** | Pointer to **string** | Human-readable name for this setting | [optional] [readonly] 
**Description** | Pointer to **string** | Short explanation of what this setting controls, shown alongside the setting | [optional] [readonly] 

## Methods

### NewGetBannerConfig

`func NewGetBannerConfig() *GetBannerConfig`

NewGetBannerConfig instantiates a new GetBannerConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetBannerConfigWithDefaults

`func NewGetBannerConfigWithDefaults() *GetBannerConfig`

NewGetBannerConfigWithDefaults instantiates a new GetBannerConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValue

`func (o *GetBannerConfig) GetValue() GetBannerConfigValue`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *GetBannerConfig) GetValueOk() (*GetBannerConfigValue, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *GetBannerConfig) SetValue(v GetBannerConfigValue)`

SetValue sets Value field to given value.

### HasValue

`func (o *GetBannerConfig) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetDisplayName

`func (o *GetBannerConfig) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *GetBannerConfig) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *GetBannerConfig) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *GetBannerConfig) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *GetBannerConfig) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GetBannerConfig) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GetBannerConfig) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GetBannerConfig) HasDescription() bool`

HasDescription returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


