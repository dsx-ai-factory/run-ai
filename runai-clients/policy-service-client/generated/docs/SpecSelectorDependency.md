# SpecSelectorDependency

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SpecSelector** | Pointer to **NullableString** | Optional parent spec selector that provides iteration context. When set, the element path is created relative to each result of the parent selector. When omitted, element is relative to the workload root.  | [optional] 
**Path** | Pointer to **string** | Dot-separated path to ensure exists below the parent. Example: \&quot;template.metadata\&quot; creates both .template and .template.metadata if missing.  | [optional] 

## Methods

### NewSpecSelectorDependency

`func NewSpecSelectorDependency() *SpecSelectorDependency`

NewSpecSelectorDependency instantiates a new SpecSelectorDependency object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecSelectorDependencyWithDefaults

`func NewSpecSelectorDependencyWithDefaults() *SpecSelectorDependency`

NewSpecSelectorDependencyWithDefaults instantiates a new SpecSelectorDependency object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSpecSelector

`func (o *SpecSelectorDependency) GetSpecSelector() string`

GetSpecSelector returns the SpecSelector field if non-nil, zero value otherwise.

### GetSpecSelectorOk

`func (o *SpecSelectorDependency) GetSpecSelectorOk() (*string, bool)`

GetSpecSelectorOk returns a tuple with the SpecSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelector

`func (o *SpecSelectorDependency) SetSpecSelector(v string)`

SetSpecSelector sets SpecSelector field to given value.

### HasSpecSelector

`func (o *SpecSelectorDependency) HasSpecSelector() bool`

HasSpecSelector returns a boolean if a field has been set.

### SetSpecSelectorNil

`func (o *SpecSelectorDependency) SetSpecSelectorNil(b bool)`

 SetSpecSelectorNil sets the value for SpecSelector to be an explicit nil

### UnsetSpecSelector
`func (o *SpecSelectorDependency) UnsetSpecSelector()`

UnsetSpecSelector ensures that no value is present for SpecSelector, not even an explicit nil
### GetPath

`func (o *SpecSelectorDependency) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *SpecSelectorDependency) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *SpecSelectorDependency) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *SpecSelectorDependency) HasPath() bool`

HasPath returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


