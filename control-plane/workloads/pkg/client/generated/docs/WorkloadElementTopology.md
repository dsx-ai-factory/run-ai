# WorkloadElementTopology

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | The name of the network topology associated with the element — either as requested via constraints, or inherited from the workload&#39;s topology constraints, or inferred from the nodepools the element&#39;s running pods are scheduled on. | [optional] 
**PreferredPlacement** | Pointer to **string** | The topology level to use as a preferred placement constraint for the element. | [optional] 
**RequiredPlacement** | Pointer to **string** | The topology level to use as a required placement constraint for the element. If the requirement cannot be satisfied at this topology level, placement will fail. | [optional] 
**ActualPlacement** | Pointer to [**NullableWorkloadTopologyActualPlacement**](WorkloadTopologyActualPlacement.md) |  | [optional] 

## Methods

### NewWorkloadElementTopology

`func NewWorkloadElementTopology() *WorkloadElementTopology`

NewWorkloadElementTopology instantiates a new WorkloadElementTopology object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadElementTopologyWithDefaults

`func NewWorkloadElementTopologyWithDefaults() *WorkloadElementTopology`

NewWorkloadElementTopologyWithDefaults instantiates a new WorkloadElementTopology object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadElementTopology) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadElementTopology) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadElementTopology) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkloadElementTopology) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPreferredPlacement

`func (o *WorkloadElementTopology) GetPreferredPlacement() string`

GetPreferredPlacement returns the PreferredPlacement field if non-nil, zero value otherwise.

### GetPreferredPlacementOk

`func (o *WorkloadElementTopology) GetPreferredPlacementOk() (*string, bool)`

GetPreferredPlacementOk returns a tuple with the PreferredPlacement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreferredPlacement

`func (o *WorkloadElementTopology) SetPreferredPlacement(v string)`

SetPreferredPlacement sets PreferredPlacement field to given value.

### HasPreferredPlacement

`func (o *WorkloadElementTopology) HasPreferredPlacement() bool`

HasPreferredPlacement returns a boolean if a field has been set.

### GetRequiredPlacement

`func (o *WorkloadElementTopology) GetRequiredPlacement() string`

GetRequiredPlacement returns the RequiredPlacement field if non-nil, zero value otherwise.

### GetRequiredPlacementOk

`func (o *WorkloadElementTopology) GetRequiredPlacementOk() (*string, bool)`

GetRequiredPlacementOk returns a tuple with the RequiredPlacement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequiredPlacement

`func (o *WorkloadElementTopology) SetRequiredPlacement(v string)`

SetRequiredPlacement sets RequiredPlacement field to given value.

### HasRequiredPlacement

`func (o *WorkloadElementTopology) HasRequiredPlacement() bool`

HasRequiredPlacement returns a boolean if a field has been set.

### GetActualPlacement

`func (o *WorkloadElementTopology) GetActualPlacement() WorkloadTopologyActualPlacement`

GetActualPlacement returns the ActualPlacement field if non-nil, zero value otherwise.

### GetActualPlacementOk

`func (o *WorkloadElementTopology) GetActualPlacementOk() (*WorkloadTopologyActualPlacement, bool)`

GetActualPlacementOk returns a tuple with the ActualPlacement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualPlacement

`func (o *WorkloadElementTopology) SetActualPlacement(v WorkloadTopologyActualPlacement)`

SetActualPlacement sets ActualPlacement field to given value.

### HasActualPlacement

`func (o *WorkloadElementTopology) HasActualPlacement() bool`

HasActualPlacement returns a boolean if a field has been set.

### SetActualPlacementNil

`func (o *WorkloadElementTopology) SetActualPlacementNil(b bool)`

 SetActualPlacementNil sets the value for ActualPlacement to be an explicit nil

### UnsetActualPlacement
`func (o *WorkloadElementTopology) UnsetActualPlacement()`

UnsetActualPlacement ensures that no value is present for ActualPlacement, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


