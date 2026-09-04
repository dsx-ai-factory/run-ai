# WorkloadV2MetadataAutoFill

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | A unique ID of the workload. | 
**Gvk** | [**GVK**](GVK.md) |  | 
**ProjectName** | **string** | The name of the project | 
**ClusterId** | **string** | The id of the cluster. | 
**TenantId** | **int32** | The id of the tenant. | 
**DepartmentId** | **string** | The id of the department. | 
**DepartmentName** | **string** | The name of the department | 
**CreatedAt** | **time.Time** | The timestamp for when the workload was created. | 
**CreatedBy** | **string** | Identifier of the user who created the workload. | 
**UpdatedAt** | **time.Time** | The timestamp for the last time the workload was updated. | 
**UpdatedBy** | **string** | Identifier of the user who last updated the workload. | 
**DeletedAt** | Pointer to **NullableTime** | The timestamp indicating when the workload was deleted. | [optional] 
**DeletedBy** | Pointer to **NullableString** | Identifier of the user who deleted the workload. | [optional] 

## Methods

### NewWorkloadV2MetadataAutoFill

`func NewWorkloadV2MetadataAutoFill(id string, gvk GVK, projectName string, clusterId string, tenantId int32, departmentId string, departmentName string, createdAt time.Time, createdBy string, updatedAt time.Time, updatedBy string, ) *WorkloadV2MetadataAutoFill`

NewWorkloadV2MetadataAutoFill instantiates a new WorkloadV2MetadataAutoFill object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2MetadataAutoFillWithDefaults

`func NewWorkloadV2MetadataAutoFillWithDefaults() *WorkloadV2MetadataAutoFill`

NewWorkloadV2MetadataAutoFillWithDefaults instantiates a new WorkloadV2MetadataAutoFill object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkloadV2MetadataAutoFill) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadV2MetadataAutoFill) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadV2MetadataAutoFill) SetId(v string)`

SetId sets Id field to given value.


### GetGvk

`func (o *WorkloadV2MetadataAutoFill) GetGvk() GVK`

GetGvk returns the Gvk field if non-nil, zero value otherwise.

### GetGvkOk

`func (o *WorkloadV2MetadataAutoFill) GetGvkOk() (*GVK, bool)`

GetGvkOk returns a tuple with the Gvk field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGvk

`func (o *WorkloadV2MetadataAutoFill) SetGvk(v GVK)`

SetGvk sets Gvk field to given value.


### GetProjectName

`func (o *WorkloadV2MetadataAutoFill) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *WorkloadV2MetadataAutoFill) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *WorkloadV2MetadataAutoFill) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetClusterId

`func (o *WorkloadV2MetadataAutoFill) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadV2MetadataAutoFill) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadV2MetadataAutoFill) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetTenantId

`func (o *WorkloadV2MetadataAutoFill) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *WorkloadV2MetadataAutoFill) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *WorkloadV2MetadataAutoFill) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetDepartmentId

`func (o *WorkloadV2MetadataAutoFill) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *WorkloadV2MetadataAutoFill) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *WorkloadV2MetadataAutoFill) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.


### GetDepartmentName

`func (o *WorkloadV2MetadataAutoFill) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *WorkloadV2MetadataAutoFill) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *WorkloadV2MetadataAutoFill) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.


### GetCreatedAt

`func (o *WorkloadV2MetadataAutoFill) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadV2MetadataAutoFill) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadV2MetadataAutoFill) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCreatedBy

`func (o *WorkloadV2MetadataAutoFill) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *WorkloadV2MetadataAutoFill) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *WorkloadV2MetadataAutoFill) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetUpdatedAt

`func (o *WorkloadV2MetadataAutoFill) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkloadV2MetadataAutoFill) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkloadV2MetadataAutoFill) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUpdatedBy

`func (o *WorkloadV2MetadataAutoFill) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *WorkloadV2MetadataAutoFill) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *WorkloadV2MetadataAutoFill) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetDeletedAt

`func (o *WorkloadV2MetadataAutoFill) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *WorkloadV2MetadataAutoFill) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *WorkloadV2MetadataAutoFill) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *WorkloadV2MetadataAutoFill) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *WorkloadV2MetadataAutoFill) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *WorkloadV2MetadataAutoFill) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetDeletedBy

`func (o *WorkloadV2MetadataAutoFill) GetDeletedBy() string`

GetDeletedBy returns the DeletedBy field if non-nil, zero value otherwise.

### GetDeletedByOk

`func (o *WorkloadV2MetadataAutoFill) GetDeletedByOk() (*string, bool)`

GetDeletedByOk returns a tuple with the DeletedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedBy

`func (o *WorkloadV2MetadataAutoFill) SetDeletedBy(v string)`

SetDeletedBy sets DeletedBy field to given value.

### HasDeletedBy

`func (o *WorkloadV2MetadataAutoFill) HasDeletedBy() bool`

HasDeletedBy returns a boolean if a field has been set.

### SetDeletedByNil

`func (o *WorkloadV2MetadataAutoFill) SetDeletedByNil(b bool)`

 SetDeletedByNil sets the value for DeletedBy to be an explicit nil

### UnsetDeletedBy
`func (o *WorkloadV2MetadataAutoFill) UnsetDeletedBy()`

UnsetDeletedBy ensures that no value is present for DeletedBy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


