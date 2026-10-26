# WorkloadTopologyActualPlacement

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Level** | Pointer to **string** | The network topology level at which placement was achieved (for example, rack, block, or zone). | [optional] 
**Value** | Pointer to **string** | The specific topology value at that level (for example, rack-1 or block-a). | [optional] 

## Methods

### NewWorkloadTopologyActualPlacement

`func NewWorkloadTopologyActualPlacement() *WorkloadTopologyActualPlacement`

NewWorkloadTopologyActualPlacement instantiates a new WorkloadTopologyActualPlacement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTopologyActualPlacementWithDefaults

`func NewWorkloadTopologyActualPlacementWithDefaults() *WorkloadTopologyActualPlacement`

NewWorkloadTopologyActualPlacementWithDefaults instantiates a new WorkloadTopologyActualPlacement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLevel

`func (o *WorkloadTopologyActualPlacement) GetLevel() string`

GetLevel returns the Level field if non-nil, zero value otherwise.

### GetLevelOk

`func (o *WorkloadTopologyActualPlacement) GetLevelOk() (*string, bool)`

GetLevelOk returns a tuple with the Level field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevel

`func (o *WorkloadTopologyActualPlacement) SetLevel(v string)`

SetLevel sets Level field to given value.

### HasLevel

`func (o *WorkloadTopologyActualPlacement) HasLevel() bool`

HasLevel returns a boolean if a field has been set.

### GetValue

`func (o *WorkloadTopologyActualPlacement) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *WorkloadTopologyActualPlacement) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *WorkloadTopologyActualPlacement) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *WorkloadTopologyActualPlacement) HasValue() bool`

HasValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


