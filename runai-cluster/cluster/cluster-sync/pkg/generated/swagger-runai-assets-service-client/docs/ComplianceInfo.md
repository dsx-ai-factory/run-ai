# ComplianceInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Imposed** | Pointer to **bool** | Whether the asset is imposed on created workloads by the administrator. | [optional] 
**Compliance** | Pointer to **bool** | Whether the asset info complies with creation of workloads in the project. | [optional] 
**Reason** | Pointer to [**[]ComplianceInfoReason**](ComplianceInfoReason.md) | For asset that does not comply, specify one or more reason why. | [optional] 

## Methods

### NewComplianceInfo

`func NewComplianceInfo() *ComplianceInfo`

NewComplianceInfo instantiates a new ComplianceInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewComplianceInfoWithDefaults

`func NewComplianceInfoWithDefaults() *ComplianceInfo`

NewComplianceInfoWithDefaults instantiates a new ComplianceInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetImposed

`func (o *ComplianceInfo) GetImposed() bool`

GetImposed returns the Imposed field if non-nil, zero value otherwise.

### GetImposedOk

`func (o *ComplianceInfo) GetImposedOk() (*bool, bool)`

GetImposedOk returns a tuple with the Imposed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImposed

`func (o *ComplianceInfo) SetImposed(v bool)`

SetImposed sets Imposed field to given value.

### HasImposed

`func (o *ComplianceInfo) HasImposed() bool`

HasImposed returns a boolean if a field has been set.

### GetCompliance

`func (o *ComplianceInfo) GetCompliance() bool`

GetCompliance returns the Compliance field if non-nil, zero value otherwise.

### GetComplianceOk

`func (o *ComplianceInfo) GetComplianceOk() (*bool, bool)`

GetComplianceOk returns a tuple with the Compliance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompliance

`func (o *ComplianceInfo) SetCompliance(v bool)`

SetCompliance sets Compliance field to given value.

### HasCompliance

`func (o *ComplianceInfo) HasCompliance() bool`

HasCompliance returns a boolean if a field has been set.

### GetReason

`func (o *ComplianceInfo) GetReason() []ComplianceInfoReason`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *ComplianceInfo) GetReasonOk() (*[]ComplianceInfoReason, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *ComplianceInfo) SetReason(v []ComplianceInfoReason)`

SetReason sets Reason field to given value.

### HasReason

`func (o *ComplianceInfo) HasReason() bool`

HasReason returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


