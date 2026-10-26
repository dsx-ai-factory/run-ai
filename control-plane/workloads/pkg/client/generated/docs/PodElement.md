# PodElement

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ElementId** | **string** | The unique identifier of the workload structure element this pod is associated with. | 
**HierarchyLevel** | **int32** | The hierarchy level of this pod&#39;s associated workload structure element. Level 1 represents the top of the hierarchy, with levels 2 and 3 representing progressively nested elements. | 

## Methods

### NewPodElement

`func NewPodElement(elementId string, hierarchyLevel int32, ) *PodElement`

NewPodElement instantiates a new PodElement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodElementWithDefaults

`func NewPodElementWithDefaults() *PodElement`

NewPodElementWithDefaults instantiates a new PodElement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetElementId

`func (o *PodElement) GetElementId() string`

GetElementId returns the ElementId field if non-nil, zero value otherwise.

### GetElementIdOk

`func (o *PodElement) GetElementIdOk() (*string, bool)`

GetElementIdOk returns a tuple with the ElementId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetElementId

`func (o *PodElement) SetElementId(v string)`

SetElementId sets ElementId field to given value.


### GetHierarchyLevel

`func (o *PodElement) GetHierarchyLevel() int32`

GetHierarchyLevel returns the HierarchyLevel field if non-nil, zero value otherwise.

### GetHierarchyLevelOk

`func (o *PodElement) GetHierarchyLevelOk() (*int32, bool)`

GetHierarchyLevelOk returns a tuple with the HierarchyLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHierarchyLevel

`func (o *PodElement) SetHierarchyLevel(v int32)`

SetHierarchyLevel sets HierarchyLevel field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


