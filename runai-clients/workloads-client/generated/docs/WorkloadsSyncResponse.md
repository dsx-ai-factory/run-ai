# WorkloadsSyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RequestInfo** | Pointer to [**RequestInfo**](RequestInfo.md) |  | [optional] 
**ResponseInfo** | Pointer to [**ResponseInfo**](ResponseInfo.md) |  | [optional] 
**SystemInfo** | Pointer to [**SystemInfo**](SystemInfo.md) |  | [optional] 
**Workloads** | Pointer to [**[]WorkloadsSyncInfo**](WorkloadsSyncInfo.md) |  | [optional] 

## Methods

### NewWorkloadsSyncResponse

`func NewWorkloadsSyncResponse() *WorkloadsSyncResponse`

NewWorkloadsSyncResponse instantiates a new WorkloadsSyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadsSyncResponseWithDefaults

`func NewWorkloadsSyncResponseWithDefaults() *WorkloadsSyncResponse`

NewWorkloadsSyncResponseWithDefaults instantiates a new WorkloadsSyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequestInfo

`func (o *WorkloadsSyncResponse) GetRequestInfo() RequestInfo`

GetRequestInfo returns the RequestInfo field if non-nil, zero value otherwise.

### GetRequestInfoOk

`func (o *WorkloadsSyncResponse) GetRequestInfoOk() (*RequestInfo, bool)`

GetRequestInfoOk returns a tuple with the RequestInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestInfo

`func (o *WorkloadsSyncResponse) SetRequestInfo(v RequestInfo)`

SetRequestInfo sets RequestInfo field to given value.

### HasRequestInfo

`func (o *WorkloadsSyncResponse) HasRequestInfo() bool`

HasRequestInfo returns a boolean if a field has been set.

### GetResponseInfo

`func (o *WorkloadsSyncResponse) GetResponseInfo() ResponseInfo`

GetResponseInfo returns the ResponseInfo field if non-nil, zero value otherwise.

### GetResponseInfoOk

`func (o *WorkloadsSyncResponse) GetResponseInfoOk() (*ResponseInfo, bool)`

GetResponseInfoOk returns a tuple with the ResponseInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseInfo

`func (o *WorkloadsSyncResponse) SetResponseInfo(v ResponseInfo)`

SetResponseInfo sets ResponseInfo field to given value.

### HasResponseInfo

`func (o *WorkloadsSyncResponse) HasResponseInfo() bool`

HasResponseInfo returns a boolean if a field has been set.

### GetSystemInfo

`func (o *WorkloadsSyncResponse) GetSystemInfo() SystemInfo`

GetSystemInfo returns the SystemInfo field if non-nil, zero value otherwise.

### GetSystemInfoOk

`func (o *WorkloadsSyncResponse) GetSystemInfoOk() (*SystemInfo, bool)`

GetSystemInfoOk returns a tuple with the SystemInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemInfo

`func (o *WorkloadsSyncResponse) SetSystemInfo(v SystemInfo)`

SetSystemInfo sets SystemInfo field to given value.

### HasSystemInfo

`func (o *WorkloadsSyncResponse) HasSystemInfo() bool`

HasSystemInfo returns a boolean if a field has been set.

### GetWorkloads

`func (o *WorkloadsSyncResponse) GetWorkloads() []WorkloadsSyncInfo`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *WorkloadsSyncResponse) GetWorkloadsOk() (*[]WorkloadsSyncInfo, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *WorkloadsSyncResponse) SetWorkloads(v []WorkloadsSyncInfo)`

SetWorkloads sets Workloads field to given value.

### HasWorkloads

`func (o *WorkloadsSyncResponse) HasWorkloads() bool`

HasWorkloads returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


