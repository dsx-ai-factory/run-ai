# ComplianceInfoReason

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Field** | Pointer to **NullableString** | For policy related issues, includes a json path to the field in the policy for which the issue applies. | [optional] 
**Details** | **string** | Explanation of the issue. | 
**Rule** | Pointer to [**NullablePolicyRuleEnum**](PolicyRuleEnum.md) |  | [optional] 

## Methods

### NewComplianceInfoReason

`func NewComplianceInfoReason(details string, ) *ComplianceInfoReason`

NewComplianceInfoReason instantiates a new ComplianceInfoReason object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComplianceInfoReasonWithDefaults

`func NewComplianceInfoReasonWithDefaults() *ComplianceInfoReason`

NewComplianceInfoReasonWithDefaults instantiates a new ComplianceInfoReason object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetField

`func (o *ComplianceInfoReason) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *ComplianceInfoReason) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *ComplianceInfoReason) SetField(v string)`

SetField sets Field field to given value.

### HasField

`func (o *ComplianceInfoReason) HasField() bool`

HasField returns a boolean if a field has been set.

### SetFieldNil

`func (o *ComplianceInfoReason) SetFieldNil(b bool)`

 SetFieldNil sets the value for Field to be an explicit nil

### UnsetField
`func (o *ComplianceInfoReason) UnsetField()`

UnsetField ensures that no value is present for Field, not even an explicit nil
### GetDetails

`func (o *ComplianceInfoReason) GetDetails() string`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *ComplianceInfoReason) GetDetailsOk() (*string, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *ComplianceInfoReason) SetDetails(v string)`

SetDetails sets Details field to given value.


### GetRule

`func (o *ComplianceInfoReason) GetRule() PolicyRuleEnum`

GetRule returns the Rule field if non-nil, zero value otherwise.

### GetRuleOk

`func (o *ComplianceInfoReason) GetRuleOk() (*PolicyRuleEnum, bool)`

GetRuleOk returns a tuple with the Rule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRule

`func (o *ComplianceInfoReason) SetRule(v PolicyRuleEnum)`

SetRule sets Rule field to given value.

### HasRule

`func (o *ComplianceInfoReason) HasRule() bool`

HasRule returns a boolean if a field has been set.

### SetRuleNil

`func (o *ComplianceInfoReason) SetRuleNil(b bool)`

 SetRuleNil sets the value for Rule to be an explicit nil

### UnsetRule
`func (o *ComplianceInfoReason) UnsetRule()`

UnsetRule ensures that no value is present for Rule, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


