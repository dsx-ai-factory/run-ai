# PvcAddedAttribute

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | **string** | The annotation key for the PVC. Must comply with Kubernetes naming rules: ≤63 characters, beginning and ending with an alphanumeric character ([A-Za-z0-9]), and may include dashes (-), underscores (_), or dots (.). | 
**Format** | Pointer to [**NullablePvcAddedAttributeFormat**](PvcAddedAttributeFormat.md) |  | [optional] 
**Display** | Pointer to **NullableString** | User-friendly display name for the attribute, shown in the UI. | [optional] 
**Description** | Pointer to **NullableString** | A description of the attribute&#39;s purpose or expected values. | [optional] 
**Placeholder** | Pointer to **NullableString** | Example or hint text for the attribute&#39;s value, displayed as a watermark or help message in the UI. | [optional] 
**Required** | Pointer to **NullableBool** | Indicates whether this attribute is mandatory. | [optional] 
**DefaultValue** | Pointer to **NullableString** | Default value assigned to the attribute when not explicitly provided. | [optional] 
**ValidationRegexp** | Pointer to **NullableString** | Regular expression used to validate user input for text attributes. | [optional] 
**MinValue** | Pointer to **NullableFloat32** | Minimum allowed value for attributes of type number. | [optional] 
**MaxValue** | Pointer to **NullableFloat32** | Maximum allowed value for attributes of type number. | [optional] 
**Step** | Pointer to **NullableFloat32** | Defines the incremental step between allowed numeric values (for example, 1, 0.1). | [optional] 
**ValidationError** | Pointer to **NullableString** | Custom validation error message shown if input fails validation. If omitted, a default error message is generated automatically. | [optional] 

## Methods

### NewPvcAddedAttribute

`func NewPvcAddedAttribute(key string, ) *PvcAddedAttribute`

NewPvcAddedAttribute instantiates a new PvcAddedAttribute object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPvcAddedAttributeWithDefaults

`func NewPvcAddedAttributeWithDefaults() *PvcAddedAttribute`

NewPvcAddedAttributeWithDefaults instantiates a new PvcAddedAttribute object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *PvcAddedAttribute) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *PvcAddedAttribute) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *PvcAddedAttribute) SetKey(v string)`

SetKey sets Key field to given value.


### GetFormat

`func (o *PvcAddedAttribute) GetFormat() PvcAddedAttributeFormat`

GetFormat returns the Format field if non-nil, zero value otherwise.

### GetFormatOk

`func (o *PvcAddedAttribute) GetFormatOk() (*PvcAddedAttributeFormat, bool)`

GetFormatOk returns a tuple with the Format field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFormat

`func (o *PvcAddedAttribute) SetFormat(v PvcAddedAttributeFormat)`

SetFormat sets Format field to given value.

### HasFormat

`func (o *PvcAddedAttribute) HasFormat() bool`

HasFormat returns a boolean if a field has been set.

### SetFormatNil

`func (o *PvcAddedAttribute) SetFormatNil(b bool)`

 SetFormatNil sets the value for Format to be an explicit nil

### UnsetFormat
`func (o *PvcAddedAttribute) UnsetFormat()`

UnsetFormat ensures that no value is present for Format, not even an explicit nil
### GetDisplay

`func (o *PvcAddedAttribute) GetDisplay() string`

GetDisplay returns the Display field if non-nil, zero value otherwise.

### GetDisplayOk

`func (o *PvcAddedAttribute) GetDisplayOk() (*string, bool)`

GetDisplayOk returns a tuple with the Display field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisplay

`func (o *PvcAddedAttribute) SetDisplay(v string)`

SetDisplay sets Display field to given value.

### HasDisplay

`func (o *PvcAddedAttribute) HasDisplay() bool`

HasDisplay returns a boolean if a field has been set.

### SetDisplayNil

`func (o *PvcAddedAttribute) SetDisplayNil(b bool)`

 SetDisplayNil sets the value for Display to be an explicit nil

### UnsetDisplay
`func (o *PvcAddedAttribute) UnsetDisplay()`

UnsetDisplay ensures that no value is present for Display, not even an explicit nil
### GetDescription

`func (o *PvcAddedAttribute) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *PvcAddedAttribute) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *PvcAddedAttribute) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *PvcAddedAttribute) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *PvcAddedAttribute) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *PvcAddedAttribute) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetPlaceholder

`func (o *PvcAddedAttribute) GetPlaceholder() string`

GetPlaceholder returns the Placeholder field if non-nil, zero value otherwise.

### GetPlaceholderOk

`func (o *PvcAddedAttribute) GetPlaceholderOk() (*string, bool)`

GetPlaceholderOk returns a tuple with the Placeholder field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlaceholder

`func (o *PvcAddedAttribute) SetPlaceholder(v string)`

SetPlaceholder sets Placeholder field to given value.

### HasPlaceholder

`func (o *PvcAddedAttribute) HasPlaceholder() bool`

HasPlaceholder returns a boolean if a field has been set.

### SetPlaceholderNil

`func (o *PvcAddedAttribute) SetPlaceholderNil(b bool)`

 SetPlaceholderNil sets the value for Placeholder to be an explicit nil

### UnsetPlaceholder
`func (o *PvcAddedAttribute) UnsetPlaceholder()`

UnsetPlaceholder ensures that no value is present for Placeholder, not even an explicit nil
### GetRequired

