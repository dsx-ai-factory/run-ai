# SpecSelectorsWithScopeValue

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Query** | Pointer to **string** | The JQ query used to select a subtree of the workload spec. Consult https://jqlang.org/manual/ for further info. | [optional] 
**ArgsInfo** | Pointer to [**map[string]MappingFieldType**](MappingFieldType.md) | The data type of each argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. | [optional] 
**ArgsEnum** | Pointer to **map[string][]string** | The allowed values for each string-typed argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. Up to 100 values per argument. | [optional] 
**FieldGroups** | Pointer to **[]string** | The field groups exposed by this spec selector. | [optional] 
**Dependency** | Pointer to [**NullableSpecSelectorDependency**](SpecSelectorDependency.md) |  | [optional] 
**Scope** | [**MappingScope**](MappingScope.md) |  | 

## Methods

### NewSpecSelectorsWithScopeValue

`func NewSpecSelectorsWithScopeValue(scope MappingScope, ) *SpecSelectorsWithScopeValue`

NewSpecSelectorsWithScopeValue instantiates a new SpecSelectorsWithScopeValue object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecSelectorsWithScopeValueWithDefaults

`func NewSpecSelectorsWithScopeValueWithDefaults() *SpecSelectorsWithScopeValue`

NewSpecSelectorsWithScopeValueWithDefaults instantiates a new SpecSelectorsWithScopeValue object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *SpecSelectorsWithScopeValue) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *SpecSelectorsWithScopeValue) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *SpecSelectorsWithScopeValue) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *SpecSelectorsWithScopeValue) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *SpecSelectorsWithScopeValue) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *SpecSelectorsWithScopeValue) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetQuery

`func (o *SpecSelectorsWithScopeValue) GetQuery() string`

GetQuery returns the Query field if non-nil, zero value otherwise.

### GetQueryOk

`func (o *SpecSelectorsWithScopeValue) GetQueryOk() (*string, bool)`

GetQueryOk returns a tuple with the Query field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuery

`func (o *SpecSelectorsWithScopeValue) SetQuery(v string)`

SetQuery sets Query field to given value.

### HasQuery

`func (o *SpecSelectorsWithScopeValue) HasQuery() bool`

HasQuery returns a boolean if a field has been set.

### GetArgsInfo

`func (o *SpecSelectorsWithScopeValue) GetArgsInfo() map[string]MappingFieldType`

GetArgsInfo returns the ArgsInfo field if non-nil, zero value otherwise.

### GetArgsInfoOk

`func (o *SpecSelectorsWithScopeValue) GetArgsInfoOk() (*map[string]MappingFieldType, bool)`

GetArgsInfoOk returns a tuple with the ArgsInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsInfo

`func (o *SpecSelectorsWithScopeValue) SetArgsInfo(v map[string]MappingFieldType)`

SetArgsInfo sets ArgsInfo field to given value.

### HasArgsInfo

`func (o *SpecSelectorsWithScopeValue) HasArgsInfo() bool`

HasArgsInfo returns a boolean if a field has been set.

### GetArgsEnum

`func (o *SpecSelectorsWithScopeValue) GetArgsEnum() map[string][]string`

GetArgsEnum returns the ArgsEnum field if non-nil, zero value otherwise.

### GetArgsEnumOk

`func (o *SpecSelectorsWithScopeValue) GetArgsEnumOk() (*map[string][]string, bool)`

GetArgsEnumOk returns a tuple with the ArgsEnum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsEnum

`func (o *SpecSelectorsWithScopeValue) SetArgsEnum(v map[string][]string)`

SetArgsEnum sets ArgsEnum field to given value.

### HasArgsEnum

`func (o *SpecSelectorsWithScopeValue) HasArgsEnum() bool`

HasArgsEnum returns a boolean if a field has been set.

### GetFieldGroups

`func (o *SpecSelectorsWithScopeValue) GetFieldGroups() []string`

GetFieldGroups returns the FieldGroups field if non-nil, zero value otherwise.

### GetFieldGroupsOk

`func (o *SpecSelectorsWithScopeValue) GetFieldGroupsOk() (*[]string, bool)`

GetFieldGroupsOk returns a tuple with the FieldGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldGroups

`func (o *SpecSelectorsWithScopeValue) SetFieldGroups(v []string)`

SetFieldGroups sets FieldGroups field to given value.

### HasFieldGroups

`func (o *SpecSelectorsWithScopeValue) HasFieldGroups() bool`

HasFieldGroups returns a boolean if a field has been set.

### GetDependency

`func (o *SpecSelectorsWithScopeValue) GetDependency() SpecSelectorDependency`

GetDependency returns the Dependency field if non-nil, zero value otherwise.

### GetDependencyOk

`func (o *SpecSelectorsWithScopeValue) GetDependencyOk() (*SpecSelectorDependency, bool)`

GetDependencyOk returns a tuple with the Dependency field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependency

`func (o *SpecSelectorsWithScopeValue) SetDependency(v SpecSelectorDependency)`

SetDependency sets Dependency field to given value.

### HasDependency

`func (o *SpecSelectorsWithScopeValue) HasDependency() bool`

HasDependency returns a boolean if a field has been set.

### SetDependencyNil

`func (o *SpecSelectorsWithScopeValue) SetDependencyNil(b bool)`

 SetDependencyNil sets the value for Dependency to be an explicit nil

### UnsetDependency
`func (o *SpecSelectorsWithScopeValue) UnsetDependency()`

UnsetDependency ensures that no value is present for Dependency, not even an explicit nil
### GetScope

`func (o *SpecSelectorsWithScopeValue) GetScope() MappingScope`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *SpecSelectorsWithScopeValue) GetScopeOk() (*MappingScope, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *SpecSelectorsWithScopeValue) SetScope(v MappingScope)`

SetScope sets Scope field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


