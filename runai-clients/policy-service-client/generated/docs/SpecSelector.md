# SpecSelector

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Query** | Pointer to **string** | The JQ query used to select a subtree of the workload spec. Consult https://jqlang.org/manual/ for further info. | [optional] 
**ArgsInfo** | Pointer to [**map[string]MappingFieldType**](MappingFieldType.md) | The data type of each argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. | [optional] 
**ArgsEnum** | Pointer to **map[string][]string** | The allowed values for each string-typed argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. Up to 100 values per argument. | [optional] 
**FieldGroups** | Pointer to **[]string** | The field groups exposed by this spec selector. | [optional] 
**Dependency** | Pointer to [**NullableSpecSelectorDependency**](SpecSelectorDependency.md) |  | [optional] 

## Methods

### NewSpecSelector

`func NewSpecSelector() *SpecSelector`

NewSpecSelector instantiates a new SpecSelector object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecSelectorWithDefaults

`func NewSpecSelectorWithDefaults() *SpecSelector`

NewSpecSelectorWithDefaults instantiates a new SpecSelector object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *SpecSelector) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *SpecSelector) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *SpecSelector) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *SpecSelector) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *SpecSelector) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *SpecSelector) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetQuery

`func (o *SpecSelector) GetQuery() string`

GetQuery returns the Query field if non-nil, zero value otherwise.

### GetQueryOk

`func (o *SpecSelector) GetQueryOk() (*string, bool)`

GetQueryOk returns a tuple with the Query field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuery

`func (o *SpecSelector) SetQuery(v string)`

SetQuery sets Query field to given value.

### HasQuery

`func (o *SpecSelector) HasQuery() bool`

HasQuery returns a boolean if a field has been set.

### GetArgsInfo

`func (o *SpecSelector) GetArgsInfo() map[string]MappingFieldType`

GetArgsInfo returns the ArgsInfo field if non-nil, zero value otherwise.

### GetArgsInfoOk

`func (o *SpecSelector) GetArgsInfoOk() (*map[string]MappingFieldType, bool)`

GetArgsInfoOk returns a tuple with the ArgsInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsInfo

`func (o *SpecSelector) SetArgsInfo(v map[string]MappingFieldType)`

SetArgsInfo sets ArgsInfo field to given value.

### HasArgsInfo

`func (o *SpecSelector) HasArgsInfo() bool`

HasArgsInfo returns a boolean if a field has been set.

### GetArgsEnum

`func (o *SpecSelector) GetArgsEnum() map[string][]string`

GetArgsEnum returns the ArgsEnum field if non-nil, zero value otherwise.

### GetArgsEnumOk

`func (o *SpecSelector) GetArgsEnumOk() (*map[string][]string, bool)`

GetArgsEnumOk returns a tuple with the ArgsEnum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsEnum

`func (o *SpecSelector) SetArgsEnum(v map[string][]string)`

SetArgsEnum sets ArgsEnum field to given value.

### HasArgsEnum

`func (o *SpecSelector) HasArgsEnum() bool`

HasArgsEnum returns a boolean if a field has been set.

### GetFieldGroups

`func (o *SpecSelector) GetFieldGroups() []string`

GetFieldGroups returns the FieldGroups field if non-nil, zero value otherwise.

### GetFieldGroupsOk

`func (o *SpecSelector) GetFieldGroupsOk() (*[]string, bool)`

GetFieldGroupsOk returns a tuple with the FieldGroups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldGroups

`func (o *SpecSelector) SetFieldGroups(v []string)`

SetFieldGroups sets FieldGroups field to given value.

### HasFieldGroups

`func (o *SpecSelector) HasFieldGroups() bool`

HasFieldGroups returns a boolean if a field has been set.

### GetDependency

`func (o *SpecSelector) GetDependency() SpecSelectorDependency`

GetDependency returns the Dependency field if non-nil, zero value otherwise.

### GetDependencyOk

`func (o *SpecSelector) GetDependencyOk() (*SpecSelectorDependency, bool)`

GetDependencyOk returns a tuple with the Dependency field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependency

`func (o *SpecSelector) SetDependency(v SpecSelectorDependency)`

SetDependency sets Dependency field to given value.

### HasDependency

`func (o *SpecSelector) HasDependency() bool`

HasDependency returns a boolean if a field has been set.

### SetDependencyNil

`func (o *SpecSelector) SetDependencyNil(b bool)`

 SetDependencyNil sets the value for Dependency to be an explicit nil

### UnsetDependency
`func (o *SpecSelector) UnsetDependency()`

UnsetDependency ensures that no value is present for Dependency, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


