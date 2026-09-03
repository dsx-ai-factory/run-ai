# DatavolumeCreationFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Description** | Pointer to **string** |  | [optional] 
**OriginPvcName** | **string** | The name of the PVC that the datavolume is based on | 
**ProjectId** | **string** | The ID of the project that in its namespace the origin pvc is located | 
**ShouldDeleteOriginalVolume** | Pointer to **bool** | If true, the original storage volume will be deleted together with the datavolume | [optional] [default to false]

## Methods

### NewDatavolumeCreationFields

`func NewDatavolumeCreationFields(name string, originPvcName string, projectId string, ) *DatavolumeCreationFields`

NewDatavolumeCreationFields instantiates a new DatavolumeCreationFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatavolumeCreationFieldsWithDefaults

`func NewDatavolumeCreationFieldsWithDefaults() *DatavolumeCreationFields`

NewDatavolumeCreationFieldsWithDefaults instantiates a new DatavolumeCreationFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DatavolumeCreationFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DatavolumeCreationFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DatavolumeCreationFields) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *DatavolumeCreationFields) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DatavolumeCreationFields) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DatavolumeCreationFields) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DatavolumeCreationFields) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOriginPvcName

`func (o *DatavolumeCreationFields) GetOriginPvcName() string`

GetOriginPvcName returns the OriginPvcName field if non-nil, zero value otherwise.

### GetOriginPvcNameOk

`func (o *DatavolumeCreationFields) GetOriginPvcNameOk() (*string, bool)`

GetOriginPvcNameOk returns a tuple with the OriginPvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPvcName

`func (o *DatavolumeCreationFields) SetOriginPvcName(v string)`

SetOriginPvcName sets OriginPvcName field to given value.


### GetProjectId

`func (o *DatavolumeCreationFields) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *DatavolumeCreationFields) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *DatavolumeCreationFields) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetShouldDeleteOriginalVolume

`func (o *DatavolumeCreationFields) GetShouldDeleteOriginalVolume() bool`

GetShouldDeleteOriginalVolume returns the ShouldDeleteOriginalVolume field if non-nil, zero value otherwise.

### GetShouldDeleteOriginalVolumeOk

`func (o *DatavolumeCreationFields) GetShouldDeleteOriginalVolumeOk() (*bool, bool)`

GetShouldDeleteOriginalVolumeOk returns a tuple with the ShouldDeleteOriginalVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetShouldDeleteOriginalVolume

`func (o *DatavolumeCreationFields) SetShouldDeleteOriginalVolume(v bool)`

SetShouldDeleteOriginalVolume sets ShouldDeleteOriginalVolume field to given value.

### HasShouldDeleteOriginalVolume

`func (o *DatavolumeCreationFields) HasShouldDeleteOriginalVolume() bool`

HasShouldDeleteOriginalVolume returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


