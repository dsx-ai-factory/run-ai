# WorkloadV2MetadataUpdateParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Priority** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Configuration** | Pointer to [**NullableWorkloadV2Configuration**](WorkloadV2Configuration.md) |  | [optional] 

## Methods

### NewWorkloadV2MetadataUpdateParams

`func NewWorkloadV2MetadataUpdateParams() *WorkloadV2MetadataUpdateParams`

NewWorkloadV2MetadataUpdateParams instantiates a new WorkloadV2MetadataUpdateParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadV2MetadataUpdateParamsWithDefaults

`func NewWorkloadV2MetadataUpdateParamsWithDefaults() *WorkloadV2MetadataUpdateParams`

NewWorkloadV2MetadataUpdateParamsWithDefaults instantiates a new WorkloadV2MetadataUpdateParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPriority

`func (o *WorkloadV2MetadataUpdateParams) GetPriority() string`

GetPriority returns the Priority field if non-nil, zero value otherwise.

### GetPriorityOk

`func (o *WorkloadV2MetadataUpdateParams) GetPriorityOk() (*string, bool)`

GetPriorityOk returns a tuple with the Priority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriority

`func (o *WorkloadV2MetadataUpdateParams) SetPriority(v string)`

SetPriority sets Priority field to given value.

### HasPriority

`func (o *WorkloadV2MetadataUpdateParams) HasPriority() bool`

HasPriority returns a boolean if a field has been set.

### SetPriorityNil

`func (o *WorkloadV2MetadataUpdateParams) SetPriorityNil(b bool)`

 SetPriorityNil sets the value for Priority to be an explicit nil

### UnsetPriority
`func (o *WorkloadV2MetadataUpdateParams) UnsetPriority()`

UnsetPriority ensures that no value is present for Priority, not even an explicit nil
### GetCategory

`func (o *WorkloadV2MetadataUpdateParams) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkloadV2MetadataUpdateParams) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkloadV2MetadataUpdateParams) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkloadV2MetadataUpdateParams) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *WorkloadV2MetadataUpdateParams) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *WorkloadV2MetadataUpdateParams) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetConfiguration

`func (o *WorkloadV2MetadataUpdateParams) GetConfiguration() WorkloadV2Configuration`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *WorkloadV2MetadataUpdateParams) GetConfigurationOk() (*WorkloadV2Configuration, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *WorkloadV2MetadataUpdateParams) SetConfiguration(v WorkloadV2Configuration)`

SetConfiguration sets Configuration field to given value.

### HasConfiguration

`func (o *WorkloadV2MetadataUpdateParams) HasConfiguration() bool`

HasConfiguration returns a boolean if a field has been set.

### SetConfigurationNil

`func (o *WorkloadV2MetadataUpdateParams) SetConfigurationNil(b bool)`

 SetConfigurationNil sets the value for Configuration to be an explicit nil

### UnsetConfiguration
`func (o *WorkloadV2MetadataUpdateParams) UnsetConfiguration()`

UnsetConfiguration ensures that no value is present for Configuration, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


