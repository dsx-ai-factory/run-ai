# WorkloadV2Metadata

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**ProjectId** | **string** | The id of the project. | 
**Priority** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**Configuration** | Pointer to [**WorkloadV2Configuration**](WorkloadV2Configuration.md) |  | [optional] 
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

### NewWorkloadV2Metadata

`func NewWorkloadV2Metadata(name string, projectId string, id string, gvk GVK, projectName string, clusterId string, tenantId int32, departmentId string, departmentName string, createdAt time.Time, createdBy string, updatedAt time.Time, updatedBy string, ) *WorkloadV2Metadata`

NewWorkloadV2Metadata instantiates a new WorkloadV2Metadata object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2MetadataWithDefaults

`func NewWorkloadV2MetadataWithDefaults() *WorkloadV2Metadata`

NewWorkloadV2MetadataWithDefaults instantiates a new WorkloadV2Metadata object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadV2Metadata) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadV2Metadata) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadV2Metadata) SetName(v string)`

SetName sets Name field to given value.


### GetProjectId

`func (o *WorkloadV2Metadata) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadV2Metadata) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadV2Metadata) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetPriority

`func (o *WorkloadV2Metadata) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *WorkloadV2Metadata) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *WorkloadV2Metadata) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *WorkloadV2Metadata) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### SetPriorityNil

`func (o *WorkloadV2Metadata) SetPriorityNil(b bool)`

 SetPriorityNil sets the value for Priority to be an explicit nil

### UnsetPriority
`func (o *WorkloadV2Metadata) UnsetPriority()`

UnsetPriority ensures that no value is present for Priority, not even an explicit nil
### GetCategory

`func (o *WorkloadV2Metadata) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkloadV2Metadata) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkloadV2Metadata) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkloadV2Metadata) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *WorkloadV2Metadata) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *WorkloadV2Metadata) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetPreemptibility

`func (o *WorkloadV2Metadata) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadV2Metadata) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadV2Metadata) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadV2Metadata) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkloadV2Metadata) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkloadV2Metadata) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetConfiguration

`func (o *WorkloadV2Metadata) GetConfiguration() WorkloadV2Configuration`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *WorkloadV2Metadata) GetConfigurationOk() (*WorkloadV2Configuration, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *WorkloadV2Metadata) SetConfiguration(v WorkloadV2Configuration)`

SetConfiguration sets Configuration field to given value.

### HasConfiguration

`func (o *WorkloadV2Metadata) HasConfiguration() bool`

HasConfiguration returns a boolean if a field has been set.

### GetId

`func (o *WorkloadV2Metadata) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkloadV2Metadata) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkloadV2Metadata) SetId(v string)`

SetId sets Id field to given value.


### GetGvk

`func (o *WorkloadV2Metadata) GetGvk() GVK`

GetGvk returns the Gvk field if non-nil, zero value otherwise.

### GetGvkOk

`func (o *WorkloadV2Metadata) GetGvkOk() (*GVK, bool)`

GetGvkOk returns a tuple with the Gvk field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGvk

`func (o *WorkloadV2Metadata) SetGvk(v GVK)`

SetGvk sets Gvk field to given value.


### GetProjectName

`func (o *WorkloadV2Metadata) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *WorkloadV2Metadata) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *WorkloadV2Metadata) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetClusterId

`func (o *WorkloadV2Metadata) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadV2Metadata) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadV2Metadata) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetTenantId

`func (o *WorkloadV2Metadata) GetTenantId() int32`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *WorkloadV2Metadata) GetTenantIdOk() (*int32, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *WorkloadV2Metadata) SetTenantId(v int32)`

SetTenantId sets TenantId field to given value.


### GetDepartmentId

`func (o *WorkloadV2Metadata) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *WorkloadV2Metadata) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *WorkloadV2Metadata) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.


### GetDepartmentName

`func (o *WorkloadV2Metadata) GetDepartmentName() string`

GetDepartmentName returns the DepartmentName field if non-nil, zero value otherwise.

### GetDepartmentNameOk

`func (o *WorkloadV2Metadata) GetDepartmentNameOk() (*string, bool)`

GetDepartmentNameOk returns a tuple with the DepartmentName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentName

`func (o *WorkloadV2Metadata) SetDepartmentName(v string)`

SetDepartmentName sets DepartmentName field to given value.


### GetCreatedAt

`func (o *WorkloadV2Metadata) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkloadV2Metadata) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkloadV2Metadata) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetCreatedBy

`func (o *WorkloadV2Metadata) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *WorkloadV2Metadata) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *WorkloadV2Metadata) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetUpdatedAt

`func (o *WorkloadV2Metadata) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkloadV2Metadata) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkloadV2Metadata) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetUpdatedBy

`func (o *WorkloadV2Metadata) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *WorkloadV2Metadata) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *WorkloadV2Metadata) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetDeletedAt

`func (o *WorkloadV2Metadata) GetDeletedAt() time.Time`

GetDeletedAt returns the DeletedAt field if non-nil, zero value otherwise.

### GetDeletedAtOk

`func (o *WorkloadV2Metadata) GetDeletedAtOk() (*time.Time, bool)`

GetDeletedAtOk returns a tuple with the DeletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedAt

`func (o *WorkloadV2Metadata) SetDeletedAt(v time.Time)`

SetDeletedAt sets DeletedAt field to given value.

### HasDeletedAt

`func (o *WorkloadV2Metadata) HasDeletedAt() bool`

HasDeletedAt returns a boolean if a field has been set.

### SetDeletedAtNil

`func (o *WorkloadV2Metadata) SetDeletedAtNil(b bool)`

 SetDeletedAtNil sets the value for DeletedAt to be an explicit nil

### UnsetDeletedAt
`func (o *WorkloadV2Metadata) UnsetDeletedAt()`

UnsetDeletedAt ensures that no value is present for DeletedAt, not even an explicit nil
### GetDeletedBy

`func (o *WorkloadV2Metadata) GetDeletedBy() string`

GetDeletedBy returns the DeletedBy field if non-nil, zero value otherwise.

### GetDeletedByOk

`func (o *WorkloadV2Metadata) GetDeletedByOk() (*string, bool)`

GetDeletedByOk returns a tuple with the DeletedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletedBy

`func (o *WorkloadV2Metadata) SetDeletedBy(v string)`

SetDeletedBy sets DeletedBy field to given value.

### HasDeletedBy

`func (o *WorkloadV2Metadata) HasDeletedBy() bool`

HasDeletedBy returns a boolean if a field has been set.

### SetDeletedByNil

`func (o *WorkloadV2Metadata) SetDeletedByNil(b bool)`

 SetDeletedByNil sets the value for DeletedBy to be an explicit nil

### UnsetDeletedBy
`func (o *WorkloadV2Metadata) UnsetDeletedBy()`

UnsetDeletedBy ensures that no value is present for DeletedBy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


