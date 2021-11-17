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

Now that you have a working `AlgorandBuffer`, you can start storing and deleting data. You can do that entirely
asynchronously.

```go
data := map[string]string{
    "match_256846": "Astralis",
    "match_256847": "Vitality",
    "match_256849": "Gambit",
}

buffer.PutElements(data)
```
Now the goroutine will write this data to the Siam app. You should see a result soon. If you want to know if the data has been written to the chain, you can use
```go
// Polling interval of one second
ok := buffer.ContainsWithin(data, time.Minute, time.Second)
``` 

If the data was successfully inserted, you can now delete some of that data

```go
if ok {
    buffer.DeleteElements("match_256849", "match_256847")
}
```
### Inspecting Data
To fetch the actual data that currently lives on the blockchain, you can use `GetBuffer()`
```go
data, err := buffer.GetBuffer()  //returns map[string]string of key-value store
if err != nil {
    log.Fatal(err)
}
```

So if you want to check if data was deleted you can write
```go
for len(data) != 1 {
    data, _ = buffer.GetBuffer()
    time.Sleep(time.Millisecond * 200)
}
``` 
## Existing Oracle Apps

TODO

## License

This project is licensed under the permissive zlib license.

## Relevant Resources

* [What is Algorand?](https://developer.algorand.org/docs/get-started/basics/why_algorand/)
* [Smart Contracts](https://developer.algorand.org/docs/get-details/dapps/smart-contracts/)
* [Parameter Tables](https://developer.algorand.org/docs/get-details/parameter_tables/#stateful-smart-contract-constraints)