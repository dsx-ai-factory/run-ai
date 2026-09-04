# WorkloadTypeStatusFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to **string** | The current phase of the workload type in the cluster. | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Version** | Pointer to **string** | The version of the resource definition&#39;s Kubernetes CRD. | [optional] 

## Methods

### NewWorkloadTypeStatusFields

`func NewWorkloadTypeStatusFields() *WorkloadTypeStatusFields`

NewWorkloadTypeStatusFields instantiates a new WorkloadTypeStatusFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeStatusFieldsWithDefaults

`func NewWorkloadTypeStatusFieldsWithDefaults() *WorkloadTypeStatusFields`

NewWorkloadTypeStatusFieldsWithDefaults instantiates a new WorkloadTypeStatusFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *WorkloadTypeStatusFields) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *WorkloadTypeStatusFields) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *WorkloadTypeStatusFields) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *WorkloadTypeStatusFields) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetConditions

`func (o *WorkloadTypeStatusFields) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *WorkloadTypeStatusFields) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *WorkloadTypeStatusFields) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *WorkloadTypeStatusFields) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetVersion

`func (o *WorkloadTypeStatusFields) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadTypeStatusFields) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadTypeStatusFields) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadTypeStatusFields) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


