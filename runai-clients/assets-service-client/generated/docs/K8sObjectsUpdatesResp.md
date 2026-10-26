# K8sObjectsUpdatesResp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RequestInfo** | Pointer to [**RequestInfo**](RequestInfo.md) |  | [optional] 
**ResponseInfo** | Pointer to [**ResponseInfo**](ResponseInfo.md) |  | [optional] 
**SystemInfo** | Pointer to [**SystemInfo**](SystemInfo.md) |  | [optional] 
**Assets** | Pointer to [**[]K8sObjectsUpdate**](K8sObjectsUpdate.md) |  | [optional] 

## Methods

### NewK8sObjectsUpdatesResp

`func NewK8sObjectsUpdatesResp() *K8sObjectsUpdatesResp`

NewK8sObjectsUpdatesResp instantiates a new K8sObjectsUpdatesResp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewK8sObjectsUpdatesRespWithDefaults

`func NewK8sObjectsUpdatesRespWithDefaults() *K8sObjectsUpdatesResp`

NewK8sObjectsUpdatesRespWithDefaults instantiates a new K8sObjectsUpdatesResp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequestInfo

`func (o *K8sObjectsUpdatesResp) GetRequestInfo() RequestInfo`

GetRequestInfo returns the RequestInfo field if non-nil, zero value otherwise.

### GetRequestInfoOk

`func (o *K8sObjectsUpdatesResp) GetRequestInfoOk() (*RequestInfo, bool)`

GetRequestInfoOk returns a tuple with the RequestInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestInfo

`func (o *K8sObjectsUpdatesResp) SetRequestInfo(v RequestInfo)`

SetRequestInfo sets RequestInfo field to given value.

### HasRequestInfo

`func (o *K8sObjectsUpdatesResp) HasRequestInfo() bool`

HasRequestInfo returns a boolean if a field has been set.

### GetResponseInfo

`func (o *K8sObjectsUpdatesResp) GetResponseInfo() ResponseInfo`

GetResponseInfo returns the ResponseInfo field if non-nil, zero value otherwise.

### GetResponseInfoOk

`func (o *K8sObjectsUpdatesResp) GetResponseInfoOk() (*ResponseInfo, bool)`

GetResponseInfoOk returns a tuple with the ResponseInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseInfo

`func (o *K8sObjectsUpdatesResp) SetResponseInfo(v ResponseInfo)`

SetResponseInfo sets ResponseInfo field to given value.

### HasResponseInfo

`func (o *K8sObjectsUpdatesResp) HasResponseInfo() bool`

HasResponseInfo returns a boolean if a field has been set.

### GetSystemInfo

`func (o *K8sObjectsUpdatesResp) GetSystemInfo() SystemInfo`

GetSystemInfo returns the SystemInfo field if non-nil, zero value otherwise.

### GetSystemInfoOk

`func (o *K8sObjectsUpdatesResp) GetSystemInfoOk() (*SystemInfo, bool)`

GetSystemInfoOk returns a tuple with the SystemInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemInfo

`func (o *K8sObjectsUpdatesResp) SetSystemInfo(v SystemInfo)`

SetSystemInfo sets SystemInfo field to given value.

### HasSystemInfo

`func (o *K8sObjectsUpdatesResp) HasSystemInfo() bool`

HasSystemInfo returns a boolean if a field has been set.

### GetAssets

`func (o *K8sObjectsUpdatesResp) GetAssets() []K8sObjectsUpdate`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *K8sObjectsUpdatesResp) GetAssetsOk() (*[]K8sObjectsUpdate, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *K8sObjectsUpdatesResp) SetAssets(v []K8sObjectsUpdate)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *K8sObjectsUpdatesResp) HasAssets() bool`

HasAssets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


