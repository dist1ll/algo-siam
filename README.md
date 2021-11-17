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

* URL of an algod endpoint (e.g. )
* API token for the endpoint (e.g. for sandbox node)
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
buffer, err := siam.CreateAlgorandBufferFromEnv(nil)
```

If you want to supply the configuration arguments manually, you can do so with the following snippet

```go
c := client.CreateAlgorandClientMock(URL, token)
buffer, err := siam.CreateAlgorandBuffer(c, base64key, nil)
```

This will create a new Siam application (or detect an existing one). If the endpoint is unreachable, the token is incorrect, or the account has not enough funds to cover transactions, an error will be returned.

The last step is to create a managing goroutine. This routine takes care of updating your data asynchronously.

```go
wg, cancel := buffer.SpawnManagingRoutine(&siam.ManageConfig{
	SleepTime:           time.Second * 20, // time to sleep after failing to connect to node
	HealthCheckInterval: time.Minute,      // how much time between mandatory node health checks
})

// your code

wg.Wait()
```
Here you can configure how frequently the `AlgorandBuffer` should talk to the node. The method returns a cancel function, which you can use to terminate the goroutine. You can then use `wg.Wait()` to wait for the goroutine to terminate. 

## Writing, Deleting and Inspecting Data

Now that you have a working `AlgorandBuffer`, you can start fetching, storing and deleting data. 
This is done asynchronously.

### Inspecting Data
To fetch the actual data that currently lives on the blockchain, you can use `GetBuffer`
```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
data, err := buffer.GetBuffer(ctx)  //returns map[string]string of key-value store
```

At the moment, `data` will be an empty map. `GetBuffer` returns the actual data stored in the Algorand
application. You can use it to check if your data has been written to the blockchain.

### Writing Data

```go
data := map[string]string{
    "match_256846": "Astralis",
    "match_256847": "Vitality",
    "match_256849": "Gambit",
}

err = buffer.PutElements(data)
```
Now the goroutine will write this data to the Siam app. You should see a result soon. 
If you want to wait for the data to be submitted, use

```go
buffer.WaitForFlush()
```

You can assert that the data was successfully written to by running

```go
correct := buffer.ContainsWithin(data, time.Minute, time.Second)
if correct {
    fmt.Println("data was correctly inserted")
}
```

### Deleting Data

Once you've confirmed that data exists on the Siam application, you can safely call `DeleteElements`

```go
// delete two matches
err = buffer.DeleteElements("match_256846", "match_256847")
buffer.WaitForFlush()
```

## Example
## Existing Oracle Apps

TODO

## License

This project is licensed under the permissive zlib license.

## Relevant Resources

* [What is Algorand?](https://developer.algorand.org/docs/get-started/basics/why_algorand/)
* [Smart Contracts](https://developer.algorand.org/docs/get-details/dapps/smart-contracts/)
* [Parameter Tables](https://developer.algorand.org/docs/get-details/parameter_tables/#stateful-smart-contract-constraints)