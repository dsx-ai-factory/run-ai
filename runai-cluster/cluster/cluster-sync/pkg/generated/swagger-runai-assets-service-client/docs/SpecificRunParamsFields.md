# SpecificRunParamsFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodeType** | Pointer to **NullableString** | Nodes (machines) or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes as explained in the Group Nodes guide at https://docs.run.ai/admin/researcher-setup/limit-to-node-group. This flag can be used in conjunction with Project-based affinity. In this case, the flag is used to refine the list of allowable node groups set in the Project. For more information consult the Projects guide at https://docs.run.ai/admin/admin-ui-setup/project-setup. | [optional] 
**NodePools** | Pointer to **[]string** | A prioritize list of node pools for the scheduler to run the workspace on. The scheduler will always try to use the first node pool before moving to the next one when the fist is not available. | [optional] 
**AllowOverQuota** | Pointer to **NullableBool** | Whether to allow the workspace to exceed the quota of the project. | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workspace. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workspace. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (Completed or Failed) will be automatically deleted. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed. | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates whether the job should be terminated, by the system, after it has been preempted. | [optional] 

## Methods

### NewSpecificRunParamsFields

`func NewSpecificRunParamsFields() *SpecificRunParamsFields`

NewSpecificRunParamsFields instantiates a new SpecificRunParamsFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSpecificRunParamsFieldsWithDefaults

`func NewSpecificRunParamsFieldsWithDefaults() *SpecificRunParamsFields`

NewSpecificRunParamsFieldsWithDefaults instantiates a new SpecificRunParamsFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodeType

`func (o *SpecificRunParamsFields) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *SpecificRunParamsFields) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *SpecificRunParamsFields) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *SpecificRunParamsFields) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *SpecificRunParamsFields) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *SpecificRunParamsFields) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *SpecificRunParamsFields) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *SpecificRunParamsFields) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *SpecificRunParamsFields) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *SpecificRunParamsFields) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *SpecificRunParamsFields) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *SpecificRunParamsFields) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetAllowOverQuota

`func (o *SpecificRunParamsFields) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *SpecificRunParamsFields) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *SpecificRunParamsFields) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *SpecificRunParamsFields) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### SetAllowOverQuotaNil

`func (o *SpecificRunParamsFields) SetAllowOverQuotaNil(b bool)`

 SetAllowOverQuotaNil sets the value for AllowOverQuota to be an explicit nil

### UnsetAllowOverQuota
`func (o *SpecificRunParamsFields) UnsetAllowOverQuota()`

UnsetAllowOverQuota ensures that no value is present for AllowOverQuota, not even an explicit nil
### GetAnnotations

`func (o *SpecificRunParamsFields) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *SpecificRunParamsFields) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *SpecificRunParamsFields) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *SpecificRunParamsFields) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetLabels

`func (o *SpecificRunParamsFields) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *SpecificRunParamsFields) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *SpecificRunParamsFields) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *SpecificRunParamsFields) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsFields) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *SpecificRunParamsFields) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsFields) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *SpecificRunParamsFields) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *SpecificRunParamsFields) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *SpecificRunParamsFields) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *SpecificRunParamsFields) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *SpecificRunParamsFields) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *SpecificRunParamsFields) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *SpecificRunParamsFields) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *SpecificRunParamsFields) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *SpecificRunParamsFields) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *SpecificRunParamsFields) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *SpecificRunParamsFields) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *SpecificRunParamsFields) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *SpecificRunParamsFields) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *SpecificRunParamsFields) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *SpecificRunParamsFields) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


