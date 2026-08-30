# ClusterReportedStatus1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**[]ClusterReportedStatus1ConditionsInner**](ClusterReportedStatus1ConditionsInner.md) |  | [optional] 
**Operands** | Pointer to [**map[string]ClusterReportedStatusOperandsValue**](ClusterReportedStatusOperandsValue.md) | OperandStatuses specifies the status of the managed operands | [optional] 
**Dependencies** | Pointer to [**ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] 

## Methods

### NewClusterReportedStatus1

`func NewClusterReportedStatus1() *ClusterReportedStatus1`

NewClusterReportedStatus1 instantiates a new ClusterReportedStatus1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterReportedStatus1WithDefaults

`func NewClusterReportedStatus1WithDefaults() *ClusterReportedStatus1`

NewClusterReportedStatus1WithDefaults instantiates a new ClusterReportedStatus1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConditions

`func (o *ClusterReportedStatus1) GetConditions() []ClusterReportedStatus1ConditionsInner`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ClusterReportedStatus1) GetConditionsOk() (*[]ClusterReportedStatus1ConditionsInner, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ClusterReportedStatus1) SetConditions(v []ClusterReportedStatus1ConditionsInner)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ClusterReportedStatus1) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetOperands

`func (o *ClusterReportedStatus1) GetOperands() map[string]ClusterReportedStatusOperandsValue`

GetOperands returns the Operands field if non-nil, zero value otherwise.

### GetOperandsOk

`func (o *ClusterReportedStatus1) GetOperandsOk() (*map[string]ClusterReportedStatusOperandsValue, bool)`

GetOperandsOk returns a tuple with the Operands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperands

`func (o *ClusterReportedStatus1) SetOperands(v map[string]ClusterReportedStatusOperandsValue)`

SetOperands sets Operands field to given value.

### HasOperands

`func (o *ClusterReportedStatus1) HasOperands() bool`

HasOperands returns a boolean if a field has been set.

### GetDependencies

`func (o *ClusterReportedStatus1) GetDependencies() ClusterDependenciesStatus`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *ClusterReportedStatus1) GetDependenciesOk() (*ClusterDependenciesStatus, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *ClusterReportedStatus1) SetDependencies(v ClusterDependenciesStatus)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *ClusterReportedStatus1) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


