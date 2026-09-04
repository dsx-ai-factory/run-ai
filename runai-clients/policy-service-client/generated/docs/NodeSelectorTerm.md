# NodeSelectorTerm

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MatchExpressions** | Pointer to [**[]MatchExpression**](MatchExpression.md) | A list of node selector requirements by node&#39;s labels. | [optional] 

## Methods

### NewNodeSelectorTerm

`func NewNodeSelectorTerm() *NodeSelectorTerm`

NewNodeSelectorTerm instantiates a new NodeSelectorTerm object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeSelectorTermWithDefaults

`func NewNodeSelectorTermWithDefaults() *NodeSelectorTerm`

NewNodeSelectorTermWithDefaults instantiates a new NodeSelectorTerm object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMatchExpressions

`func (o *NodeSelectorTerm) GetMatchExpressions() []MatchExpression`

GetMatchExpressions returns the MatchExpressions field if non-nil, zero value otherwise.

### GetMatchExpressionsOk

`func (o *NodeSelectorTerm) GetMatchExpressionsOk() (*[]MatchExpression, bool)`

GetMatchExpressionsOk returns a tuple with the MatchExpressions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMatchExpressions

`func (o *NodeSelectorTerm) SetMatchExpressions(v []MatchExpression)`

SetMatchExpressions sets MatchExpressions field to given value.

### HasMatchExpressions

`func (o *NodeSelectorTerm) HasMatchExpressions() bool`

HasMatchExpressions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


