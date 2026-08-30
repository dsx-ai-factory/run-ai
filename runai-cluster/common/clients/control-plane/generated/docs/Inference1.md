# Inference1

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
**Spec** | Pointer to [**InferenceSpecSpec**](InferenceSpecSpec.md) |  | [optional] 

## Methods

### NewInference1

`func NewInference1(name string, requestedName string, workloadId string, projectId string, clusterId string, createdBy string, createdAt time.Time, desiredPhase WorkloadDesiredPhase, ) *Inference1`

NewInference1 instantiates a new Inference1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInference1WithDefaults

`func NewInference1WithDefaults() *Inference1`

NewInference1WithDefaults instantiates a new Inference1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Inference1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Inference1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Inference1) SetName(v string)`

SetName sets Name field to given value.


### GetRequestedName

`func (o *Inference1) GetRequestedName() string`

GetRequestedName returns the RequestedName field if non-nil, zero value otherwise.

### GetRequestedNameOk

`func (o *Inference1) GetRequestedNameOk() (*string, bool)`

GetRequestedNameOk returns a tuple with the RequestedName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedName

`func (o *Inference1) SetRequestedName(v string)`

SetRequestedName sets RequestedName field to given value.


### GetWorkloadId

`func (o *Inference1) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *Inference1) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *Inference1) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetProjectId

`func (o *Inference1) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Inference1) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Inference1) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetDepartmentId

`func (o *Inference1) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Inference1) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Inference1) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Inference1) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *Inference1) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Inference1) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Inference1) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetCreatedBy

`func (o *Inference1) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Inference1) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Inference1) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *Inference1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Inference1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Inference1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetDesiredPhase

`func (o *Inference1) GetDesiredPhase() WorkloadDesiredPhase`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *Inference1) GetDesiredPhaseOk() (*WorkloadDesiredPhase, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *Inference1) SetDesiredPhase(v WorkloadDesiredPhase)`

SetDesiredPhase sets DesiredPhase field to given value.


### GetActualPhase

`func (o *Inference1) GetActualPhase() Phase`

GetActualPhase returns the ActualPhase field if non-nil, zero value otherwise.

### GetActualPhaseOk

`func (o *Inference1) GetActualPhaseOk() (*Phase, bool)`

GetActualPhaseOk returns a tuple with the ActualPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualPhase

`func (o *Inference1) SetActualPhase(v Phase)`

SetActualPhase sets ActualPhase field to given value.

### HasActualPhase

`func (o *Inference1) HasActualPhase() bool`

HasActualPhase returns a boolean if a field has been set.

### GetSpec

`func (o *Inference1) GetSpec() InferenceSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Inference1) GetSpecOk() (*InferenceSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Inference1) SetSpec(v InferenceSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *Inference1) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


