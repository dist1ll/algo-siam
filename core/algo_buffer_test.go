package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/transaction"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
)

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

	_, _, err = crypto.SignTransaction(account.PrivateKey, txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return
	}
}
