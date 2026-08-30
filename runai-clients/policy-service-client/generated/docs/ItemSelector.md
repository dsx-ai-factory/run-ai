# ItemSelector

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **NullableString** | A free-text description of the element. | [optional] 
**Query** | Pointer to **string** | The JQ query used to filter items of an itemized field. Consult https://jqlang.org/manual/ for further info. | [optional] 
**ArgsInfo** | Pointer to [**map[string]MappingFieldType**](MappingFieldType.md) | The data type of each argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. | [optional] 
**ArgsEnum** | Pointer to **map[string][]string** | The allowed values for each string-typed argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. Up to 100 values per argument. | [optional] 

## Methods

### NewItemSelector

`func NewItemSelector() *ItemSelector`

NewItemSelector instantiates a new ItemSelector object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewItemSelectorWithDefaults

`func NewItemSelectorWithDefaults() *ItemSelector`

NewItemSelectorWithDefaults instantiates a new ItemSelector object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *ItemSelector) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ItemSelector) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ItemSelector) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ItemSelector) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### SetDescriptionNil

`func (o *ItemSelector) SetDescriptionNil(b bool)`

 SetDescriptionNil sets the value for Description to be an explicit nil

### UnsetDescription
`func (o *ItemSelector) UnsetDescription()`

UnsetDescription ensures that no value is present for Description, not even an explicit nil
### GetQuery

`func (o *ItemSelector) GetQuery() string`

GetQuery returns the Query field if non-nil, zero value otherwise.

### GetQueryOk

`func (o *ItemSelector) GetQueryOk() (*string, bool)`

GetQueryOk returns a tuple with the Query field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuery

`func (o *ItemSelector) SetQuery(v string)`

SetQuery sets Query field to given value.

### HasQuery

`func (o *ItemSelector) HasQuery() bool`

HasQuery returns a boolean if a field has been set.

### GetArgsInfo

`func (o *ItemSelector) GetArgsInfo() map[string]MappingFieldType`

GetArgsInfo returns the ArgsInfo field if non-nil, zero value otherwise.

### GetArgsInfoOk

`func (o *ItemSelector) GetArgsInfoOk() (*map[string]MappingFieldType, bool)`

GetArgsInfoOk returns a tuple with the ArgsInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsInfo

`func (o *ItemSelector) SetArgsInfo(v map[string]MappingFieldType)`

SetArgsInfo sets ArgsInfo field to given value.

### HasArgsInfo

`func (o *ItemSelector) HasArgsInfo() bool`

HasArgsInfo returns a boolean if a field has been set.

### GetArgsEnum

`func (o *ItemSelector) GetArgsEnum() map[string][]string`

GetArgsEnum returns the ArgsEnum field if non-nil, zero value otherwise.

### GetArgsEnumOk

`func (o *ItemSelector) GetArgsEnumOk() (*map[string][]string, bool)`

GetArgsEnumOk returns a tuple with the ArgsEnum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsEnum

`func (o *ItemSelector) SetArgsEnum(v map[string][]string)`

SetArgsEnum sets ArgsEnum field to given value.

### HasArgsEnum

`func (o *ItemSelector) HasArgsEnum() bool`

HasArgsEnum returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


