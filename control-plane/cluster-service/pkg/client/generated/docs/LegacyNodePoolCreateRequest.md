# LegacyNodePoolCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Node Pool Name | [optional] 
**LabelKey** | Pointer to **string** | Label key for associated nodes to the Node Pool (with value as in labelValue) | [optional] 
**LabelValue** | Pointer to **string** | Label value for associated nodes to the Node Pool (with key as in labelKey) | [optional] 
**OverProvisioningRatio** | Pointer to **NullableInt32** |  | [optional] 
**PlacementStrategy** | Pointer to [**LegacyPlacementStrategy**](LegacyPlacementStrategy.md) |  | [optional] 

## Methods

### NewLegacyNodePoolCreateRequest

`func NewLegacyNodePoolCreateRequest() *LegacyNodePoolCreateRequest`

NewLegacyNodePoolCreateRequest instantiates a new LegacyNodePoolCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyNodePoolCreateRequestWithDefaults

`func NewLegacyNodePoolCreateRequestWithDefaults() *LegacyNodePoolCreateRequest`

NewLegacyNodePoolCreateRequestWithDefaults instantiates a new LegacyNodePoolCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *LegacyNodePoolCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *LegacyNodePoolCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *LegacyNodePoolCreateRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *LegacyNodePoolCreateRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetLabelKey

`func (o *LegacyNodePoolCreateRequest) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *LegacyNodePoolCreateRequest) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *LegacyNodePoolCreateRequest) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *LegacyNodePoolCreateRequest) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### GetLabelValue

`func (o *LegacyNodePoolCreateRequest) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *LegacyNodePoolCreateRequest) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *LegacyNodePoolCreateRequest) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *LegacyNodePoolCreateRequest) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.

### GetOverProvisioningRatio

`func (o *LegacyNodePoolCreateRequest) GetOverProvisioningRatio() int32`

GetOverProvisioningRatio returns the OverProvisioningRatio field if non-nil, zero value otherwise.

### GetOverProvisioningRatioOk

`func (o *LegacyNodePoolCreateRequest) GetOverProvisioningRatioOk() (*int32, bool)`

GetOverProvisioningRatioOk returns a tuple with the OverProvisioningRatio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOverProvisioningRatio

`func (o *LegacyNodePoolCreateRequest) SetOverProvisioningRatio(v int32)`

SetOverProvisioningRatio sets OverProvisioningRatio field to given value.

### HasOverProvisioningRatio

`func (o *LegacyNodePoolCreateRequest) HasOverProvisioningRatio() bool`

HasOverProvisioningRatio returns a boolean if a field has been set.

### SetOverProvisioningRatioNil

`func (o *LegacyNodePoolCreateRequest) SetOverProvisioningRatioNil(b bool)`

 SetOverProvisioningRatioNil sets the value for OverProvisioningRatio to be an explicit nil

### UnsetOverProvisioningRatio
`func (o *LegacyNodePoolCreateRequest) UnsetOverProvisioningRatio()`

UnsetOverProvisioningRatio ensures that no value is present for OverProvisioningRatio, not even an explicit nil
### GetPlacementStrategy

`func (o *LegacyNodePoolCreateRequest) GetPlacementStrategy() LegacyPlacementStrategy`

GetPlacementStrategy returns the PlacementStrategy field if non-nil, zero value otherwise.

### GetPlacementStrategyOk

`func (o *LegacyNodePoolCreateRequest) GetPlacementStrategyOk() (*LegacyPlacementStrategy, bool)`

GetPlacementStrategyOk returns a tuple with the PlacementStrategy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacementStrategy

`func (o *LegacyNodePoolCreateRequest) SetPlacementStrategy(v LegacyPlacementStrategy)`

SetPlacementStrategy sets PlacementStrategy field to given value.

### HasPlacementStrategy

`func (o *LegacyNodePoolCreateRequest) HasPlacementStrategy() bool`

HasPlacementStrategy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


