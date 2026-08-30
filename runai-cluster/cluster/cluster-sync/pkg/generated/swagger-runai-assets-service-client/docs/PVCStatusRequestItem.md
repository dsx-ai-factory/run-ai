# PVCStatusRequestItem

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | ID of the cluster who sent the request | 
**ClaimName** | **string** | The name of the pvc in the cluster | 
**StatusInfo** | [**PVCStatusSyncStatusInfo**](PVCStatusSyncStatusInfo.md) |  | 
**ScopeType** | [**Scope**](Scope.md) |  | 
**ScopeId** | **string** | The ID of the scope the asset is belong to | 

## Methods

### NewPVCStatusRequestItem

`func NewPVCStatusRequestItem(clusterId string, claimName string, statusInfo PVCStatusSyncStatusInfo, scopeType Scope, scopeId string, ) *PVCStatusRequestItem`

NewPVCStatusRequestItem instantiates a new PVCStatusRequestItem object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCStatusRequestItemWithDefaults

`func NewPVCStatusRequestItemWithDefaults() *PVCStatusRequestItem`

NewPVCStatusRequestItemWithDefaults instantiates a new PVCStatusRequestItem object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *PVCStatusRequestItem) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PVCStatusRequestItem) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PVCStatusRequestItem) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetClaimName

`func (o *PVCStatusRequestItem) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *PVCStatusRequestItem) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *PVCStatusRequestItem) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.


### GetStatusInfo

`func (o *PVCStatusRequestItem) GetStatusInfo() PVCStatusSyncStatusInfo`

GetStatusInfo returns the StatusInfo field if non-nil, zero value otherwise.

### GetStatusInfoOk

`func (o *PVCStatusRequestItem) GetStatusInfoOk() (*PVCStatusSyncStatusInfo, bool)`

GetStatusInfoOk returns a tuple with the StatusInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusInfo

`func (o *PVCStatusRequestItem) SetStatusInfo(v PVCStatusSyncStatusInfo)`

SetStatusInfo sets StatusInfo field to given value.


### GetScopeType

`func (o *PVCStatusRequestItem) GetScopeType() Scope`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *PVCStatusRequestItem) GetScopeTypeOk() (*Scope, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *PVCStatusRequestItem) SetScopeType(v Scope)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *PVCStatusRequestItem) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *PVCStatusRequestItem) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *PVCStatusRequestItem) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


