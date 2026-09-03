# Object

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Object** | **map[string]interface{}** | The raw k8s object after sanitization (metadata.managedFields and last-applied-configuration annotation are stripped). | 
**RetrievedAt** | **time.Time** | RFC3339 timestamp captured immediately after the k8s API server returned the object. | 

## Methods

### NewObject

`func NewObject(object map[string]interface{}, retrievedAt time.Time, ) *Object`

NewObject instantiates a new Object object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewObjectWithDefaults

`func NewObjectWithDefaults() *Object`

NewObjectWithDefaults instantiates a new Object object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetObject

`func (o *Object) GetObject() map[string]interface{}`

GetObject returns the Object field if non-nil, zero value otherwise.

### GetObjectOk

`func (o *Object) GetObjectOk() (*map[string]interface{}, bool)`

GetObjectOk returns a tuple with the Object field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObject

`func (o *Object) SetObject(v map[string]interface{})`

SetObject sets Object field to given value.


### GetRetrievedAt

`func (o *Object) GetRetrievedAt() time.Time`

GetRetrievedAt returns the RetrievedAt field if non-nil, zero value otherwise.

### GetRetrievedAtOk

`func (o *Object) GetRetrievedAtOk() (*time.Time, bool)`

GetRetrievedAtOk returns a tuple with the RetrievedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetrievedAt

`func (o *Object) SetRetrievedAt(v time.Time)`

SetRetrievedAt sets RetrievedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


