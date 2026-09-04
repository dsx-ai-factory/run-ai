# NodePoolCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Node Pool Name | [optional] 
**OverProvisioningRatio** | Pointer to **int32** |  | [optional] [default to 1]
**LabelKey** | Pointer to **string** | Label key for associated nodes to the Node Pool (with value as in labelValue) | [optional] 
**LabelValue** | Pointer to **string** | Label value for associated nodes to the Node Pool (with key as in labelKey) | [optional] 
**PlacementStrategy** | Pointer to [**PlacementStrategy**](PlacementStrategy.md) |  | [optional] 

## Methods

### NewNodePoolCreateRequest

`func NewNodePoolCreateRequest() *NodePoolCreateRequest`

NewNodePoolCreateRequest instantiates a new NodePoolCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodePoolCreateRequestWithDefaults

`func NewNodePoolCreateRequestWithDefaults() *NodePoolCreateRequest`

NewNodePoolCreateRequestWithDefaults instantiates a new NodePoolCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *NodePoolCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *NodePoolCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *NodePoolCreateRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *NodePoolCreateRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetOverProvisioningRatio

`func (o *NodePoolCreateRequest) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *NodePoolCreateRequest) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *NodePoolCreateRequest) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *NodePoolCreateRequest) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### GetLabelKey

`func (o *NodePoolCreateRequest) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NodePoolCreateRequest) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NodePoolCreateRequest) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *NodePoolCreateRequest) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### GetLabelValue

`func (o *NodePoolCreateRequest) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NodePoolCreateRequest) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NodePoolCreateRequest) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *NodePoolCreateRequest) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### GetPlacementStrategy

`func (o *NodePoolCreateRequest) GetPlacementStrategy() PlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *NodePoolCreateRequest) GetPlacementStrategyOk() (*PlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *NodePoolCreateRequest) SetPlacementStrategy(v PlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *NodePoolCreateRequest) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


