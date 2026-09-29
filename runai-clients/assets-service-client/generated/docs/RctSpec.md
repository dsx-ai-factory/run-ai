# RctSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ObjectMeta** | Pointer to [**NullableK8sObjectMeta**](K8sObjectMeta.md) |  | [optional] 
**ObjectSpec** | Pointer to **map[string]interface{}** | The raw Kubernetes spec of the ResourceClaimTemplate. | [optional] 

## Methods

### NewRctSpec

`func NewRctSpec() *RctSpec`

NewRctSpec instantiates a new RctSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRctSpecWithDefaults

`func NewRctSpecWithDefaults() *RctSpec`

NewRctSpecWithDefaults instantiates a new RctSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetObjectMeta

`func (o *RctSpec) GetObjectMeta() K8sObjectMeta`

GetObjectMeta returns the ObjectMeta field if non-nil, zero value otherwise.

### GetObjectMetaOk

`func (o *RctSpec) GetObjectMetaOk() (*K8sObjectMeta, bool)`

GetObjectMetaOk returns a tuple with the ObjectMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectMeta

`func (o *RctSpec) SetObjectMeta(v K8sObjectMeta)`

SetObjectMeta sets ObjectMeta field to given value.

### HasObjectMeta

`func (o *RctSpec) HasObjectMeta() bool`

HasObjectMeta returns a boolean if a field has been set.

### SetObjectMetaNil

`func (o *RctSpec) SetObjectMetaNil(b bool)`

 SetObjectMetaNil sets the value for ObjectMeta to be an explicit nil

### UnsetObjectMeta
`func (o *RctSpec) UnsetObjectMeta()`

UnsetObjectMeta ensures that no value is present for ObjectMeta, not even an explicit nil
### GetObjectSpec

`func (o *RctSpec) GetObjectSpec() map[string]interface{}`

GetObjectSpec returns the ObjectSpec field if non-nil, zero value otherwise.

### GetObjectSpecOk

`func (o *RctSpec) GetObjectSpecOk() (*map[string]interface{}, bool)`

GetObjectSpecOk returns a tuple with the ObjectSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectSpec

`func (o *RctSpec) SetObjectSpec(v map[string]interface{})`

SetObjectSpec sets ObjectSpec field to given value.

### HasObjectSpec

`func (o *RctSpec) HasObjectSpec() bool`

HasObjectSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


