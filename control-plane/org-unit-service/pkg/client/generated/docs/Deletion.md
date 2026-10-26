# Deletion

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | [**DeletionTypes**](DeletionTypes.md) |  | 
**Timestamp** | **time.Time** |  | 
**Force** | **bool** |  | [default to false]

## Methods

### NewDeletion

`func NewDeletion(type_ DeletionTypes, timestamp time.Time, force bool, ) *Deletion`

NewDeletion instantiates a new Deletion object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeletionWithDefaults

`func NewDeletionWithDefaults() *Deletion`

NewDeletionWithDefaults instantiates a new Deletion object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *Deletion) GetType() DeletionTypes`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *Deletion) GetTypeOk() (*DeletionTypes, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *Deletion) SetType(v DeletionTypes)`

SetType sets Type field to given value.


### GetTimestamp

`func (o *Deletion) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *Deletion) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *Deletion) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.


### GetForce

`func (o *Deletion) GetForce() bool`

GetForce returns the Force field if non-nil, zero value otherwise.

### GetForceOk

`func (o *Deletion) GetForceOk() (*bool, bool)`

GetForceOk returns a tuple with the Force field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForce

`func (o *Deletion) SetForce(v bool)`

SetForce sets Force field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


