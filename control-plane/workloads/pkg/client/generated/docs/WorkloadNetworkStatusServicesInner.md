# WorkloadNetworkStatusServicesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Name of the service. | [optional] 
**Type** | Pointer to **string** | Type of the service (e.g., ClusterIP, NodePort, etc.). | [optional] 
**Ports** | Pointer to [**[]WorkloadNetworkStatusServicesInnerPortsInner**](WorkloadNetworkStatusServicesInnerPortsInner.md) | List of ports exposed by the service. | [optional] 
**LoadBalancer** | Pointer to [**WorkloadNetworkStatusServicesInnerLoadBalancer**](WorkloadNetworkStatusServicesInnerLoadBalancer.md) |  | [optional] 
**Urls** | Pointer to [**[]WorkloadNetworkStatusServicesInnerUrlsInner**](WorkloadNetworkStatusServicesInnerUrlsInner.md) | URLs reachable through this service. | [optional] 

## Methods

### NewWorkloadNetworkStatusServicesInner

`func NewWorkloadNetworkStatusServicesInner() *WorkloadNetworkStatusServicesInner`

NewWorkloadNetworkStatusServicesInner instantiates a new WorkloadNetworkStatusServicesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadNetworkStatusServicesInnerWithDefaults

`func NewWorkloadNetworkStatusServicesInnerWithDefaults() *WorkloadNetworkStatusServicesInner`

NewWorkloadNetworkStatusServicesInnerWithDefaults instantiates a new WorkloadNetworkStatusServicesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadNetworkStatusServicesInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadNetworkStatusServicesInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadNetworkStatusServicesInner) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkloadNetworkStatusServicesInner) HasName() bool`

HasName returns a boolean if a field has been set.

### GetType

`func (o *WorkloadNetworkStatusServicesInner) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *WorkloadNetworkStatusServicesInner) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *WorkloadNetworkStatusServicesInner) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *WorkloadNetworkStatusServicesInner) HasType() bool`

HasType returns a boolean if a field has been set.

### GetPorts

`func (o *WorkloadNetworkStatusServicesInner) GetPorts() []WorkloadNetworkStatusServicesInnerPortsInner`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *WorkloadNetworkStatusServicesInner) GetPortsOk() (*[]WorkloadNetworkStatusServicesInnerPortsInner, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *WorkloadNetworkStatusServicesInner) SetPorts(v []WorkloadNetworkStatusServicesInnerPortsInner)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *WorkloadNetworkStatusServicesInner) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### GetLoadBalancer

`func (o *WorkloadNetworkStatusServicesInner) GetLoadBalancer() WorkloadNetworkStatusServicesInnerLoadBalancer`

GetLoadBalancer returns the LoadBalancer field if non-nil, zero value otherwise.

### GetLoadBalancerOk

`func (o *WorkloadNetworkStatusServicesInner) GetLoadBalancerOk() (*WorkloadNetworkStatusServicesInnerLoadBalancer, bool)`

GetLoadBalancerOk returns a tuple with the LoadBalancer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLoadBalancer

`func (o *WorkloadNetworkStatusServicesInner) SetLoadBalancer(v WorkloadNetworkStatusServicesInnerLoadBalancer)`

SetLoadBalancer sets LoadBalancer field to given value.

### HasLoadBalancer

`func (o *WorkloadNetworkStatusServicesInner) HasLoadBalancer() bool`

HasLoadBalancer returns a boolean if a field has been set.

### GetUrls

`func (o *WorkloadNetworkStatusServicesInner) GetUrls() []WorkloadNetworkStatusServicesInnerUrlsInner`

GetUrls returns the Urls field if non-nil, zero value otherwise.

### GetUrlsOk

`func (o *WorkloadNetworkStatusServicesInner) GetUrlsOk() (*[]WorkloadNetworkStatusServicesInnerUrlsInner, bool)`

GetUrlsOk returns a tuple with the Urls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrls

`func (o *WorkloadNetworkStatusServicesInner) SetUrls(v []WorkloadNetworkStatusServicesInnerUrlsInner)`

SetUrls sets Urls field to given value.

### HasUrls

`func (o *WorkloadNetworkStatusServicesInner) HasUrls() bool`

HasUrls returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


