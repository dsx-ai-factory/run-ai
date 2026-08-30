# GetAutoRedirectSsoSetting

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Value** | Pointer to [**GetAutoRedirectSsoSettingValue**](GetAutoRedirectSsoSettingValue.md) |  | [optional] 
**DisplayName** | Pointer to **string** | Human-readable name for this setting | [optional] [readonly] 
**Description** | Pointer to **string** | Description of what this setting controls | [optional] [readonly] 

## Methods

### NewGetAutoRedirectSsoSetting

`func NewGetAutoRedirectSsoSetting() *GetAutoRedirectSsoSetting`

NewGetAutoRedirectSsoSetting instantiates a new GetAutoRedirectSsoSetting object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetAutoRedirectSsoSettingWithDefaults

`func NewGetAutoRedirectSsoSettingWithDefaults() *GetAutoRedirectSsoSetting`

NewGetAutoRedirectSsoSettingWithDefaults instantiates a new GetAutoRedirectSsoSetting object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetValue

`func (o *GetAutoRedirectSsoSetting) GetValue() GetAutoRedirectSsoSettingValue`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *GetAutoRedirectSsoSetting) GetValueOk() (*GetAutoRedirectSsoSettingValue, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *GetAutoRedirectSsoSetting) SetValue(v GetAutoRedirectSsoSettingValue)`

SetValue sets Value field to given value.

### HasValue

`func (o *GetAutoRedirectSsoSetting) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetDisplayName

`func (o *GetAutoRedirectSsoSetting) GetDisplayName() string`

GetDisplayName returns the DisplayName field if non-nil, zero value otherwise.

### GetDisplayNameOk

`func (o *GetAutoRedirectSsoSetting) GetDisplayNameOk() (*string, bool)`

GetDisplayNameOk returns a tuple with the DisplayName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplayName

`func (o *GetAutoRedirectSsoSetting) SetDisplayName(v string)`

SetDisplayName sets DisplayName field to given value.

### HasDisplayName

`func (o *GetAutoRedirectSsoSetting) HasDisplayName() bool`

HasDisplayName returns a boolean if a field has been set.

### GetDescription

`func (o *GetAutoRedirectSsoSetting) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *GetAutoRedirectSsoSetting) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *GetAutoRedirectSsoSetting) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *GetAutoRedirectSsoSetting) HasDescription() bool`

HasDescription returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


