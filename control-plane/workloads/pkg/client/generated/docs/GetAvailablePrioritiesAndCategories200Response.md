# GetAvailablePrioritiesAndCategories200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Priorities** | [**[]ClusterPriority**](ClusterPriority.md) |  | 
**Categories** | [**[]WorkloadCategory**](WorkloadCategory.md) |  | 

## Methods

### NewGetAvailablePrioritiesAndCategories200Response

`func NewGetAvailablePrioritiesAndCategories200Response(priorities []ClusterPriority, categories []WorkloadCategory, ) *GetAvailablePrioritiesAndCategories200Response`

NewGetAvailablePrioritiesAndCategories200Response instantiates a new GetAvailablePrioritiesAndCategories200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGetAvailablePrioritiesAndCategories200ResponseWithDefaults

`func NewGetAvailablePrioritiesAndCategories200ResponseWithDefaults() *GetAvailablePrioritiesAndCategories200Response`

NewGetAvailablePrioritiesAndCategories200ResponseWithDefaults instantiates a new GetAvailablePrioritiesAndCategories200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPriorities

`func (o *GetAvailablePrioritiesAndCategories200Response) GetPriorities() []ClusterPriority`

GetPriorities returns the Priorities field if non-nil, zero value otherwise.

### GetPrioritiesOk

`func (o *GetAvailablePrioritiesAndCategories200Response) GetPrioritiesOk() (*[]ClusterPriority, bool)`

GetPrioritiesOk returns a tuple with the Priorities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorities

`func (o *GetAvailablePrioritiesAndCategories200Response) SetPriorities(v []ClusterPriority)`

SetPriorities sets Priorities field to given value.


### GetCategories

`func (o *GetAvailablePrioritiesAndCategories200Response) GetCategories() []WorkloadCategory`

GetCategories returns the Categories field if non-nil, zero value otherwise.

### GetCategoriesOk

`func (o *GetAvailablePrioritiesAndCategories200Response) GetCategoriesOk() (*[]WorkloadCategory, bool)`

GetCategoriesOk returns a tuple with the Categories field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategories

`func (o *GetAvailablePrioritiesAndCategories200Response) SetCategories(v []WorkloadCategory)`

SetCategories sets Categories field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


