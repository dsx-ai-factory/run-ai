# BreakdownItemSelector

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Query** | Pointer to **string** | The JQ query used to filter items of an itemized field. Consult https://jqlang.org/manual/ for further info. | [optional] 
**ArgsInfo** | Pointer to [**map[string]MappingFieldType**](MappingFieldType.md) | The data type of each argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. | [optional] 
**ArgsEnum** | Pointer to **map[string][]string** | The allowed values for each string-typed argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. Up to 100 values per argument. | [optional] 

## Methods

### NewBreakdownItemSelector

`func NewBreakdownItemSelector() *BreakdownItemSelector`

NewBreakdownItemSelector instantiates a new BreakdownItemSelector object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBreakdownItemSelectorWithDefaults

`func NewBreakdownItemSelectorWithDefaults() *BreakdownItemSelector`

NewBreakdownItemSelectorWithDefaults instantiates a new BreakdownItemSelector object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetQuery

`func (o *BreakdownItemSelector) GetQuery() string`

GetQuery returns the Query field if non-nil, zero value otherwise.

### GetQueryOk

`func (o *BreakdownItemSelector) GetQueryOk() (*string, bool)`

GetQueryOk returns a tuple with the Query field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuery

`func (o *BreakdownItemSelector) SetQuery(v string)`

SetQuery sets Query field to given value.

### HasQuery

`func (o *BreakdownItemSelector) HasQuery() bool`

HasQuery returns a boolean if a field has been set.

### GetArgsInfo

`func (o *BreakdownItemSelector) GetArgsInfo() map[string]MappingFieldType`

GetArgsInfo returns the ArgsInfo field if non-nil, zero value otherwise.

### GetArgsInfoOk

`func (o *BreakdownItemSelector) GetArgsInfoOk() (*map[string]MappingFieldType, bool)`

GetArgsInfoOk returns a tuple with the ArgsInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsInfo

`func (o *BreakdownItemSelector) SetArgsInfo(v map[string]MappingFieldType)`

SetArgsInfo sets ArgsInfo field to given value.

### HasArgsInfo

`func (o *BreakdownItemSelector) HasArgsInfo() bool`

HasArgsInfo returns a boolean if a field has been set.

### GetArgsEnum

`func (o *BreakdownItemSelector) GetArgsEnum() map[string][]string`

GetArgsEnum returns the ArgsEnum field if non-nil, zero value otherwise.

### GetArgsEnumOk

`func (o *BreakdownItemSelector) GetArgsEnumOk() (*map[string][]string, bool)`

GetArgsEnumOk returns a tuple with the ArgsEnum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsEnum

`func (o *BreakdownItemSelector) SetArgsEnum(v map[string][]string)`

SetArgsEnum sets ArgsEnum field to given value.

### HasArgsEnum

`func (o *BreakdownItemSelector) HasArgsEnum() bool`

HasArgsEnum returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


