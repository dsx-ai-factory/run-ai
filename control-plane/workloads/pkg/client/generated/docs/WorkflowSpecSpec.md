# WorkflowSpecSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Groups** | Pointer to [**[]WorkflowGroupsInner**](WorkflowGroupsInner.md) | Definitions of the workflow groups. | [optional] 
**Compute** | Pointer to [**[]WorkflowComputeProfile**](WorkflowComputeProfile.md) | Defines the workflow compute profiles. | [optional] 
**Placement** | Pointer to [**[]WorkflowPlacementProfile**](WorkflowPlacementProfile.md) | Defines the workflow placement profiles. | [optional] 
**Storage** | Pointer to [**[]WorkflowStorageProfile**](WorkflowStorageProfile.md) | Defines the workflow storage profiles. | [optional] 

## Methods

### NewWorkflowSpecSpec

`func NewWorkflowSpecSpec() *WorkflowSpecSpec`

NewWorkflowSpecSpec instantiates a new WorkflowSpecSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowSpecSpecWithDefaults

`func NewWorkflowSpecSpecWithDefaults() *WorkflowSpecSpec`

NewWorkflowSpecSpecWithDefaults instantiates a new WorkflowSpecSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCategory

`func (o *WorkflowSpecSpec) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *WorkflowSpecSpec) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *WorkflowSpecSpec) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *WorkflowSpecSpec) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *WorkflowSpecSpec) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *WorkflowSpecSpec) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetNodePools

`func (o *WorkflowSpecSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *WorkflowSpecSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *WorkflowSpecSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *WorkflowSpecSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *WorkflowSpecSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *WorkflowSpecSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPreemptibility

`func (o *WorkflowSpecSpec) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *WorkflowSpecSpec) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *WorkflowSpecSpec) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *WorkflowSpecSpec) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *WorkflowSpecSpec) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *WorkflowSpecSpec) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *WorkflowSpecSpec) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *WorkflowSpecSpec) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *WorkflowSpecSpec) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *WorkflowSpecSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *WorkflowSpecSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *WorkflowSpecSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetGroups

`func (o *WorkflowSpecSpec) GetGroups() []WorkflowGroupsInner`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *WorkflowSpecSpec) GetGroupsOk() (*[]WorkflowGroupsInner, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *WorkflowSpecSpec) SetGroups(v []WorkflowGroupsInner)`

SetGroups sets Groups field to given value.

### HasGroups

`func (o *WorkflowSpecSpec) HasGroups() bool`

HasGroups returns a boolean if a field has been set.

### GetCompute

`func (o *WorkflowSpecSpec) GetCompute() []WorkflowComputeProfile`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *WorkflowSpecSpec) GetComputeOk() (*[]WorkflowComputeProfile, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *WorkflowSpecSpec) SetCompute(v []WorkflowComputeProfile)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *WorkflowSpecSpec) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### GetPlacement

`func (o *WorkflowSpecSpec) GetPlacement() []WorkflowPlacementProfile`

GetPlacement returns the Placement field if non-nil, zero value otherwise.

### GetPlacementOk

`func (o *WorkflowSpecSpec) GetPlacementOk() (*[]WorkflowPlacementProfile, bool)`

GetPlacementOk returns a tuple with the Placement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacement

`func (o *WorkflowSpecSpec) SetPlacement(v []WorkflowPlacementProfile)`

SetPlacement sets Placement field to given value.

### HasPlacement

`func (o *WorkflowSpecSpec) HasPlacement() bool`

HasPlacement returns a boolean if a field has been set.

### GetStorage

`func (o *WorkflowSpecSpec) GetStorage() []WorkflowStorageProfile`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkflowSpecSpec) GetStorageOk() (*[]WorkflowStorageProfile, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkflowSpecSpec) SetStorage(v []WorkflowStorageProfile)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *WorkflowSpecSpec) HasStorage() bool`

HasStorage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


