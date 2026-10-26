# TemplateComplianceInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ComplianceIssues** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | A list of issues preventing the use of a template or asset for creating workloads in a given project. | [optional] 

## Methods

### NewTemplateComplianceInfo

`func NewTemplateComplianceInfo() *TemplateComplianceInfo`

NewTemplateComplianceInfo instantiates a new TemplateComplianceInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTemplateComplianceInfoWithDefaults

`func NewTemplateComplianceInfoWithDefaults() *TemplateComplianceInfo`

NewTemplateComplianceInfoWithDefaults instantiates a new TemplateComplianceInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetComplianceIssues

`func (o *TemplateComplianceInfo) GetComplianceIssues() []ComplianceInfoReason`

GetComplianceIssues returns the ComplianceIssues field if non-nil, zero value otherwise.

### GetComplianceIssuesOk

`func (o *TemplateComplianceInfo) GetComplianceIssuesOk() (*[]ComplianceInfoReason, bool)`

GetComplianceIssuesOk returns a tuple with the ComplianceIssues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComplianceIssues

`func (o *TemplateComplianceInfo) SetComplianceIssues(v []ComplianceInfoReason)`

SetComplianceIssues sets ComplianceIssues field to given value.

### HasComplianceIssues

`func (o *TemplateComplianceInfo) HasComplianceIssues() bool`

HasComplianceIssues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


