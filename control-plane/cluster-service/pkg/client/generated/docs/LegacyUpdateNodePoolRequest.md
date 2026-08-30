# LegacyUpdateNodePoolRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LabelKey** | Pointer to **string** | key of node label for pool | [optional] 
**LabelValue** | Pointer to **string** | value of node label for pool | [optional] 
**OverProvisioningRatio** | Pointer to **NullableInt32** |  | [optional] 
**PlacementStrategy** | Pointer to [**LegacyPlacementStrategy**](LegacyPlacementStrategy.md) |  | [optional] 

## Methods

### NewLegacyUpdateNodePoolRequest

`func NewLegacyUpdateNodePoolRequest() *LegacyUpdateNodePoolRequest`

NewLegacyUpdateNodePoolRequest instantiates a new LegacyUpdateNodePoolRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyUpdateNodePoolRequestWithDefaults

`func NewLegacyUpdateNodePoolRequestWithDefaults() *LegacyUpdateNodePoolRequest`

NewLegacyUpdateNodePoolRequestWithDefaults instantiates a new LegacyUpdateNodePoolRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabelKey

`func (o *LegacyUpdateNodePoolRequest) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *LegacyUpdateNodePoolRequest) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *LegacyUpdateNodePoolRequest) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *LegacyUpdateNodePoolRequest) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### GetLabelValue

`func (o *LegacyUpdateNodePoolRequest) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *LegacyUpdateNodePoolRequest) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *LegacyUpdateNodePoolRequest) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *LegacyUpdateNodePoolRequest) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### GetOverProvisioningRatio

`func (o *LegacyUpdateNodePoolRequest) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *LegacyUpdateNodePoolRequest) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *LegacyUpdateNodePoolRequest) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *LegacyUpdateNodePoolRequest) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### SetOverProvisioningRatioNil

`func (o *LegacyUpdateNodePoolRequest) SetOverProvisioningRatioNil(b bool)`

 SetOverProvisioningRatioNil sets the value for OverProvisioningRatio to be an explicit nil

### UnsetOverProvisioningRatio
`func (o *LegacyUpdateNodePoolRequest) UnsetOverProvisioningRatio()`

UnsetOverProvisioningRatio ensures that no value is present for OverProvisioningRatio, not even an explicit nil
### GetPlacementStrategy

`func (o *LegacyUpdateNodePoolRequest) GetPlacementStrategy() LegacyPlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *LegacyUpdateNodePoolRequest) GetPlacementStrategyOk() (*LegacyPlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *LegacyUpdateNodePoolRequest) SetPlacementStrategy(v LegacyPlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *LegacyUpdateNodePoolRequest) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


