package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/algorand/go-algorand-sdk/types"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
)

func compileProgram(client *algod.Client, programSource string) (compiledProgram []byte) {
	compileResponse, err := client.TealCompile([]byte(programSource)).Do(context.Background())
	if err != nil {
		fmt.Printf("Issue with compile: %s\n", err)
		return
	}
	compiledProgram, _ = base64.StdEncoding.DecodeString(compileResponse.Result)
	return compiledProgram
}
func TestChainConnection(t *testing.T) {

	algobuf, err := NewAlgorandBufferFromEnv()
	if err != nil {
		t.Errorf("error creating AlgorandBuffer: %s", err)
	}

	err = algobuf.VerifyToken()
	if err != nil {
		t.Error(err)
	}
	_, err = algobuf.Client.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return
	}

	// declare application state storage (immutable)
	const localInts = 1
	const localBytes = 1
	const globalInts = 1
	const globalBytes = 0

	// define schema
	globalSchema := types.StateSchema{NumUint: uint64(globalInts), NumByteSlice: uint64(globalBytes)}
	localSchema := types.StateSchema{NumUint: uint64(localInts), NumByteSlice: uint64(localBytes)}
	// get transaction suggested parameters
	params, _ := algobuf.Client.SuggestedParams().Do(context.Background())
	// comment out the next two (2) lines to use suggested fees
	params.FlatFee = true
	params.Fee = 1000

	// create unsigned transaction
	//b, err := ioutil.ReadFile("./scripts/approval.teal") // just pass the file name
    //if err != nil {
    //    fmt.Print(err)
    //}
    appr := compileProgram(algobuf.Client, "#pragma version 4\nint 1")
    clear := compileProgram(algobuf.Client, "#pragma version 4\nint 1")

    app, err := algobuf.Client.GetApplicationByID(5).Do(context.Background())
    fmt.Println(app)
	return
	txn, _ := future.MakeApplicationCreateTx(false, appr, clear, globalSchema, localSchema,
		nil, nil, nil, nil, params, algobuf.Account.Address, nil,
		types.Digest{}, [32]byte{}, types.Address{})

	txn, _ = future.MakeApplicationDeleteTx(6, nil, nil, nil, nil,
		params, algobuf.Account.Address, nil, types.Digest{}, [32]byte{}, types.Address{})
	// Sign the transaction
	txID, signedTxn, _ := crypto.SignTransaction(algobuf.Account.PrivateKey, txn)
	fmt.Printf("Signed txid: %s\n", txID)

	// Submit the transaction
	sendResponse, _ := algobuf.Client.SendRawTransaction(signedTxn).Do(context.Background())
	fmt.Printf("Submitted transaction %s\n", sendResponse)

	// Wait for confirmation
	waitForConfirmation(txID, algobuf.Client, 5)

	// display results
	confirmedTxn, _, _ := algobuf.Client.PendingTransactionInformation(txID).Do(context.Background())
	appId := confirmedTxn.ApplicationIndex
	fmt.Printf("Created new app-id: %d\n", appId)
}
