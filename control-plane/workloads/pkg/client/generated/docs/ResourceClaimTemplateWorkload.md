# ResourceClaimTemplateWorkload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkloadId** | **string** | A unique ID of the workload. | 
**Name** | **string** | The name of the workload. | 
**Type** | **string** | The type of the workload. Any workload type the cluster reports, not only the NVIDIA Run:ai native ones. | 
**ProjectId** | **string** | The id of the project. | 
**Pods** | [**[]ResourceClaimTemplateWorkloadPod**](ResourceClaimTemplateWorkloadPod.md) | The pods of the workload that reference the resource claim template. | 

## Methods

### NewResourceClaimTemplateWorkload

`func NewResourceClaimTemplateWorkload(workloadId string, name string, type_ string, projectId string, pods []ResourceClaimTemplateWorkloadPod, ) *ResourceClaimTemplateWorkload`

NewResourceClaimTemplateWorkload instantiates a new ResourceClaimTemplateWorkload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceClaimTemplateWorkloadWithDefaults

`func NewResourceClaimTemplateWorkloadWithDefaults() *ResourceClaimTemplateWorkload`

NewResourceClaimTemplateWorkloadWithDefaults instantiates a new ResourceClaimTemplateWorkload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkloadId

`func (o *ResourceClaimTemplateWorkload) GetWorkloadId() string`

GetWorkloadId returns the WorkloadId field if non-nil, zero value otherwise.

### GetWorkloadIdOk

`func (o *ResourceClaimTemplateWorkload) GetWorkloadIdOk() (*string, bool)`

GetWorkloadIdOk returns a tuple with the WorkloadId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadId

`func (o *ResourceClaimTemplateWorkload) SetWorkloadId(v string)`

SetWorkloadId sets WorkloadId field to given value.


### GetName

`func (o *ResourceClaimTemplateWorkload) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ResourceClaimTemplateWorkload) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ResourceClaimTemplateWorkload) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *ResourceClaimTemplateWorkload) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ResourceClaimTemplateWorkload) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ResourceClaimTemplateWorkload) SetType(v string)`

SetType sets Type field to given value.


### GetProjectId

`func (o *ResourceClaimTemplateWorkload) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *ResourceClaimTemplateWorkload) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *ResourceClaimTemplateWorkload) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetPods

`func (o *ResourceClaimTemplateWorkload) GetPods() []ResourceClaimTemplateWorkloadPod`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *ResourceClaimTemplateWorkload) GetPodsOk() (*[]ResourceClaimTemplateWorkloadPod, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *ResourceClaimTemplateWorkload) SetPods(v []ResourceClaimTemplateWorkloadPod)`

SetPods sets Pods field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


