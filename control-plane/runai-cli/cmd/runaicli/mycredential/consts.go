/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

const (
	ShortMsg = "manage personal credentials associated with the current user"

	ListShortMsg              = "list personal credentials"
	CreateShortMsg            = "create a personal credential"
	DeleteShortMsg            = "delete a personal credential"
	FailedToListCredentials   = "Failed to list credentials"
	FailedToCreateCredentials = "Failed to create credential"
	FailedToDeleteCredential  = "Failed to delete personal credential"
	TypeFlag                  = "type"
	TypeFlagUsage             = "Filter by credential type (genericSecret, dockerRegistry, ngcApiKey)"
	TypeFlagUsageCreate       = "Credential type: genericSecret, dockerRegistry, or ngcApiKey (required)"
	ItemFlag                  = "item"
	ItemFlagUsage             = "Key=value pairs; required keys depend on --type (see above). For genericSecret only, repeat to add multiple pairs to one credential."
	DescriptionFlag           = "description"
	DescriptionFlagUsage      = "Optional description for the credential (e.g. \"My API config\" or a single word)."
	ListExample               = `# List all personal credentials
runai my-credential list

# List only Docker registry credentials
runai my-credential list --type dockerRegistry`
	CreateUse  = "create SECRET_NAME --type <genericSecret|dockerRegistry|ngcApiKey> [flags]"
	CreateLong = `Create a personal credential. Set --type and pass data with --item using the keys below.

  genericSecret   Keys: key, value (repeat --item for multiple pairs). Requires cluster version 2.22+.
  dockerRegistry  Keys: user, password, url. Requires cluster version 2.22+.
  ngcApiKey       Keys: ngc-api-key. Requires cluster version 2.23+.`
	CreateExample = `  runai my-credential create app-config --type genericSecret --item key=API_URL,value=https://api.example.com --item key=DB_PASS,value=secret
  runai my-credential create dockerhub-creds --type dockerRegistry --item user=myuser,password=secret,url=https://index.docker.io/v1/
  runai my-credential create ngc-prod --type ngcApiKey --item ngc-api-key=YOUR_KEY`
	DeleteUse     = "delete SECRET_NAME --type <genericSecret|dockerRegistry|ngcApiKey>"
	DeleteLong    = "Delete a personal credential by name and type."
	DeleteExample = `  runai my-credential delete app-config --type genericSecret
  runai my-credential delete dockerhub-creds --type dockerRegistry
  runai my-credential delete ngc-prod --type ngcApiKey`
)
