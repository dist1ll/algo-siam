package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
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


func TestChainAppCreationDeletion(t *testing.T) {
	a, err := NewAlgorandBufferFromEnv()
	//  func()*NoApplication{return &NoApplication{}}()

	if e := &(NoApplication{}); errors.As(err, &e) {
		t.Errorf("Zero Apps registered %s", e.Account.Address)
	}

	_, err = a.Client.SuggestedParams(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return
	}
	return
	// declare application state storage (immutable)
	const localInts = 0
	const localBytes = 0
	const globalInts = 0
	const globalBytes = 8192

	// define schema
	//globalSchema := types.StateSchema{NumUint: uint64(globalInts), NumByteSlice: uint64(globalBytes)}
	//localSchema := types.StateSchema{NumUint: uint64(localInts), NumByteSlice: uint64(localBytes)}
	// get transaction suggested parameters
	params, _ := a.Client.SuggestedParams(context.Background())
	// comment out the next two (2) lines to use suggested fees
	params.FlatFee = true
	params.Fee = 1000

	// create unsigned transaction
	//b, err := ioutil.ReadFile("./scripts/approval.teal") // just pass the file name
    //if err != nil {
    //    fmt.Print(err)
    //}
    //appr := compileProgram(a.Client, "#pragma version 4\nint 1")
    //clear := compileProgram(a.Client, "#pragma version 4\nint 1")
	//
	//txn, _ := future.MakeApplicationCreateTx(false, appr, clear, globalSchema, localSchema,
	//	nil, nil, nil, nil, params, a.Account.Address, nil,
	//	types.Digest{}, [32]byte{}, types.Address{})
	//
	//txn, _ = future.MakeApplicationDeleteTx(5, nil, nil, nil, nil,
	//	params, a.Account.Address, nil, types.Digest{}, [32]byte{}, types.Address{})
	//
	//// Sign the transaction
	//txID, signedTxn, _ := crypto.SignTransaction(a.Account.PrivateKey, txn)
	//fmt.Printf("Signed txid: %s\n", txID)
	//
	//// Submit the transaction
	//sendResponse, _ := a.Client.SendRawTransaction(signedTxn).Do(context.Background())
	//fmt.Printf("Submitted transaction %s\n", sendResponse)
	//
	//// Wait for confirmation
	//waitForConfirmation(txID, a.Client, 5)
	//
	//// display results
	//confirmedTxn, _, _ := a.Client.PendingTransactionInformation(txID).Do(context.Background())
	//appId := confirmedTxn.ApplicationIndex
	//fmt.Printf("Created new app-id: %d\n", appId)
}
/*
func TestChainConnection(t *testing.T) {
	return
	algobuf, err := NewAlgorandBufferFromEnv()
	if err != nil {
		t.Errorf("error creating AlgorandBuffer: %s", err)
	}

	_, err = algobuf.Client.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return
	}

	// declare application state storage (immutable)
	const localInts = 0
	const localBytes = 0
	const globalInts = 0
	const globalBytes = 8192

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

*/