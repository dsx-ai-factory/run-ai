# WorkloadTypeStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to **string** | The current phase of the workload type in the cluster. | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Version** | Pointer to **string** | The version of the resource definition&#39;s Kubernetes CRD. | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**UpdatedAt** | Pointer to **time.Time** | The date and time when the workload type status of this cluster was last updated. | [optional] 

## Methods

### NewWorkloadTypeStatus

`func NewWorkloadTypeStatus() *WorkloadTypeStatus`

NewWorkloadTypeStatus instantiates a new WorkloadTypeStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeStatusWithDefaults

`func NewWorkloadTypeStatusWithDefaults() *WorkloadTypeStatus`

NewWorkloadTypeStatusWithDefaults instantiates a new WorkloadTypeStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *WorkloadTypeStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *WorkloadTypeStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *WorkloadTypeStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *WorkloadTypeStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetConditions

`func (o *WorkloadTypeStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *WorkloadTypeStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *WorkloadTypeStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *WorkloadTypeStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetVersion

`func (o *WorkloadTypeStatus) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadTypeStatus) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadTypeStatus) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadTypeStatus) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetClusterId

`func (o *WorkloadTypeStatus) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadTypeStatus) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadTypeStatus) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *WorkloadTypeStatus) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkloadTypeStatus) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkloadTypeStatus) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkloadTypeStatus) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkloadTypeStatus) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


