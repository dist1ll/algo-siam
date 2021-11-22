# Siam

[![Go Report Card](https://goreportcard.com/badge/github.com/m2q/algo-siam)](https://goreportcard.com/report/github.com/m2q/algo-siam)
[![License: Zlib](https://img.shields.io/badge/License-Zlib-blue.svg)](https://opensource.org/licenses/Zlib)

Siam provides an easy interface for storing Oracle data inside Algorand applications, and is written in Go. Siam stores
data into the global state of the application, which can then be read by other parties in the Algorand chain. The Siam
application uses [this](./client/approval.teal) TEAL contract.

You can install the necessary dependency with the following command.

```
go get github.com/m2q/algo-siam
```

## Configuration

The library needs three things in order to work:

* URL of an algod endpoint
* API token for the endpoint
* The base64-encoded private key of an account with sufficient funds. Note that any existing applications **will be
  deleted**. It is recommended to create a new account just for this purpose.

These can be supplied as environment variables:

| Environment Variable      | Example value |
| ----------- | ----------- |
| SIAM_URL_NODE      | `https://testnet.algoexplorerapi.io`       |
| SIAM_ALGOD_TOKEN   | `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`        |
| SIAM_PRIVATE_KEY | `z2BGxfLJhB67Rwm/FP9su+M9VnfZvJXGhpwghlujZcWFWZbaa0jgJ4eO1IWsvNKRFw8bLQUnK2nRa+YmLNvQCA==`

Alternatively, you can pass these values as arguments inside the code.

## Getting Started

To write and delete data, you need to create an `siam.AlgorandBuffer`. If you configured Siam via environment variables,
you can create an AlgorandBuffer with one line:

```go
buffer, err := siam.CreateAlgorandBufferFromEnv()
```

If you want to supply the configuration arguments manually, you can do so with the following snippet

```go
c := client.CreateAlgorandClientMock(URL, token)
buffer, err := siam.CreateAlgorandBuffer(c, base64key)
```

This will create a new Siam application (or detect an existing one). If the endpoint is unreachable, the token is incorrect, or the account has not enough funds to cover transactions, an error will be returned.

## Writing, Deleting and Inspecting Data

Now that you have a working `AlgorandBuffer`, you can start fetching, storing and deleting data. All
calls receive a context object, which you can use to set timeouts or cancel requests. 

### Inspecting Data
To fetch the actual data that currently lives on the blockchain, you can use `GetBuffer`
```go
data, err := buffer.GetBuffer(context.Background())  //returns map[string]string of key-value store
```

At the moment, `data` will be an empty map. `GetBuffer` returns the actual data stored in the Algorand
application. You can use it to check if what data has been written to the blockchain. There's also a 
convenience function:

```go
contains, err := buffer.Contains(context.Background(), data)
``` 

### Writing Data

To write data to the global state, simply write:
```go
data := map[string]string{
    "match_256846": "Astralis",
    "match_256847": "Vitality",
    "match_256849": "Gambit",
}

err = buffer.PutElements(context.Background(), data)
if err != nil { 
    // data was not written
}
```
If no error is returned, the data was successfully written to the blockchain. If you want 
to *update* existing data, you can just use the same method. If you want to store raw `[]byte` data
instead of strings, use `PutElementsRaw` and `GetBufferRaw` (which will 
use `map[string][]byte` instead).

### Deleting Data

To delete keys from the global state, call `DeleteElements`

```go
// delete two matches
err = buffer.DeleteElements(context.Background(), "match_256846", "match_256847")
```

If `err == nil`, the data was deleted. Note that this method will *not* return an error if you 
supply keys that don't exist. The transaction will still be published, it just won't change the 
global state.  

## Existing Oracle Apps

An example usage can be found here

* (siam-cs)[https://www.github.com/m2q/siam-cs]

## License

This project is licensed under the permissive zlib license.

## Relevant Resources

* [What is Algorand?](https://developer.algorand.org/docs/get-started/basics/why_algorand/)
* [Smart Contracts](https://developer.algorand.org/docs/get-details/dapps/smart-contracts/)
* [Parameter Tables](https://developer.algorand.org/docs/get-details/parameter_tables/#stateful-smart-contract-constraints)