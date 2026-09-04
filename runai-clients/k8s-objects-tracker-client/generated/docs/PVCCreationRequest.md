# PVCCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Origin** | Pointer to [**NullableOrigin**](Origin.md) |  | [optional] 
**Issues** | Pointer to [**[]Issue**](Issue.md) | list of issues that was found while syncing the resource to the cluster | [optional] 
**Phase** | Pointer to [**NullablePvcPhaseOptional**](PvcPhaseOptional.md) |  | [optional] 
**RequestedCapacity** | Pointer to **NullableString** | The capacity of the PVC. | [optional] 
**ActualCapacity** | Pointer to **NullableString** | The capacity of the PVC. | [optional] 
**StorageClassName** | Pointer to **NullableString** | The name of the storage class. | [optional] 
**DataSharing** | Pointer to **NullableBool** | A Flag to mark if the PVC data is shared. | [optional] 

## Methods

### NewPVCCreationRequest

`func NewPVCCreationRequest() *PVCCreationRequest`

NewPVCCreationRequest instantiates a new PVCCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCCreationRequestWithDefaults

`func NewPVCCreationRequestWithDefaults() *PVCCreationRequest`

NewPVCCreationRequestWithDefaults instantiates a new PVCCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOrigin

`func (o *PVCCreationRequest) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *PVCCreationRequest) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *PVCCreationRequest) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.

### HasOrigin

`func (o *PVCCreationRequest) HasOrigin() bool`

HasOrigin returns a boolean if a field has been set.

### SetOriginNil

`func (o *PVCCreationRequest) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *PVCCreationRequest) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetIssues

`func (o *PVCCreationRequest) GetIssues() []Issue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *PVCCreationRequest) GetIssuesOk() (*[]Issue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *PVCCreationRequest) SetIssues(v []Issue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *PVCCreationRequest) HasIssues() bool`

HasIssues returns a boolean if a field has been set.

### GetPhase

`func (o *PVCCreationRequest) GetPhase() PvcPhaseOptional`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PVCCreationRequest) GetPhaseOk() (*PvcPhaseOptional, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PVCCreationRequest) SetPhase(v PvcPhaseOptional)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PVCCreationRequest) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### SetPhaseNil

`func (o *PVCCreationRequest) SetPhaseNil(b bool)`

 SetPhaseNil sets the value for Phase to be an explicit nil

### UnsetPhase
`func (o *PVCCreationRequest) UnsetPhase()`

UnsetPhase ensures that no value is present for Phase, not even an explicit nil
### GetRequestedCapacity

`func (o *PVCCreationRequest) GetRequestedCapacity() string`

GetRequestedCapacity returns the RequestedCapacity field if non-nil, zero value otherwise.

### GetRequestedCapacityOk

`func (o *PVCCreationRequest) GetRequestedCapacityOk() (*string, bool)`

GetRequestedCapacityOk returns a tuple with the RequestedCapacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedCapacity

`func (o *PVCCreationRequest) SetRequestedCapacity(v string)`

SetRequestedCapacity sets RequestedCapacity field to given value.

### HasRequestedCapacity

`func (o *PVCCreationRequest) HasRequestedCapacity() bool`

HasRequestedCapacity returns a boolean if a field has been set.

### SetRequestedCapacityNil

`func (o *PVCCreationRequest) SetRequestedCapacityNil(b bool)`

 SetRequestedCapacityNil sets the value for RequestedCapacity to be an explicit nil

### UnsetRequestedCapacity
`func (o *PVCCreationRequest) UnsetRequestedCapacity()`

UnsetRequestedCapacity ensures that no value is present for RequestedCapacity, not even an explicit nil
### GetActualCapacity

`func (o *PVCCreationRequest) GetActualCapacity() string`

GetActualCapacity returns the ActualCapacity field if non-nil, zero value otherwise.

### GetActualCapacityOk

`func (o *PVCCreationRequest) GetActualCapacityOk() (*string, bool)`

GetActualCapacityOk returns a tuple with the ActualCapacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualCapacity

`func (o *PVCCreationRequest) SetActualCapacity(v string)`

SetActualCapacity sets ActualCapacity field to given value.

### HasActualCapacity

`func (o *PVCCreationRequest) HasActualCapacity() bool`

HasActualCapacity returns a boolean if a field has been set.

### SetActualCapacityNil

`func (o *PVCCreationRequest) SetActualCapacityNil(b bool)`

 SetActualCapacityNil sets the value for ActualCapacity to be an explicit nil

### UnsetActualCapacity
`func (o *PVCCreationRequest) UnsetActualCapacity()`

UnsetActualCapacity ensures that no value is present for ActualCapacity, not even an explicit nil
### GetStorageClassName

`func (o *PVCCreationRequest) GetStorageClassName() string`

GetStorageClassName returns the StorageClassName field if non-nil, zero value otherwise.

### GetStorageClassNameOk

`func (o *PVCCreationRequest) GetStorageClassNameOk() (*string, bool)`

GetStorageClassNameOk returns a tuple with the StorageClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassName

`func (o *PVCCreationRequest) SetStorageClassName(v string)`

SetStorageClassName sets StorageClassName field to given value.

### HasStorageClassName

`func (o *PVCCreationRequest) HasStorageClassName() bool`

HasStorageClassName returns a boolean if a field has been set.

### SetStorageClassNameNil

`func (o *PVCCreationRequest) SetStorageClassNameNil(b bool)`

 SetStorageClassNameNil sets the value for StorageClassName to be an explicit nil

### UnsetStorageClassName
`func (o *PVCCreationRequest) UnsetStorageClassName()`

UnsetStorageClassName ensures that no value is present for StorageClassName, not even an explicit nil
### GetDataSharing

`func (o *PVCCreationRequest) GetDataSharing() bool`

GetDataSharing returns the DataSharing field if non-nil, zero value otherwise.

### GetDataSharingOk

`func (o *PVCCreationRequest) GetDataSharingOk() (*bool, bool)`

GetDataSharingOk returns a tuple with the DataSharing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataSharing

`func (o *PVCCreationRequest) SetDataSharing(v bool)`

SetDataSharing sets DataSharing field to given value.

### HasDataSharing

`func (o *PVCCreationRequest) HasDataSharing() bool`

HasDataSharing returns a boolean if a field has been set.

### SetDataSharingNil

`func (o *PVCCreationRequest) SetDataSharingNil(b bool)`

 SetDataSharingNil sets the value for DataSharing to be an explicit nil

### UnsetDataSharing
`func (o *PVCCreationRequest) UnsetDataSharing()`

UnsetDataSharing ensures that no value is present for DataSharing, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


