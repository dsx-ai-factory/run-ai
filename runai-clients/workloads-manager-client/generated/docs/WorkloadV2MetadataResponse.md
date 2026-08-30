# WorkloadV2MetadataResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**ProjectId** | **string** | The id of the project. | 
**Priority** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**Configuration** | Pointer to [**WorkloadV2Configuration**](WorkloadV2Configuration.md) |  | [optional] 

## Methods

### NewWorkloadV2MetadataResponse

`func NewWorkloadV2MetadataResponse(name string, projectId string, ) *WorkloadV2MetadataResponse`

NewWorkloadV2MetadataResponse instantiates a new WorkloadV2MetadataResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2MetadataResponseWithDefaults

`func NewWorkloadV2MetadataResponseWithDefaults() *WorkloadV2MetadataResponse`

NewWorkloadV2MetadataResponseWithDefaults instantiates a new WorkloadV2MetadataResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadV2MetadataResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadV2MetadataResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadV2MetadataResponse) SetName(v string)`

SetName sets Name field to given value.


### GetProjectId

`func (o *WorkloadV2MetadataResponse) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadV2MetadataResponse) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadV2MetadataResponse) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetPriority

`func (o *WorkloadV2MetadataResponse) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *WorkloadV2MetadataResponse) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *WorkloadV2MetadataResponse) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *WorkloadV2MetadataResponse) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### SetPriorityNil

`func (o *WorkloadV2MetadataResponse) SetPriorityNil(b bool)`

 SetPriorityNil sets the value for Priority to be an explicit nil

### UnsetPriority
`func (o *WorkloadV2MetadataResponse) UnsetPriority()`

UnsetPriority ensures that no value is present for Priority, not even an explicit nil
### GetCategory

`func (o *WorkloadV2MetadataResponse) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkloadV2MetadataResponse) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkloadV2MetadataResponse) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkloadV2MetadataResponse) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *WorkloadV2MetadataResponse) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *WorkloadV2MetadataResponse) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetPreemptibility

`func (o *WorkloadV2MetadataResponse) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadV2MetadataResponse) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadV2MetadataResponse) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadV2MetadataResponse) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkloadV2MetadataResponse) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkloadV2MetadataResponse) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetConfiguration

`func (o *WorkloadV2MetadataResponse) GetConfiguration() WorkloadV2Configuration`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *WorkloadV2MetadataResponse) GetConfigurationOk() (*WorkloadV2Configuration, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *WorkloadV2MetadataResponse) SetConfiguration(v WorkloadV2Configuration)`

SetConfiguration sets Configuration field to given value.

### HasConfiguration

`func (o *WorkloadV2MetadataResponse) HasConfiguration() bool`

HasConfiguration returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


