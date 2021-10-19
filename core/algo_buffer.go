package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
	core "github.com/m2q/aema/core/client"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
)

// Schema of AlgorandBuffer.
const localInts = 0
const localBytes = 0
const globalInts = 0
const globalBytes = 64

const AlgorandDefaultTimeout time.Duration = time.Second * 5

type AlgorandBuffer struct {
	// AppId is the ID of Algorand application this buffer publishes to.
	AppId         uint64
	// Account is the owner of the buffer's Algorand application.
	Account       crypto.Account
	// Client is the wrapping interface for communicating with the node
	Client        AlgorandClient
	// timeoutLength is the default duration for Client requests like
	// Health() or Status() to timeout.
	timeoutLength time.Duration
}

// NewAlgorandBuffer creates a new instance of AlgorandBuffer. base64key is the
func NewAlgorandBuffer(c AlgorandClient, base64key string) (*AlgorandBuffer, error) {
	// Decode Base64 private key
	pk, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		return nil, err
	}

	account, err := crypto.AccountFromPrivateKey(pk)
	if err != nil {
		return nil, err
	}

	buffer := &AlgorandBuffer{
		Client:        c,
		Account:       account,
		timeoutLength: AlgorandDefaultTimeout,
	}

	err = buffer.Health()
	if err != nil {
		return buffer, fmt.Errorf("error checking health: %w", err)
	}

	err = buffer.VerifyToken()
	if err != nil {
		return buffer, fmt.Errorf("error verifying token: %w", err)
	}

	return buffer, err
}

// NewAlgorandBufferFromEnv creates an AlgorandBuffer from environment
// variables. See README.md for more information.
func NewAlgorandBufferFromEnv() (*AlgorandBuffer, error) {
	url, token, base64key := GetAlgorandEnvironmentVars()
	a, err := core.CreateAlgorandClientWrapper(url, token)
	if err != nil {
		return nil, err
	}
	return NewAlgorandBuffer(a, base64key)
}

// VerifyToken checks whether the URL and provided API token resolve to a correct
// Algorand node instance.
func (ab *AlgorandBuffer) VerifyToken() error {
	// status requires valid API token
	_, err := ab.Client.Status(context.Background())
	return err
}

// Health returns nil if node is online and healthy
func (ab *AlgorandBuffer) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), ab.timeoutLength)
	err := ab.Client.HealthCheck(ctx)
	cancel()
	// Null-response of health check means node is ok!
	if _, ok := err.(*json.InvalidUnmarshalError); ok {
		return nil
	}
	return err
}

// GetApplication returns the application that handles the algo buffer. Returns an error
// if the associated address URL has zero or more than one application.
func (ab *AlgorandBuffer) GetApplication() (models.Application, error) {
	info, err := ab.Client.AccountInformation(ab.Account.Address.String(), context.Background())
	if err != nil {
		return models.Application{}, err
	}
	if len(info.CreatedApps) == 0 {
		return models.Application{}, &NoApplication{Account: ab.Account}
	}
	if len(info.CreatedApps) > 1 {
		return models.Application{}, &TooManyApplications{Account: ab.Account, Apps: info.CreatedApps}
	}

	return info.CreatedApps[0], nil
}

// GetBuffer returns the stored global state of this buffers algorand application
func (ab *AlgorandBuffer) GetBuffer() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ab.timeoutLength)
	app, err := ab.Client.GetApplicationByID(ab.AppId, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, kv := range app.Params.GlobalState {
		m[kv.Key] = kv.Value.Bytes
	}
	return m, nil
}

func (ab *AlgorandBuffer) CreateApplication() error {
	_, err := ab.Client.SuggestedParams(context.Background())
	if err != nil {
		return fmt.Errorf("error getting suggested tx params: %s", err)
	}

	globalSchema := types.StateSchema{NumUint: uint64(globalInts), NumByteSlice: uint64(globalBytes)}
	localSchema := types.StateSchema{NumUint: uint64(localInts), NumByteSlice: uint64(localBytes)}
	params, _ := ab.Client.SuggestedParams(context.Background())
	// comment out the next two (2) lines to use suggested fees
	params.FlatFee = true
	params.Fee = 1000

	//b, err := ioutil.ReadFile("./scripts/approval.teal") // just pass the file name
    //if err != nil {
    //   fmt.Print(err)
    //}
	appr := compileProgram(ab.Client, []byte("#pragma version 4\nint 1"))
    clear := compileProgram(ab.Client, []byte("#pragma version 4\nint 1"))

	txn, _ := future.MakeApplicationCreateTx(false, appr, clear, globalSchema, localSchema,
		nil, nil, nil, nil, params, ab.Account.Address, nil,
		types.Digest{}, [32]byte{}, types.Address{})

	//txn, _ = future.MakeApplicationDeleteTx(5, nil, nil, nil, nil,
	//	params, ab.Account.Address, nil, types.Digest{}, [32]byte{}, types.Address{})

	// Sign the transaction
	txID, signedTxn, _ := crypto.SignTransaction(ab.Account.PrivateKey, txn)
	fmt.Printf("Signed txid: %s\n", txID)

	// Submit the transaction
	sendResponse, _ := ab.Client.SendRawTransaction(signedTxn, context.Background())
	fmt.Printf("Submitted transaction %s\n", sendResponse)

	// Wait for confirmation
	waitForConfirmation(txID, ab.Client, 5)

	// display results
	confirmedTxn, _, _ := ab.Client.PendingTransactionInformation(txID, context.Background())
	appId := confirmedTxn.ApplicationIndex
	fmt.Printf("Created new app-id: %d\n", appId)
	return nil
}

func (ab *AlgorandBuffer) DeleteApplication(appId uint64) error {
	_, err := ab.Client.SuggestedParams(context.Background())
	if err != nil {
		return fmt.Errorf("error getting suggested tx params: %s", err)
	}

	params, _ := ab.Client.SuggestedParams(context.Background())
	params.FlatFee = true
	params.Fee = 1000

	txn, _ := future.MakeApplicationDeleteTx(appId, nil, nil, nil, nil,
		params, ab.Account.Address, nil, types.Digest{}, [32]byte{}, types.Address{})

	// Sign the transaction
	txID, signedTxn, _ := crypto.SignTransaction(ab.Account.PrivateKey, txn)
	fmt.Printf("Signed txid: %s\n", txID)

	// Submit the transaction
	sendResponse, _ := ab.Client.SendRawTransaction(signedTxn, context.Background())
	fmt.Printf("Submitted transaction %s\n", sendResponse)

	// Wait for confirmation
	waitForConfirmation(txID, ab.Client, 5)

	// display results
	_, _, err = ab.Client.PendingTransactionInformation(txID, context.Background())
	return err
}