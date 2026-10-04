# WorkloadTypesForSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FullSync** | Pointer to **bool** | Upsert all items and delete all missing items | [optional] 
**WorkloadTypes** | Pointer to [**[]WorkloadTypeForSync**](WorkloadTypeForSync.md) |  | [optional] 

## Methods

### NewWorkloadTypesForSync

`func NewWorkloadTypesForSync() *WorkloadTypesForSync`

NewWorkloadTypesForSync instantiates a new WorkloadTypesForSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadTypesForSyncWithDefaults

`func NewWorkloadTypesForSyncWithDefaults() *WorkloadTypesForSync`

NewWorkloadTypesForSyncWithDefaults instantiates a new WorkloadTypesForSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFullSync

`func (o *WorkloadTypesForSync) GetFullSync() bool`

GetFullSync returns the FullSync field if non-nil, zero value otherwise.

### GetFullSyncOk

`func (o *WorkloadTypesForSync) GetFullSyncOk() (*bool, bool)`

GetFullSyncOk returns a tuple with the FullSync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullSync

`func (o *WorkloadTypesForSync) SetFullSync(v bool)`

SetFullSync sets FullSync field to given value.

### HasFullSync

`func (o *WorkloadTypesForSync) HasFullSync() bool`

HasFullSync returns a boolean if a field has been set.

### GetWorkloadTypes

`func (o *WorkloadTypesForSync) GetWorkloadTypes() []WorkloadTypeForSync`

GetWorkloadTypes returns the WorkloadTypes field if non-nil, zero value otherwise.

### GetWorkloadTypesOk

`func (o *WorkloadTypesForSync) GetWorkloadTypesOk() (*[]WorkloadTypeForSync, bool)`

GetWorkloadTypesOk returns a tuple with the WorkloadTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkloadTypes

`func (o *WorkloadTypesForSync) SetWorkloadTypes(v []WorkloadTypeForSync)`

SetWorkloadTypes sets WorkloadTypes field to given value.

### HasWorkloadTypes

`func (o *WorkloadTypesForSync) HasWorkloadTypes() bool`

HasWorkloadTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


