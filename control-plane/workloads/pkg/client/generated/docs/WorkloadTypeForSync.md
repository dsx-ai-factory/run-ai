# WorkloadTypeForSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to **string** | The current phase of the workload type in the cluster. | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Version** | Pointer to **string** | The version of the resource definition&#39;s Kubernetes CRD. | [optional] 
**Id** | **string** | The unique identifier of the workload type. | 

## Methods

### NewWorkloadTypeForSync

`func NewWorkloadTypeForSync(id string, ) *WorkloadTypeForSync`

NewWorkloadTypeForSync instantiates a new WorkloadTypeForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypeForSyncWithDefaults

`func NewWorkloadTypeForSyncWithDefaults() *WorkloadTypeForSync`

NewWorkloadTypeForSyncWithDefaults instantiates a new WorkloadTypeForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *WorkloadTypeForSync) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *WorkloadTypeForSync) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *WorkloadTypeForSync) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *WorkloadTypeForSync) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetConditions

`func (o *WorkloadTypeForSync) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *WorkloadTypeForSync) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *WorkloadTypeForSync) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *WorkloadTypeForSync) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetVersion

`func (o *WorkloadTypeForSync) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkloadTypeForSync) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkloadTypeForSync) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkloadTypeForSync) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetId

`func (o *WorkloadTypeForSync) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadTypeForSync) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadTypeForSync) SetId(v string)`

SetId sets Id field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


