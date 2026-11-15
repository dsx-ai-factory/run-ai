# ResourceClaimTemplateWorkloadPod

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PodId** | **string** | The id of the pod. | 
**PodName** | **string** |  | 
**ResourceClaims** | [**[]PodResourceClaimBinding**](PodResourceClaimBinding.md) | The pod&#39;s claims that were generated from the resource claim template. A pod usually has one, but may list several when more than one of its claim entries references the same template. | 

## Methods

### NewResourceClaimTemplateWorkloadPod

`func NewResourceClaimTemplateWorkloadPod(podId string, podName string, resourceClaims []PodResourceClaimBinding, ) *ResourceClaimTemplateWorkloadPod`

NewResourceClaimTemplateWorkloadPod instantiates a new ResourceClaimTemplateWorkloadPod object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceClaimTemplateWorkloadPodWithDefaults

`func NewResourceClaimTemplateWorkloadPodWithDefaults() *ResourceClaimTemplateWorkloadPod`

NewResourceClaimTemplateWorkloadPodWithDefaults instantiates a new ResourceClaimTemplateWorkloadPod object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPodId

`func (o *ResourceClaimTemplateWorkloadPod) GetPodId() string`

GetPodId returns the PodId field if non-nil, zero value otherwise.

### GetPodIdOk

`func (o *ResourceClaimTemplateWorkloadPod) GetPodIdOk() (*string, bool)`

GetPodIdOk returns a tuple with the PodId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodId

`func (o *ResourceClaimTemplateWorkloadPod) SetPodId(v string)`

SetPodId sets PodId field to given value.


### GetPodName

`func (o *ResourceClaimTemplateWorkloadPod) GetPodName() string`

GetPodName returns the PodName field if non-nil, zero value otherwise.

### GetPodNameOk

`func (o *ResourceClaimTemplateWorkloadPod) GetPodNameOk() (*string, bool)`

GetPodNameOk returns a tuple with the PodName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodName

`func (o *ResourceClaimTemplateWorkloadPod) SetPodName(v string)`

SetPodName sets PodName field to given value.


### GetResourceClaims

`func (o *ResourceClaimTemplateWorkloadPod) GetResourceClaims() []PodResourceClaimBinding`

GetResourceClaims returns the ResourceClaims field if non-nil, zero value otherwise.

### GetResourceClaimsOk

`func (o *ResourceClaimTemplateWorkloadPod) GetResourceClaimsOk() (*[]PodResourceClaimBinding, bool)`

GetResourceClaimsOk returns a tuple with the ResourceClaims field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceClaims

`func (o *ResourceClaimTemplateWorkloadPod) SetResourceClaims(v []PodResourceClaimBinding)`

SetResourceClaims sets ResourceClaims field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


