# Annotation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the annotation (mandatory) | [optional] 
**Value** | Pointer to **NullableString** | The value of the annotation. | [optional] 
**Exclude** | Pointer to **NullableBool** | Use &#39;true&#39; in case the annotation is defined in defaults of the policy, and you wish to exclude it from the workload. | [optional] [default to false]

## Methods

### NewAnnotation

`func NewAnnotation() *Annotation`

NewAnnotation instantiates a new Annotation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAnnotationWithDefaults

`func NewAnnotationWithDefaults() *Annotation`

NewAnnotationWithDefaults instantiates a new Annotation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Annotation) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Annotation) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Annotation) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Annotation) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *Annotation) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *Annotation) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetValue

`func (o *Annotation) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *Annotation) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *Annotation) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *Annotation) HasValue() bool`

HasValue returns a boolean if a field has been set.

### SetValueNil

`func (o *Annotation) SetValueNil(b bool)`

 SetValueNil sets the value for Value to be an explicit nil

### UnsetValue
`func (o *Annotation) UnsetValue()`

UnsetValue ensures that no value is present for Value, not even an explicit nil
### GetExclude

`func (o *Annotation) GetExclude() bool`

GetExclude returns the Exclude field if non-nil, zero value otherwise.

### GetExcludeOk

`func (o *Annotation) GetExcludeOk() (*bool, bool)`

GetExcludeOk returns a tuple with the Exclude field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExclude

`func (o *Annotation) SetExclude(v bool)`

SetExclude sets Exclude field to given value.

### HasExclude

`func (o *Annotation) HasExclude() bool`

HasExclude returns a boolean if a field has been set.

### SetExcludeNil

`func (o *Annotation) SetExcludeNil(b bool)`

 SetExcludeNil sets the value for Exclude to be an explicit nil

### UnsetExclude
`func (o *Annotation) UnsetExclude()`

UnsetExclude ensures that no value is present for Exclude, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


