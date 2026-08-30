# WorkloadTemplateCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**AssetCreationRequest**](AssetCreationRequest.md) |  | 
**Spec** | [**SpecificRunCreationFields**](SpecificRunCreationFields.md) |  | 

## Methods

### NewWorkloadTemplateCreationRequest

`func NewWorkloadTemplateCreationRequest(meta AssetCreationRequest, spec SpecificRunCreationFields, ) *WorkloadTemplateCreationRequest`

NewWorkloadTemplateCreationRequest instantiates a new WorkloadTemplateCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTemplateCreationRequestWithDefaults

`func NewWorkloadTemplateCreationRequestWithDefaults() *WorkloadTemplateCreationRequest`

NewWorkloadTemplateCreationRequestWithDefaults instantiates a new WorkloadTemplateCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkloadTemplateCreationRequest) GetMeta() AssetCreationRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkloadTemplateCreationRequest) GetMetaOk() (*AssetCreationRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkloadTemplateCreationRequest) SetMeta(v AssetCreationRequest)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *WorkloadTemplateCreationRequest) GetSpec() SpecificRunCreationFields`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *WorkloadTemplateCreationRequest) GetSpecOk() (*SpecificRunCreationFields, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *WorkloadTemplateCreationRequest) SetSpec(v SpecificRunCreationFields)`

SetSpec sets Spec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


