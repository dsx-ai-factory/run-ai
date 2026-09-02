# ExternalToolInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ToolType** | [**ExternalToolType**](ExternalToolType.md) |  | 
**ExternalUrl** | **string** | The external url for connecting to the external tool. The url can include environment variables that will be replaced with the value provided for those variables when the workspace is created. | 

## Methods

### NewExternalToolInfo

`func NewExternalToolInfo(toolType ExternalToolType, externalUrl string, ) *ExternalToolInfo`

NewExternalToolInfo instantiates a new ExternalToolInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExternalToolInfoWithDefaults

`func NewExternalToolInfoWithDefaults() *ExternalToolInfo`

NewExternalToolInfoWithDefaults instantiates a new ExternalToolInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToolType

`func (o *ExternalToolInfo) GetToolType() ExternalToolType`

GetToolType returns the ToolType field if non-nil, zero value otherwise.

### GetToolTypeOk

`func (o *ExternalToolInfo) GetToolTypeOk() (*ExternalToolType, bool)`

GetToolTypeOk returns a tuple with the ToolType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToolType

`func (o *ExternalToolInfo) SetToolType(v ExternalToolType)`

SetToolType sets ToolType field to given value.


### GetExternalUrl

`func (o *ExternalToolInfo) GetExternalUrl() string`

GetExternalUrl returns the ExternalUrl field if non-nil, zero value otherwise.

### GetExternalUrlOk

`func (o *ExternalToolInfo) GetExternalUrlOk() (*string, bool)`

GetExternalUrlOk returns a tuple with the ExternalUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalUrl

`func (o *ExternalToolInfo) SetExternalUrl(v string)`

SetExternalUrl sets ExternalUrl field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


