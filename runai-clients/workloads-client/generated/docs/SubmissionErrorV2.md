# SubmissionErrorV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | **int32** |  | 
**Message** | **string** |  | 
**Details** | Pointer to **string** |  | [optional] 
**ComplianceIssues** | Pointer to [**[]ComplianceIssuesV2ComplianceIssuesInner**](ComplianceIssuesV2ComplianceIssuesInner.md) |  | [optional] 

## Methods

### NewSubmissionErrorV2

`func NewSubmissionErrorV2(code int32, message string, ) *SubmissionErrorV2`

NewSubmissionErrorV2 instantiates a new SubmissionErrorV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSubmissionErrorV2WithDefaults

`func NewSubmissionErrorV2WithDefaults() *SubmissionErrorV2`

NewSubmissionErrorV2WithDefaults instantiates a new SubmissionErrorV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *SubmissionErrorV2) GetCode() int32`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *SubmissionErrorV2) GetCodeOk() (*int32, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *SubmissionErrorV2) SetCode(v int32)`

SetCode sets Code field to given value.


### GetMessage

`func (o *SubmissionErrorV2) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *SubmissionErrorV2) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *SubmissionErrorV2) SetMessage(v string)`

SetMessage sets Message field to given value.


### GetDetails

`func (o *SubmissionErrorV2) GetDetails() string`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *SubmissionErrorV2) GetDetailsOk() (*string, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *SubmissionErrorV2) SetDetails(v string)`

SetDetails sets Details field to given value.

### HasDetails

`func (o *SubmissionErrorV2) HasDetails() bool`

HasDetails returns a boolean if a field has been set.

### GetComplianceIssues

`func (o *SubmissionErrorV2) GetComplianceIssues() []ComplianceIssuesV2ComplianceIssuesInner`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *SubmissionErrorV2) GetComplianceIssuesOk() (*[]ComplianceIssuesV2ComplianceIssuesInner, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *SubmissionErrorV2) SetComplianceIssues(v []ComplianceIssuesV2ComplianceIssuesInner)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *SubmissionErrorV2) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


