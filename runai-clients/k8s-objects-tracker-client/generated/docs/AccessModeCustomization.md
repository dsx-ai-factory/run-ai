# AccessModeCustomization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Required** | Pointer to **NullableBool** | If true, the accessMode field is mandatory when creating a PVC that uses this storage class. | [optional] 
**Default** | Pointer to [**NullablePvcAccessModes**](PvcAccessModes.md) |  | [optional] 
**SupportedValues** | Pointer to [**NullablePvcSupportedAccessModes**](PvcSupportedAccessModes.md) |  | [optional] 

## Methods

### NewAccessModeCustomization

`func NewAccessModeCustomization() *AccessModeCustomization`

NewAccessModeCustomization instantiates a new AccessModeCustomization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAccessModeCustomizationWithDefaults

`func NewAccessModeCustomizationWithDefaults() *AccessModeCustomization`

NewAccessModeCustomizationWithDefaults instantiates a new AccessModeCustomization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequired

`func (o *AccessModeCustomization) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *AccessModeCustomization) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *AccessModeCustomization) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *AccessModeCustomization) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *AccessModeCustomization) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *AccessModeCustomization) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetDefault

`func (o *AccessModeCustomization) GetDefault() PvcAccessModes`

GetDefault returns the Default field if non-nil, zero value otherwise.

### GetDefaultOk

`func (o *AccessModeCustomization) GetDefaultOk() (*PvcAccessModes, bool)`

GetDefaultOk returns a tuple with the Default field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefault

`func (o *AccessModeCustomization) SetDefault(v PvcAccessModes)`

SetDefault sets Default field to given value.

### HasDefault

`func (o *AccessModeCustomization) HasDefault() bool`

HasDefault returns a boolean if a field has been set.

### SetDefaultNil

`func (o *AccessModeCustomization) SetDefaultNil(b bool)`

 SetDefaultNil sets the value for Default to be an explicit nil

### UnsetDefault
`func (o *AccessModeCustomization) UnsetDefault()`

UnsetDefault ensures that no value is present for Default, not even an explicit nil
### GetSupportedValues

`func (o *AccessModeCustomization) GetSupportedValues() PvcSupportedAccessModes`

GetSupportedValues returns the SupportedValues field if non-nil, zero value otherwise.

### GetSupportedValuesOk

`func (o *AccessModeCustomization) GetSupportedValuesOk() (*PvcSupportedAccessModes, bool)`

GetSupportedValuesOk returns a tuple with the SupportedValues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSupportedValues

`func (o *AccessModeCustomization) SetSupportedValues(v PvcSupportedAccessModes)`

SetSupportedValues sets SupportedValues field to given value.

### HasSupportedValues

`func (o *AccessModeCustomization) HasSupportedValues() bool`

HasSupportedValues returns a boolean if a field has been set.

### SetSupportedValuesNil

`func (o *AccessModeCustomization) SetSupportedValuesNil(b bool)`

 SetSupportedValuesNil sets the value for SupportedValues to be an explicit nil

### UnsetSupportedValues
`func (o *AccessModeCustomization) UnsetSupportedValues()`

UnsetSupportedValues ensures that no value is present for SupportedValues, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


