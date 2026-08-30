# How to Authenticate to the API

NVIDIA Run:ai APIs are accessed using **bearer tokens**. A token can be obtained using an access key, which serves as the authentication method for both users and service accounts.

An access key consists of a client ID and client secret, which are used to generate an access token for API authentication. Once a token is obtained, it can be included in the `Authorization` header of subsequent API requests.

* To create service accounts for your organization, see [Service accounts](https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/infrastructure-setup/authentication/service-accounts)
* To create your own access keys, see [User access keys](https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/settings/user-settings/user-access-keys)

## Assign Access Rules

After you have created a new service account, you will need to assign it access rules. To assign access rules to the service account, see [Access rules](https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/infrastructure-setup/authentication/accessrules). Make sure you assign the correct rules to your service account. Use the [Roles](https://app.gitbook.com/s/MeXTfFihZ2OyRbNxZBY3/infrastructure-setup/authentication/roles) to assign the correct access rules.

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
  'https://<COMPANY-URL>/api/v1/token' \
  --header 'Accept: */*' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "grantType":"client_credentials",
  "clientId":"<CLIENT ID>",
  "clientSecret" : "<CLIENT SECRET>"
}'
```

</Tab>
<Tab title="Python">

**Python**

```python
import requests
import json
reqUrl = "https://<COMPANY-URL>/api/v1/token"
headersList = {
 "Accept": "*/*",
 "Content-Type": "application/json"
}
payload = json.dumps({
  "grantType":"client_credentials",
  "clientId":"<CLIENT ID>",
  "clientSecret" : "<CLIENT SECRET>"
})
response = requests.request("POST", reqUrl, data=payload, headers=headersList)
print(response.text)
```

</Tab>
<Tab title="Go">

**Go**

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
)

func main() {
    url := "https://<COMPANY-URL>/api/v1/token"
    payload := map[string]string{
        "grantType":    "client_credentials",
        "clientId":     "<CLIENT ID>",
        "clientSecret": "<CLIENT SECRET>",
    }
    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        panic(err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        panic(err)
    }
    req.Header.Set("Accept", "*/*")
    req.Header.Set("Content-Type", "application/json")

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
  "accessToken": "<TOKEN>", 
}
```

To call NVIDIA Run:ai REST APIs, the application must pass the retrieved `accessToken` as a _Bearer token_ in the Authorization header of your HTTP request.
