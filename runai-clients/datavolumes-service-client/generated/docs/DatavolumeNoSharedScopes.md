# DatavolumeNoSharedScopes

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
**Status** | [**DatavolumeStatus**](DatavolumeStatus.md) |  | 

## Methods

### NewDatavolumeNoSharedScopes

`func NewDatavolumeNoSharedScopes(id string, name string, originPvcName string, projectId string, projectName string, status DatavolumeStatus, ) *DatavolumeNoSharedScopes`

NewDatavolumeNoSharedScopes instantiates a new DatavolumeNoSharedScopes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeNoSharedScopesWithDefaults

`func NewDatavolumeNoSharedScopesWithDefaults() *DatavolumeNoSharedScopes`

NewDatavolumeNoSharedScopesWithDefaults instantiates a new DatavolumeNoSharedScopes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DatavolumeNoSharedScopes) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DatavolumeNoSharedScopes) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DatavolumeNoSharedScopes) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedAt

`func (o *DatavolumeNoSharedScopes) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DatavolumeNoSharedScopes) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DatavolumeNoSharedScopes) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DatavolumeNoSharedScopes) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *DatavolumeNoSharedScopes) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DatavolumeNoSharedScopes) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DatavolumeNoSharedScopes) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *DatavolumeNoSharedScopes) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetCreatedBy

`func (o *DatavolumeNoSharedScopes) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *DatavolumeNoSharedScopes) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *DatavolumeNoSharedScopes) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *DatavolumeNoSharedScopes) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetUpdatedBy

`func (o *DatavolumeNoSharedScopes) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *DatavolumeNoSharedScopes) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *DatavolumeNoSharedScopes) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.

### HasUpdatedBy

`func (o *DatavolumeNoSharedScopes) HasUpdatedBy() bool`

HasUpdatedBy returns a boolean if a field has been set.

### GetName

`func (o *DatavolumeNoSharedScopes) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DatavolumeNoSharedScopes) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DatavolumeNoSharedScopes) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *DatavolumeNoSharedScopes) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DatavolumeNoSharedScopes) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DatavolumeNoSharedScopes) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DatavolumeNoSharedScopes) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOriginPvcName

`func (o *DatavolumeNoSharedScopes) GetOriginPvcName() string`

GetOriginPvcName returns the OriginPvcName field if non-nil, zero value otherwise.

### GetOriginPvcNameOk

`func (o *DatavolumeNoSharedScopes) GetOriginPvcNameOk() (*string, bool)`

GetOriginPvcNameOk returns a tuple with the OriginPvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPvcName

`func (o *DatavolumeNoSharedScopes) SetOriginPvcName(v string)`

SetOriginPvcName sets OriginPvcName field to given value.


### GetProjectId

`func (o *DatavolumeNoSharedScopes) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *DatavolumeNoSharedScopes) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *DatavolumeNoSharedScopes) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetShouldDeleteOriginalVolume

`func (o *DatavolumeNoSharedScopes) GetShouldDeleteOriginalVolume() bool`

GetShouldDeleteOriginalVolume returns the ShouldDeleteOriginalVolume field if non-nil, zero value otherwise.

### GetShouldDeleteOriginalVolumeOk

`func (o *DatavolumeNoSharedScopes) GetShouldDeleteOriginalVolumeOk() (*bool, bool)`

GetShouldDeleteOriginalVolumeOk returns a tuple with the ShouldDeleteOriginalVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShouldDeleteOriginalVolume

`func (o *DatavolumeNoSharedScopes) SetShouldDeleteOriginalVolume(v bool)`

SetShouldDeleteOriginalVolume sets ShouldDeleteOriginalVolume field to given value.

### HasShouldDeleteOriginalVolume

`func (o *DatavolumeNoSharedScopes) HasShouldDeleteOriginalVolume() bool`

HasShouldDeleteOriginalVolume returns a boolean if a field has been set.

### GetUseOriginalVolume

`func (o *DatavolumeNoSharedScopes) GetUseOriginalVolume() bool`

GetUseOriginalVolume returns the UseOriginalVolume field if non-nil, zero value otherwise.

### GetUseOriginalVolumeOk

`func (o *DatavolumeNoSharedScopes) GetUseOriginalVolumeOk() (*bool, bool)`

GetUseOriginalVolumeOk returns a tuple with the UseOriginalVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseOriginalVolume

`func (o *DatavolumeNoSharedScopes) SetUseOriginalVolume(v bool)`

SetUseOriginalVolume sets UseOriginalVolume field to given value.

### HasUseOriginalVolume

`func (o *DatavolumeNoSharedScopes) HasUseOriginalVolume() bool`

HasUseOriginalVolume returns a boolean if a field has been set.

### SetUseOriginalVolumeNil

`func (o *DatavolumeNoSharedScopes) SetUseOriginalVolumeNil(b bool)`

 SetUseOriginalVolumeNil sets the value for UseOriginalVolume to be an explicit nil

### UnsetUseOriginalVolume
`func (o *DatavolumeNoSharedScopes) UnsetUseOriginalVolume()`

UnsetUseOriginalVolume ensures that no value is present for UseOriginalVolume, not even an explicit nil
### GetProjectName

`func (o *DatavolumeNoSharedScopes) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *DatavolumeNoSharedScopes) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *DatavolumeNoSharedScopes) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetDepartmentId

`func (o *DatavolumeNoSharedScopes) GetDepartmentId() string`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *DatavolumeNoSharedScopes) GetDepartmentIdOk() (*string, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *DatavolumeNoSharedScopes) SetDepartmentId(v string)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *DatavolumeNoSharedScopes) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *DatavolumeNoSharedScopes) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *DatavolumeNoSharedScopes) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *DatavolumeNoSharedScopes) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *DatavolumeNoSharedScopes) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetStatus

`func (o *DatavolumeNoSharedScopes) GetStatus() DatavolumeStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DatavolumeNoSharedScopes) GetStatusOk() (*DatavolumeStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DatavolumeNoSharedScopes) SetStatus(v DatavolumeStatus)`

SetStatus sets Status field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


