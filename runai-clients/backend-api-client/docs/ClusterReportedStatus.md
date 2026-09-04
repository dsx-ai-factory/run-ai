# ClusterReportedStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**[]ClusterReportedStatusConditionsInner**](ClusterReportedStatusConditionsInner.md) |  | [optional] 
**Operands** | Pointer to [**map[string]ClusterReportedStatusOperandsValue**](ClusterReportedStatusOperandsValue.md) | OperandStatuses specifies the status of the managed operands | [optional] 
**Dependencies** | Pointer to [**ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] 

## Methods

### NewClusterReportedStatus

`func NewClusterReportedStatus() *ClusterReportedStatus`

NewClusterReportedStatus instantiates a new ClusterReportedStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedStatusWithDefaults

`func NewClusterReportedStatusWithDefaults() *ClusterReportedStatus`

NewClusterReportedStatusWithDefaults instantiates a new ClusterReportedStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConditions

`func (o *ClusterReportedStatus) GetConditions() []ClusterReportedStatusConditionsInner`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ClusterReportedStatus) GetConditionsOk() (*[]ClusterReportedStatusConditionsInner, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ClusterReportedStatus) SetConditions(v []ClusterReportedStatusConditionsInner)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ClusterReportedStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetOperands

`func (o *ClusterReportedStatus) GetOperands() map[string]ClusterReportedStatusOperandsValue`

GetOperands returns the Operands field if non-nil, zero value otherwise.

### GetOperandsOk

`func (o *ClusterReportedStatus) GetOperandsOk() (*map[string]ClusterReportedStatusOperandsValue, bool)`

GetOperandsOk returns a tuple with the Operands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperands

`func (o *ClusterReportedStatus) SetOperands(v map[string]ClusterReportedStatusOperandsValue)`

SetOperands sets Operands field to given value.

### HasOperands

`func (o *ClusterReportedStatus) HasOperands() bool`

HasOperands returns a boolean if a field has been set.

### GetDependencies

`func (o *ClusterReportedStatus) GetDependencies() ClusterDependenciesStatus`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *ClusterReportedStatus) GetDependenciesOk() (*ClusterDependenciesStatus, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *ClusterReportedStatus) SetDependencies(v ClusterDependenciesStatus)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *ClusterReportedStatus) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


