# PolicySyncResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RequestInfo** | Pointer to [**PolicySyncResponseRequestInfo**](PolicySyncResponseRequestInfo.md) |  | [optional] 
**ResponseInfo** | Pointer to [**PolicySyncResponseResponseInfo**](PolicySyncResponseResponseInfo.md) |  | [optional] 
**SystemInfo** | Pointer to [**PolicySyncResponseSystemInfo**](PolicySyncResponseSystemInfo.md) |  | [optional] 
**Policies** | Pointer to [**[]PolicySyncResponsePoliciesInner**](PolicySyncResponsePoliciesInner.md) |  | [optional] 

## Methods

### NewPolicySyncResponse

`func NewPolicySyncResponse() *PolicySyncResponse`

NewPolicySyncResponse instantiates a new PolicySyncResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicySyncResponseWithDefaults

`func NewPolicySyncResponseWithDefaults() *PolicySyncResponse`

NewPolicySyncResponseWithDefaults instantiates a new PolicySyncResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequestInfo

`func (o *PolicySyncResponse) GetRequestInfo() PolicySyncResponseRequestInfo`

GetRequestInfo returns the RequestInfo field if non-nil, zero value otherwise.

### GetRequestInfoOk

`func (o *PolicySyncResponse) GetRequestInfoOk() (*PolicySyncResponseRequestInfo, bool)`

GetRequestInfoOk returns a tuple with the RequestInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestInfo

`func (o *PolicySyncResponse) SetRequestInfo(v PolicySyncResponseRequestInfo)`

SetRequestInfo sets RequestInfo field to given value.

### HasRequestInfo

`func (o *PolicySyncResponse) HasRequestInfo() bool`

HasRequestInfo returns a boolean if a field has been set.

### GetResponseInfo

`func (o *PolicySyncResponse) GetResponseInfo() PolicySyncResponseResponseInfo`

GetResponseInfo returns the ResponseInfo field if non-nil, zero value otherwise.

### GetResponseInfoOk

`func (o *PolicySyncResponse) GetResponseInfoOk() (*PolicySyncResponseResponseInfo, bool)`

GetResponseInfoOk returns a tuple with the ResponseInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseInfo

`func (o *PolicySyncResponse) SetResponseInfo(v PolicySyncResponseResponseInfo)`

SetResponseInfo sets ResponseInfo field to given value.

### HasResponseInfo

`func (o *PolicySyncResponse) HasResponseInfo() bool`

HasResponseInfo returns a boolean if a field has been set.

### GetSystemInfo

`func (o *PolicySyncResponse) GetSystemInfo() PolicySyncResponseSystemInfo`

GetSystemInfo returns the SystemInfo field if non-nil, zero value otherwise.

### GetSystemInfoOk

`func (o *PolicySyncResponse) GetSystemInfoOk() (*PolicySyncResponseSystemInfo, bool)`

GetSystemInfoOk returns a tuple with the SystemInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSystemInfo

`func (o *PolicySyncResponse) SetSystemInfo(v PolicySyncResponseSystemInfo)`

SetSystemInfo sets SystemInfo field to given value.

### HasSystemInfo

`func (o *PolicySyncResponse) HasSystemInfo() bool`

HasSystemInfo returns a boolean if a field has been set.

### GetPolicies

`func (o *PolicySyncResponse) GetPolicies() []PolicySyncResponsePoliciesInner`

GetPolicies returns the Policies field if non-nil, zero value otherwise.

### GetPoliciesOk

`func (o *PolicySyncResponse) GetPoliciesOk() (*[]PolicySyncResponsePoliciesInner, bool)`

GetPoliciesOk returns a tuple with the Policies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicies

`func (o *PolicySyncResponse) SetPolicies(v []PolicySyncResponsePoliciesInner)`

SetPolicies sets Policies field to given value.

### HasPolicies

`func (o *PolicySyncResponse) HasPolicies() bool`

HasPolicies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


