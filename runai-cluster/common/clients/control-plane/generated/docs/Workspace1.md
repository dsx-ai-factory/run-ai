# Workspace1

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
**Spec** | Pointer to [**WorkspaceSpecSpec**](WorkspaceSpecSpec.md) |  | [optional] 

## Methods

### NewWorkspace1

`func NewWorkspace1(name string, requestedName string, workloadId string, projectId string, clusterId string, createdBy string, createdAt time.Time, desiredPhase WorkloadDesiredPhase, ) *Workspace1`

NewWorkspace1 instantiates a new Workspace1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspace1WithDefaults

`func NewWorkspace1WithDefaults() *Workspace1`

NewWorkspace1WithDefaults instantiates a new Workspace1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *Workspace1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Workspace1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Workspace1) SetName(v string)`

SetName sets Name field to given value.


### GetRequestedName

`func (o *Workspace1) GetRequestedName() string`

GetRequestedName returns the RequestedName field if non-nil, zero value otherwise.

### GetRequestedNameOk

`func (o *Workspace1) GetRequestedNameOk() (*string, bool)`

GetRequestedNameOk returns a tuple with the RequestedName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedName

`func (o *Workspace1) SetRequestedName(v string)`

SetRequestedName sets RequestedName field to given value.


### GetWorkloadId

`func (o *Workspace1) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *Workspace1) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *Workspace1) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetProjectId

`func (o *Workspace1) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Workspace1) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Workspace1) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetDepartmentId

`func (o *Workspace1) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Workspace1) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Workspace1) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Workspace1) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *Workspace1) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Workspace1) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Workspace1) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetCreatedBy

`func (o *Workspace1) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Workspace1) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Workspace1) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *Workspace1) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Workspace1) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Workspace1) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetDesiredPhase

`func (o *Workspace1) GetDesiredPhase() WorkloadDesiredPhase`

GetDesiredPhase returns the DesiredPhase field if non-nil, zero value otherwise.

### GetDesiredPhaseOk

`func (o *Workspace1) GetDesiredPhaseOk() (*WorkloadDesiredPhase, bool)`

GetDesiredPhaseOk returns a tuple with the DesiredPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDesiredPhase

`func (o *Workspace1) SetDesiredPhase(v WorkloadDesiredPhase)`

SetDesiredPhase sets DesiredPhase field to given value.


### GetActualPhase

`func (o *Workspace1) GetActualPhase() Phase`

GetActualPhase returns the ActualPhase field if non-nil, zero value otherwise.

### GetActualPhaseOk

`func (o *Workspace1) GetActualPhaseOk() (*Phase, bool)`

GetActualPhaseOk returns a tuple with the ActualPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualPhase

`func (o *Workspace1) SetActualPhase(v Phase)`

SetActualPhase sets ActualPhase field to given value.

### HasActualPhase

`func (o *Workspace1) HasActualPhase() bool`

HasActualPhase returns a boolean if a field has been set.

### GetSpec

`func (o *Workspace1) GetSpec() WorkspaceSpecSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Workspace1) GetSpecOk() (*WorkspaceSpecSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Workspace1) SetSpec(v WorkspaceSpecSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *Workspace1) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


