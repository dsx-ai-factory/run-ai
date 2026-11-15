# WorkloadResourceClaimsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Next** | Pointer to **int32** | Offset for the next page of results. Present only when there are more pods beyond the current page (i.e., the number of returned pods equals the limit).  | [optional] 
**Pods** | [**[]WorkloadPodResourceClaims**](WorkloadPodResourceClaims.md) | The pods of the workload that reference at least one claim. | 

## Methods

### NewWorkloadResourceClaimsResponse

`func NewWorkloadResourceClaimsResponse(pods []WorkloadPodResourceClaims, ) *WorkloadResourceClaimsResponse`

NewWorkloadResourceClaimsResponse instantiates a new WorkloadResourceClaimsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadResourceClaimsResponseWithDefaults

`func NewWorkloadResourceClaimsResponseWithDefaults() *WorkloadResourceClaimsResponse`

NewWorkloadResourceClaimsResponseWithDefaults instantiates a new WorkloadResourceClaimsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNext

`func (o *WorkloadResourceClaimsResponse) GetNext() int32`

GetNext returns the Next field if non-nil, zero value otherwise.

### GetNextOk

`func (o *WorkloadResourceClaimsResponse) GetNextOk() (*int32, bool)`

GetNextOk returns a tuple with the Next field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNext

`func (o *WorkloadResourceClaimsResponse) SetNext(v int32)`

SetNext sets Next field to given value.

### HasNext

`func (o *WorkloadResourceClaimsResponse) HasNext() bool`

HasNext returns a boolean if a field has been set.

### GetPods

`func (o *WorkloadResourceClaimsResponse) GetPods() []WorkloadPodResourceClaims`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *WorkloadResourceClaimsResponse) GetPodsOk() (*[]WorkloadPodResourceClaims, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *WorkloadResourceClaimsResponse) SetPods(v []WorkloadPodResourceClaims)`

SetPods sets Pods field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


