# DataVolumeInstance

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **NullableString** | The unique identifier of the data volume. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | The path where the data volume will be mounted. (mandatory) | [optional] 

## Methods

### NewDataVolumeInstance

`func NewDataVolumeInstance() *DataVolumeInstance`

NewDataVolumeInstance instantiates a new DataVolumeInstance object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataVolumeInstanceWithDefaults

`func NewDataVolumeInstanceWithDefaults() *DataVolumeInstance`

NewDataVolumeInstanceWithDefaults instantiates a new DataVolumeInstance object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DataVolumeInstance) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DataVolumeInstance) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DataVolumeInstance) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DataVolumeInstance) HasId() bool`

HasId returns a boolean if a field has been set.

### SetIdNil

`func (o *DataVolumeInstance) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *DataVolumeInstance) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil
### GetMountPath

`func (o *DataVolumeInstance) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *DataVolumeInstance) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *DataVolumeInstance) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *DataVolumeInstance) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *DataVolumeInstance) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *DataVolumeInstance) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


