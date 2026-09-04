# MatchExpression

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Key** | **string** | The label key that the selector applies to (mandatory). | 
**Operator** | [**MatchExpressionOperator**](MatchExpressionOperator.md) |  | 
**Values** | Pointer to **[]string** | An array of string values. If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. If the operator is Gt or Lt, the values array must have a single element, which will be interpreted as an integer. | [optional] 

## Methods

### NewMatchExpression

`func NewMatchExpression(key string, operator MatchExpressionOperator, ) *MatchExpression`

NewMatchExpression instantiates a new MatchExpression object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMatchExpressionWithDefaults

`func NewMatchExpressionWithDefaults() *MatchExpression`

NewMatchExpressionWithDefaults instantiates a new MatchExpression object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKey

`func (o *MatchExpression) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *MatchExpression) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *MatchExpression) SetKey(v string)`

SetKey sets Key field to given value.


### GetOperator

`func (o *MatchExpression) GetOperator() MatchExpressionOperator`

GetOperator returns the Operator field if non-nil, zero value otherwise.

### GetOperatorOk

`func (o *MatchExpression) GetOperatorOk() (*MatchExpressionOperator, bool)`

GetOperatorOk returns a tuple with the Operator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperator

`func (o *MatchExpression) SetOperator(v MatchExpressionOperator)`

SetOperator sets Operator field to given value.


### GetValues

`func (o *MatchExpression) GetValues() []string`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *MatchExpression) GetValuesOk() (*[]string, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *MatchExpression) SetValues(v []string)`

SetValues sets Values field to given value.

### HasValues

`func (o *MatchExpression) HasValues() bool`

HasValues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


