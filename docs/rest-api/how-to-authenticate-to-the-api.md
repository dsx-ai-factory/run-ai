# How to Authenticate to the API

NVIDIA Run:ai APIs are accessed using **bearer tokens**. A token can be obtained using an access key, which serves as the authentication method for both users and service accounts.

An access key consists of a client ID and client secret, which are used to generate an access token for API authentication. Once a token is obtained, it can be included in the `Authorization` header of subsequent API requests.

* To create service accounts for your organization, see [Service accounts](../platform-management/authentication/service-accounts.md)
* To create your own access keys, see [User access keys](../settings/user-settings/user-access-keys.md)

## Assign Access Rules

After you have created a new service account, you will need to assign it access rules. To assign access rules to the service account, see [Access rules](../platform-management/authentication/accessrules.md). Make sure you assign the correct rules to your service account. Use the [Roles](../platform-management/authentication/roles.md) to assign the correct access rules.

## Request an API Token

Use the client credentials created to get a temporary token to access NVIDIA Run:ai as follows.

### Example Command to Get an API Token

Replace `<COMPANY-URL>` below with:

* For self-hosted use the NVIDIA Run:ai user interface URL.

<Tabs>

<Tab title="cURL">

**cURL**

```bash
curl -X POST \
  'https://<COMPANY-URL>/api/v2/token' \
  --header 'Accept: */*' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=client_credentials' \
  --data-urlencode 'client_id=<CLIENT ID>' \
  --data-urlencode 'client_secret=<CLIENT SECRET>'
```

</Tab>
<Tab title="Python">

**Python**

```python
import requests
reqUrl = "https://<COMPANY-URL>/api/v2/token"
headersList = {
 "Accept": "*/*",
 "Content-Type": "application/x-www-form-urlencoded"
}
payload = {
  "grant_type": "client_credentials",
  "client_id": "<CLIENT ID>",
  "client_secret": "<CLIENT SECRET>"
}
response = requests.request("POST", reqUrl, data=payload, headers=headersList)
print(response.text)
```

</Tab>
<Tab title="Go">

**Go**

```go
package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
)

func main() {
    apiUrl := "https://<COMPANY-URL>/api/v2/token"
    payload := url.Values{}
    payload.Set("grant_type", "client_credentials")
    payload.Set("client_id", "<CLIENT ID>")
    payload.Set("client_secret", "<CLIENT SECRET>")

    req, err := http.NewRequest("POST", apiUrl, strings.NewReader(payload.Encode()))
    if err != nil {
        panic(err)
    }
    req.Header.Set("Accept", "*/*")
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(body))
}
```

</Tab>

</Tabs>

### Response

The API response will look as follows:

```json
{
  "access_token": "<TOKEN>",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

To call NVIDIA Run:ai REST APIs, the application must pass the retrieved `access_token` as a _Bearer token_ in the Authorization header of your HTTP request.
