# DatavolumeInternal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**OriginPvcName** | **string** | The name of the PVC that the datavolume is based on | 
**ProjectId** | **string** | The ID of the project that in its namespace the origin pvc is located | 
**ShouldDeleteOriginalVolume** | Pointer to **bool** | If true, the original storage volume will be deleted together with the datavolume | [optional] [default to false]
**ProjectName** | **string** |  | 
**DepartmentId** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**SharedScopes** | Pointer to [**[]SharedScope**](SharedScope.md) | Will be returned only if the user has the required permissions to view those scopes | [optional] 
**Phase** | Pointer to [**DatavolumePhase**](DatavolumePhase.md) |  | [optional] 

## Methods

### NewDatavolumeInternal

`func NewDatavolumeInternal(id string, name string, originPvcName string, projectId string, projectName string, ) *DatavolumeInternal`

NewDatavolumeInternal instantiates a new DatavolumeInternal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeInternalWithDefaults

`func NewDatavolumeInternalWithDefaults() *DatavolumeInternal`

NewDatavolumeInternalWithDefaults instantiates a new DatavolumeInternal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DatavolumeInternal) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DatavolumeInternal) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DatavolumeInternal) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *DatavolumeInternal) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DatavolumeInternal) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DatavolumeInternal) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DatavolumeInternal) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *DatavolumeInternal) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DatavolumeInternal) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DatavolumeInternal) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *DatavolumeInternal) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *DatavolumeInternal) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *DatavolumeInternal) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *DatavolumeInternal) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *DatavolumeInternal) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *DatavolumeInternal) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *DatavolumeInternal) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *DatavolumeInternal) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *DatavolumeInternal) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetName

`func (o *DatavolumeInternal) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DatavolumeInternal) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DatavolumeInternal) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *DatavolumeInternal) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DatavolumeInternal) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DatavolumeInternal) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DatavolumeInternal) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOriginPvcName

`func (o *DatavolumeInternal) GetOriginPvcName() string`

GetOriginPvcName returns the OriginPvcName field if non-nil, zero value otherwise.

### GetOriginPvcNameOk

`func (o *DatavolumeInternal) GetOriginPvcNameOk() (*string, bool)`

GetOriginPvcNameOk returns a tuple with the OriginPvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPvcName

`func (o *DatavolumeInternal) SetOriginPvcName(v string)`

SetOriginPvcName sets OriginPvcName field to given value.


### GetProjectId

`func (o *DatavolumeInternal) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *DatavolumeInternal) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *DatavolumeInternal) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetShouldDeleteOriginalVolume

`func (o *DatavolumeInternal) GetShouldDeleteOriginalVolume() bool`

GetShouldDeleteOriginalVolume returns the ShouldDeleteOriginalVolume field if non-nil, zero value otherwise.

### GetShouldDeleteOriginalVolumeOk

`func (o *DatavolumeInternal) GetShouldDeleteOriginalVolumeOk() (*bool, bool)`

GetShouldDeleteOriginalVolumeOk returns a tuple with the ShouldDeleteOriginalVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShouldDeleteOriginalVolume

`func (o *DatavolumeInternal) SetShouldDeleteOriginalVolume(v bool)`

SetShouldDeleteOriginalVolume sets ShouldDeleteOriginalVolume field to given value.

### HasShouldDeleteOriginalVolume

`func (o *DatavolumeInternal) HasShouldDeleteOriginalVolume() bool`

HasShouldDeleteOriginalVolume returns a boolean if a field has been set.

### GetProjectName

`func (o *DatavolumeInternal) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *DatavolumeInternal) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *DatavolumeInternal) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetDepartmentId

`func (o *DatavolumeInternal) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *DatavolumeInternal) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *DatavolumeInternal) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *DatavolumeInternal) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *DatavolumeInternal) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DatavolumeInternal) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DatavolumeInternal) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *DatavolumeInternal) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetSharedScopes

`func (o *DatavolumeInternal) GetSharedScopes() []SharedScope`

GetSharedScopes returns the SharedScopes field if non-nil, zero value otherwise.

### GetSharedScopesOk

`func (o *DatavolumeInternal) GetSharedScopesOk() (*[]SharedScope, bool)`

GetSharedScopesOk returns a tuple with the SharedScopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSharedScopes

`func (o *DatavolumeInternal) SetSharedScopes(v []SharedScope)`

SetSharedScopes sets SharedScopes field to given value.

### HasSharedScopes

`func (o *DatavolumeInternal) HasSharedScopes() bool`

HasSharedScopes returns a boolean if a field has been set.

### GetPhase

`func (o *DatavolumeInternal) GetPhase() DatavolumePhase`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *DatavolumeInternal) GetPhaseOk() (*DatavolumePhase, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *DatavolumeInternal) SetPhase(v DatavolumePhase)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *DatavolumeInternal) HasPhase() bool`

HasPhase returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


