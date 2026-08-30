# List objects with CLI

NVIDIA Run:ai CLI list commands now offer more flexibility. Use these options to control the data you retrieve:

## Key Flags
* `--next-token:` Use this to pick up where you left off from a previous command. Include the next-token value from that previous response. This replaces the old `--offset` flag.
* `--max-items`: Sets a cap on the total number of records to retrieve. Default is -1, which means no limit (fetch all records).
* `--page-size`: Sets how many records to get in one go (i.e., the page size). The default is 50, and the maximum is 500. This replaces the old `--limit` flag.
* `--no-pagination`: Fetches all records in a single request. Can be used with `--page-size` and `--next-token.`

## Examples
* Get results starting from a specific point: `runai workload list --next-token <token>`
* Limit the total number of results: `runai workload list --max-items 100`
* Get results in groups of 20: `runai workload list --page-size 20`
* Get all results in one request (up to the default page size): `runai workload list --no-pagination`
* Get all results in one request (with a specific page size): `runai workload list --no-pagination --page-size 25`

## Important Notes
* Some old flags are being phased out:
  * `--offset`: Replaced by `--next-token`
  * `--limit:` Replaced by `--page-size`
* Some flags can't be used together :
  * Using both `--offset` and `--next-token`will cause an error.
  * `--page-size` greater than 500 will cause an error.
