# WorkloadsV2Updates

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RequestInfo** | Pointer to [**RequestInfo**](RequestInfo.md) |  | [optional] 
**ResponseInfo** | Pointer to [**ResponseInfo**](ResponseInfo.md) |  | [optional] 
**SystemInfo** | Pointer to [**SystemInfo**](SystemInfo.md) |  | [optional] 
**Workloads** | Pointer to [**[]WorkloadV2Update**](WorkloadV2Update.md) |  | [optional] 

## Methods

### NewWorkloadsV2Updates

`func NewWorkloadsV2Updates() *WorkloadsV2Updates`

NewWorkloadsV2Updates instantiates a new WorkloadsV2Updates object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadsV2UpdatesWithDefaults

`func NewWorkloadsV2UpdatesWithDefaults() *WorkloadsV2Updates`

NewWorkloadsV2UpdatesWithDefaults instantiates a new WorkloadsV2Updates object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequestInfo

`func (o *WorkloadsV2Updates) GetRequestInfo() RequestInfo`

GetRequestInfo returns the RequestInfo field if non-nil, zero value otherwise.

### GetRequestInfoOk

`func (o *WorkloadsV2Updates) GetRequestInfoOk() (*RequestInfo, bool)`

GetRequestInfoOk returns a tuple with the RequestInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestInfo

`func (o *WorkloadsV2Updates) SetRequestInfo(v RequestInfo)`

SetRequestInfo sets RequestInfo field to given value.

### HasRequestInfo

`func (o *WorkloadsV2Updates) HasRequestInfo() bool`

HasRequestInfo returns a boolean if a field has been set.

### GetResponseInfo

`func (o *WorkloadsV2Updates) GetResponseInfo() ResponseInfo`

GetResponseInfo returns the ResponseInfo field if non-nil, zero value otherwise.

### GetResponseInfoOk

`func (o *WorkloadsV2Updates) GetResponseInfoOk() (*ResponseInfo, bool)`

GetResponseInfoOk returns a tuple with the ResponseInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseInfo

`func (o *WorkloadsV2Updates) SetResponseInfo(v ResponseInfo)`

SetResponseInfo sets ResponseInfo field to given value.

### HasResponseInfo

`func (o *WorkloadsV2Updates) HasResponseInfo() bool`

HasResponseInfo returns a boolean if a field has been set.

### GetSystemInfo

`func (o *WorkloadsV2Updates) GetSystemInfo() SystemInfo`

GetSystemInfo returns the SystemInfo field if non-nil, zero value otherwise.

### GetSystemInfoOk

`func (o *WorkloadsV2Updates) GetSystemInfoOk() (*SystemInfo, bool)`

GetSystemInfoOk returns a tuple with the SystemInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemInfo

`func (o *WorkloadsV2Updates) SetSystemInfo(v SystemInfo)`

SetSystemInfo sets SystemInfo field to given value.

### HasSystemInfo

`func (o *WorkloadsV2Updates) HasSystemInfo() bool`

HasSystemInfo returns a boolean if a field has been set.

### GetWorkloads

`func (o *WorkloadsV2Updates) GetWorkloads() []WorkloadV2Update`

GetWorkloads returns the Workloads field if non-nil, zero value otherwise.

### GetWorkloadsOk

`func (o *WorkloadsV2Updates) GetWorkloadsOk() (*[]WorkloadV2Update, bool)`

GetWorkloadsOk returns a tuple with the Workloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloads

`func (o *WorkloadsV2Updates) SetWorkloads(v []WorkloadV2Update)`

SetWorkloads sets Workloads field to given value.

### HasWorkloads

`func (o *WorkloadsV2Updates) HasWorkloads() bool`

HasWorkloads returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


