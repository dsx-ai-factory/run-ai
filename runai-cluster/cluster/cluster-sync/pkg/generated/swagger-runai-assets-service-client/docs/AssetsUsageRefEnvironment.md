# AssetsUsageRefEnvironment

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the asset. | 
**Name** | **string** | The name of the asset. | 
**ToolTypes** | Pointer to [**[]ToolType**](ToolType.md) | the type of tools as defined in the connections of the environment asset. | [optional] 

## Methods

### NewAssetsUsageRefEnvironment

`func NewAssetsUsageRefEnvironment(id string, name string, ) *AssetsUsageRefEnvironment`

NewAssetsUsageRefEnvironment instantiates a new AssetsUsageRefEnvironment object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsUsageRefEnvironmentWithDefaults

`func NewAssetsUsageRefEnvironmentWithDefaults() *AssetsUsageRefEnvironment`

NewAssetsUsageRefEnvironmentWithDefaults instantiates a new AssetsUsageRefEnvironment object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AssetsUsageRefEnvironment) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AssetsUsageRefEnvironment) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AssetsUsageRefEnvironment) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *AssetsUsageRefEnvironment) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AssetsUsageRefEnvironment) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AssetsUsageRefEnvironment) SetName(v string)`

SetName sets Name field to given value.


### GetToolTypes

`func (o *AssetsUsageRefEnvironment) GetToolTypes() []ToolType`

GetToolTypes returns the ToolTypes field if non-nil, zero value otherwise.

### GetToolTypesOk

`func (o *AssetsUsageRefEnvironment) GetToolTypesOk() (*[]ToolType, bool)`

GetToolTypesOk returns a tuple with the ToolTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolTypes

`func (o *AssetsUsageRefEnvironment) SetToolTypes(v []ToolType)`

SetToolTypes sets ToolTypes field to given value.

### HasToolTypes

`func (o *AssetsUsageRefEnvironment) HasToolTypes() bool`

HasToolTypes returns a boolean if a field has been set.

### SetToolTypesNil

`func (o *AssetsUsageRefEnvironment) SetToolTypesNil(b bool)`

 SetToolTypesNil sets the value for ToolTypes to be an explicit nil

### UnsetToolTypes
`func (o *AssetsUsageRefEnvironment) UnsetToolTypes()`

UnsetToolTypes ensures that no value is present for ToolTypes, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


