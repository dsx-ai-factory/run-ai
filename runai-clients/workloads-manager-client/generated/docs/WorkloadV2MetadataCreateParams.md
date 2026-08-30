# WorkloadV2MetadataCreateParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | The name of the workload. | [optional] 
**UseGivenNameAsPrefix** | Pointer to **bool** | When true, the requested name will be treated as a prefix. The final name of the workload will be composed of the name followed by a random set of characters. | [optional] [default to false]
**ProjectId** | Pointer to **string** | The id of the project. | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**Priority** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**Configuration** | Pointer to [**WorkloadV2Configuration**](WorkloadV2Configuration.md) |  | [optional] 

## Methods

### NewWorkloadV2MetadataCreateParams

`func NewWorkloadV2MetadataCreateParams() *WorkloadV2MetadataCreateParams`

NewWorkloadV2MetadataCreateParams instantiates a new WorkloadV2MetadataCreateParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2MetadataCreateParamsWithDefaults

`func NewWorkloadV2MetadataCreateParamsWithDefaults() *WorkloadV2MetadataCreateParams`

NewWorkloadV2MetadataCreateParamsWithDefaults instantiates a new WorkloadV2MetadataCreateParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadV2MetadataCreateParams) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadV2MetadataCreateParams) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadV2MetadataCreateParams) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkloadV2MetadataCreateParams) HasName() bool`

HasName returns a boolean if a field has been set.

### GetUseGivenNameAsPrefix

`func (o *WorkloadV2MetadataCreateParams) GetUseGivenNameAsPrefix() bool`

GetUseGivenNameAsPrefix returns the UseGivenNameAsPrefix field if non-nil, zero value otherwise.

### GetUseGivenNameAsPrefixOk

`func (o *WorkloadV2MetadataCreateParams) GetUseGivenNameAsPrefixOk() (*bool, bool)`

GetUseGivenNameAsPrefixOk returns a tuple with the UseGivenNameAsPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseGivenNameAsPrefix

`func (o *WorkloadV2MetadataCreateParams) SetUseGivenNameAsPrefix(v bool)`

SetUseGivenNameAsPrefix sets UseGivenNameAsPrefix field to given value.

### HasUseGivenNameAsPrefix

`func (o *WorkloadV2MetadataCreateParams) HasUseGivenNameAsPrefix() bool`

HasUseGivenNameAsPrefix returns a boolean if a field has been set.

### GetProjectId

`func (o *WorkloadV2MetadataCreateParams) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadV2MetadataCreateParams) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadV2MetadataCreateParams) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *WorkloadV2MetadataCreateParams) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetClusterId

`func (o *WorkloadV2MetadataCreateParams) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadV2MetadataCreateParams) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadV2MetadataCreateParams) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *WorkloadV2MetadataCreateParams) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetPriority

`func (o *WorkloadV2MetadataCreateParams) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *WorkloadV2MetadataCreateParams) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *WorkloadV2MetadataCreateParams) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *WorkloadV2MetadataCreateParams) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### SetPriorityNil

`func (o *WorkloadV2MetadataCreateParams) SetPriorityNil(b bool)`

 SetPriorityNil sets the value for Priority to be an explicit nil

### UnsetPriority
`func (o *WorkloadV2MetadataCreateParams) UnsetPriority()`

UnsetPriority ensures that no value is present for Priority, not even an explicit nil
### GetCategory

`func (o *WorkloadV2MetadataCreateParams) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkloadV2MetadataCreateParams) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkloadV2MetadataCreateParams) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkloadV2MetadataCreateParams) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *WorkloadV2MetadataCreateParams) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *WorkloadV2MetadataCreateParams) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetPreemptibility

`func (o *WorkloadV2MetadataCreateParams) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkloadV2MetadataCreateParams) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkloadV2MetadataCreateParams) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkloadV2MetadataCreateParams) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkloadV2MetadataCreateParams) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkloadV2MetadataCreateParams) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetConfiguration

`func (o *WorkloadV2MetadataCreateParams) GetConfiguration() WorkloadV2Configuration`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *WorkloadV2MetadataCreateParams) GetConfigurationOk() (*WorkloadV2Configuration, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *WorkloadV2MetadataCreateParams) SetConfiguration(v WorkloadV2Configuration)`

SetConfiguration sets Configuration field to given value.

### HasConfiguration

`func (o *WorkloadV2MetadataCreateParams) HasConfiguration() bool`

HasConfiguration returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


