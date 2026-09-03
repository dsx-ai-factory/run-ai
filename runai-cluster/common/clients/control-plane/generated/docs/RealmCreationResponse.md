# RealmCreationResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RealmName** | **string** |  | 
**Domain** | **string** |  | 
**Clients** | [**[]Client**](Client.md) |  | 

## Methods

### NewRealmCreationResponse

`func NewRealmCreationResponse(realmName string, domain string, clients []Client, ) *RealmCreationResponse`

NewRealmCreationResponse instantiates a new RealmCreationResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRealmCreationResponseWithDefaults

`func NewRealmCreationResponseWithDefaults() *RealmCreationResponse`

NewRealmCreationResponseWithDefaults instantiates a new RealmCreationResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRealmName

`func (o *RealmCreationResponse) GetRealmName() string`

GetRealmName returns the RealmName field if non-nil, zero value otherwise.

### GetRealmNameOk

`func (o *RealmCreationResponse) GetRealmNameOk() (*string, bool)`

GetRealmNameOk returns a tuple with the RealmName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRealmName

`func (o *RealmCreationResponse) SetRealmName(v string)`

SetRealmName sets RealmName field to given value.


### GetDomain

`func (o *RealmCreationResponse) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *RealmCreationResponse) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *RealmCreationResponse) SetDomain(v string)`

SetDomain sets Domain field to given value.


### GetClients

`func (o *RealmCreationResponse) GetClients() []Client`

GetClients returns the Clients field if non-nil, zero value otherwise.

### GetClientsOk

`func (o *RealmCreationResponse) GetClientsOk() (*[]Client, bool)`

GetClientsOk returns a tuple with the Clients field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClients

`func (o *RealmCreationResponse) SetClients(v []Client)`

SetClients sets Clients field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


