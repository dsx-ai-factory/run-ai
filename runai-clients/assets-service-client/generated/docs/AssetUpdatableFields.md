# AssetUpdatableFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the asset. | 
**Description** | Pointer to **NullableString** | The description of the asset. | [optional] 

## Methods

### NewAssetUpdatableFields

`func NewAssetUpdatableFields(name string, ) *AssetUpdatableFields`

NewAssetUpdatableFields instantiates a new AssetUpdatableFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetUpdatableFieldsWithDefaults

`func NewAssetUpdatableFieldsWithDefaults() *AssetUpdatableFields`

NewAssetUpdatableFieldsWithDefaults instantiates a new AssetUpdatableFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *AssetUpdatableFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AssetUpdatableFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AssetUpdatableFields) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *AssetUpdatableFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *AssetUpdatableFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *AssetUpdatableFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *AssetUpdatableFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *AssetUpdatableFields) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *AssetUpdatableFields) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


