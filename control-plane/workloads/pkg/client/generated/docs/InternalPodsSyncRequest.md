# InternalPodsSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**Pods** | [**[]PodSyncFields**](PodSyncFields.md) |  | 

## Methods

### NewInternalPodsSyncRequest

`func NewInternalPodsSyncRequest(clusterId string, pods []PodSyncFields, ) *InternalPodsSyncRequest`

NewInternalPodsSyncRequest instantiates a new InternalPodsSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalPodsSyncRequestWithDefaults

`func NewInternalPodsSyncRequestWithDefaults() *InternalPodsSyncRequest`

NewInternalPodsSyncRequestWithDefaults instantiates a new InternalPodsSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *InternalPodsSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *InternalPodsSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *InternalPodsSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetPods

`func (o *InternalPodsSyncRequest) GetPods() []PodSyncFields`

GetPods returns the Pods field if non-nil, zero value otherwise.

### GetPodsOk

`func (o *InternalPodsSyncRequest) GetPodsOk() (*[]PodSyncFields, bool)`

GetPodsOk returns a tuple with the Pods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPods

`func (o *InternalPodsSyncRequest) SetPods(v []PodSyncFields)`

SetPods sets Pods field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


