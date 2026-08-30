# DistributedWorkload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**RequestedName** | **string** | The name as was requested for the workload. If useGivenNameAsPrefix, in the creation request, is false, name and requestedName should be identical. Otherwise, name should be composed of requestedName followed by a suffix of random characters. | 
**WorkloadId** | **string** | A unique ID of the workload. | 
**ProjectId** | **string** | The id of the project. | 
**DepartmentId** | Pointer to **string** | The id of the department. | [optional] 
**ClusterId** | **string** | The id of the cluster. | 
**CreatedBy** | **string** | The user who created the workload | 
**CreatedAt** | **time.Time** | The creation time of the workload. | 
**DesiredPhase** | [**WorkloadDesiredPhase**](WorkloadDesiredPhase.md) |  | 
**ActualPhase** | Pointer to [**Phase**](Phase.md) |  | [optional] 
**Spec** | Pointer to [**DistributedSpecSpec**](DistributedSpecSpec.md) |  | [optional] 
**MasterSpecSameAsWorker** | Pointer to **NullableBool** | used for distributed workloads to indicate that the master spec should be the same as the worker spec. in this case, masterSpec should not be specified. | [optional] 
**MasterSpec** | Pointer to [**NullableDistributedSpecMasterSpec**](DistributedSpecMasterSpec.md) |  | [optional] 

## Methods

### NewDistributedWorkload

`func NewDistributedWorkload(name string, requestedName string, workloadId string, projectId string, clusterId string, createdBy string, createdAt time.Time, desiredPhase WorkloadDesiredPhase, ) *DistributedWorkload`

NewDistributedWorkload instantiates a new DistributedWorkload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedWorkloadWithDefaults

`func NewDistributedWorkloadWithDefaults() *DistributedWorkload`

NewDistributedWorkloadWithDefaults instantiates a new DistributedWorkload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DistributedWorkload) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DistributedWorkload) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DistributedWorkload) SetName(v string)`

SetName sets Name field to given value.


### GetRequestedName

`func (o *DistributedWorkload) GetRequestedName() string`

GetRequestedName returns the RequestedName field if non-nil, zero value otherwise.

### GetRequestedNameOk

`func (o *DistributedWorkload) GetRequestedNameOk() (*string, bool)`

GetRequestedNameOk returns a tuple with the RequestedName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedName

`func (o *DistributedWorkload) SetRequestedName(v string)`

SetRequestedName sets RequestedName field to given value.


### GetWorkloadId

`func (o *DistributedWorkload) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *DistributedWorkload) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *DistributedWorkload) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetProjectId

`func (o *DistributedWorkload) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *DistributedWorkload) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *DistributedWorkload) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetDepartmentId

`func (o *DistributedWorkload) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *DistributedWorkload) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *DistributedWorkload) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *DistributedWorkload) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *DistributedWorkload) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DistributedWorkload) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DistributedWorkload) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetCreatedBy

`func (o *DistributedWorkload) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *DistributedWorkload) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *DistributedWorkload) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *DistributedWorkload) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DistributedWorkload) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DistributedWorkload) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetDesiredPhase

`func (o *DistributedWorkload) GetDesiredPhase() WorkloadDesiredPhase`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *DistributedWorkload) GetDesiredPhaseOk() (*WorkloadDesiredPhase, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *DistributedWorkload) SetDesiredPhase(v WorkloadDesiredPhase)`

SetDesiredPhase sets DesiredPhase field to given value.


### GetActualPhase

`func (o *DistributedWorkload) GetActualPhase() Phase`

GetActualPhase returns the ActualPhase field if non-nil, zero value otherwise.

### GetActualPhaseOk

`func (o *DistributedWorkload) GetActualPhaseOk() (*Phase, bool)`

GetActualPhaseOk returns a tuple with the ActualPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualPhase

`func (o *DistributedWorkload) SetActualPhase(v Phase)`

SetActualPhase sets ActualPhase field to given value.

### HasActualPhase

`func (o *DistributedWorkload) HasActualPhase() bool`

HasActualPhase returns a boolean if a field has been set.

### GetSpec

`func (o *DistributedWorkload) GetSpec() DistributedSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedWorkload) GetSpecOk() (*DistributedSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedWorkload) SetSpec(v DistributedSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedWorkload) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetMasterSpecSameAsWorker

`func (o *DistributedWorkload) GetMasterSpecSameAsWorker() bool`

GetMasterSpecSameAsWorker returns the MasterSpecSameAsWorker field if non-nil, zero value otherwise.

### GetMasterSpecSameAsWorkerOk

`func (o *DistributedWorkload) GetMasterSpecSameAsWorkerOk() (*bool, bool)`

GetMasterSpecSameAsWorkerOk returns a tuple with the MasterSpecSameAsWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpecSameAsWorker

`func (o *DistributedWorkload) SetMasterSpecSameAsWorker(v bool)`

SetMasterSpecSameAsWorker sets MasterSpecSameAsWorker field to given value.

### HasMasterSpecSameAsWorker

`func (o *DistributedWorkload) HasMasterSpecSameAsWorker() bool`

HasMasterSpecSameAsWorker returns a boolean if a field has been set.

### SetMasterSpecSameAsWorkerNil

`func (o *DistributedWorkload) SetMasterSpecSameAsWorkerNil(b bool)`

 SetMasterSpecSameAsWorkerNil sets the value for MasterSpecSameAsWorker to be an explicit nil

### UnsetMasterSpecSameAsWorker
`func (o *DistributedWorkload) UnsetMasterSpecSameAsWorker()`

UnsetMasterSpecSameAsWorker ensures that no value is present for MasterSpecSameAsWorker, not even an explicit nil
### GetMasterSpec

`func (o *DistributedWorkload) GetMasterSpec() DistributedSpecMasterSpec`

GetMasterSpec returns the MasterSpec field if non-nil, zero value otherwise.

### GetMasterSpecOk

`func (o *DistributedWorkload) GetMasterSpecOk() (*DistributedSpecMasterSpec, bool)`

GetMasterSpecOk returns a tuple with the MasterSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMasterSpec

`func (o *DistributedWorkload) SetMasterSpec(v DistributedSpecMasterSpec)`

SetMasterSpec sets MasterSpec field to given value.

### HasMasterSpec

`func (o *DistributedWorkload) HasMasterSpec() bool`

HasMasterSpec returns a boolean if a field has been set.

### SetMasterSpecNil

`func (o *DistributedWorkload) SetMasterSpecNil(b bool)`

 SetMasterSpecNil sets the value for MasterSpec to be an explicit nil

### UnsetMasterSpec
`func (o *DistributedWorkload) UnsetMasterSpec()`

UnsetMasterSpec ensures that no value is present for MasterSpec, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


