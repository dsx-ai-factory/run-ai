# ClaimSizeCustomization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Default** | Pointer to **NullableString** | Requested size for the PVC. Mandatory when existingPvc is false. Recommended sizes: TB/GB/MB/TIB/GIB/MIB | [optional] 
**SupportedUnits** | Pointer to **[]string** | Units supported for specifying PVC sizes. Valid values: \&quot;MB\&quot;, \&quot;MiB\&quot;, \&quot;GB\&quot;, \&quot;GiB\&quot;, \&quot;TB\&quot;, \&quot;TiB\&quot;. | [optional] 
**Min** | Pointer to **NullableString** | Minimum claim size allowed for PVCs created with this storage class. Mandatory when existingPvc &#x3D; false. Recommended values: TB, GB, MB, TIB, GIB, MIB. | [optional] 
**Max** | Pointer to **NullableString** | Maximum claim size allowed for PVCs created with this storage class. Mandatory when existingPvc &#x3D; false. Recommended values: TB, GB, MB, TIB, GIB, MIB. | [optional] 
**Step** | Pointer to **NullableString** | Defines the incremental step between allowed numeric values (for example, 1, 0.1). | [optional] 

## Methods

### NewClaimSizeCustomization

`func NewClaimSizeCustomization() *ClaimSizeCustomization`

NewClaimSizeCustomization instantiates a new ClaimSizeCustomization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClaimSizeCustomizationWithDefaults

`func NewClaimSizeCustomizationWithDefaults() *ClaimSizeCustomization`

NewClaimSizeCustomizationWithDefaults instantiates a new ClaimSizeCustomization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefault

`func (o *ClaimSizeCustomization) GetDefault() string`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *ClaimSizeCustomization) GetDefaultOk() (*string, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *ClaimSizeCustomization) SetDefault(v string)`

SetDefault sets Default field to given value.

### HasDefault

`func (o *ClaimSizeCustomization) HasDefault() bool`

HasDefault returns a boolean if a field has been set.

### SetDefaultNil

`func (o *ClaimSizeCustomization) SetDefaultNil(b bool)`

 SetDefaultNil sets the value for Default to be an explicit nil

### UnsetDefault
`func (o *ClaimSizeCustomization) UnsetDefault()`

UnsetDefault ensures that no value is present for Default, not even an explicit nil
### GetSupportedUnits

`func (o *ClaimSizeCustomization) GetSupportedUnits() []string`

GetSupportedUnits returns the SupportedUnits field if non-nil, zero value otherwise.

### GetSupportedUnitsOk

`func (o *ClaimSizeCustomization) GetSupportedUnitsOk() (*[]string, bool)`

GetSupportedUnitsOk returns a tuple with the SupportedUnits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupportedUnits

`func (o *ClaimSizeCustomization) SetSupportedUnits(v []string)`

SetSupportedUnits sets SupportedUnits field to given value.

### HasSupportedUnits

`func (o *ClaimSizeCustomization) HasSupportedUnits() bool`

HasSupportedUnits returns a boolean if a field has been set.

### GetMin

`func (o *ClaimSizeCustomization) GetMin() string`

GetMin returns the Min field if non-nil, zero value otherwise.

### GetMinOk

`func (o *ClaimSizeCustomization) GetMinOk() (*string, bool)`

GetMinOk returns a tuple with the Min field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMin

`func (o *ClaimSizeCustomization) SetMin(v string)`

SetMin sets Min field to given value.

### HasMin

`func (o *ClaimSizeCustomization) HasMin() bool`

HasMin returns a boolean if a field has been set.

### SetMinNil

`func (o *ClaimSizeCustomization) SetMinNil(b bool)`

 SetMinNil sets the value for Min to be an explicit nil

### UnsetMin
`func (o *ClaimSizeCustomization) UnsetMin()`

UnsetMin ensures that no value is present for Min, not even an explicit nil
### GetMax

`func (o *ClaimSizeCustomization) GetMax() string`

GetMax returns the Max field if non-nil, zero value otherwise.

### GetMaxOk

`func (o *ClaimSizeCustomization) GetMaxOk() (*string, bool)`

GetMaxOk returns a tuple with the Max field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMax

`func (o *ClaimSizeCustomization) SetMax(v string)`

SetMax sets Max field to given value.

### HasMax

`func (o *ClaimSizeCustomization) HasMax() bool`

HasMax returns a boolean if a field has been set.

### SetMaxNil

`func (o *ClaimSizeCustomization) SetMaxNil(b bool)`

 SetMaxNil sets the value for Max to be an explicit nil

### UnsetMax
`func (o *ClaimSizeCustomization) UnsetMax()`

UnsetMax ensures that no value is present for Max, not even an explicit nil
### GetStep

`func (o *ClaimSizeCustomization) GetStep() string`

GetStep returns the Step field if non-nil, zero value otherwise.

### GetStepOk

`func (o *ClaimSizeCustomization) GetStepOk() (*string, bool)`

GetStepOk returns a tuple with the Step field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStep

`func (o *ClaimSizeCustomization) SetStep(v string)`

SetStep sets Step field to given value.

### HasStep

`func (o *ClaimSizeCustomization) HasStep() bool`

HasStep returns a boolean if a field has been set.

### SetStepNil

`func (o *ClaimSizeCustomization) SetStepNil(b bool)`

 SetStepNil sets the value for Step to be an explicit nil

### UnsetStep
`func (o *ClaimSizeCustomization) UnsetStep()`

UnsetStep ensures that no value is present for Step, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


