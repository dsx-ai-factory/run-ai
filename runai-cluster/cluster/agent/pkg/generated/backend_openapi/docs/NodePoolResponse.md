# NodePoolResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **float32** | The NodePool id | [optional] [readonly] 
**LabelKey** | Pointer to **string** | Key of node label for pool | [optional] 
**LabelValue** | Pointer to **string** | Value of node label for pool | [optional] 
**PlacementStrategy** | Pointer to [**PlacementStrategy**](PlacementStrategy.md) |  | [optional] 
**IsDefault** | Pointer to **bool** | Is the Node Pool default for cluster | [optional] 
**Status** | Pointer to **string** | Status of Node Pool | [optional] 
**StatusMessage** | Pointer to **string** | Message for status of Node Pool | [optional] 
**Nodes** | Pointer to **string** | List of Nodes that are assigned to this nodepool - as json string | [optional] 

## Methods

### NewNodePoolResponse

`func NewNodePoolResponse() *NodePoolResponse`

NewNodePoolResponse instantiates a new NodePoolResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolResponseWithDefaults

`func NewNodePoolResponseWithDefaults() *NodePoolResponse`

NewNodePoolResponseWithDefaults instantiates a new NodePoolResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NodePoolResponse) GetId() float32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NodePoolResponse) GetIdOk() (*float32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NodePoolResponse) SetId(v float32)`

SetId sets Id field to given value.

### HasId

`func (o *NodePoolResponse) HasId() bool`

HasId returns a boolean if a field has been set.

### GetLabelKey

`func (o *NodePoolResponse) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodePoolResponse) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodePoolResponse) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *NodePoolResponse) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### GetLabelValue

`func (o *NodePoolResponse) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodePoolResponse) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodePoolResponse) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *NodePoolResponse) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### GetPlacementStrategy

`func (o *NodePoolResponse) GetPlacementStrategy() PlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *NodePoolResponse) GetPlacementStrategyOk() (*PlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *NodePoolResponse) SetPlacementStrategy(v PlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *NodePoolResponse) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.

### GetIsDefault

`func (o *NodePoolResponse) GetIsDefault() bool`

GetIsDefault returns the IsDefault field if non-nil, zero value otherwise.

### GetIsDefaultOk

`func (o *NodePoolResponse) GetIsDefaultOk() (*bool, bool)`

GetIsDefaultOk returns a tuple with the IsDefault field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDefault

`func (o *NodePoolResponse) SetIsDefault(v bool)`

SetIsDefault sets IsDefault field to given value.

### HasIsDefault

`func (o *NodePoolResponse) HasIsDefault() bool`

HasIsDefault returns a boolean if a field has been set.

### GetStatus

`func (o *NodePoolResponse) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *NodePoolResponse) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *NodePoolResponse) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *NodePoolResponse) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetStatusMessage

`func (o *NodePoolResponse) GetStatusMessage() string`

GetStatusMessage returns the StatusMessage field if non-nil, zero value otherwise.

### GetStatusMessageOk

`func (o *NodePoolResponse) GetStatusMessageOk() (*string, bool)`

GetStatusMessageOk returns a tuple with the StatusMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusMessage

`func (o *NodePoolResponse) SetStatusMessage(v string)`

SetStatusMessage sets StatusMessage field to given value.

### HasStatusMessage

`func (o *NodePoolResponse) HasStatusMessage() bool`

HasStatusMessage returns a boolean if a field has been set.

### GetNodes

`func (o *NodePoolResponse) GetNodes() string`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *NodePoolResponse) GetNodesOk() (*string, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *NodePoolResponse) SetNodes(v string)`

SetNodes sets Nodes field to given value.

### HasNodes

`func (o *NodePoolResponse) HasNodes() bool`

HasNodes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


