# CredentialsUpdatesResp

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RequestInfo** | Pointer to [**RequestInfo**](RequestInfo.md) |  | [optional] 
**ResponseInfo** | Pointer to [**ResponseInfo**](ResponseInfo.md) |  | [optional] 
**SystemInfo** | Pointer to [**SystemInfo**](SystemInfo.md) |  | [optional] 
**Assets** | Pointer to [**[]CredentialsUpdate**](CredentialsUpdate.md) |  | [optional] 

## Methods

### NewCredentialsUpdatesResp

`func NewCredentialsUpdatesResp() *CredentialsUpdatesResp`

NewCredentialsUpdatesResp instantiates a new CredentialsUpdatesResp object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCredentialsUpdatesRespWithDefaults

`func NewCredentialsUpdatesRespWithDefaults() *CredentialsUpdatesResp`

NewCredentialsUpdatesRespWithDefaults instantiates a new CredentialsUpdatesResp object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequestInfo

`func (o *CredentialsUpdatesResp) GetRequestInfo() RequestInfo`

GetRequestInfo returns the RequestInfo field if non-nil, zero value otherwise.

### GetRequestInfoOk

`func (o *CredentialsUpdatesResp) GetRequestInfoOk() (*RequestInfo, bool)`

GetRequestInfoOk returns a tuple with the RequestInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestInfo

`func (o *CredentialsUpdatesResp) SetRequestInfo(v RequestInfo)`

SetRequestInfo sets RequestInfo field to given value.

### HasRequestInfo

`func (o *CredentialsUpdatesResp) HasRequestInfo() bool`

HasRequestInfo returns a boolean if a field has been set.

### GetResponseInfo

`func (o *CredentialsUpdatesResp) GetResponseInfo() ResponseInfo`

GetResponseInfo returns the ResponseInfo field if non-nil, zero value otherwise.

### GetResponseInfoOk

`func (o *CredentialsUpdatesResp) GetResponseInfoOk() (*ResponseInfo, bool)`

GetResponseInfoOk returns a tuple with the ResponseInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseInfo

`func (o *CredentialsUpdatesResp) SetResponseInfo(v ResponseInfo)`

SetResponseInfo sets ResponseInfo field to given value.

### HasResponseInfo

`func (o *CredentialsUpdatesResp) HasResponseInfo() bool`

HasResponseInfo returns a boolean if a field has been set.

### GetSystemInfo

`func (o *CredentialsUpdatesResp) GetSystemInfo() SystemInfo`

GetSystemInfo returns the SystemInfo field if non-nil, zero value otherwise.

### GetSystemInfoOk

`func (o *CredentialsUpdatesResp) GetSystemInfoOk() (*SystemInfo, bool)`

GetSystemInfoOk returns a tuple with the SystemInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemInfo

`func (o *CredentialsUpdatesResp) SetSystemInfo(v SystemInfo)`

SetSystemInfo sets SystemInfo field to given value.

### HasSystemInfo

`func (o *CredentialsUpdatesResp) HasSystemInfo() bool`

HasSystemInfo returns a boolean if a field has been set.

### GetAssets

`func (o *CredentialsUpdatesResp) GetAssets() []CredentialsUpdate`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *CredentialsUpdatesResp) GetAssetsOk() (*[]CredentialsUpdate, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *CredentialsUpdatesResp) SetAssets(v []CredentialsUpdate)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *CredentialsUpdatesResp) HasAssets() bool`

HasAssets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


