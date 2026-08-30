# FormatterField

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SetExpr** | Pointer to **NullableString** | The JQ expression used to write the field value into the workload spec. Consult https://jqlang.org/manual/ for further info. | [optional] 
**GetExpr** | Pointer to **NullableString** | The JQ expression used to read the field value from the workload spec. Consult https://jqlang.org/manual/ for further info. | [optional] 
**ArrayKey** | Pointer to [**NullableFormatterArrayKey**](FormatterArrayKey.md) |  | [optional] 
**ArgsInfo** | Pointer to [**map[string]MappingFieldType**](MappingFieldType.md) | The data type of each argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. | [optional] 
**ArgsEnum** | Pointer to **map[string][]string** | The allowed values for each string-typed argument referenced by the JQ expression as $ARGS.named.&lt;argName&gt;. Up to 100 values per argument. | [optional] 

## Methods

### NewFormatterField

`func NewFormatterField() *FormatterField`

NewFormatterField instantiates a new FormatterField object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFormatterFieldWithDefaults

`func NewFormatterFieldWithDefaults() *FormatterField`

NewFormatterFieldWithDefaults instantiates a new FormatterField object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSetExpr

`func (o *FormatterField) GetSetExpr() string`

GetSetExpr returns the SetExpr field if non-nil, zero value otherwise.

### GetSetExprOk

`func (o *FormatterField) GetSetExprOk() (*string, bool)`

GetSetExprOk returns a tuple with the SetExpr field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSetExpr

`func (o *FormatterField) SetSetExpr(v string)`

SetSetExpr sets SetExpr field to given value.

### HasSetExpr

`func (o *FormatterField) HasSetExpr() bool`

HasSetExpr returns a boolean if a field has been set.

### SetSetExprNil

`func (o *FormatterField) SetSetExprNil(b bool)`

 SetSetExprNil sets the value for SetExpr to be an explicit nil

### UnsetSetExpr
`func (o *FormatterField) UnsetSetExpr()`

UnsetSetExpr ensures that no value is present for SetExpr, not even an explicit nil
### GetGetExpr

`func (o *FormatterField) GetGetExpr() string`

GetGetExpr returns the GetExpr field if non-nil, zero value otherwise.

### GetGetExprOk

`func (o *FormatterField) GetGetExprOk() (*string, bool)`

GetGetExprOk returns a tuple with the GetExpr field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGetExpr

`func (o *FormatterField) SetGetExpr(v string)`

SetGetExpr sets GetExpr field to given value.

### HasGetExpr

`func (o *FormatterField) HasGetExpr() bool`

HasGetExpr returns a boolean if a field has been set.

### SetGetExprNil

`func (o *FormatterField) SetGetExprNil(b bool)`

 SetGetExprNil sets the value for GetExpr to be an explicit nil

### UnsetGetExpr
`func (o *FormatterField) UnsetGetExpr()`

UnsetGetExpr ensures that no value is present for GetExpr, not even an explicit nil
### GetArrayKey

`func (o *FormatterField) GetArrayKey() FormatterArrayKey`

GetArrayKey returns the ArrayKey field if non-nil, zero value otherwise.

### GetArrayKeyOk

`func (o *FormatterField) GetArrayKeyOk() (*FormatterArrayKey, bool)`

GetArrayKeyOk returns a tuple with the ArrayKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArrayKey

`func (o *FormatterField) SetArrayKey(v FormatterArrayKey)`

SetArrayKey sets ArrayKey field to given value.

### HasArrayKey

`func (o *FormatterField) HasArrayKey() bool`

HasArrayKey returns a boolean if a field has been set.

### SetArrayKeyNil

`func (o *FormatterField) SetArrayKeyNil(b bool)`

 SetArrayKeyNil sets the value for ArrayKey to be an explicit nil

### UnsetArrayKey
`func (o *FormatterField) UnsetArrayKey()`

UnsetArrayKey ensures that no value is present for ArrayKey, not even an explicit nil
### GetArgsInfo

`func (o *FormatterField) GetArgsInfo() map[string]MappingFieldType`

GetArgsInfo returns the ArgsInfo field if non-nil, zero value otherwise.

### GetArgsInfoOk

`func (o *FormatterField) GetArgsInfoOk() (*map[string]MappingFieldType, bool)`

GetArgsInfoOk returns a tuple with the ArgsInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsInfo

`func (o *FormatterField) SetArgsInfo(v map[string]MappingFieldType)`

SetArgsInfo sets ArgsInfo field to given value.

### HasArgsInfo

`func (o *FormatterField) HasArgsInfo() bool`

HasArgsInfo returns a boolean if a field has been set.

### GetArgsEnum

`func (o *FormatterField) GetArgsEnum() map[string][]string`

GetArgsEnum returns the ArgsEnum field if non-nil, zero value otherwise.

### GetArgsEnumOk

`func (o *FormatterField) GetArgsEnumOk() (*map[string][]string, bool)`

GetArgsEnumOk returns a tuple with the ArgsEnum field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgsEnum

`func (o *FormatterField) SetArgsEnum(v map[string][]string)`

SetArgsEnum sets ArgsEnum field to given value.

### HasArgsEnum

`func (o *FormatterField) HasArgsEnum() bool`

HasArgsEnum returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


