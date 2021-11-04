package client

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/types"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)


// Schema of AlgorandBuffer.
const localInts = 0
const localBytes = 0
const globalInts = 0
const globalBytes = 64

// Arguments
const MaxKVArgs = 15

const AlgorandDefaultTimeout time.Duration = time.Second * 20
const AlgorandDefaultMinSleep time.Duration = time.Second * 5

// AlgorandClient provides a wrapper interface around the go-algorand-sdk client.
// It also provides several useful abstractions for maintaining consistent
// application state.
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

	// DeleteApplication deletes an application with given ID from a given account.
	// If the account has no apps, or none of its apps have the correct ID, then an
	// error is returned.
	DeleteApplication(crypto.Account, uint64) error

	// CreateApplication creates a new application with given teal code. It will wait
	// for a confirmation from the node, and is blocking. Returns AppId.
	CreateApplication(acc crypto.Account, approval string, clear string) (uint64, error)

	// StoreGlobals stores a given array of TEAL key-value pairs
	StoreGlobals(crypto.Account, uint64, []models.TealKeyValue) error
}

// GeneratePrivateKey64 returns a random, base64-encoded private key.
func GeneratePrivateKey64() string {
	acc := crypto.GenerateAccount()
	return base64.StdEncoding.EncodeToString(acc.PrivateKey)
}

// ValidAccount returns true if the given account is a valid AlgorandBuffer target
// and ready to store data in a single application
func ValidAccount(account models.Account) bool {
	return len(account.CreatedApps) == 1 && FulfillsSchema(account.CreatedApps[0])
}

func GenerateSchemas() (types.StateSchema, types.StateSchema) {
	globalSchema := types.StateSchema{NumUint: uint64(globalInts), NumByteSlice: uint64(globalBytes)}
	localSchema := types.StateSchema{NumUint: uint64(localInts), NumByteSlice: uint64(localBytes)}
	return localSchema, globalSchema
}

func GenerateSchemasModel() (models.ApplicationStateSchema, models.ApplicationStateSchema) {
	g, l := GenerateSchemas()
	globalSchema := models.ApplicationStateSchema{NumUint: g.NumUint, NumByteSlice: g.NumByteSlice}
	localSchema := models.ApplicationStateSchema{NumUint: l.NumUint, NumByteSlice: l.NumByteSlice}
	return localSchema, globalSchema
}

func FulfillsSchema(app models.Application) bool {
	if app.Id == 0 {
		return false
	}
	if app.Params.GlobalStateSchema.NumByteSlice != 64 {
		return false
	}
	if app.Params.GlobalStateSchema.NumUint != 0 {
		return false
	}
	return true
}

func CompileProgram(client AlgorandClient, program []byte) (compiledProgram []byte) {
	compileResponse, err := client.TealCompile(program, context.Background())
	if err != nil {
		fmt.Printf("Issue with compile: %s\n", err)
		return
	}
	compiledProgram, _ = base64.StdEncoding.DecodeString(compileResponse.Result)
	return compiledProgram
}

// Utility function that waits for a given txId to be confirmed by the network
func WaitForConfirmation(txID string, client AlgorandClient, timeout uint64) (models.PendingTransactionInfoResponse, error) {
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
