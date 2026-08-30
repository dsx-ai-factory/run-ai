# JobsNodeAffinityInteractive

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AffinityType** | Pointer to **string** | The type of affinity of the jobs on the nodes. | [optional] 
**SelectedTypes** | Pointer to [**[]NodeTypeForProject**](NodeTypeForProject.md) |  | [optional] 

## Methods

### NewJobsNodeAffinityInteractive

`func NewJobsNodeAffinityInteractive() *JobsNodeAffinityInteractive`

NewJobsNodeAffinityInteractive instantiates a new JobsNodeAffinityInteractive object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewJobsNodeAffinityInteractiveWithDefaults

`func NewJobsNodeAffinityInteractiveWithDefaults() *JobsNodeAffinityInteractive`

NewJobsNodeAffinityInteractiveWithDefaults instantiates a new JobsNodeAffinityInteractive object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAffinityType

`func (o *JobsNodeAffinityInteractive) GetAffinityType() string`

GetAffinityType returns the AffinityType field if non-nil, zero value otherwise.

### GetAffinityTypeOk

`func (o *JobsNodeAffinityInteractive) GetAffinityTypeOk() (*string, bool)`

GetAffinityTypeOk returns a tuple with the AffinityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffinityType

`func (o *JobsNodeAffinityInteractive) SetAffinityType(v string)`

SetAffinityType sets AffinityType field to given value.

### HasAffinityType

`func (o *JobsNodeAffinityInteractive) HasAffinityType() bool`

HasAffinityType returns a boolean if a field has been set.

### GetSelectedTypes

`func (o *JobsNodeAffinityInteractive) GetSelectedTypes() []NodeTypeForProject`

GetSelectedTypes returns the SelectedTypes field if non-nil, zero value otherwise.

### GetSelectedTypesOk

`func (o *JobsNodeAffinityInteractive) GetSelectedTypesOk() (*[]NodeTypeForProject, bool)`

GetSelectedTypesOk returns a tuple with the SelectedTypes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectedTypes

`func (o *JobsNodeAffinityInteractive) SetSelectedTypes(v []NodeTypeForProject)`

SetSelectedTypes sets SelectedTypes field to given value.

### HasSelectedTypes

`func (o *JobsNodeAffinityInteractive) HasSelectedTypes() bool`

HasSelectedTypes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


