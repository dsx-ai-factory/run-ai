# MappingCreUpdRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadType** | [**GVK**](GVK.md) | The Kubernetes group, version, and kind of the workload type this mapping applies to. | 
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**SpecSelectors** | Pointer to [**map[string]SpecSelector**](SpecSelector.md) | A map of spec selectors. Map **keys** are selector names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;SpecSelector&#x60;); it does **not** validate keys. The API enforces each key to match &#x60;SpecSelectorName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 

## Methods

### NewMappingCreUpdRequest

`func NewMappingCreUpdRequest(workloadType GVK, ) *MappingCreUpdRequest`

NewMappingCreUpdRequest instantiates a new MappingCreUpdRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMappingCreUpdRequestWithDefaults

`func NewMappingCreUpdRequestWithDefaults() *MappingCreUpdRequest`

NewMappingCreUpdRequestWithDefaults instantiates a new MappingCreUpdRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadType

`func (o *MappingCreUpdRequest) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *MappingCreUpdRequest) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *MappingCreUpdRequest) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.


### GetDescription

`func (o *MappingCreUpdRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *MappingCreUpdRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *MappingCreUpdRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *MappingCreUpdRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *MappingCreUpdRequest) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *MappingCreUpdRequest) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetSpecSelectors

`func (o *MappingCreUpdRequest) GetSpecSelectors() map[string]SpecSelector`

GetSpecSelectors returns the SpecSelectors field if non-nil, zero value otherwise.

### GetSpecSelectorsOk

`func (o *MappingCreUpdRequest) GetSpecSelectorsOk() (*map[string]SpecSelector, bool)`

GetSpecSelectorsOk returns a tuple with the SpecSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelectors

`func (o *MappingCreUpdRequest) SetSpecSelectors(v map[string]SpecSelector)`

SetSpecSelectors sets SpecSelectors field to given value.

### HasSpecSelectors

`func (o *MappingCreUpdRequest) HasSpecSelectors() bool`

HasSpecSelectors returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


