# WorkloadPodResourceClaims

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PodId** | **string** | The id of the pod. | 
**PodName** | **string** |  | 
**Task** | Pointer to **NullableString** | Identifies the function of the pod within a distributed workload (for example, leader, worker, or master). | [optional] 
**ResourceClaims** | [**[]WorkloadResourceClaim**](WorkloadResourceClaim.md) | The claims the pod references, in the order they appear in &#x60;pod.spec.resourceClaims&#x60;. | 

## Methods

### NewWorkloadPodResourceClaims

`func NewWorkloadPodResourceClaims(podId string, podName string, resourceClaims []WorkloadResourceClaim, ) *WorkloadPodResourceClaims`

NewWorkloadPodResourceClaims instantiates a new WorkloadPodResourceClaims object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadPodResourceClaimsWithDefaults

`func NewWorkloadPodResourceClaimsWithDefaults() *WorkloadPodResourceClaims`

NewWorkloadPodResourceClaimsWithDefaults instantiates a new WorkloadPodResourceClaims object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPodId

`func (o *WorkloadPodResourceClaims) GetPodId() string`

GetPodId returns the PodId field if non-nil, zero value otherwise.

### GetPodIdOk

`func (o *WorkloadPodResourceClaims) GetPodIdOk() (*string, bool)`

GetPodIdOk returns a tuple with the PodId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodId

`func (o *WorkloadPodResourceClaims) SetPodId(v string)`

SetPodId sets PodId field to given value.


### GetPodName

`func (o *WorkloadPodResourceClaims) GetPodName() string`

GetPodName returns the PodName field if non-nil, zero value otherwise.

### GetPodNameOk

`func (o *WorkloadPodResourceClaims) GetPodNameOk() (*string, bool)`

GetPodNameOk returns a tuple with the PodName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodName

`func (o *WorkloadPodResourceClaims) SetPodName(v string)`

SetPodName sets PodName field to given value.


### GetTask

`func (o *WorkloadPodResourceClaims) GetTask() string`

GetTask returns the Task field if non-nil, zero value otherwise.

### GetTaskOk

`func (o *WorkloadPodResourceClaims) GetTaskOk() (*string, bool)`

GetTaskOk returns a tuple with the Task field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTask

`func (o *WorkloadPodResourceClaims) SetTask(v string)`

SetTask sets Task field to given value.

### HasTask

`func (o *WorkloadPodResourceClaims) HasTask() bool`

HasTask returns a boolean if a field has been set.

### SetTaskNil

`func (o *WorkloadPodResourceClaims) SetTaskNil(b bool)`

 SetTaskNil sets the value for Task to be an explicit nil

### UnsetTask
`func (o *WorkloadPodResourceClaims) UnsetTask()`

UnsetTask ensures that no value is present for Task, not even an explicit nil
### GetResourceClaims

`func (o *WorkloadPodResourceClaims) GetResourceClaims() []WorkloadResourceClaim`

GetResourceClaims returns the ResourceClaims field if non-nil, zero value otherwise.

### GetResourceClaimsOk

`func (o *WorkloadPodResourceClaims) GetResourceClaimsOk() (*[]WorkloadResourceClaim, bool)`

GetResourceClaimsOk returns a tuple with the ResourceClaims field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceClaims

`func (o *WorkloadPodResourceClaims) SetResourceClaims(v []WorkloadResourceClaim)`

SetResourceClaims sets ResourceClaims field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


