# InternalWorkloadsSyncRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**Workloads** | [**[]WorkloadSyncFields**](WorkloadSyncFields.md) |  | 

## Methods

### NewInternalWorkloadsSyncRequest

`func NewInternalWorkloadsSyncRequest(clusterId string, workloads []WorkloadSyncFields, ) *InternalWorkloadsSyncRequest`

NewInternalWorkloadsSyncRequest instantiates a new InternalWorkloadsSyncRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInternalWorkloadsSyncRequestWithDefaults

`func NewInternalWorkloadsSyncRequestWithDefaults() *InternalWorkloadsSyncRequest`

NewInternalWorkloadsSyncRequestWithDefaults instantiates a new InternalWorkloadsSyncRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *InternalWorkloadsSyncRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *InternalWorkloadsSyncRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *InternalWorkloadsSyncRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetWorkloads

`func (o *InternalWorkloadsSyncRequest) GetWorkloads() []WorkloadSyncFields`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *InternalWorkloadsSyncRequest) GetWorkloadsOk() (*[]WorkloadSyncFields, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *InternalWorkloadsSyncRequest) SetWorkloads(v []WorkloadSyncFields)`

SetWorkloads sets Workloads field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


