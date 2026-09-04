# PVCStatusSyncStatusInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Status** | Pointer to [**NullableAssetStatus**](AssetStatus.md) |  | [optional] 
**ErrorMessage** | Pointer to **string** | An error message if something wrong happened in the cluster | [optional] 

## Methods

### NewPVCStatusSyncStatusInfo

`func NewPVCStatusSyncStatusInfo() *PVCStatusSyncStatusInfo`

NewPVCStatusSyncStatusInfo instantiates a new PVCStatusSyncStatusInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCStatusSyncStatusInfoWithDefaults

`func NewPVCStatusSyncStatusInfoWithDefaults() *PVCStatusSyncStatusInfo`

NewPVCStatusSyncStatusInfoWithDefaults instantiates a new PVCStatusSyncStatusInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatus

`func (o *PVCStatusSyncStatusInfo) GetStatus() AssetStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PVCStatusSyncStatusInfo) GetStatusOk() (*AssetStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PVCStatusSyncStatusInfo) SetStatus(v AssetStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *PVCStatusSyncStatusInfo) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *PVCStatusSyncStatusInfo) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *PVCStatusSyncStatusInfo) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetErrorMessage

`func (o *PVCStatusSyncStatusInfo) GetErrorMessage() string`

GetErrorMessage returns the ErrorMessage field if non-nil, zero value otherwise.

### GetErrorMessageOk

`func (o *PVCStatusSyncStatusInfo) GetErrorMessageOk() (*string, bool)`

GetErrorMessageOk returns a tuple with the ErrorMessage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorMessage

`func (o *PVCStatusSyncStatusInfo) SetErrorMessage(v string)`

SetErrorMessage sets ErrorMessage field to given value.

### HasErrorMessage

`func (o *PVCStatusSyncStatusInfo) HasErrorMessage() bool`

HasErrorMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


