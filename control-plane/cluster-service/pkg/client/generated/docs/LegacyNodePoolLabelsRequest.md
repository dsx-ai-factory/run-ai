# LegacyNodePoolLabelsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LabelKey** | Pointer to **string** | key of node label for pool | [optional] 
**LabelValue** | Pointer to **string** | value of node label for pool | [optional] 

## Methods

### NewLegacyNodePoolLabelsRequest

`func NewLegacyNodePoolLabelsRequest() *LegacyNodePoolLabelsRequest`

NewLegacyNodePoolLabelsRequest instantiates a new LegacyNodePoolLabelsRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyNodePoolLabelsRequestWithDefaults

`func NewLegacyNodePoolLabelsRequestWithDefaults() *LegacyNodePoolLabelsRequest`

NewLegacyNodePoolLabelsRequestWithDefaults instantiates a new LegacyNodePoolLabelsRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabelKey

`func (o *LegacyNodePoolLabelsRequest) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *LegacyNodePoolLabelsRequest) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *LegacyNodePoolLabelsRequest) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.

### HasLabelKey

`func (o *LegacyNodePoolLabelsRequest) HasLabelKey() bool`

HasLabelKey returns a boolean if a field has been set.

### GetLabelValue

`func (o *LegacyNodePoolLabelsRequest) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *LegacyNodePoolLabelsRequest) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *LegacyNodePoolLabelsRequest) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.

### HasLabelValue

`func (o *LegacyNodePoolLabelsRequest) HasLabelValue() bool`

HasLabelValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


