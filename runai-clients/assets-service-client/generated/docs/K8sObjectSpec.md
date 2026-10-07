# K8sObjectSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ObjectMeta** | Pointer to [**NullableK8sObjectMeta**](K8sObjectMeta.md) |  | [optional] 
**ObjectSpec** | **map[string]interface{}** | The raw Kubernetes spec of the object. | 

## Methods

### NewK8sObjectSpec

`func NewK8sObjectSpec(objectSpec map[string]interface{}, ) *K8sObjectSpec`

NewK8sObjectSpec instantiates a new K8sObjectSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewK8sObjectSpecWithDefaults

`func NewK8sObjectSpecWithDefaults() *K8sObjectSpec`

NewK8sObjectSpecWithDefaults instantiates a new K8sObjectSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetObjectMeta

`func (o *K8sObjectSpec) GetObjectMeta() K8sObjectMeta`

GetObjectMeta returns the ObjectMeta field if non-nil, zero value otherwise.

### GetObjectMetaOk

`func (o *K8sObjectSpec) GetObjectMetaOk() (*K8sObjectMeta, bool)`

GetObjectMetaOk returns a tuple with the ObjectMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectMeta

`func (o *K8sObjectSpec) SetObjectMeta(v K8sObjectMeta)`

SetObjectMeta sets ObjectMeta field to given value.

### HasObjectMeta

`func (o *K8sObjectSpec) HasObjectMeta() bool`

HasObjectMeta returns a boolean if a field has been set.

### SetObjectMetaNil

`func (o *K8sObjectSpec) SetObjectMetaNil(b bool)`

 SetObjectMetaNil sets the value for ObjectMeta to be an explicit nil

### UnsetObjectMeta
`func (o *K8sObjectSpec) UnsetObjectMeta()`

UnsetObjectMeta ensures that no value is present for ObjectMeta, not even an explicit nil
### GetObjectSpec

`func (o *K8sObjectSpec) GetObjectSpec() map[string]interface{}`

GetObjectSpec returns the ObjectSpec field if non-nil, zero value otherwise.

### GetObjectSpecOk

`func (o *K8sObjectSpec) GetObjectSpecOk() (*map[string]interface{}, bool)`

GetObjectSpecOk returns a tuple with the ObjectSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectSpec

`func (o *K8sObjectSpec) SetObjectSpec(v map[string]interface{})`

SetObjectSpec sets ObjectSpec field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


