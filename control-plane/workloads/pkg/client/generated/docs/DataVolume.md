# DataVolume

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **NullableString** | The unique identifier of the data volume. (mandatory) | [optional] 
**MountPath** | Pointer to **NullableString** | The path where the data volume will be mounted. (mandatory) | [optional] 

## Methods

### NewDataVolume

`func NewDataVolume() *DataVolume`

NewDataVolume instantiates a new DataVolume object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataVolumeWithDefaults

`func NewDataVolumeWithDefaults() *DataVolume`

NewDataVolumeWithDefaults instantiates a new DataVolume object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DataVolume) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DataVolume) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DataVolume) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DataVolume) HasId() bool`

HasId returns a boolean if a field has been set.

### SetIdNil

`func (o *DataVolume) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *DataVolume) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil
### GetMountPath

`func (o *DataVolume) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *DataVolume) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *DataVolume) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.

### HasMountPath

`func (o *DataVolume) HasMountPath() bool`

HasMountPath returns a boolean if a field has been set.

### SetMountPathNil

`func (o *DataVolume) SetMountPathNil(b bool)`

 SetMountPathNil sets the value for MountPath to be an explicit nil

### UnsetMountPath
`func (o *DataVolume) UnsetMountPath()`

UnsetMountPath ensures that no value is present for MountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


