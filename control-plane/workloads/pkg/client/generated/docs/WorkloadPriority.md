# WorkloadPriority

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier of the workload priority | 
**Source** | Pointer to **string** | The source of the priority class. Indicates where the priority definition originated. Currently, the value is either Run:ai or empty. | [optional] 
**Name** | **string** | The name of the workload priority (e.g., low, medium, high) | 
**Value** | **int32** | The scheduling value assigned to this priority | 
**CreatedAt** | Pointer to **time.Time** | The timestamp for when the priority was created | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The timestamp for when the priority was updated | [optional] 
**AggregatedPhase** | Pointer to **string** | Combined status of the priority across all clusters | [optional] 
**Status** | Pointer to [**[]PriorityStatus**](PriorityStatus.md) | Per-cluster status details | [optional] 
**ScopeType** | Pointer to **string** | The type of scope the workload priority applies to. | [optional] 
**ScopeId** | Pointer to **string** | The unique identifier of the scope the workload priority applies to. | [optional] 

## Methods

### NewWorkloadPriority

`func NewWorkloadPriority(id string, name string, value int32, ) *WorkloadPriority`

NewWorkloadPriority instantiates a new WorkloadPriority object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadPriorityWithDefaults

`func NewWorkloadPriorityWithDefaults() *WorkloadPriority`

NewWorkloadPriorityWithDefaults instantiates a new WorkloadPriority object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadPriority) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadPriority) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadPriority) SetId(v string)`

SetId sets Id field to given value.


### GetSource

`func (o *WorkloadPriority) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *WorkloadPriority) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *WorkloadPriority) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *WorkloadPriority) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetName

`func (o *WorkloadPriority) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadPriority) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadPriority) SetName(v string)`

SetName sets Name field to given value.


### GetValue

`func (o *WorkloadPriority) GetValue() int32`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *WorkloadPriority) GetValueOk() (*int32, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *WorkloadPriority) SetValue(v int32)`

SetValue sets Value field to given value.


### GetCreatedAt

`func (o *WorkloadPriority) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadPriority) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadPriority) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkloadPriority) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkloadPriority) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkloadPriority) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkloadPriority) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkloadPriority) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetAggregatedPhase

`func (o *WorkloadPriority) GetAggregatedPhase() string`

GetAggregatedPhase returns the AggregatedPhase field if non-nil, zero value otherwise.

### GetAggregatedPhaseOk

`func (o *WorkloadPriority) GetAggregatedPhaseOk() (*string, bool)`

GetAggregatedPhaseOk returns a tuple with the AggregatedPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggregatedPhase

`func (o *WorkloadPriority) SetAggregatedPhase(v string)`

SetAggregatedPhase sets AggregatedPhase field to given value.

### HasAggregatedPhase

`func (o *WorkloadPriority) HasAggregatedPhase() bool`

HasAggregatedPhase returns a boolean if a field has been set.

### GetStatus

`func (o *WorkloadPriority) GetStatus() []PriorityStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkloadPriority) GetStatusOk() (*[]PriorityStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkloadPriority) SetStatus(v []PriorityStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *WorkloadPriority) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetScopeType

`func (o *WorkloadPriority) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *WorkloadPriority) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *WorkloadPriority) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.

### HasScopeType

`func (o *WorkloadPriority) HasScopeType() bool`

HasScopeType returns a boolean if a field has been set.

### GetScopeId

`func (o *WorkloadPriority) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *WorkloadPriority) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *WorkloadPriority) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.

### HasScopeId

`func (o *WorkloadPriority) HasScopeId() bool`

HasScopeId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


