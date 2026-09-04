# SchedulingConditionQueueDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OrgType** | [**OrgType**](OrgType.md) |  | 
**ExceedingResources** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 
**QueueRequestedResources** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 
**QueueDeservedResources** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 
**QueueAllocatedResources** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 
**QueueAllocatedNonPreemptibleResources** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 
**QueueResourceLimits** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 
**PodGroupRequestedResources** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 
**PodGroupRequestedNonPreemptibleResources** | Pointer to [**QueueResources**](QueueResources.md) |  | [optional] 

## Methods

### NewSchedulingConditionQueueDetails

`func NewSchedulingConditionQueueDetails(orgType OrgType, ) *SchedulingConditionQueueDetails`

NewSchedulingConditionQueueDetails instantiates a new SchedulingConditionQueueDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSchedulingConditionQueueDetailsWithDefaults

`func NewSchedulingConditionQueueDetailsWithDefaults() *SchedulingConditionQueueDetails`

NewSchedulingConditionQueueDetailsWithDefaults instantiates a new SchedulingConditionQueueDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOrgType

`func (o *SchedulingConditionQueueDetails) GetOrgType() OrgType`

GetOrgType returns the OrgType field if non-nil, zero value otherwise.

### GetOrgTypeOk

`func (o *SchedulingConditionQueueDetails) GetOrgTypeOk() (*OrgType, bool)`

GetOrgTypeOk returns a tuple with the OrgType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgType

`func (o *SchedulingConditionQueueDetails) SetOrgType(v OrgType)`

SetOrgType sets OrgType field to given value.


### GetExceedingResources

`func (o *SchedulingConditionQueueDetails) GetExceedingResources() QueueResources`

GetExceedingResources returns the ExceedingResources field if non-nil, zero value otherwise.

### GetExceedingResourcesOk

`func (o *SchedulingConditionQueueDetails) GetExceedingResourcesOk() (*QueueResources, bool)`

GetExceedingResourcesOk returns a tuple with the ExceedingResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExceedingResources

`func (o *SchedulingConditionQueueDetails) SetExceedingResources(v QueueResources)`

SetExceedingResources sets ExceedingResources field to given value.

### HasExceedingResources

`func (o *SchedulingConditionQueueDetails) HasExceedingResources() bool`

HasExceedingResources returns a boolean if a field has been set.

### GetQueueRequestedResources

`func (o *SchedulingConditionQueueDetails) GetQueueRequestedResources() QueueResources`

GetQueueRequestedResources returns the QueueRequestedResources field if non-nil, zero value otherwise.

### GetQueueRequestedResourcesOk

`func (o *SchedulingConditionQueueDetails) GetQueueRequestedResourcesOk() (*QueueResources, bool)`

GetQueueRequestedResourcesOk returns a tuple with the QueueRequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueueRequestedResources

`func (o *SchedulingConditionQueueDetails) SetQueueRequestedResources(v QueueResources)`

SetQueueRequestedResources sets QueueRequestedResources field to given value.

### HasQueueRequestedResources

`func (o *SchedulingConditionQueueDetails) HasQueueRequestedResources() bool`

HasQueueRequestedResources returns a boolean if a field has been set.

### GetQueueDeservedResources

`func (o *SchedulingConditionQueueDetails) GetQueueDeservedResources() QueueResources`

GetQueueDeservedResources returns the QueueDeservedResources field if non-nil, zero value otherwise.

### GetQueueDeservedResourcesOk

`func (o *SchedulingConditionQueueDetails) GetQueueDeservedResourcesOk() (*QueueResources, bool)`

GetQueueDeservedResourcesOk returns a tuple with the QueueDeservedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueueDeservedResources

`func (o *SchedulingConditionQueueDetails) SetQueueDeservedResources(v QueueResources)`

SetQueueDeservedResources sets QueueDeservedResources field to given value.

### HasQueueDeservedResources

`func (o *SchedulingConditionQueueDetails) HasQueueDeservedResources() bool`

HasQueueDeservedResources returns a boolean if a field has been set.

### GetQueueAllocatedResources

`func (o *SchedulingConditionQueueDetails) GetQueueAllocatedResources() QueueResources`

GetQueueAllocatedResources returns the QueueAllocatedResources field if non-nil, zero value otherwise.

### GetQueueAllocatedResourcesOk

`func (o *SchedulingConditionQueueDetails) GetQueueAllocatedResourcesOk() (*QueueResources, bool)`

GetQueueAllocatedResourcesOk returns a tuple with the QueueAllocatedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueueAllocatedResources

