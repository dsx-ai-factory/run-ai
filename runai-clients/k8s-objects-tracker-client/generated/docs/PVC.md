# PVC

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | name of the pvc. | 
**Origin** | Pointer to [**NullableOrigin**](Origin.md) |  | [optional] 
**Issues** | Pointer to [**[]Issue**](Issue.md) | list of issues that was found while syncing the resource to the cluster | [optional] 
**Phase** | Pointer to [**NullablePvcPhaseOptional**](PvcPhaseOptional.md) |  | [optional] 
**RequestedCapacity** | Pointer to **NullableString** | The capacity of the PVC. | [optional] 
**ActualCapacity** | Pointer to **NullableString** | The capacity of the PVC. | [optional] 
**StorageClassName** | Pointer to **NullableString** | The name of the storage class. | [optional] 
**DataSharing** | Pointer to **NullableBool** | A Flag to mark if the PVC data is shared. | [optional] 

## Methods

### NewPVC

`func NewPVC(name string, ) *PVC`

NewPVC instantiates a new PVC object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCWithDefaults

`func NewPVCWithDefaults() *PVC`

NewPVCWithDefaults instantiates a new PVC object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PVC) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PVC) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PVC) SetName(v string)`

SetName sets Name field to given value.


### GetOrigin

`func (o *PVC) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *PVC) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *PVC) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.

### HasOrigin

`func (o *PVC) HasOrigin() bool`

HasOrigin returns a boolean if a field has been set.

### SetOriginNil

`func (o *PVC) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *PVC) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetIssues

`func (o *PVC) GetIssues() []Issue`

GetIssues returns the Issues field if non-nil, zero value otherwise.

### GetIssuesOk

`func (o *PVC) GetIssuesOk() (*[]Issue, bool)`

GetIssuesOk returns a tuple with the Issues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIssues

`func (o *PVC) SetIssues(v []Issue)`

SetIssues sets Issues field to given value.

### HasIssues

`func (o *PVC) HasIssues() bool`

HasIssues returns a boolean if a field has been set.

### GetPhase

`func (o *PVC) GetPhase() PvcPhaseOptional`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PVC) GetPhaseOk() (*PvcPhaseOptional, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PVC) SetPhase(v PvcPhaseOptional)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PVC) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### SetPhaseNil

`func (o *PVC) SetPhaseNil(b bool)`

 SetPhaseNil sets the value for Phase to be an explicit nil

### UnsetPhase
`func (o *PVC) UnsetPhase()`

UnsetPhase ensures that no value is present for Phase, not even an explicit nil
### GetRequestedCapacity

`func (o *PVC) GetRequestedCapacity() string`

GetRequestedCapacity returns the RequestedCapacity field if non-nil, zero value otherwise.

### GetRequestedCapacityOk

`func (o *PVC) GetRequestedCapacityOk() (*string, bool)`

GetRequestedCapacityOk returns a tuple with the RequestedCapacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedCapacity

`func (o *PVC) SetRequestedCapacity(v string)`

SetRequestedCapacity sets RequestedCapacity field to given value.

### HasRequestedCapacity

`func (o *PVC) HasRequestedCapacity() bool`

HasRequestedCapacity returns a boolean if a field has been set.

### SetRequestedCapacityNil

`func (o *PVC) SetRequestedCapacityNil(b bool)`

 SetRequestedCapacityNil sets the value for RequestedCapacity to be an explicit nil

### UnsetRequestedCapacity
`func (o *PVC) UnsetRequestedCapacity()`

UnsetRequestedCapacity ensures that no value is present for RequestedCapacity, not even an explicit nil
### GetActualCapacity

`func (o *PVC) GetActualCapacity() string`

GetActualCapacity returns the ActualCapacity field if non-nil, zero value otherwise.

### GetActualCapacityOk

`func (o *PVC) GetActualCapacityOk() (*string, bool)`

GetActualCapacityOk returns a tuple with the ActualCapacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualCapacity

`func (o *PVC) SetActualCapacity(v string)`

SetActualCapacity sets ActualCapacity field to given value.

### HasActualCapacity

`func (o *PVC) HasActualCapacity() bool`

HasActualCapacity returns a boolean if a field has been set.

### SetActualCapacityNil

`func (o *PVC) SetActualCapacityNil(b bool)`

 SetActualCapacityNil sets the value for ActualCapacity to be an explicit nil

### UnsetActualCapacity
`func (o *PVC) UnsetActualCapacity()`

UnsetActualCapacity ensures that no value is present for ActualCapacity, not even an explicit nil
### GetStorageClassName

`func (o *PVC) GetStorageClassName() string`

GetStorageClassName returns the StorageClassName field if non-nil, zero value otherwise.

### GetStorageClassNameOk

`func (o *PVC) GetStorageClassNameOk() (*string, bool)`

GetStorageClassNameOk returns a tuple with the StorageClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassName

`func (o *PVC) SetStorageClassName(v string)`

SetStorageClassName sets StorageClassName field to given value.

### HasStorageClassName

`func (o *PVC) HasStorageClassName() bool`

HasStorageClassName returns a boolean if a field has been set.

### SetStorageClassNameNil

`func (o *PVC) SetStorageClassNameNil(b bool)`

 SetStorageClassNameNil sets the value for StorageClassName to be an explicit nil

### UnsetStorageClassName
`func (o *PVC) UnsetStorageClassName()`

UnsetStorageClassName ensures that no value is present for StorageClassName, not even an explicit nil
### GetDataSharing

`func (o *PVC) GetDataSharing() bool`

GetDataSharing returns the DataSharing field if non-nil, zero value otherwise.

### GetDataSharingOk

`func (o *PVC) GetDataSharingOk() (*bool, bool)`

GetDataSharingOk returns a tuple with the DataSharing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataSharing

`func (o *PVC) SetDataSharing(v bool)`

SetDataSharing sets DataSharing field to given value.

### HasDataSharing

`func (o *PVC) HasDataSharing() bool`

HasDataSharing returns a boolean if a field has been set.

### SetDataSharingNil

`func (o *PVC) SetDataSharingNil(b bool)`

 SetDataSharingNil sets the value for DataSharing to be an explicit nil

### UnsetDataSharing
`func (o *PVC) UnsetDataSharing()`

UnsetDataSharing ensures that no value is present for DataSharing, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


