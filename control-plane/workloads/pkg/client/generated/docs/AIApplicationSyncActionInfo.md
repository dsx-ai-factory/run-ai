# AIApplicationSyncActionInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to **string** |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 
**Started** | Pointer to **string** |  | [optional] 
**Completed** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Error** | Pointer to **string** |  | [optional] 
**HelmError** | Pointer to **string** |  | [optional] 
**HelmPhase** | Pointer to **string** |  | [optional] 
**ChartUrl** | Pointer to **string** |  | [optional] 
**ReleaseVersion** | Pointer to **int32** |  | [optional] 

## Methods

### NewAIApplicationSyncActionInfo

`func NewAIApplicationSyncActionInfo() *AIApplicationSyncActionInfo`

NewAIApplicationSyncActionInfo instantiates a new AIApplicationSyncActionInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAIApplicationSyncActionInfoWithDefaults

`func NewAIApplicationSyncActionInfoWithDefaults() *AIApplicationSyncActionInfo`

NewAIApplicationSyncActionInfoWithDefaults instantiates a new AIApplicationSyncActionInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *AIApplicationSyncActionInfo) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *AIApplicationSyncActionInfo) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *AIApplicationSyncActionInfo) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *AIApplicationSyncActionInfo) HasType() bool`

HasType returns a boolean if a field has been set.

### GetStatus

`func (o *AIApplicationSyncActionInfo) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *AIApplicationSyncActionInfo) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *AIApplicationSyncActionInfo) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *AIApplicationSyncActionInfo) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetStarted

`func (o *AIApplicationSyncActionInfo) GetStarted() string`

GetStarted returns the Started field if non-nil, zero value otherwise.

### GetStartedOk

`func (o *AIApplicationSyncActionInfo) GetStartedOk() (*string, bool)`

GetStartedOk returns a tuple with the Started field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStarted

`func (o *AIApplicationSyncActionInfo) SetStarted(v string)`

SetStarted sets Started field to given value.

### HasStarted

`func (o *AIApplicationSyncActionInfo) HasStarted() bool`

HasStarted returns a boolean if a field has been set.

### GetCompleted

`func (o *AIApplicationSyncActionInfo) GetCompleted() string`

GetCompleted returns the Completed field if non-nil, zero value otherwise.

### GetCompletedOk

`func (o *AIApplicationSyncActionInfo) GetCompletedOk() (*string, bool)`

GetCompletedOk returns a tuple with the Completed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompleted

`func (o *AIApplicationSyncActionInfo) SetCompleted(v string)`

SetCompleted sets Completed field to given value.

### HasCompleted

`func (o *AIApplicationSyncActionInfo) HasCompleted() bool`

HasCompleted returns a boolean if a field has been set.

### GetMessage

`func (o *AIApplicationSyncActionInfo) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *AIApplicationSyncActionInfo) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *AIApplicationSyncActionInfo) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *AIApplicationSyncActionInfo) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetError

`func (o *AIApplicationSyncActionInfo) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *AIApplicationSyncActionInfo) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *AIApplicationSyncActionInfo) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *AIApplicationSyncActionInfo) HasError() bool`

HasError returns a boolean if a field has been set.

### GetHelmError

`func (o *AIApplicationSyncActionInfo) GetHelmError() string`

GetHelmError returns the HelmError field if non-nil, zero value otherwise.

### GetHelmErrorOk

`func (o *AIApplicationSyncActionInfo) GetHelmErrorOk() (*string, bool)`

GetHelmErrorOk returns a tuple with the HelmError field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHelmError

`func (o *AIApplicationSyncActionInfo) SetHelmError(v string)`

SetHelmError sets HelmError field to given value.

### HasHelmError

`func (o *AIApplicationSyncActionInfo) HasHelmError() bool`

HasHelmError returns a boolean if a field has been set.

### GetHelmPhase

`func (o *AIApplicationSyncActionInfo) GetHelmPhase() string`

GetHelmPhase returns the HelmPhase field if non-nil, zero value otherwise.

### GetHelmPhaseOk

`func (o *AIApplicationSyncActionInfo) GetHelmPhaseOk() (*string, bool)`

GetHelmPhaseOk returns a tuple with the HelmPhase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHelmPhase

`func (o *AIApplicationSyncActionInfo) SetHelmPhase(v string)`

SetHelmPhase sets HelmPhase field to given value.

### HasHelmPhase

`func (o *AIApplicationSyncActionInfo) HasHelmPhase() bool`

HasHelmPhase returns a boolean if a field has been set.

### GetChartUrl

`func (o *AIApplicationSyncActionInfo) GetChartUrl() string`

GetChartUrl returns the ChartUrl field if non-nil, zero value otherwise.

### GetChartUrlOk

`func (o *AIApplicationSyncActionInfo) GetChartUrlOk() (*string, bool)`

GetChartUrlOk returns a tuple with the ChartUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChartUrl

`func (o *AIApplicationSyncActionInfo) SetChartUrl(v string)`

SetChartUrl sets ChartUrl field to given value.

### HasChartUrl

`func (o *AIApplicationSyncActionInfo) HasChartUrl() bool`

HasChartUrl returns a boolean if a field has been set.

### GetReleaseVersion

`func (o *AIApplicationSyncActionInfo) GetReleaseVersion() int32`

GetReleaseVersion returns the ReleaseVersion field if non-nil, zero value otherwise.

### GetReleaseVersionOk

`func (o *AIApplicationSyncActionInfo) GetReleaseVersionOk() (*int32, bool)`

GetReleaseVersionOk returns a tuple with the ReleaseVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleaseVersion

`func (o *AIApplicationSyncActionInfo) SetReleaseVersion(v int32)`

SetReleaseVersion sets ReleaseVersion field to given value.

### HasReleaseVersion

`func (o *AIApplicationSyncActionInfo) HasReleaseVersion() bool`

HasReleaseVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


