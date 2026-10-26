# SyncPatchRequestInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**SyncFailed** | Pointer to **bool** |  | [optional] 

## Methods

### NewSyncPatchRequestInner

`func NewSyncPatchRequestInner() *SyncPatchRequestInner`

NewSyncPatchRequestInner instantiates a new SyncPatchRequestInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSyncPatchRequestInnerWithDefaults

`func NewSyncPatchRequestInnerWithDefaults() *SyncPatchRequestInner`

NewSyncPatchRequestInnerWithDefaults instantiates a new SyncPatchRequestInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SyncPatchRequestInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SyncPatchRequestInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SyncPatchRequestInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *SyncPatchRequestInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetSyncFailed

`func (o *SyncPatchRequestInner) GetSyncFailed() bool`

GetSyncFailed returns the SyncFailed field if non-nil, zero value otherwise.

### GetSyncFailedOk

`func (o *SyncPatchRequestInner) GetSyncFailedOk() (*bool, bool)`

GetSyncFailedOk returns a tuple with the SyncFailed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSyncFailed

`func (o *SyncPatchRequestInner) SetSyncFailed(v bool)`

SetSyncFailed sets SyncFailed field to given value.

### HasSyncFailed

`func (o *SyncPatchRequestInner) HasSyncFailed() bool`

HasSyncFailed returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


