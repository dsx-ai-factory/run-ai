# PVCListInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClusterId** | **string** | The id of the cluster. | 
**ScopeType** | **string** | the scope type of the resource. | 
**ScopeId** | **string** | The id of the tenant, cluster, department or project, depending on the scope type. | 
**Name** | **string** | name of the pvc. | 
**Origin** | [**NullableOrigin**](Origin.md) |  | 
**Phase** | Pointer to [**NullablePvcPhaseOptional**](PvcPhaseOptional.md) |  | [optional] 
**RequestedCapacity** | Pointer to **NullableString** | The capacity of the PVC. | [optional] 
**ActualCapacity** | Pointer to **NullableString** | The capacity of the PVC. | [optional] 
**StorageClassName** | Pointer to **NullableString** | The name of the storage class. | [optional] 
**DataSharing** | Pointer to **NullableBool** | A Flag to mark if the PVC data is shared. | [optional] 

## Methods

### NewPVCListInfo

`func NewPVCListInfo(clusterId string, scopeType string, scopeId string, name string, origin NullableOrigin, ) *PVCListInfo`

NewPVCListInfo instantiates a new PVCListInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCListInfoWithDefaults

`func NewPVCListInfoWithDefaults() *PVCListInfo`

NewPVCListInfoWithDefaults instantiates a new PVCListInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClusterId

