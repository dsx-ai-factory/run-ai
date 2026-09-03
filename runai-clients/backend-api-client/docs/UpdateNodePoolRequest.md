# UpdateNodePoolRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LabelKey** | Pointer to **string** | key of node label for pool | [optional] 
**LabelValue** | Pointer to **string** | value of node label for pool | [optional] 
**OverProvisioningRatio** | Pointer to **int32** |  | [optional] [default to 1]
**PlacementStrategy** | Pointer to [**PlacementStrategy**](PlacementStrategy.md) |  | [optional] 

## Methods

### NewUpdateNodePoolRequest

`func NewUpdateNodePoolRequest() *UpdateNodePoolRequest`

NewUpdateNodePoolRequest instantiates a new UpdateNodePoolRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateNodePoolRequestWithDefaults

`func NewUpdateNodePoolRequestWithDefaults() *UpdateNodePoolRequest`

NewUpdateNodePoolRequestWithDefaults instantiates a new UpdateNodePoolRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabelKey

`func (o *UpdateNodePoolRequest) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *UpdateNodePoolRequest) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *UpdateNodePoolRequest) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *UpdateNodePoolRequest) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### GetLabelValue

`func (o *UpdateNodePoolRequest) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *UpdateNodePoolRequest) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *UpdateNodePoolRequest) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *UpdateNodePoolRequest) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### GetOverProvisioningRatio

`func (o *UpdateNodePoolRequest) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *UpdateNodePoolRequest) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *UpdateNodePoolRequest) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *UpdateNodePoolRequest) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### GetPlacementStrategy

`func (o *UpdateNodePoolRequest) GetPlacementStrategy() PlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *UpdateNodePoolRequest) GetPlacementStrategyOk() (*PlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *UpdateNodePoolRequest) SetPlacementStrategy(v PlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *UpdateNodePoolRequest) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


