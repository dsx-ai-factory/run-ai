# ServiceSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The Kubernetes Service name. | 
**Type** | Pointer to [**NullableServiceType**](ServiceType.md) |  | [optional] [default to SERVICETYPE_CLUSTER_IP]
**Selector** | **map[string]string** | Label selector to identify the pods that this service routes traffic to. | 
**Labels** | Pointer to **map[string]string** | Labels to apply to the Kubernetes Service. | [optional] 
**Ports** | [**[]ServicePort**](ServicePort.md) | Ports exposed by the service. | 

## Methods

### NewServiceSpec

`func NewServiceSpec(name string, selector map[string]string, ports []ServicePort, ) *ServiceSpec`

NewServiceSpec instantiates a new ServiceSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServiceSpecWithDefaults

`func NewServiceSpecWithDefaults() *ServiceSpec`

NewServiceSpecWithDefaults instantiates a new ServiceSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ServiceSpec) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ServiceSpec) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ServiceSpec) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *ServiceSpec) GetType() ServiceType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ServiceSpec) GetTypeOk() (*ServiceType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ServiceSpec) SetType(v ServiceType)`

SetType sets Type field to given value.

### HasType

`func (o *ServiceSpec) HasType() bool`

HasType returns a boolean if a field has been set.

### SetTypeNil

`func (o *ServiceSpec) SetTypeNil(b bool)`

 SetTypeNil sets the value for Type to be an explicit nil

### UnsetType
`func (o *ServiceSpec) UnsetType()`

UnsetType ensures that no value is present for Type, not even an explicit nil
### GetSelector

`func (o *ServiceSpec) GetSelector() map[string]string`

GetSelector returns the Selector field if non-nil, zero value otherwise.

### GetSelectorOk

`func (o *ServiceSpec) GetSelectorOk() (*map[string]string, bool)`

GetSelectorOk returns a tuple with the Selector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelector

`func (o *ServiceSpec) SetSelector(v map[string]string)`

SetSelector sets Selector field to given value.


### SetSelectorNil

`func (o *ServiceSpec) SetSelectorNil(b bool)`

 SetSelectorNil sets the value for Selector to be an explicit nil

### UnsetSelector
`func (o *ServiceSpec) UnsetSelector()`

UnsetSelector ensures that no value is present for Selector, not even an explicit nil
### GetLabels

`func (o *ServiceSpec) GetLabels() map[string]string`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *ServiceSpec) GetLabelsOk() (*map[string]string, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *ServiceSpec) SetLabels(v map[string]string)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *ServiceSpec) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetPorts

`func (o *ServiceSpec) GetPorts() []ServicePort`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *ServiceSpec) GetPortsOk() (*[]ServicePort, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *ServiceSpec) SetPorts(v []ServicePort)`

SetPorts sets Ports field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