`func (o *PVCListInfo) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PVCListInfo) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PVCListInfo) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetScopeType

`func (o *PVCListInfo) GetScopeType() string`

GetScopeType returns the ScopeType field if non-nil, zero value otherwise.

### GetScopeTypeOk

`func (o *PVCListInfo) GetScopeTypeOk() (*string, bool)`

GetScopeTypeOk returns a tuple with the ScopeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeType

`func (o *PVCListInfo) SetScopeType(v string)`

SetScopeType sets ScopeType field to given value.


### GetScopeId

`func (o *PVCListInfo) GetScopeId() string`

GetScopeId returns the ScopeId field if non-nil, zero value otherwise.

### GetScopeIdOk

`func (o *PVCListInfo) GetScopeIdOk() (*string, bool)`

GetScopeIdOk returns a tuple with the ScopeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopeId

`func (o *PVCListInfo) SetScopeId(v string)`

SetScopeId sets ScopeId field to given value.


### GetName

`func (o *PVCListInfo) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PVCListInfo) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PVCListInfo) SetName(v string)`

SetName sets Name field to given value.


### GetOrigin

`func (o *PVCListInfo) GetOrigin() Origin`

GetOrigin returns the Origin field if non-nil, zero value otherwise.

### GetOriginOk

`func (o *PVCListInfo) GetOriginOk() (*Origin, bool)`

GetOriginOk returns a tuple with the Origin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrigin

`func (o *PVCListInfo) SetOrigin(v Origin)`

SetOrigin sets Origin field to given value.


### SetOriginNil

`func (o *PVCListInfo) SetOriginNil(b bool)`

 SetOriginNil sets the value for Origin to be an explicit nil

### UnsetOrigin
`func (o *PVCListInfo) UnsetOrigin()`

UnsetOrigin ensures that no value is present for Origin, not even an explicit nil
### GetPhase

`func (o *PVCListInfo) GetPhase() PvcPhaseOptional`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PVCListInfo) GetPhaseOk() (*PvcPhaseOptional, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PVCListInfo) SetPhase(v PvcPhaseOptional)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PVCListInfo) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### SetPhaseNil

`func (o *PVCListInfo) SetPhaseNil(b bool)`

 SetPhaseNil sets the value for Phase to be an explicit nil

### UnsetPhase
`func (o *PVCListInfo) UnsetPhase()`

UnsetPhase ensures that no value is present for Phase, not even an explicit nil
### GetRequestedCapacity

`func (o *PVCListInfo) GetRequestedCapacity() string`

GetRequestedCapacity returns the RequestedCapacity field if non-nil, zero value otherwise.

### GetRequestedCapacityOk

`func (o *PVCListInfo) GetRequestedCapacityOk() (*string, bool)`

GetRequestedCapacityOk returns a tuple with the RequestedCapacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestedCapacity

`func (o *PVCListInfo) SetRequestedCapacity(v string)`

SetRequestedCapacity sets RequestedCapacity field to given value.

### HasRequestedCapacity

`func (o *PVCListInfo) HasRequestedCapacity() bool`

HasRequestedCapacity returns a boolean if a field has been set.

### SetRequestedCapacityNil

`func (o *PVCListInfo) SetRequestedCapacityNil(b bool)`

 SetRequestedCapacityNil sets the value for RequestedCapacity to be an explicit nil

### UnsetRequestedCapacity
`func (o *PVCListInfo) UnsetRequestedCapacity()`

UnsetRequestedCapacity ensures that no value is present for RequestedCapacity, not even an explicit nil
### GetActualCapacity

`func (o *PVCListInfo) GetActualCapacity() string`

GetActualCapacity returns the ActualCapacity field if non-nil, zero value otherwise.

### GetActualCapacityOk

`func (o *PVCListInfo) GetActualCapacityOk() (*string, bool)`

GetActualCapacityOk returns a tuple with the ActualCapacity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActualCapacity

`func (o *PVCListInfo) SetActualCapacity(v string)`

SetActualCapacity sets ActualCapacity field to given value.

### HasActualCapacity

`func (o *PVCListInfo) HasActualCapacity() bool`

HasActualCapacity returns a boolean if a field has been set.

### SetActualCapacityNil

`func (o *PVCListInfo) SetActualCapacityNil(b bool)`

 SetActualCapacityNil sets the value for ActualCapacity to be an explicit nil

### UnsetActualCapacity
`func (o *PVCListInfo) UnsetActualCapacity()`

UnsetActualCapacity ensures that no value is present for ActualCapacity, not even an explicit nil
### GetStorageClassName

`func (o *PVCListInfo) GetStorageClassName() string`

GetStorageClassName returns the StorageClassName field if non-nil, zero value otherwise.

### GetStorageClassNameOk

`func (o *PVCListInfo) GetStorageClassNameOk() (*string, bool)`

GetStorageClassNameOk returns a tuple with the StorageClassName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClassName

`func (o *PVCListInfo) SetStorageClassName(v string)`

SetStorageClassName sets StorageClassName field to given value.

### HasStorageClassName

`func (o *PVCListInfo) HasStorageClassName() bool`

HasStorageClassName returns a boolean if a field has been set.

### SetStorageClassNameNil

`func (o *PVCListInfo) SetStorageClassNameNil(b bool)`

 SetStorageClassNameNil sets the value for StorageClassName to be an explicit nil

### UnsetStorageClassName
`func (o *PVCListInfo) UnsetStorageClassName()`

UnsetStorageClassName ensures that no value is present for StorageClassName, not even an explicit nil
### GetDataSharing

`func (o *PVCListInfo) GetDataSharing() bool`

GetDataSharing returns the DataSharing field if non-nil, zero value otherwise.

### GetDataSharingOk

`func (o *PVCListInfo) GetDataSharingOk() (*bool, bool)`

GetDataSharingOk returns a tuple with the DataSharing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataSharing

`func (o *PVCListInfo) SetDataSharing(v bool)`

SetDataSharing sets DataSharing field to given value.

### HasDataSharing

`func (o *PVCListInfo) HasDataSharing() bool`

HasDataSharing returns a boolean if a field has been set.

### SetDataSharingNil

`func (o *PVCListInfo) SetDataSharingNil(b bool)`

 SetDataSharingNil sets the value for DataSharing to be an explicit nil

### UnsetDataSharing
`func (o *PVCListInfo) UnsetDataSharing()`

UnsetDataSharing ensures that no value is present for DataSharing, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


