# AIApplicationSyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RequestInfo** | Pointer to [**RequestInfo**](RequestInfo.md) |  | [optional] 
**ResponseInfo** | Pointer to [**ResponseInfo**](ResponseInfo.md) |  | [optional] 
**SystemInfo** | Pointer to [**SystemInfo**](SystemInfo.md) |  | [optional] 
**Applications** | Pointer to [**[]AIApplicationSyncInfo**](AIApplicationSyncInfo.md) |  | [optional] 

## Methods

### NewAIApplicationSyncResponse

`func NewAIApplicationSyncResponse() *AIApplicationSyncResponse`

NewAIApplicationSyncResponse instantiates a new AIApplicationSyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationSyncResponseWithDefaults

`func NewAIApplicationSyncResponseWithDefaults() *AIApplicationSyncResponse`

NewAIApplicationSyncResponseWithDefaults instantiates a new AIApplicationSyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequestInfo

`func (o *AIApplicationSyncResponse) GetRequestInfo() RequestInfo`

GetRequestInfo returns the RequestInfo field if non-nil, zero value otherwise.

### GetRequestInfoOk

`func (o *AIApplicationSyncResponse) GetRequestInfoOk() (*RequestInfo, bool)`

GetRequestInfoOk returns a tuple with the RequestInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestInfo

`func (o *AIApplicationSyncResponse) SetRequestInfo(v RequestInfo)`

SetRequestInfo sets RequestInfo field to given value.

### HasRequestInfo

`func (o *AIApplicationSyncResponse) HasRequestInfo() bool`

HasRequestInfo returns a boolean if a field has been set.

### GetResponseInfo

`func (o *AIApplicationSyncResponse) GetResponseInfo() ResponseInfo`

GetResponseInfo returns the ResponseInfo field if non-nil, zero value otherwise.

### GetResponseInfoOk

`func (o *AIApplicationSyncResponse) GetResponseInfoOk() (*ResponseInfo, bool)`

GetResponseInfoOk returns a tuple with the ResponseInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseInfo

`func (o *AIApplicationSyncResponse) SetResponseInfo(v ResponseInfo)`

SetResponseInfo sets ResponseInfo field to given value.

### HasResponseInfo

`func (o *AIApplicationSyncResponse) HasResponseInfo() bool`

HasResponseInfo returns a boolean if a field has been set.

### GetSystemInfo

`func (o *AIApplicationSyncResponse) GetSystemInfo() SystemInfo`

GetSystemInfo returns the SystemInfo field if non-nil, zero value otherwise.

### GetSystemInfoOk

`func (o *AIApplicationSyncResponse) GetSystemInfoOk() (*SystemInfo, bool)`

GetSystemInfoOk returns a tuple with the SystemInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemInfo

`func (o *AIApplicationSyncResponse) SetSystemInfo(v SystemInfo)`

SetSystemInfo sets SystemInfo field to given value.

### HasSystemInfo

`func (o *AIApplicationSyncResponse) HasSystemInfo() bool`

HasSystemInfo returns a boolean if a field has been set.

### GetApplications

`func (o *AIApplicationSyncResponse) GetApplications() []AIApplicationSyncInfo`

GetApplications returns the Applications field if non-nil, zero value otherwise.

### GetApplicationsOk

`func (o *AIApplicationSyncResponse) GetApplicationsOk() (*[]AIApplicationSyncInfo, bool)`

GetApplicationsOk returns a tuple with the Applications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApplications

`func (o *AIApplicationSyncResponse) SetApplications(v []AIApplicationSyncInfo)`

SetApplications sets Applications field to given value.

### HasApplications

`func (o *AIApplicationSyncResponse) HasApplications() bool`

HasApplications returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