`func (o *PvcAddedAttribute) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *PvcAddedAttribute) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *PvcAddedAttribute) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *PvcAddedAttribute) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *PvcAddedAttribute) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *PvcAddedAttribute) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetDefaultValue

`func (o *PvcAddedAttribute) GetDefaultValue() string`

GetDefaultValue returns the DefaultValue field if non-nil, zero value otherwise.

### GetDefaultValueOk

`func (o *PvcAddedAttribute) GetDefaultValueOk() (*string, bool)`

GetDefaultValueOk returns a tuple with the DefaultValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultValue

`func (o *PvcAddedAttribute) SetDefaultValue(v string)`

SetDefaultValue sets DefaultValue field to given value.

### HasDefaultValue

`func (o *PvcAddedAttribute) HasDefaultValue() bool`

HasDefaultValue returns a boolean if a field has been set.

### SetDefaultValueNil

`func (o *PvcAddedAttribute) SetDefaultValueNil(b bool)`

 SetDefaultValueNil sets the value for DefaultValue to be an explicit nil

### UnsetDefaultValue
`func (o *PvcAddedAttribute) UnsetDefaultValue()`

UnsetDefaultValue ensures that no value is present for DefaultValue, not even an explicit nil
### GetValidationRegexp

`func (o *PvcAddedAttribute) GetValidationRegexp() string`

GetValidationRegexp returns the ValidationRegexp field if non-nil, zero value otherwise.

### GetValidationRegexpOk

`func (o *PvcAddedAttribute) GetValidationRegexpOk() (*string, bool)`

GetValidationRegexpOk returns a tuple with the ValidationRegexp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidationRegexp

`func (o *PvcAddedAttribute) SetValidationRegexp(v string)`

SetValidationRegexp sets ValidationRegexp field to given value.

### HasValidationRegexp

`func (o *PvcAddedAttribute) HasValidationRegexp() bool`

HasValidationRegexp returns a boolean if a field has been set.

### SetValidationRegexpNil

`func (o *PvcAddedAttribute) SetValidationRegexpNil(b bool)`

 SetValidationRegexpNil sets the value for ValidationRegexp to be an explicit nil

### UnsetValidationRegexp
`func (o *PvcAddedAttribute) UnsetValidationRegexp()`

UnsetValidationRegexp ensures that no value is present for ValidationRegexp, not even an explicit nil
### GetMinValue

`func (o *PvcAddedAttribute) GetMinValue() float32`

GetMinValue returns the MinValue field if non-nil, zero value otherwise.

### GetMinValueOk

`func (o *PvcAddedAttribute) GetMinValueOk() (*float32, bool)`

GetMinValueOk returns a tuple with the MinValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinValue

`func (o *PvcAddedAttribute) SetMinValue(v float32)`

SetMinValue sets MinValue field to given value.

### HasMinValue

`func (o *PvcAddedAttribute) HasMinValue() bool`

HasMinValue returns a boolean if a field has been set.

### SetMinValueNil

`func (o *PvcAddedAttribute) SetMinValueNil(b bool)`

 SetMinValueNil sets the value for MinValue to be an explicit nil

### UnsetMinValue
`func (o *PvcAddedAttribute) UnsetMinValue()`

UnsetMinValue ensures that no value is present for MinValue, not even an explicit nil
### GetMaxValue

`func (o *PvcAddedAttribute) GetMaxValue() float32`

GetMaxValue returns the MaxValue field if non-nil, zero value otherwise.

### GetMaxValueOk

`func (o *PvcAddedAttribute) GetMaxValueOk() (*float32, bool)`

GetMaxValueOk returns a tuple with the MaxValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxValue

`func (o *PvcAddedAttribute) SetMaxValue(v float32)`

SetMaxValue sets MaxValue field to given value.

### HasMaxValue

`func (o *PvcAddedAttribute) HasMaxValue() bool`

HasMaxValue returns a boolean if a field has been set.

### SetMaxValueNil

`func (o *PvcAddedAttribute) SetMaxValueNil(b bool)`

 SetMaxValueNil sets the value for MaxValue to be an explicit nil

### UnsetMaxValue
`func (o *PvcAddedAttribute) UnsetMaxValue()`

UnsetMaxValue ensures that no value is present for MaxValue, not even an explicit nil
### GetStep

`func (o *PvcAddedAttribute) GetStep() float32`

GetStep returns the Step field if non-nil, zero value otherwise.

### GetStepOk

`func (o *PvcAddedAttribute) GetStepOk() (*float32, bool)`

GetStepOk returns a tuple with the Step field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStep

`func (o *PvcAddedAttribute) SetStep(v float32)`

SetStep sets Step field to given value.

### HasStep

`func (o *PvcAddedAttribute) HasStep() bool`

HasStep returns a boolean if a field has been set.

### SetStepNil

`func (o *PvcAddedAttribute) SetStepNil(b bool)`

 SetStepNil sets the value for Step to be an explicit nil

### UnsetStep
`func (o *PvcAddedAttribute) UnsetStep()`

UnsetStep ensures that no value is present for Step, not even an explicit nil
### GetValidationError

`func (o *PvcAddedAttribute) GetValidationError() string`

GetValidationError returns the ValidationError field if non-nil, zero value otherwise.

### GetValidationErrorOk

`func (o *PvcAddedAttribute) GetValidationErrorOk() (*string, bool)`

GetValidationErrorOk returns a tuple with the ValidationError field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidationError

`func (o *PvcAddedAttribute) SetValidationError(v string)`

SetValidationError sets ValidationError field to given value.

### HasValidationError

`func (o *PvcAddedAttribute) HasValidationError() bool`

HasValidationError returns a boolean if a field has been set.

### SetValidationErrorNil

`func (o *PvcAddedAttribute) SetValidationErrorNil(b bool)`

 SetValidationErrorNil sets the value for ValidationError to be an explicit nil

### UnsetValidationError
`func (o *PvcAddedAttribute) UnsetValidationError()`

UnsetValidationError ensures that no value is present for ValidationError, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


