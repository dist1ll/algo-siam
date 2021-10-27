package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"

	"github.com/m2q/aema/core/client"

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
	// storeArguments is consumed by the Manage goroutine and writes kv pairs
	// regularly to the blockchain
	storeArguments chan models.TealKeyValue
	// timeoutLength is the default duration for Client requests like
	// Health() or Status() to timeout.
	timeoutLength time.Duration
}

type ManageConfig struct {
	// SleepTime is the minimum amount of time the Manage routine will sleep aften
	// failing to execute a blockchain action
	SleepTime time.Duration
	// HealthCheckInterval determines the interval between node and application
	// health checks.
	HealthCheckInterval time.Duration
	// ChannelPollFrequency determines how much time passes between subsequent
	// checks of the store/delete channels.
	ChannelPollFrequency time.Duration
}

func GetDefaultManageConfig() *ManageConfig {
	return &ManageConfig{
		SleepTime:            client.AlgorandDefaultMinSleep,
		HealthCheckInterval:  time.Millisecond,
		ChannelPollFrequency: time.Microsecond,
	}
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
		Client:         c,
		AccountCrypt:   account,
		storeArguments: make(chan models.TealKeyValue, 64),
		timeoutLength:  client.AlgorandDefaultTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.AlgorandDefaultTimeout)
	err = buffer.ensureRemoteValid(ctx)
	cancel()
	if err != nil {
		return buffer, err
	}

	return buffer, err
}

// ensureRemoteValid ensures the node is healthy and the target account is in a valid
// state. To achieve this, it will verify, create or delete applications and store the
// updated results in the AlgorandBuffer. Call this when initializing the AlgorandBuffer
// (see CreateAlgorandBuffer).
//
// If the target account is valid and the node is healthy, this function does nothing.
func (ab *AlgorandBuffer) ensureRemoteValid(ctx context.Context) error {
	// Connectivity check
	err := ab.checkConnection()
	if err != nil {
		return err
	}

	// Deletion Routine
	err = ab.manageDeletion()
	if err != nil {
		return err
	}

	// Creation Routine
	err = ab.manageCreation()
	if err != nil {
		return err
	}

	// Set AppID correctly
	ctx, cancel := context.WithTimeout(context.Background(), ab.timeoutLength)
	info, err := ab.Client.AccountInformation(ab.AccountCrypt.Address.String(), ctx)
	cancel()
	if err != nil {
		return err
	}
	ab.AppId = info.CreatedApps[0].Id
	return nil
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

// PutElements stores given key-value pairs. Existing keys will be overridden,
// non-existing keys will be created.
func (ab *AlgorandBuffer) PutElements(data map[string]string) {
	for k, v := range data {
		ab.storeArguments <- models.TealKeyValue{Key: k, Value: models.TealValue{Bytes: v}}
	}
}

// Manage is a constantly running routine that manages the lifecycle of the
// AlgorandBuffer. It performs continuous checks against the node, smart
// contract, application state and funding amount. Manage takes care of
// asynchronous buffer writes, by queueing and writing them when the node
// is available.
//
// The config parameter describes the behavior of the Manage routine, which
// include sleep time after unsuccessful node calls.
func (ab *AlgorandBuffer) Manage(ctx context.Context, config *ManageConfig) {
	if config == nil {
		config = GetDefaultManageConfig()
	}

	// ALL proposed key-value pairs for application
	kvArray := make([]models.TealKeyValue, 0, 1000)

	for ctx.Err() == nil {
		err := ab.ensureRemoteValid(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(config.SleepTime):
				continue
			}
		}

		// If no arguments, wait for arguments (or until a health check
		// needs to be made)
		if len(kvArray) == 0 {
			select {
			case <- ctx.Done():
				return
			case kv := <- ab.storeArguments:
				kvArray = append(kvArray, kv)
			case <- time.After(config.HealthCheckInterval):
				continue
			}
		}


		// unwrap the rest of the store args
		for len(ab.storeArguments) != 0 && len(kvArray) <= client.MaxKVArgs {
			kvArray = append(kvArray, <-ab.storeArguments)
		}

		// attempt to store data
		err = ab.Client.StoreGlobals(ab.AccountCrypt, kvArray)
		if err != nil {
			select {
			case <- ctx.Done():
				return
			case <- time.After(config.SleepTime):
				continue
			}
		}

		// if successful, delete kvArray
		kvArray = make([]models.TealKeyValue, 0, 1000)
	}
}

// manageCreation creates an Algorand application for the target account.
// For this to work, the account needs to be valid (i.e. have no registered
// app and enough funding).
func (ab *AlgorandBuffer) manageCreation() error {
	info, err := ab.Client.AccountInformation(ab.AccountCrypt.Address.String(), context.Background())
	if err != nil {
		return err
	}

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

	ab.AppId = appId
	return nil
}

// manageDeletion removes applications tied to the target account, if they
// don't fulfil the specs of the Algorand buffer (e.g. wrong schema). If
// the account has several valid applications, then the one with the smallest
// CreatedAtRound-parameter will be kept. All others will be deleted.
func (ab *AlgorandBuffer) manageDeletion() error {
	info, err := ab.Client.AccountInformation(ab.AccountCrypt.Address.String(), context.Background())
	if err != nil {
		return err
	}
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
		}
	}
	return nil
}

// checkConnection is a helper function that checks node connectivity and
// verifies that the API token is correct. Ideally this is done on a regular
// basis and used to monitor the app.
func (ab *AlgorandBuffer) checkConnection() error {
	err := ab.Health()
	if err != nil {
		return err
	}
	return ab.VerifyToken()
}
