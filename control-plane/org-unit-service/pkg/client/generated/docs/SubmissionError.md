# SubmissionError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | **int32** |  | 
**Message** | **string** |  | 
**Details** | Pointer to **string** |  | [optional] 
**ComplianceIssues** | Pointer to [**[]ComplianceIssuesComplianceIssuesInner**](ComplianceIssuesComplianceIssuesInner.md) |  | [optional] 

## Methods

### NewSubmissionError

`func NewSubmissionError(code int32, message string, ) *SubmissionError`

NewSubmissionError instantiates a new SubmissionError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmissionErrorWithDefaults

`func NewSubmissionErrorWithDefaults() *SubmissionError`

NewSubmissionErrorWithDefaults instantiates a new SubmissionError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *SubmissionError) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *SubmissionError) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *SubmissionError) SetCode(v int32)`

SetCode sets Code field to given value.


### GetMessage

`func (o *SubmissionError) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *SubmissionError) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *SubmissionError) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetDetails

`func (o *SubmissionError) GetDetails() string`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *SubmissionError) GetDetailsOk() (*string, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *SubmissionError) SetDetails(v string)`

SetDetails sets Details field to given value.

### HasDetails

`func (o *SubmissionError) HasDetails() bool`

HasDetails returns a boolean if a field has been set.

### GetComplianceIssues

`func (o *SubmissionError) GetComplianceIssues() []ComplianceIssuesComplianceIssuesInner`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *SubmissionError) GetComplianceIssuesOk() (*[]ComplianceIssuesComplianceIssuesInner, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *SubmissionError) SetComplianceIssues(v []ComplianceIssuesComplianceIssuesInner)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *SubmissionError) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


