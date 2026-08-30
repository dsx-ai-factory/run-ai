# PodTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Containers** | [**[]Container**](Container.md) |  | 
**ResourceRequest** | **map[string]float32** |  | 

## Methods

### NewPodTemplate

`func NewPodTemplate(containers []Container, resourceRequest map[string]float32, ) *PodTemplate`

NewPodTemplate instantiates a new PodTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodTemplateWithDefaults

`func NewPodTemplateWithDefaults() *PodTemplate`

NewPodTemplateWithDefaults instantiates a new PodTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContainers

`func (o *PodTemplate) GetContainers() []Container`

GetContainers returns the Containers field if non-nil, zero value otherwise.

### GetContainersOk

`func (o *PodTemplate) GetContainersOk() (*[]Container, bool)`

GetContainersOk returns a tuple with the Containers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainers

`func (o *PodTemplate) SetContainers(v []Container)`

SetContainers sets Containers field to given value.


### GetResourceRequest

`func (o *PodTemplate) GetResourceRequest() map[string]float32`

GetResourceRequest returns the ResourceRequest field if non-nil, zero value otherwise.

### GetResourceRequestOk

`func (o *PodTemplate) GetResourceRequestOk() (*map[string]float32, bool)`

GetResourceRequestOk returns a tuple with the ResourceRequest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceRequest

`func (o *PodTemplate) SetResourceRequest(v map[string]float32)`

SetResourceRequest sets ResourceRequest field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


