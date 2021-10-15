package core

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/transaction"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
)

func TestAlgorandBufferCreation(t *testing.T) {
	//_, private := GenerateBase64Keypair()
	//ab, err := NewAlgorandBuffer("", "", private)

}
func TestChainConnection(t *testing.T) {

	addr, token, _ := GetAlgorandEnvironmentVars()

	account, err := crypto.AccountFromPrivateKey(nil)

	if err != nil {
		t.Fatal("Error reading private key. Empty or not the right format.")
	}

	algodClient, err := algod.MakeClient(addr, token)
	if err != nil {
		t.Fatalf("Issue with creating algod client: %s\n", err)
		return
	}

	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("Error getting suggested tx params: %s\n", err)
		return
	}
	fromAddr := account.Address.String()
	toAddr := "TBCRT2557QUKJTQQHS2AWRXO7BUJQGB7ZAAKCWCM5G4SLFRBW5K5LXOTN4"
	var amount uint64 = 5000000
	var minFee uint64 = 1000
	note := []byte("Hello World")
	genID := txParams.GenesisID
	genHash := txParams.GenesisHash
	firstValidRound := uint64(txParams.FirstRoundValid)
	lastValidRound := uint64(txParams.LastRoundValid)
	txn, err := transaction.MakePaymentTxnWithFlatFee(fromAddr, toAddr, minFee, amount, firstValidRound, lastValidRound, note, "", genID, genHash)
	if err != nil {
		fmt.Printf("Error creating transaction: %s\n", err)
		return
	}

	/// Sign the transaction
	txID, signedTxn, err := crypto.SignTransaction(account.PrivateKey, txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return
	}
	fmt.Printf("Signed txid: %s\n", txID)

	// Submit the transaction
	sendResponse, err := algodClient.SendRawTransaction(signedTxn).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return
	}
	fmt.Printf("Submitted transaction %s\n", sendResponse)

	// Wait for confirmation
	confirmedTxn, err := waitForConfirmation(txID, algodClient, 4)
	if err != nil {
		fmt.Printf("Error waiting for confirmation on txID: %s\n", txID)
		return
	}

	// Display completed transaction
	txnJSON, err := json.MarshalIndent(confirmedTxn.Transaction.Txn, "", "\t")
	if err != nil {
		fmt.Printf("Can not marshall txn data: %s\n", err)
	}
	fmt.Printf("Transaction information: %s\n", txnJSON)
	fmt.Printf("Decoded note: %s\n", string(confirmedTxn.Transaction.Txn.Note))
}
