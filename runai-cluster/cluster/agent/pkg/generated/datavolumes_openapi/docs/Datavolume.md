# Datavolume

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** |  | 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 
**UpdatedBy** | Pointer to **string** |  | [optional] 
**Name** | **string** | The name of the new data volume. | 
**Description** | Pointer to **string** | A brief description of the data volume and its purpose. | [optional] 
**OriginPvcName** | **string** | The name of the existing PVC that the new data volume is based on. | 
**ProjectId** | **string** | The ID of the project whose namespace contains the origin PVC. | 
**ShouldDeleteOriginalVolume** | Pointer to **bool** | True is not supported. This parameter will be removed in a future release. | [optional] [default to false]
**UseOriginalVolume** | Pointer to **NullableBool** | If true, the original storage volume will be used together with the data volume. Default true as long as \&quot;shouldDeleteOriginalVolume\&quot; is false. | [optional] [default to true]
**ProjectName** | **string** |  | 
**DepartmentId** | Pointer to **string** |  | [optional] 
**ClusterId** | Pointer to **string** | The id of the cluster. | [optional] 
**SharedScopes** | Pointer to [**[]SharedScope**](SharedScope.md) | Will be returned only if the user has the required permissions to view those scopes | [optional] 
**Status** | [**DatavolumeStatus**](DatavolumeStatus.md) |  | 

## Methods

### NewDatavolume

`func NewDatavolume(id string, name string, originPvcName string, projectId string, projectName string, status DatavolumeStatus, ) *Datavolume`

NewDatavolume instantiates a new Datavolume object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeWithDefaults

`func NewDatavolumeWithDefaults() *Datavolume`

NewDatavolumeWithDefaults instantiates a new Datavolume object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Datavolume) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Datavolume) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Datavolume) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *Datavolume) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Datavolume) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Datavolume) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Datavolume) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *Datavolume) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Datavolume) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Datavolume) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *Datavolume) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *Datavolume) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *Datavolume) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *Datavolume) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *Datavolume) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *Datavolume) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *Datavolume) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *Datavolume) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *Datavolume) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetName

`func (o *Datavolume) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Datavolume) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Datavolume) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *Datavolume) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *Datavolume) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *Datavolume) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *Datavolume) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOriginPvcName

`func (o *Datavolume) GetOriginPvcName() string`

GetOriginPvcName returns the OriginPvcName field if non-nil, zero value otherwise.

### GetOriginPvcNameOk

`func (o *Datavolume) GetOriginPvcNameOk() (*string, bool)`

GetOriginPvcNameOk returns a tuple with the OriginPvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPvcName

`func (o *Datavolume) SetOriginPvcName(v string)`

SetOriginPvcName sets OriginPvcName field to given value.


### GetProjectId

`func (o *Datavolume) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *Datavolume) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *Datavolume) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetShouldDeleteOriginalVolume

`func (o *Datavolume) GetShouldDeleteOriginalVolume() bool`

GetShouldDeleteOriginalVolume returns the ShouldDeleteOriginalVolume field if non-nil, zero value otherwise.

### GetShouldDeleteOriginalVolumeOk

`func (o *Datavolume) GetShouldDeleteOriginalVolumeOk() (*bool, bool)`

GetShouldDeleteOriginalVolumeOk returns a tuple with the ShouldDeleteOriginalVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShouldDeleteOriginalVolume

`func (o *Datavolume) SetShouldDeleteOriginalVolume(v bool)`

SetShouldDeleteOriginalVolume sets ShouldDeleteOriginalVolume field to given value.

### HasShouldDeleteOriginalVolume

`func (o *Datavolume) HasShouldDeleteOriginalVolume() bool`

HasShouldDeleteOriginalVolume returns a boolean if a field has been set.

### GetUseOriginalVolume

`func (o *Datavolume) GetUseOriginalVolume() bool`

GetUseOriginalVolume returns the UseOriginalVolume field if non-nil, zero value otherwise.

### GetUseOriginalVolumeOk

`func (o *Datavolume) GetUseOriginalVolumeOk() (*bool, bool)`

GetUseOriginalVolumeOk returns a tuple with the UseOriginalVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseOriginalVolume

`func (o *Datavolume) SetUseOriginalVolume(v bool)`

SetUseOriginalVolume sets UseOriginalVolume field to given value.

### HasUseOriginalVolume

`func (o *Datavolume) HasUseOriginalVolume() bool`

HasUseOriginalVolume returns a boolean if a field has been set.

### SetUseOriginalVolumeNil

`func (o *Datavolume) SetUseOriginalVolumeNil(b bool)`

 SetUseOriginalVolumeNil sets the value for UseOriginalVolume to be an explicit nil

### UnsetUseOriginalVolume
`func (o *Datavolume) UnsetUseOriginalVolume()`

UnsetUseOriginalVolume ensures that no value is present for UseOriginalVolume, not even an explicit nil
### GetProjectName

`func (o *Datavolume) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *Datavolume) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *Datavolume) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetDepartmentId

`func (o *Datavolume) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *Datavolume) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *Datavolume) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *Datavolume) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *Datavolume) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *Datavolume) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *Datavolume) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *Datavolume) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetSharedScopes

`func (o *Datavolume) GetSharedScopes() []SharedScope`

GetSharedScopes returns the SharedScopes field if non-nil, zero value otherwise.

### GetSharedScopesOk

`func (o *Datavolume) GetSharedScopesOk() (*[]SharedScope, bool)`

GetSharedScopesOk returns a tuple with the SharedScopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSharedScopes

`func (o *Datavolume) SetSharedScopes(v []SharedScope)`

SetSharedScopes sets SharedScopes field to given value.

### HasSharedScopes

`func (o *Datavolume) HasSharedScopes() bool`

HasSharedScopes returns a boolean if a field has been set.

### GetStatus

`func (o *Datavolume) GetStatus() DatavolumeStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *Datavolume) GetStatusOk() (*DatavolumeStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *Datavolume) SetStatus(v DatavolumeStatus)`

SetStatus sets Status field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


