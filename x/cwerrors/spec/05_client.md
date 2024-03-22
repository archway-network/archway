# Client

Section describes interaction with the module by the user

## CLI

### Query

The `query` commands alllows a user to query the module state

Use the `-h`/`--help` flag to get a help description of a command.

`archwayd q cwerrors -h`

> You can add the `-o json` for the JSON output format

#### params

Get the current module parameters

Usage:

`archwayd q cwerrors params [flags]`

Example output:

```yaml
error_stored_time: "302400"
subscription_fee:
  amount: "0"
  denom: aarch
subscription_period: "302400"
```

#### errors

List all the errors for the given contract

Usage:

`archwayd q cwerrors errors [contract-address]`

Example:

`archway q cwerrors errors archway1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqukxvuk`

Example output:

```yaml
errors:
- module_name: "callback"
  error_code: 2
  contract_address: archway1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqukxvuk
  input_payload: "{'job_id':1}"
  error_message: "Out of gas"
```

#### is-subscribed

Lists if the given contract is subscribed to error callbacks and the block height it is valid till

Usage:

`archwayd q cwerrors is-subscribed [contract-address]`

Example:

`archway q cwerrors is-subscribed archway1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqukxvuk`

Example output:

```yaml
subscribed: true
subscription_valid_till: 1234
```

### TX

The `tx` commands allows a user to interact with the module.

Use the `-h`/`--help` flag to get a help description of a command.

`archwayd tx cwerrors -h`


#### subscribe-to-error

Create a new subscription which will register a contract for a sudo callback on errors

Usage: 

`archwayd tx cwerrors subscribe-to-error [contract-address] [fee-amount] [flags]`

Example:

`archwayd tx cwerrors subscribe-to-error archway1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqukxvuk 7000aarch --from myAccountKey`