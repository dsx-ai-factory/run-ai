# VolumeModeCustomizationSupportedValues

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Filesystem** | Pointer to **NullableBool** | Indicates whether filesystem mode is supported. Filesystem allows the volume to be mounted as a filesystem, enabling the usage of directories and files. | [optional] [default to true]
**Block** | Pointer to **NullableBool** | Indicates whether block mode is supported. Exposes the volume as a block storage, which can be formatted or used by applications directly without a filesystem. | [optional] [default to true]

## Methods

### NewVolumeModeCustomizationSupportedValues

`func NewVolumeModeCustomizationSupportedValues() *VolumeModeCustomizationSupportedValues`

NewVolumeModeCustomizationSupportedValues instantiates a new VolumeModeCustomizationSupportedValues object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeModeCustomizationSupportedValuesWithDefaults

`func NewVolumeModeCustomizationSupportedValuesWithDefaults() *VolumeModeCustomizationSupportedValues`

NewVolumeModeCustomizationSupportedValuesWithDefaults instantiates a new VolumeModeCustomizationSupportedValues object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFilesystem

`func (o *VolumeModeCustomizationSupportedValues) GetFilesystem() bool`

GetFilesystem returns the Filesystem field if non-nil, zero value otherwise.

### GetFilesystemOk

`func (o *VolumeModeCustomizationSupportedValues) GetFilesystemOk() (*bool, bool)`

GetFilesystemOk returns a tuple with the Filesystem field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilesystem

`func (o *VolumeModeCustomizationSupportedValues) SetFilesystem(v bool)`

SetFilesystem sets Filesystem field to given value.

### HasFilesystem

`func (o *VolumeModeCustomizationSupportedValues) HasFilesystem() bool`

HasFilesystem returns a boolean if a field has been set.

### SetFilesystemNil

`func (o *VolumeModeCustomizationSupportedValues) SetFilesystemNil(b bool)`

 SetFilesystemNil sets the value for Filesystem to be an explicit nil

### UnsetFilesystem
`func (o *VolumeModeCustomizationSupportedValues) UnsetFilesystem()`

UnsetFilesystem ensures that no value is present for Filesystem, not even an explicit nil
### GetBlock

`func (o *VolumeModeCustomizationSupportedValues) GetBlock() bool`

GetBlock returns the Block field if non-nil, zero value otherwise.

### GetBlockOk

`func (o *VolumeModeCustomizationSupportedValues) GetBlockOk() (*bool, bool)`

GetBlockOk returns a tuple with the Block field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlock

`func (o *VolumeModeCustomizationSupportedValues) SetBlock(v bool)`

SetBlock sets Block field to given value.

### HasBlock

`func (o *VolumeModeCustomizationSupportedValues) HasBlock() bool`

HasBlock returns a boolean if a field has been set.

### SetBlockNil

`func (o *VolumeModeCustomizationSupportedValues) SetBlockNil(b bool)`

 SetBlockNil sets the value for Block to be an explicit nil

### UnsetBlock
`func (o *VolumeModeCustomizationSupportedValues) UnsetBlock()`

UnsetBlock ensures that no value is present for Block, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


