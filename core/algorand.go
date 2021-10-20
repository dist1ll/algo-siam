package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/types"
	"os"
	"strings"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// Environment variable names for Algorand
const envURLNode = "AEMA_URL_NODE"
const envAlgodToken = "AEMA_ALGOD_TOKEN"

// envPrivateKey is the Base64 encoded private key of our targeted account or application.
const envPrivateKey = "AEMA_PRIVATE_KEY"

// AlgorandClient provides a wrapper interface around the go-algorand-sdk client.
type AlgorandClient interface {
	SuggestedParams(context.Context) (types.SuggestedParams, error)
	HealthCheck(context.Context) error
	Status(context.Context) (models.NodeStatus, error)
	StatusAfterBlock(uint64, context.Context) (models.NodeStatus, error)
	AccountInformation(string, context.Context) (models.Account, error)
	GetApplicationByID(uint64, context.Context) (models.Application, error)
	SendRawTransaction([]byte, context.Context) (string, error)
	PendingTransactionInformation(string, context.Context) (models.PendingTransactionInfoResponse, types.SignedTxn, error)
	TealCompile([]byte, context.Context) (models.CompileResponse, error)
}

// GetAlgorandEnvironmentVars returns a config tuple needed to interact with the Algorand node.
//  addr: URL of the Algorand node
//  token: API token for the algod endpoint
//  base64key: Base64-encoded private key of the Algorand application
func GetAlgorandEnvironmentVars() (URL string, token string, base64key string) {
	URL = os.Getenv(envURLNode)
	token = os.Getenv(envAlgodToken)
	base64key = os.Getenv(envPrivateKey)
	return URL, token, base64key
}

// GeneratePrivateKey64 returns a random, base64-encoded private key.
func GeneratePrivateKey64() string {
	acc := crypto.GenerateAccount()
	return base64.StdEncoding.EncodeToString(acc.PrivateKey)
}

func compileProgram(client AlgorandClient, program []byte) (compiledProgram []byte) {
	compileResponse, err := client.TealCompile(program, context.Background())
	if err != nil {
		fmt.Printf("Issue with compile: %s\n", err)
		return
	}
	compiledProgram, _ = base64.StdEncoding.DecodeString(compileResponse.Result)
	return compiledProgram
}

// Utility function that waits for a given txId to be confirmed by the network
func waitForConfirmation(txID string, client AlgorandClient, timeout uint64) (models.PendingTransactionInfoResponse, error) {
	pt := new(models.PendingTransactionInfoResponse)
	if client == nil || txID == "" || timeout < 0 {
		fmt.Printf("Bad arguments for waitForConfirmation")
		var msg = errors.New("Bad arguments for waitForConfirmation")
		return *pt, msg

	}

	status, err := client.Status(context.Background())
	if err != nil {
		fmt.Printf("error getting algod status: %s\n", err)
		var msg = errors.New(strings.Join([]string{"error getting algod status: "}, err.Error()))
		return *pt, msg
	}
	startRound := status.LastRound + 1
	currentRound := startRound

	for currentRound < (startRound + timeout) {
		*pt, _, err = client.PendingTransactionInformation(txID, context.Background())
		if err != nil {
			fmt.Printf("error getting pending transaction: %s\n", err)
			var msg = errors.New(strings.Join([]string{"error getting pending transaction: "}, err.Error()))
			return *pt, msg
		}
		if pt.ConfirmedRound > 0 {
			fmt.Printf("Transaction "+txID+" confirmed in round %d\n", pt.ConfirmedRound)
			return *pt, nil
		}
		if pt.PoolError != "" {
			fmt.Printf("There was a pool error, then the transaction has been rejected!")
			var msg = errors.New("There was a pool error, then the transaction has been rejected")
			return *pt, msg
		}
		fmt.Printf("waiting for confirmation\n")
		status, err = client.StatusAfterBlock(currentRound, context.Background())
		currentRound++
	}
	msg := errors.New("Tx not found in round range")
	return *pt, msg
}
