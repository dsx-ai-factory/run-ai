# ClusterStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** | The cluster&#39;s state. | [optional] 
**LastConnected** | Pointer to **time.Time** | The last time the cluster was connected to the control plane. | [optional] 
**Conditions** | Pointer to [**[]ClusterReportedStatus1ConditionsInner**](ClusterReportedStatus1ConditionsInner.md) |  | [optional] 
**Operands** | Pointer to [**map[string]ClusterReportedStatusOperandsValue**](ClusterReportedStatusOperandsValue.md) | OperandStatuses specifies the status of the managed operands | [optional] 
**Dependencies** | Pointer to [**ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] 

## Methods

### NewClusterStatus

`func NewClusterStatus() *ClusterStatus`

NewClusterStatus instantiates a new ClusterStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterStatusWithDefaults

`func NewClusterStatusWithDefaults() *ClusterStatus`

NewClusterStatusWithDefaults instantiates a new ClusterStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *ClusterStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ClusterStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ClusterStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ClusterStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetLastConnected

`func (o *ClusterStatus) GetLastConnected() time.Time`

GetLastConnected returns the LastConnected field if non-nil, zero value otherwise.

### GetLastConnectedOk

`func (o *ClusterStatus) GetLastConnectedOk() (*time.Time, bool)`

GetLastConnectedOk returns a tuple with the LastConnected field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastConnected

`func (o *ClusterStatus) SetLastConnected(v time.Time)`

SetLastConnected sets LastConnected field to given value.

### HasLastConnected

`func (o *ClusterStatus) HasLastConnected() bool`

HasLastConnected returns a boolean if a field has been set.

### GetConditions

`func (o *ClusterStatus) GetConditions() []ClusterReportedStatus1ConditionsInner`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ClusterStatus) GetConditionsOk() (*[]ClusterReportedStatus1ConditionsInner, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ClusterStatus) SetConditions(v []ClusterReportedStatus1ConditionsInner)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ClusterStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetOperands

`func (o *ClusterStatus) GetOperands() map[string]ClusterReportedStatusOperandsValue`

GetOperands returns the Operands field if non-nil, zero value otherwise.

### GetOperandsOk

`func (o *ClusterStatus) GetOperandsOk() (*map[string]ClusterReportedStatusOperandsValue, bool)`

GetOperandsOk returns a tuple with the Operands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperands

`func (o *ClusterStatus) SetOperands(v map[string]ClusterReportedStatusOperandsValue)`

SetOperands sets Operands field to given value.

### HasOperands

`func (o *ClusterStatus) HasOperands() bool`

HasOperands returns a boolean if a field has been set.

### GetDependencies

`func (o *ClusterStatus) GetDependencies() ClusterDependenciesStatus`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *ClusterStatus) GetDependenciesOk() (*ClusterDependenciesStatus, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *ClusterStatus) SetDependencies(v ClusterDependenciesStatus)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *ClusterStatus) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