`func (o *SchedulingConditionQueueDetails) SetQueueAllocatedResources(v QueueResources)`

SetQueueAllocatedResources sets QueueAllocatedResources field to given value.

### HasQueueAllocatedResources

`func (o *SchedulingConditionQueueDetails) HasQueueAllocatedResources() bool`

HasQueueAllocatedResources returns a boolean if a field has been set.

### GetQueueAllocatedNonPreemptibleResources

`func (o *SchedulingConditionQueueDetails) GetQueueAllocatedNonPreemptibleResources() QueueResources`

GetQueueAllocatedNonPreemptibleResources returns the QueueAllocatedNonPreemptibleResources field if non-nil, zero value otherwise.

### GetQueueAllocatedNonPreemptibleResourcesOk

`func (o *SchedulingConditionQueueDetails) GetQueueAllocatedNonPreemptibleResourcesOk() (*QueueResources, bool)`

GetQueueAllocatedNonPreemptibleResourcesOk returns a tuple with the QueueAllocatedNonPreemptibleResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueueAllocatedNonPreemptibleResources

`func (o *SchedulingConditionQueueDetails) SetQueueAllocatedNonPreemptibleResources(v QueueResources)`

SetQueueAllocatedNonPreemptibleResources sets QueueAllocatedNonPreemptibleResources field to given value.

### HasQueueAllocatedNonPreemptibleResources

`func (o *SchedulingConditionQueueDetails) HasQueueAllocatedNonPreemptibleResources() bool`

HasQueueAllocatedNonPreemptibleResources returns a boolean if a field has been set.

### GetQueueResourceLimits

`func (o *SchedulingConditionQueueDetails) GetQueueResourceLimits() QueueResources`

GetQueueResourceLimits returns the QueueResourceLimits field if non-nil, zero value otherwise.

### GetQueueResourceLimitsOk

`func (o *SchedulingConditionQueueDetails) GetQueueResourceLimitsOk() (*QueueResources, bool)`

GetQueueResourceLimitsOk returns a tuple with the QueueResourceLimits field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQueueResourceLimits

`func (o *SchedulingConditionQueueDetails) SetQueueResourceLimits(v QueueResources)`

SetQueueResourceLimits sets QueueResourceLimits field to given value.

### HasQueueResourceLimits

`func (o *SchedulingConditionQueueDetails) HasQueueResourceLimits() bool`

HasQueueResourceLimits returns a boolean if a field has been set.

### GetPodGroupRequestedResources

`func (o *SchedulingConditionQueueDetails) GetPodGroupRequestedResources() QueueResources`

GetPodGroupRequestedResources returns the PodGroupRequestedResources field if non-nil, zero value otherwise.

### GetPodGroupRequestedResourcesOk

`func (o *SchedulingConditionQueueDetails) GetPodGroupRequestedResourcesOk() (*QueueResources, bool)`

GetPodGroupRequestedResourcesOk returns a tuple with the PodGroupRequestedResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroupRequestedResources

`func (o *SchedulingConditionQueueDetails) SetPodGroupRequestedResources(v QueueResources)`

SetPodGroupRequestedResources sets PodGroupRequestedResources field to given value.

### HasPodGroupRequestedResources

`func (o *SchedulingConditionQueueDetails) HasPodGroupRequestedResources() bool`

HasPodGroupRequestedResources returns a boolean if a field has been set.

### GetPodGroupRequestedNonPreemptibleResources

`func (o *SchedulingConditionQueueDetails) GetPodGroupRequestedNonPreemptibleResources() QueueResources`

GetPodGroupRequestedNonPreemptibleResources returns the PodGroupRequestedNonPreemptibleResources field if non-nil, zero value otherwise.

### GetPodGroupRequestedNonPreemptibleResourcesOk

`func (o *SchedulingConditionQueueDetails) GetPodGroupRequestedNonPreemptibleResourcesOk() (*QueueResources, bool)`

GetPodGroupRequestedNonPreemptibleResourcesOk returns a tuple with the PodGroupRequestedNonPreemptibleResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodGroupRequestedNonPreemptibleResources

`func (o *SchedulingConditionQueueDetails) SetPodGroupRequestedNonPreemptibleResources(v QueueResources)`

SetPodGroupRequestedNonPreemptibleResources sets PodGroupRequestedNonPreemptibleResources field to given value.

### HasPodGroupRequestedNonPreemptibleResources

`func (o *SchedulingConditionQueueDetails) HasPodGroupRequestedNonPreemptibleResources() bool`

HasPodGroupRequestedNonPreemptibleResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


