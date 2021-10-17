package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	"os"
	"strings"
)

// Environment variable names for Algorand
const envURLNode = "AEMA_URL_NODE"
const envAlgodToken = "AEMA_ALGOD_TOKEN"

// envPrivateKey is the Base64 encoded private key of our targeted account or application.
const envPrivateKey = "AEMA_PRIVATE_KEY"

type AlgorandClient interface {
	VerifyToken() error
	Health() error
}

// GetAlgorandEnvironmentVars returns a config tuple needed to interact with the Algorand node.
//  addr: URL of the Algorand node
//  token: API token for the algod endpoint
//  base64key: Base64-encoded private key of the Algorand application
func GetAlgorandEnvironmentVars() (addr string, token string, base64key string) {
	addr =  os.Getenv(envURLNode)
	token = os.Getenv(envAlgodToken)
	base64key = os.Getenv(envPrivateKey)
	return addr, token, base64key
}

func GenerateBase64Keypair() (public string, private string) {
	account := crypto.GenerateAccount()
	return account.Address.String(), base64.StdEncoding.EncodeToString(account.PrivateKey)
}

// Utility function that waits for a given txId to be confirmed by the network
func waitForConfirmation(txID string, client *algod.Client, timeout uint64) (models.PendingTransactionInfoResponse, error) {
    pt := new(models.PendingTransactionInfoResponse)
    if client == nil || txID == "" || timeout < 0 {
        fmt.Printf("Bad arguments for waitForConfirmation")
        var msg = errors.New("Bad arguments for waitForConfirmation")
        return *pt, msg

    }

    status, err := client.Status().Do(context.Background())
    if err != nil {
        fmt.Printf("error getting algod status: %s\n", err)
        var msg = errors.New(strings.Join([]string{"error getting algod status: "}, err.Error()))
        return *pt, msg
    }
    startRound := status.LastRound + 1
    currentRound := startRound

    for currentRound < (startRound + timeout) {
        *pt, _, err = client.PendingTransactionInformation(txID).Do(context.Background())
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
        status, err = client.StatusAfterBlock(currentRound).Do(context.Background())
        currentRound++
    }
    msg := errors.New("Tx not found in round range")
    return *pt, msg
}