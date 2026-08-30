# Annotation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the annotation. | 
**Value** | **string** | The value to set the annotation to. | 
**Deleted** | Pointer to **bool** | Whether to exclude this annotation from the workload. This is necessary in case the annotation definition is inherited from a policy. | [optional] 

## Methods

### NewAnnotation

`func NewAnnotation(name string, value string, ) *Annotation`

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


### GetDeleted

`func (o *Annotation) GetDeleted() bool`

GetDeleted returns the Deleted field if non-nil, zero value otherwise.

### GetDeletedOk

`func (o *Annotation) GetDeletedOk() (*bool, bool)`

GetDeletedOk returns a tuple with the Deleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeleted

`func (o *Annotation) SetDeleted(v bool)`

SetDeleted sets Deleted field to given value.

### HasDeleted

`func (o *Annotation) HasDeleted() bool`

HasDeleted returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


