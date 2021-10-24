package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/m2q/aema/core/client"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
)

const ApprovalProg = "#pragma version 4\n" +
	"addr Sender\n" +
	"addr CreatorAddress\n" +
	"=="

type AlgorandBuffer struct {
	// AppId is the ID of Algorand application this buffer publishes to.
	AppId uint64
	// AccountCrypt is the owner of the buffer's Algorand application.
	AccountCrypt crypto.Account
	// Client is the wrapping interface for communicating with the node
	Client client.AlgorandClient
	// AppChannel returns information every time the application state of
	// this buffer's account has been mutated (i.e. deleted/created app).
	// See routine Manage.
	AppChannel chan string
	// minSleep is the minimum amount of time the Manage routine will sleep
	// after failing to execute a blockchain action
	MinSleep time.Duration
	// timeoutLength is the default duration for Client requests like
	// Health() or Status() to timeout.
	timeoutLength time.Duration
	// currentlyManaged is true when the Manage()
	currentlyManaged bool
	//
	init bool
}

// CreateAlgorandBuffer creates a new instance of AlgorandBuffer. base64key is the
func CreateAlgorandBuffer(c client.AlgorandClient, base64key string) (*AlgorandBuffer, error) {
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
		AccountCrypt:  account,
		AppChannel:    make(chan string),
		MinSleep:      client.AlgorandDefaultMinSleep,
		timeoutLength: client.AlgorandDefaultTimeout,
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

// CreateAlgorandBufferFromEnv creates an AlgorandBuffer from environment
// variables. See README.md for more information.
func CreateAlgorandBufferFromEnv() (*AlgorandBuffer, error) {
	url, token, base64key := GetAlgorandEnvironmentVars()
	a, err := client.CreateAlgorandClientWrapper(url, token)
	if err != nil {
		return nil, err
	}
	return CreateAlgorandBuffer(a, base64key)
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
	info, err := ab.Client.AccountInformation(ab.AccountCrypt.Address.String(), context.Background())
	if err != nil {
		return models.Application{}, err
	}
	if len(info.CreatedApps) == 0 {
		return models.Application{}, &NoApplication{Account: ab.AccountCrypt}
	}
	if len(info.CreatedApps) > 1 {
		return models.Application{}, &TooManyApplications{Account: ab.AccountCrypt, Apps: info.CreatedApps}
	}

	return info.CreatedApps[0], nil
}

// GetBuffer returns the stored global state of this buffers algorand application
func (ab *AlgorandBuffer) GetBuffer() (map[string]string, error) {
	if !ab.currentlyManaged {
		panic("need to run 'go buffer.Manage()' before being able to store")
	}
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

func (ab *AlgorandBuffer) PutElements(elements map[string]string) {
	if !ab.currentlyManaged {
		panic("need to run 'go buffer.Manage()' before being able to store")
	}
}

// Manage is a constantly running routine that manages the lifecycle of
// the AlgorandBuffer. It performs continuous checks against the node,
// smart contract, application state and funding amount. Manage takes care
// of asynchronous buffer writes, by queueing and writing them when the node
// is available.
//
//
func (ab *AlgorandBuffer) Manage() {
	ab.currentlyManaged = true
	for {
		err := ab.checkConnection()
		if err != nil {
			time.Sleep(ab.MinSleep)
			continue
		}

		err = ab.manageApplications()
		if err != nil {
			time.Sleep(ab.MinSleep)
			continue
		}
	}
}

// manageApplications takes care of deleting and creating applications
// to make the target account valid.
func (ab *AlgorandBuffer) manageApplications() error {
	info, err := ab.Client.AccountInformation(ab.AccountCrypt.Address.String(), context.Background())
	if err != nil {
		return err
	}

	// Deletion Routine
	err = ab.manageDeletion(info)
	if err != nil {
		return err
	}

	// Creation Routine
	err = ab.manageCreation(info)
	if err != nil {
		return err
	}

	return nil
}

func (ab *AlgorandBuffer) manageCreation(info models.Account) error {
	if client.ValidAccount(info) {
		return nil
	}
	if len(info.CreatedApps) > 0 {
		return errors.New("must delete invalid applications before creating new one")
	}

	appId, err := ab.Client.CreateApplication(ab.AccountCrypt, "#pragma version 4\nint 1", "#pragma version 4\nint 1")
	if err != nil {
		return err
	}
	ab.AppChannel <- fmt.Sprintf("created app with ID: <%d>", appId)
	ab.AppId = appId
	return nil
}

func (ab *AlgorandBuffer) manageDeletion(info models.Account) error {
	// If no apps exist, no deletion necessary
	if len(info.CreatedApps) == 0 {
		return nil
	}
	// Find out if there exists an app that's already "valid" (i.e. right schema)
	validApp := -1
	earliestValidApp := uint64(math.MaxUint64)
	for i, val := range info.CreatedApps {
		if client.FulfillsSchema(val) && val.CreatedAtRound < earliestValidApp {
			validApp = i
			earliestValidApp = val.CreatedAtRound
		}
	}

	// Delete apps if there's at least one incorrect app
	if !client.ValidAccount(info) {
		for i := len(info.CreatedApps) - 1; i >= 0; i-- {
			if i == validApp {
				continue
			}
			err := ab.Client.DeleteApplication(ab.AccountCrypt, info.CreatedApps[i].Id)
			if err != nil {

				return err
			}
			ab.AppChannel <- "deleted successfully"
		}
	}
	return nil
}

func (ab *AlgorandBuffer) checkConnection() error {
	err := ab.Health()
	if err != nil {
		return err
	}
	return ab.VerifyToken()
}
