# NetworkTopologyLabel

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LabelKey** | **string** | The Kubernetes node label key that identifies a topology level (for example, topology.kubernetes.io/zone or cloud.provider.com/topology-rack). | 
**LabelValue** | **string** |  | 

## Methods

### NewNetworkTopologyLabel

`func NewNetworkTopologyLabel(labelKey string, labelValue string, ) *NetworkTopologyLabel`

NewNetworkTopologyLabel instantiates a new NetworkTopologyLabel object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNetworkTopologyLabelWithDefaults

`func NewNetworkTopologyLabelWithDefaults() *NetworkTopologyLabel`

NewNetworkTopologyLabelWithDefaults instantiates a new NetworkTopologyLabel object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetLabelKey

`func (o *NetworkTopologyLabel) GetLabelKey() string`

GetLabelKey returns the LabelKey field if non-nil, zero value otherwise.

### GetLabelKeyOk

`func (o *NetworkTopologyLabel) GetLabelKeyOk() (*string, bool)`

GetLabelKeyOk returns a tuple with the LabelKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelKey

`func (o *NetworkTopologyLabel) SetLabelKey(v string)`

SetLabelKey sets LabelKey field to given value.


### GetLabelValue

`func (o *NetworkTopologyLabel) GetLabelValue() string`

GetLabelValue returns the LabelValue field if non-nil, zero value otherwise.

### GetLabelValueOk

`func (o *NetworkTopologyLabel) GetLabelValueOk() (*string, bool)`

GetLabelValueOk returns a tuple with the LabelValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabelValue

`func (o *NetworkTopologyLabel) SetLabelValue(v string)`

SetLabelValue sets LabelValue field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


