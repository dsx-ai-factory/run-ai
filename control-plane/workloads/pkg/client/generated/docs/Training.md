# Training

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
**DeletedAt** | Pointer to **NullableTime** | The deletion time of the workload. | [optional] 
**DesiredPhase** | [**WorkloadDesiredPhase**](WorkloadDesiredPhase.md) |  | 
**ActualPhase** | Pointer to [**Phase**](Phase.md) |  | [optional] 
**Spec** | Pointer to [**TrainingSpecSpec**](TrainingSpecSpec.md) |  | [optional] 

## Methods

### NewTraining

`func NewTraining(name string, requestedName string, workloadId string, projectId string, clusterId string, createdBy string, createdAt time.Time, desiredPhase WorkloadDesiredPhase, ) *Training`

NewTraining instantiates a new Training object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingWithDefaults

`func NewTrainingWithDefaults() *Training`

NewTrainingWithDefaults instantiates a new Training object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Training) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Training) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Training) SetName(v string)`

SetName sets Name field to given value.


### GetRequestedName

`func (o *Training) GetRequestedName() string`

GetRequestedName returns the RequestedName field if non-nil, zero value otherwise.

### GetRequestedNameOk

`func (o *Training) GetRequestedNameOk() (*string, bool)`

GetRequestedNameOk returns a tuple with the RequestedName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedName

`func (o *Training) SetRequestedName(v string)`

SetRequestedName sets RequestedName field to given value.


### GetWorkloadId

`func (o *Training) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *Training) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *Training) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetProjectId

`func (o *Training) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Training) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Training) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetDepartmentId

`func (o *Training) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Training) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Training) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Training) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *Training) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Training) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Training) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetCreatedBy

`func (o *Training) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Training) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Training) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *Training) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Training) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Training) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetDeletedAt

`func (o *Training) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *Training) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *Training) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *Training) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *Training) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *Training) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetDesiredPhase

`func (o *Training) GetDesiredPhase() WorkloadDesiredPhase`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *Training) GetDesiredPhaseOk() (*WorkloadDesiredPhase, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *Training) SetDesiredPhase(v WorkloadDesiredPhase)`

SetDesiredPhase sets DesiredPhase field to given value.


### GetActualPhase

`func (o *Training) GetActualPhase() Phase`

GetActualPhase returns the ActualPhase field if non-nil, zero value otherwise.

### GetActualPhaseOk

`func (o *Training) GetActualPhaseOk() (*Phase, bool)`

GetActualPhaseOk returns a tuple with the ActualPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualPhase

`func (o *Training) SetActualPhase(v Phase)`

SetActualPhase sets ActualPhase field to given value.

### HasActualPhase

`func (o *Training) HasActualPhase() bool`

HasActualPhase returns a boolean if a field has been set.

### GetSpec

`func (o *Training) GetSpec() TrainingSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Training) GetSpecOk() (*TrainingSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Training) SetSpec(v TrainingSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *Training) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


