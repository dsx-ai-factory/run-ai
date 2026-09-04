# SharedScopesPatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Add** | Pointer to [**[]SharedScopeBase**](SharedScopeBase.md) |  | [optional] 
**Remove** | Pointer to [**[]SharedScopeBase**](SharedScopeBase.md) |  | [optional] 

## Methods

### NewSharedScopesPatchRequest

`func NewSharedScopesPatchRequest() *SharedScopesPatchRequest`

NewSharedScopesPatchRequest instantiates a new SharedScopesPatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSharedScopesPatchRequestWithDefaults

`func NewSharedScopesPatchRequestWithDefaults() *SharedScopesPatchRequest`

NewSharedScopesPatchRequestWithDefaults instantiates a new SharedScopesPatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAdd

`func (o *SharedScopesPatchRequest) GetAdd() []SharedScopeBase`

GetAdd returns the Add field if non-nil, zero value otherwise.

### GetAddOk

`func (o *SharedScopesPatchRequest) GetAddOk() (*[]SharedScopeBase, bool)`

GetAddOk returns a tuple with the Add field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdd

`func (o *SharedScopesPatchRequest) SetAdd(v []SharedScopeBase)`

SetAdd sets Add field to given value.

### HasAdd

`func (o *SharedScopesPatchRequest) HasAdd() bool`

HasAdd returns a boolean if a field has been set.

### GetRemove

`func (o *SharedScopesPatchRequest) GetRemove() []SharedScopeBase`

GetRemove returns the Remove field if non-nil, zero value otherwise.

### GetRemoveOk

`func (o *SharedScopesPatchRequest) GetRemoveOk() (*[]SharedScopeBase, bool)`

GetRemoveOk returns a tuple with the Remove field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRemove

`func (o *SharedScopesPatchRequest) SetRemove(v []SharedScopeBase)`

SetRemove sets Remove field to given value.

### HasRemove

`func (o *SharedScopesPatchRequest) HasRemove() bool`

HasRemove returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


