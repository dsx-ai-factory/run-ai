# ModelAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Title** | Pointer to **string** | the title of the model that will be displayed in the UI | [optional] 
**Description** | Pointer to **NullableString** | the description of the model that will be displayed in the UI | [optional] 
**Thumbnail** | Pointer to **NullableString** | the thumbnail of the model that will be displayed in the UI | [optional] 
**HuggingfaceId** | Pointer to **NullableString** | the huggingface id of the model | [optional] 
**Assets** | Pointer to [**AssetsIds**](AssetsIds.md) |  | [optional] 
**License** | Pointer to **NullableString** | the license of the model | [optional] 

## Methods

### NewModelAssetSpec

`func NewModelAssetSpec() *ModelAssetSpec`

NewModelAssetSpec instantiates a new ModelAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelAssetSpecWithDefaults

`func NewModelAssetSpecWithDefaults() *ModelAssetSpec`

NewModelAssetSpecWithDefaults instantiates a new ModelAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTitle

`func (o *ModelAssetSpec) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *ModelAssetSpec) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *ModelAssetSpec) SetTitle(v string)`

SetTitle sets Title field to given value.

### HasTitle

`func (o *ModelAssetSpec) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### GetDescription

`func (o *ModelAssetSpec) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ModelAssetSpec) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ModelAssetSpec) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ModelAssetSpec) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *ModelAssetSpec) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *ModelAssetSpec) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetThumbnail

`func (o *ModelAssetSpec) GetThumbnail() string`

GetThumbnail returns the Thumbnail field if non-nil, zero value otherwise.

### GetThumbnailOk

`func (o *ModelAssetSpec) GetThumbnailOk() (*string, bool)`

GetThumbnailOk returns a tuple with the Thumbnail field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThumbnail

`func (o *ModelAssetSpec) SetThumbnail(v string)`

SetThumbnail sets Thumbnail field to given value.

### HasThumbnail

`func (o *ModelAssetSpec) HasThumbnail() bool`

HasThumbnail returns a boolean if a field has been set.

### SetThumbnailNil

`func (o *ModelAssetSpec) SetThumbnailNil(b bool)`

 SetThumbnailNil sets the value for Thumbnail to be an explicit nil

### UnsetThumbnail
`func (o *ModelAssetSpec) UnsetThumbnail()`

UnsetThumbnail ensures that no value is present for Thumbnail, not even an explicit nil
### GetHuggingfaceId

`func (o *ModelAssetSpec) GetHuggingfaceId() string`

GetHuggingfaceId returns the HuggingfaceId field if non-nil, zero value otherwise.

### GetHuggingfaceIdOk

`func (o *ModelAssetSpec) GetHuggingfaceIdOk() (*string, bool)`

GetHuggingfaceIdOk returns a tuple with the HuggingfaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHuggingfaceId

`func (o *ModelAssetSpec) SetHuggingfaceId(v string)`

SetHuggingfaceId sets HuggingfaceId field to given value.

### HasHuggingfaceId

`func (o *ModelAssetSpec) HasHuggingfaceId() bool`

HasHuggingfaceId returns a boolean if a field has been set.

### SetHuggingfaceIdNil

`func (o *ModelAssetSpec) SetHuggingfaceIdNil(b bool)`

 SetHuggingfaceIdNil sets the value for HuggingfaceId to be an explicit nil

### UnsetHuggingfaceId
`func (o *ModelAssetSpec) UnsetHuggingfaceId()`

UnsetHuggingfaceId ensures that no value is present for HuggingfaceId, not even an explicit nil
### GetAssets

`func (o *ModelAssetSpec) GetAssets() AssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *ModelAssetSpec) GetAssetsOk() (*AssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *ModelAssetSpec) SetAssets(v AssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *ModelAssetSpec) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetLicense

`func (o *ModelAssetSpec) GetLicense() string`

GetLicense returns the License field if non-nil, zero value otherwise.

### GetLicenseOk

`func (o *ModelAssetSpec) GetLicenseOk() (*string, bool)`

GetLicenseOk returns a tuple with the License field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLicense

`func (o *ModelAssetSpec) SetLicense(v string)`

SetLicense sets License field to given value.

### HasLicense

`func (o *ModelAssetSpec) HasLicense() bool`

HasLicense returns a boolean if a field has been set.

### SetLicenseNil

`func (o *ModelAssetSpec) SetLicenseNil(b bool)`

 SetLicenseNil sets the value for License to be an explicit nil

### UnsetLicense
`func (o *ModelAssetSpec) UnsetLicense()`

UnsetLicense ensures that no value is present for License, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


