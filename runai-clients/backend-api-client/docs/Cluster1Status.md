# Cluster1Status

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** | The cluster&#39;s state. | [optional] 
**LastConnected** | Pointer to **time.Time** | The last time the cluster was connected to the control plane. | [optional] 
**Conditions** | Pointer to [**[]ClusterReportedStatusConditionsInner**](ClusterReportedStatusConditionsInner.md) |  | [optional] 
**Operands** | Pointer to [**map[string]ClusterReportedStatusOperandsValue**](ClusterReportedStatusOperandsValue.md) | OperandStatuses specifies the status of the managed operands | [optional] 
**Dependencies** | Pointer to [**ClusterDependenciesStatus**](ClusterDependenciesStatus.md) |  | [optional] 

## Methods

### NewCluster1Status

`func NewCluster1Status() *Cluster1Status`

NewCluster1Status instantiates a new Cluster1Status object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCluster1StatusWithDefaults

`func NewCluster1StatusWithDefaults() *Cluster1Status`

NewCluster1StatusWithDefaults instantiates a new Cluster1Status object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *Cluster1Status) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *Cluster1Status) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *Cluster1Status) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *Cluster1Status) HasState() bool`

HasState returns a boolean if a field has been set.

### GetLastConnected

`func (o *Cluster1Status) GetLastConnected() time.Time`

GetLastConnected returns the LastConnected field if non-nil, zero value otherwise.

### GetLastConnectedOk

`func (o *Cluster1Status) GetLastConnectedOk() (*time.Time, bool)`

GetLastConnectedOk returns a tuple with the LastConnected field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastConnected

`func (o *Cluster1Status) SetLastConnected(v time.Time)`

SetLastConnected sets LastConnected field to given value.

### HasLastConnected

`func (o *Cluster1Status) HasLastConnected() bool`

HasLastConnected returns a boolean if a field has been set.

### GetConditions

`func (o *Cluster1Status) GetConditions() []ClusterReportedStatusConditionsInner`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *Cluster1Status) GetConditionsOk() (*[]ClusterReportedStatusConditionsInner, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *Cluster1Status) SetConditions(v []ClusterReportedStatusConditionsInner)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *Cluster1Status) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetOperands

`func (o *Cluster1Status) GetOperands() map[string]ClusterReportedStatusOperandsValue`

GetOperands returns the Operands field if non-nil, zero value otherwise.

### GetOperandsOk

`func (o *Cluster1Status) GetOperandsOk() (*map[string]ClusterReportedStatusOperandsValue, bool)`

GetOperandsOk returns a tuple with the Operands field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOperands

`func (o *Cluster1Status) SetOperands(v map[string]ClusterReportedStatusOperandsValue)`

SetOperands sets Operands field to given value.

### HasOperands

`func (o *Cluster1Status) HasOperands() bool`

HasOperands returns a boolean if a field has been set.

### GetDependencies

`func (o *Cluster1Status) GetDependencies() ClusterDependenciesStatus`

GetDependencies returns the Dependencies field if non-nil, zero value otherwise.

### GetDependenciesOk

`func (o *Cluster1Status) GetDependenciesOk() (*ClusterDependenciesStatus, bool)`

GetDependenciesOk returns a tuple with the Dependencies field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependencies

`func (o *Cluster1Status) SetDependencies(v ClusterDependenciesStatus)`

SetDependencies sets Dependencies field to given value.

### HasDependencies

`func (o *Cluster1Status) HasDependencies() bool`

HasDependencies returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


