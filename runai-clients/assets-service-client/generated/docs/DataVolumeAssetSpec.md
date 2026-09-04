# DataVolumeAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **NullableString** | The unique identifier of the data volume. (mandatory) | [optional] 
**OriginPvcName** | Pointer to **NullableString** | The name of the PVC that the datavolume is based on | [optional] 

## Methods

### NewDataVolumeAssetSpec

`func NewDataVolumeAssetSpec() *DataVolumeAssetSpec`

NewDataVolumeAssetSpec instantiates a new DataVolumeAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataVolumeAssetSpecWithDefaults

`func NewDataVolumeAssetSpecWithDefaults() *DataVolumeAssetSpec`

NewDataVolumeAssetSpecWithDefaults instantiates a new DataVolumeAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *DataVolumeAssetSpec) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DataVolumeAssetSpec) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DataVolumeAssetSpec) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DataVolumeAssetSpec) HasId() bool`

HasId returns a boolean if a field has been set.

### SetIdNil

`func (o *DataVolumeAssetSpec) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *DataVolumeAssetSpec) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil
### GetOriginPvcName

`func (o *DataVolumeAssetSpec) GetOriginPvcName() string`

GetOriginPvcName returns the OriginPvcName field if non-nil, zero value otherwise.

### GetOriginPvcNameOk

`func (o *DataVolumeAssetSpec) GetOriginPvcNameOk() (*string, bool)`

GetOriginPvcNameOk returns a tuple with the OriginPvcName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOriginPvcName

`func (o *DataVolumeAssetSpec) SetOriginPvcName(v string)`

SetOriginPvcName sets OriginPvcName field to given value.

### HasOriginPvcName

`func (o *DataVolumeAssetSpec) HasOriginPvcName() bool`

HasOriginPvcName returns a boolean if a field has been set.

### SetOriginPvcNameNil

`func (o *DataVolumeAssetSpec) SetOriginPvcNameNil(b bool)`

 SetOriginPvcNameNil sets the value for OriginPvcName to be an explicit nil

### UnsetOriginPvcName
`func (o *DataVolumeAssetSpec) UnsetOriginPvcName()`

UnsetOriginPvcName ensures that no value is present for OriginPvcName, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


