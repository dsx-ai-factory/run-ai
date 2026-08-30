# WorkloadTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetMeta**](AssetMeta.md) |  | 
**Spec** | [**SpecificRunInfoFields**](SpecificRunInfoFields.md) |  | 

## Methods

### NewWorkloadTemplate

`func NewWorkloadTemplate(meta AssetMeta, spec SpecificRunInfoFields, ) *WorkloadTemplate`

NewWorkloadTemplate instantiates a new WorkloadTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTemplateWithDefaults

`func NewWorkloadTemplateWithDefaults() *WorkloadTemplate`

NewWorkloadTemplateWithDefaults instantiates a new WorkloadTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkloadTemplate) GetMeta() AssetMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkloadTemplate) GetMetaOk() (*AssetMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkloadTemplate) SetMeta(v AssetMeta)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *WorkloadTemplate) GetSpec() SpecificRunInfoFields`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *WorkloadTemplate) GetSpecOk() (*SpecificRunInfoFields, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *WorkloadTemplate) SetSpec(v SpecificRunInfoFields)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


