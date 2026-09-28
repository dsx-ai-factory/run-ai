# SubmitWithTemplateId

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TemplateId** | Pointer to **NullableString** | The unique identifier of the template to use for submitting this workload. The combined values provided in the spec, template and assets will be used to create the workload. | [optional] 

## Methods

### NewSubmitWithTemplateId

`func NewSubmitWithTemplateId() *SubmitWithTemplateId`

NewSubmitWithTemplateId instantiates a new SubmitWithTemplateId object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmitWithTemplateIdWithDefaults

`func NewSubmitWithTemplateIdWithDefaults() *SubmitWithTemplateId`

NewSubmitWithTemplateIdWithDefaults instantiates a new SubmitWithTemplateId object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTemplateId

`func (o *SubmitWithTemplateId) GetTemplateId() string`

GetTemplateId returns the TemplateId field if non-nil, zero value otherwise.

### GetTemplateIdOk

`func (o *SubmitWithTemplateId) GetTemplateIdOk() (*string, bool)`

GetTemplateIdOk returns a tuple with the TemplateId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplateId

`func (o *SubmitWithTemplateId) SetTemplateId(v string)`

SetTemplateId sets TemplateId field to given value.

### HasTemplateId

`func (o *SubmitWithTemplateId) HasTemplateId() bool`

HasTemplateId returns a boolean if a field has been set.

### SetTemplateIdNil

`func (o *SubmitWithTemplateId) SetTemplateIdNil(b bool)`

 SetTemplateIdNil sets the value for TemplateId to be an explicit nil

### UnsetTemplateId
`func (o *SubmitWithTemplateId) UnsetTemplateId()`

UnsetTemplateId ensures that no value is present for TemplateId, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


