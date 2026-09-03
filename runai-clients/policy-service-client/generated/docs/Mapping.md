# Mapping

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadType** | [**GVK**](GVK.md) | The Kubernetes group, version, and kind of the workload type this mapping applies to. | 
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**SpecSelectors** | Pointer to [**map[string]SpecSelectorsWithScopeValue**](SpecSelectorsWithScopeValue.md) | A map of spec selectors. Map **keys** are selector names. OpenAPI &#x60;additionalProperties&#x60; only constrains the **value** shape (&#x60;SpecSelector&#x60;); it does **not** validate keys. The API enforces each key to match &#x60;SpecSelectorName&#x60; (1–63 chars, &#x60;^[A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9])?$&#x60; — alphanumeric with optional internal hyphens; no leading/trailing hyphen).  | [optional] 
**Id** | **string** | The unique identifier for the mapping. | 
**Scope** | [**MappingScope**](MappingScope.md) |  | 
**CreatedBy** | Pointer to **string** | The identifier of the user who created the mapping. | [optional] 
**CreatedAt** | Pointer to **time.Time** | The time at which the mapping was created. | [optional] 
**UpdatedBy** | Pointer to **string** | The identifier of the user who last updated the mapping. | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The time at which the mapping was last updated. | [optional] 

## Methods

### NewMapping

`func NewMapping(workloadType GVK, id string, scope MappingScope, ) *Mapping`

NewMapping instantiates a new Mapping object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMappingWithDefaults

`func NewMappingWithDefaults() *Mapping`

NewMappingWithDefaults instantiates a new Mapping object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadType

`func (o *Mapping) GetWorkloadType() GVK`

GetWorkloadType returns the WorkloadType field if non-nil, zero value otherwise.

### GetWorkloadTypeOk

`func (o *Mapping) GetWorkloadTypeOk() (*GVK, bool)`

GetWorkloadTypeOk returns a tuple with the WorkloadType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadType

`func (o *Mapping) SetWorkloadType(v GVK)`

SetWorkloadType sets WorkloadType field to given value.


### GetDescription

`func (o *Mapping) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Mapping) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Mapping) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Mapping) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *Mapping) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *Mapping) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetSpecSelectors

`func (o *Mapping) GetSpecSelectors() map[string]SpecSelectorsWithScopeValue`

GetSpecSelectors returns the SpecSelectors field if non-nil, zero value otherwise.

### GetSpecSelectorsOk

`func (o *Mapping) GetSpecSelectorsOk() (*map[string]SpecSelectorsWithScopeValue, bool)`

GetSpecSelectorsOk returns a tuple with the SpecSelectors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelectors

`func (o *Mapping) SetSpecSelectors(v map[string]SpecSelectorsWithScopeValue)`

SetSpecSelectors sets SpecSelectors field to given value.

### HasSpecSelectors

`func (o *Mapping) HasSpecSelectors() bool`

HasSpecSelectors returns a boolean if a field has been set.

### GetId

`func (o *Mapping) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Mapping) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Mapping) SetId(v string)`

SetId sets Id field to given value.


### GetScope

`func (o *Mapping) GetScope() MappingScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *Mapping) GetScopeOk() (*MappingScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *Mapping) SetScope(v MappingScope)`

SetScope sets Scope field to given value.


### GetCreatedBy

`func (o *Mapping) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Mapping) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Mapping) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *Mapping) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Mapping) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Mapping) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Mapping) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Mapping) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *Mapping) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *Mapping) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *Mapping) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *Mapping) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *Mapping) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Mapping) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Mapping) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *Mapping) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


