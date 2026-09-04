# VolumeModeCustomization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Required** | Pointer to **NullableBool** | Indicates whether the volumeMode field is mandatory when creating a PVC. | [optional] 
**Default** | Pointer to [**NullablePvcVolumeMode**](PvcVolumeMode.md) |  | [optional] 
**SupportedValues** | Pointer to [**NullableVolumeModeCustomizationSupportedValues**](VolumeModeCustomizationSupportedValues.md) |  | [optional] 

## Methods

### NewVolumeModeCustomization

`func NewVolumeModeCustomization() *VolumeModeCustomization`

NewVolumeModeCustomization instantiates a new VolumeModeCustomization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeModeCustomizationWithDefaults

`func NewVolumeModeCustomizationWithDefaults() *VolumeModeCustomization`

NewVolumeModeCustomizationWithDefaults instantiates a new VolumeModeCustomization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequired

`func (o *VolumeModeCustomization) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *VolumeModeCustomization) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *VolumeModeCustomization) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *VolumeModeCustomization) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *VolumeModeCustomization) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *VolumeModeCustomization) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetDefault

`func (o *VolumeModeCustomization) GetDefault() PvcVolumeMode`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *VolumeModeCustomization) GetDefaultOk() (*PvcVolumeMode, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *VolumeModeCustomization) SetDefault(v PvcVolumeMode)`

SetDefault sets Default field to given value.

### HasDefault

`func (o *VolumeModeCustomization) HasDefault() bool`

HasDefault returns a boolean if a field has been set.

### SetDefaultNil

`func (o *VolumeModeCustomization) SetDefaultNil(b bool)`

 SetDefaultNil sets the value for Default to be an explicit nil

### UnsetDefault
`func (o *VolumeModeCustomization) UnsetDefault()`

UnsetDefault ensures that no value is present for Default, not even an explicit nil
### GetSupportedValues

`func (o *VolumeModeCustomization) GetSupportedValues() VolumeModeCustomizationSupportedValues`

GetSupportedValues returns the SupportedValues field if non-nil, zero value otherwise.

### GetSupportedValuesOk

`func (o *VolumeModeCustomization) GetSupportedValuesOk() (*VolumeModeCustomizationSupportedValues, bool)`

GetSupportedValuesOk returns a tuple with the SupportedValues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupportedValues

`func (o *VolumeModeCustomization) SetSupportedValues(v VolumeModeCustomizationSupportedValues)`

SetSupportedValues sets SupportedValues field to given value.

### HasSupportedValues

`func (o *VolumeModeCustomization) HasSupportedValues() bool`

HasSupportedValues returns a boolean if a field has been set.

### SetSupportedValuesNil

`func (o *VolumeModeCustomization) SetSupportedValuesNil(b bool)`

 SetSupportedValuesNil sets the value for SupportedValues to be an explicit nil

### UnsetSupportedValues
`func (o *VolumeModeCustomization) UnsetSupportedValues()`

UnsetSupportedValues ensures that no value is present for SupportedValues, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


