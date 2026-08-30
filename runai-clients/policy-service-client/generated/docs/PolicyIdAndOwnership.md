# PolicyIdAndOwnership

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The unique identifier for the policy. | 
**CreatedBy** | **string** | Identifier of the user who created the policy. | 
**CreatedAt** | **time.Time** | The timestamp for when the policy was created. | 
**UpdatedBy** | **string** | Identifier of the user who last updated the policy. | 
**UpdatedAt** | **time.Time** | The timestamp for when the policy was last updated. | 

## Methods

### NewPolicyIdAndOwnership

`func NewPolicyIdAndOwnership(id string, createdBy string, createdAt time.Time, updatedBy string, updatedAt time.Time, ) *PolicyIdAndOwnership`

NewPolicyIdAndOwnership instantiates a new PolicyIdAndOwnership object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyIdAndOwnershipWithDefaults

`func NewPolicyIdAndOwnershipWithDefaults() *PolicyIdAndOwnership`

NewPolicyIdAndOwnershipWithDefaults instantiates a new PolicyIdAndOwnership object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *PolicyIdAndOwnership) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PolicyIdAndOwnership) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PolicyIdAndOwnership) SetId(v string)`

SetId sets Id field to given value.


### GetCreatedBy

`func (o *PolicyIdAndOwnership) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *PolicyIdAndOwnership) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *PolicyIdAndOwnership) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.


### GetCreatedAt

`func (o *PolicyIdAndOwnership) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PolicyIdAndOwnership) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PolicyIdAndOwnership) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedBy

`func (o *PolicyIdAndOwnership) GetUpdatedBy() string`

GetUpdatedBy returns the UpdatedBy field if non-nil, zero value otherwise.

### GetUpdatedByOk

`func (o *PolicyIdAndOwnership) GetUpdatedByOk() (*string, bool)`

GetUpdatedByOk returns a tuple with the UpdatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedBy

`func (o *PolicyIdAndOwnership) SetUpdatedBy(v string)`

SetUpdatedBy sets UpdatedBy field to given value.


### GetUpdatedAt

`func (o *PolicyIdAndOwnership) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *PolicyIdAndOwnership) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *PolicyIdAndOwnership) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


