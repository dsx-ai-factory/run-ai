# AIApplicationChartInspectionResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChartName** | **string** | Helm chart name. | 
**ChartVersion** | **string** | Helm chart version. | 
**ArchiveSHA256** | **string** | Lowercase hexadecimal SHA-256 digest of the exact inspected archive bytes. | 
**HasQuestions** | **bool** | Whether the chart contains a root questions.yaml or questions.yml file, including an unusable file. | 
**QuestionsStatus** | [**AIApplicationChartQuestionsStatus**](AIApplicationChartQuestionsStatus.md) |  | 
**Groups** | **[]map[string]interface{}** | Normalized group declarations in author order; empty for absent or unusable declarations. | 
**Questions** | **[]map[string]interface{}** | Normalized root questions in author order. Each question preserves supported declared metadata and may contain recursive subquestions through four child levels (five levels total), with roots at depth zero. Supported compatibility forms are normalized. Unknown question types become string with a warning; unsupported optional properties are ignored with a warning. Invalid mandatory metadata, bindings, options, defaults, conditions, or excessive nesting make the entire declaration unusable. Credential and Secret question types are not supported. This response contains no raw declaration YAML or complete values.yaml defaults.  | 
**Warnings** | [**[]AIApplicationChartInspectionWarning**](AIApplicationChartInspectionWarning.md) | Diagnostics for recoverable metadata and unusable declarations. | 

## Methods

### NewAIApplicationChartInspectionResponse

`func NewAIApplicationChartInspectionResponse(chartName string, chartVersion string, archiveSHA256 string, hasQuestions bool, questionsStatus AIApplicationChartQuestionsStatus, groups []map[string]interface{}, questions []map[string]interface{}, warnings []AIApplicationChartInspectionWarning, ) *AIApplicationChartInspectionResponse`

NewAIApplicationChartInspectionResponse instantiates a new AIApplicationChartInspectionResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationChartInspectionResponseWithDefaults

`func NewAIApplicationChartInspectionResponseWithDefaults() *AIApplicationChartInspectionResponse`

NewAIApplicationChartInspectionResponseWithDefaults instantiates a new AIApplicationChartInspectionResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChartName

`func (o *AIApplicationChartInspectionResponse) GetChartName() string`

GetChartName returns the ChartName field if non-nil, zero value otherwise.

### GetChartNameOk

`func (o *AIApplicationChartInspectionResponse) GetChartNameOk() (*string, bool)`

GetChartNameOk returns a tuple with the ChartName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartName

`func (o *AIApplicationChartInspectionResponse) SetChartName(v string)`

SetChartName sets ChartName field to given value.


### GetChartVersion

`func (o *AIApplicationChartInspectionResponse) GetChartVersion() string`

GetChartVersion returns the ChartVersion field if non-nil, zero value otherwise.

### GetChartVersionOk

`func (o *AIApplicationChartInspectionResponse) GetChartVersionOk() (*string, bool)`

GetChartVersionOk returns a tuple with the ChartVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartVersion

`func (o *AIApplicationChartInspectionResponse) SetChartVersion(v string)`

SetChartVersion sets ChartVersion field to given value.


### GetArchiveSHA256

`func (o *AIApplicationChartInspectionResponse) GetArchiveSHA256() string`

GetArchiveSHA256 returns the ArchiveSHA256 field if non-nil, zero value otherwise.

### GetArchiveSHA256Ok

`func (o *AIApplicationChartInspectionResponse) GetArchiveSHA256Ok() (*string, bool)`

GetArchiveSHA256Ok returns a tuple with the ArchiveSHA256 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArchiveSHA256

`func (o *AIApplicationChartInspectionResponse) SetArchiveSHA256(v string)`

SetArchiveSHA256 sets ArchiveSHA256 field to given value.


### GetHasQuestions

`func (o *AIApplicationChartInspectionResponse) GetHasQuestions() bool`

GetHasQuestions returns the HasQuestions field if non-nil, zero value otherwise.

### GetHasQuestionsOk

`func (o *AIApplicationChartInspectionResponse) GetHasQuestionsOk() (*bool, bool)`

GetHasQuestionsOk returns a tuple with the HasQuestions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasQuestions

`func (o *AIApplicationChartInspectionResponse) SetHasQuestions(v bool)`

SetHasQuestions sets HasQuestions field to given value.


### GetQuestionsStatus

`func (o *AIApplicationChartInspectionResponse) GetQuestionsStatus() AIApplicationChartQuestionsStatus`

GetQuestionsStatus returns the QuestionsStatus field if non-nil, zero value otherwise.

### GetQuestionsStatusOk

`func (o *AIApplicationChartInspectionResponse) GetQuestionsStatusOk() (*AIApplicationChartQuestionsStatus, bool)`

GetQuestionsStatusOk returns a tuple with the QuestionsStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuestionsStatus

`func (o *AIApplicationChartInspectionResponse) SetQuestionsStatus(v AIApplicationChartQuestionsStatus)`

SetQuestionsStatus sets QuestionsStatus field to given value.


### GetGroups

`func (o *AIApplicationChartInspectionResponse) GetGroups() []map[string]interface{}`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *AIApplicationChartInspectionResponse) GetGroupsOk() (*[]map[string]interface{}, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *AIApplicationChartInspectionResponse) SetGroups(v []map[string]interface{})`

SetGroups sets Groups field to given value.


### GetQuestions

`func (o *AIApplicationChartInspectionResponse) GetQuestions() []map[string]interface{}`

GetQuestions returns the Questions field if non-nil, zero value otherwise.

### GetQuestionsOk

`func (o *AIApplicationChartInspectionResponse) GetQuestionsOk() (*[]map[string]interface{}, bool)`

GetQuestionsOk returns a tuple with the Questions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetQuestions

`func (o *AIApplicationChartInspectionResponse) SetQuestions(v []map[string]interface{})`

SetQuestions sets Questions field to given value.


### GetWarnings

`func (o *AIApplicationChartInspectionResponse) GetWarnings() []AIApplicationChartInspectionWarning`

GetWarnings returns the Warnings field if non-nil, zero value otherwise.

### GetWarningsOk

`func (o *AIApplicationChartInspectionResponse) GetWarningsOk() (*[]AIApplicationChartInspectionWarning, bool)`

GetWarningsOk returns a tuple with the Warnings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWarnings

`func (o *AIApplicationChartInspectionResponse) SetWarnings(v []AIApplicationChartInspectionWarning)`

SetWarnings sets Warnings field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


