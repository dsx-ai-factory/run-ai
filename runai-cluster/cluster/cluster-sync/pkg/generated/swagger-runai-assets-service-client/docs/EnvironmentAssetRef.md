# EnvironmentAssetRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | Unique identifier of the asset. | 
**Name** | **string** | The name of the asset. | 
**ToolTypes** | Pointer to [**[]ToolType**](ToolType.md) | the type of tools as defined in the connections of the environment asset. | [optional] 

## Methods

### NewEnvironmentAssetRef

`func NewEnvironmentAssetRef(id string, name string, ) *EnvironmentAssetRef`

NewEnvironmentAssetRef instantiates a new EnvironmentAssetRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvironmentAssetRefWithDefaults

`func NewEnvironmentAssetRefWithDefaults() *EnvironmentAssetRef`

NewEnvironmentAssetRefWithDefaults instantiates a new EnvironmentAssetRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *EnvironmentAssetRef) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *EnvironmentAssetRef) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *EnvironmentAssetRef) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *EnvironmentAssetRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EnvironmentAssetRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EnvironmentAssetRef) SetName(v string)`

SetName sets Name field to given value.


### GetToolTypes

`func (o *EnvironmentAssetRef) GetToolTypes() []ToolType`

GetToolTypes returns the ToolTypes field if non-nil, zero value otherwise.

### GetToolTypesOk

`func (o *EnvironmentAssetRef) GetToolTypesOk() (*[]ToolType, bool)`

GetToolTypesOk returns a tuple with the ToolTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolTypes

`func (o *EnvironmentAssetRef) SetToolTypes(v []ToolType)`

SetToolTypes sets ToolTypes field to given value.

### HasToolTypes

`func (o *EnvironmentAssetRef) HasToolTypes() bool`

HasToolTypes returns a boolean if a field has been set.

### SetToolTypesNil

`func (o *EnvironmentAssetRef) SetToolTypesNil(b bool)`

 SetToolTypesNil sets the value for ToolTypes to be an explicit nil

### UnsetToolTypes
`func (o *EnvironmentAssetRef) UnsetToolTypes()`

UnsetToolTypes ensures that no value is present for ToolTypes, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


