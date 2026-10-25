# AIApplicationChartInspectionWarning

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | **string** | Stable machine-readable diagnostic identifier. | 
**Message** | **string** | Explanation of the problem and available recovery. | 
**Path** | **string** | Location within the declaration, or an empty string for a document-wide diagnostic. | 

## Methods

### NewAIApplicationChartInspectionWarning

`func NewAIApplicationChartInspectionWarning(code string, message string, path string, ) *AIApplicationChartInspectionWarning`

NewAIApplicationChartInspectionWarning instantiates a new AIApplicationChartInspectionWarning object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationChartInspectionWarningWithDefaults

`func NewAIApplicationChartInspectionWarningWithDefaults() *AIApplicationChartInspectionWarning`

NewAIApplicationChartInspectionWarningWithDefaults instantiates a new AIApplicationChartInspectionWarning object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *AIApplicationChartInspectionWarning) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AIApplicationChartInspectionWarning) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AIApplicationChartInspectionWarning) SetCode(v string)`

SetCode sets Code field to given value.


### GetMessage

`func (o *AIApplicationChartInspectionWarning) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AIApplicationChartInspectionWarning) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AIApplicationChartInspectionWarning) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetPath

`func (o *AIApplicationChartInspectionWarning) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *AIApplicationChartInspectionWarning) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *AIApplicationChartInspectionWarning) SetPath(v string)`

SetPath sets Path field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


