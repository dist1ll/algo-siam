package siam

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"

	"github.com/m2q/algo-siam/client"

	"github.com/algorand/go-algorand-sdk/crypto"
)

// AlgorandBuffer implements the Buffer interface. The underlying storage mechanism is
// the Algorand blockchain. To create an AlgorandBuffer you can use the methods
// CreateAlgorandBuffer or CreateAlgorandBufferFromEnv.
//
// In general you'll have to supply a URL that represents the endpoint of an Algorand node,
// an API token for this node, and a base64 encoded private key of an account that will create
// and manage the buffer. These variables are provided either as environment variables, or
// given explicitly by creating an client.AlgorandClient via client.CreateAlgorandClientWrapper.
//
// An example of how to instantiate an AlgorandBuffer:
//   // These are the three important config values
//   url := "191.162.6.16:1337"
//   token := "efc54yxeda5o9apret6nermlar2ehn6tsikrh5oea5atnirs56klorki"
//   privKey := "z2BGxfLJ...1IWsvNKRFw8bLQUnK2nRa+YmLNvQCA=="
//
//   client, err := client.CreateAlgorandClientWrapper(url, token)
//   buffer, err := CreateAlgorandBuffer(client, privKey)
type AlgorandBuffer struct {
	// AppId is the ID of Algorand application this buffer publishes to.
	AppId uint64

	// AccountCrypt is the owner of the buffer's Algorand application.
	AccountCrypt crypto.Account

	// Client is the wrapping interface for communicating with the node
	Client client.AlgorandClient

	// storeArguments is consumed by the Manage goroutine and writes kv pairs
	// regularly to the blockchain app storage
	storeArguments chan models.TealKeyValue

	// DeleteElements is consumed by the Manage goroutine and deletes given
	// keys from the blockchain application storage
	deleteArguments chan string

	// timeoutLength is the default duration for Client requests like
	// Health() or Status() to timeout.
	timeoutLength time.Duration

	// logWrapper is used to enable or disable logs
	logWrapper *LogWrapper
	logger     *log.Logger
}

// CreateAlgorandBufferFromEnv creates an AlgorandBuffer from environment variables.
// The environment variables contain configuration to connect to an Algorand node.
// You can find explanations in the README. Alternatively, check out the implementation
// in client.GetAlgorandEnvironmentVars. You can pass your own logger to be used. If you
// pass nil, no logging will occur.
//
// This method uses the client.CreateAlgorandClientWrapper implementation. If you want to
// use your own implementation of client.AlgorandClient, use CreateAlgorandBuffer instead.
func CreateAlgorandBufferFromEnv(logger *log.Logger) (*AlgorandBuffer, error) {
	if !client.HasEnvironmentVars() {
		return nil, errors.New("configuration variables are not set. See README")
	}
	url, token, base64key := client.GetAlgorandEnvironmentVars()
	a, err := client.CreateAlgorandClientWrapper(url, token)
	if err != nil {
		return nil, err
	}
	return CreateAlgorandBuffer(a, base64key, logger)
}

// CreateAlgorandBuffer creates a new instance of AlgorandBuffer. The buffer requires an
// client.AlgorandClient to perform persistence and setup operations on the Algorand blockchain.
// base64key is the base64-encoded private key of the 'target account'. The target account
// creates and maintains the applications state on the blockchain.
func CreateAlgorandBuffer(c client.AlgorandClient, b64key string, logger *log.Logger) (*AlgorandBuffer, error) {
	if logger == nil {
		logger = log.New(ioutil.Discard, "", 0)
	}
	wrapper := NewLogWrapper(logger.Writer())
	logger.SetOutput(wrapper)

	// Decode Base64 private key
	pk, err := base64.StdEncoding.DecodeString(b64key)
	if err != nil {
		return nil, err
	}

	account, err := crypto.AccountFromPrivateKey(pk)
	if err != nil {
		return nil, err
	}

	buffer := &AlgorandBuffer{
		Client:          c,
		AccountCrypt:    account,
		deleteArguments: make(chan string, 64),
		storeArguments:  make(chan models.TealKeyValue, 64),
		timeoutLength:   client.AlgorandDefaultTimeout,
		logger:          logger,
		logWrapper:      wrapper,
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.AlgorandDefaultTimeout)
	err = buffer.ensureRemoteValid(ctx)
	cancel()
	if err != nil {
		return buffer, err
	}
	logger.Println("API endpoint and Algorand dApp ready.")
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

// GetBuffer returns the stored global state of this buffer's associated Algorand application.
func (ab *AlgorandBuffer) GetBuffer(ctx context.Context) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, ab.timeoutLength)
	app, err := ab.Client.GetApplicationByID(ab.AppId, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, kv := range app.Params.GlobalState {
		decodedKey, _ := base64.StdEncoding.DecodeString(kv.Key)
		decodedVal, _ := base64.StdEncoding.DecodeString(kv.Value.Bytes)
		m[string(decodedKey)] = string(decodedVal)
	}
	return m, nil
}

// PutElements stores given key-value pairs. Existing keys will be overridden,
// non-existing keys will be created.
func (ab *AlgorandBuffer) PutElements(ctx context.Context, data map[string]string) error {
	for k, v := range data {
		if len(k)+len(v) > 128 {
			return errors.New("kv pair cannot exceed 128 bytes")
		}
	}
	// if the number of kv pairs exceed client.MaxKVArgs, we need to split them up
	// into partitions. One txn for each partition
	partitions := PartitionMap(data, client.MaxKVArgs)
	for _, p := range partitions {
		kvArray := make([]models.TealKeyValue, 0, client.MaxKVArgs)
		for k, v := range p {
			tkv := models.TealKeyValue{Key: k, Value: models.TealValue{Bytes: v}}
			kvArray = append(kvArray, tkv)
		}
		err := ab.Client.StoreGlobals(ab.AccountCrypt, ab.AppId, kvArray)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ab *AlgorandBuffer) DeleteElements(ctx context.Context, keys ...string) error {
	for _, k := range keys {
		if len(k) > 128 {
			return errors.New("key can't exceed 128 bytes")
		}
	}
	delArray := make([]string, 0)
	for _, k := range keys {
		if len(delArray) == client.MaxArgs {
			err := ab.Client.DeleteGlobals(ab.AccountCrypt, ab.AppId, delArray...)
			if err != nil {
				return err
			}
			delArray = make([]string, 0)
		}
		delArray = append(delArray, k)
	}
	if len(delArray) > 0 {
		err := ab.Client.DeleteGlobals(ab.AccountCrypt, ab.AppId, delArray...)
		if err != nil {
			return err
		}
	}
	return nil
}

// ContainsWithin returns true if the AlgorandBuffer contains the given data within time.
// The polling interval determines how often the endpoint is pinged for new data.
func (ab *AlgorandBuffer) ContainsWithin(m map[string]string, t time.Duration, pollingInterval time.Duration) bool {
	if len(m) > client.GlobalBytes {
		return false
	}
	now := time.Now()
	for time.Now().Sub(now) < t {
		// only remaining time left
		ctx, cancel := context.WithTimeout(context.Background(), t-time.Now().Sub(now))
		data, err := ab.GetBuffer(ctx)
		cancel()
		if err != nil {
			return false
		}
		if mapContainsMap(data, m) {
			return true
		}
		time.Sleep(pollingInterval)
	}
	return false
}

// Contains returns true if the AlgorandBuffer contains the given data. Returns
// an error if the request to the Algorand node failed.
func (ab *AlgorandBuffer) Contains(ctx context.Context, m map[string]string) (bool, error) {
	if len(m) > client.GlobalBytes {
		return false, nil
	}
	data, err := ab.GetBuffer(ctx)
	if err != nil {
		return false, err
	}
	if mapContainsMap(data, m) {
		return true, nil
	}
	return false, nil
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

	appId, err := ab.Client.CreateApplication(ab.AccountCrypt, client.ApproveTeal, client.ClearTeal)
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
		return fmt.Errorf("failed on node health check. bad url? %s", err)
	}
	err = ab.VerifyToken()
	if err != nil {
		// note: for some reason, even a malformed URL can pass the health call.
		return fmt.Errorf("failed on verifying token. bad token or URL has trailing slash. %s", err)
	}
	return err
}

func (ab *AlgorandBuffer) EnableLogger() {
	ab.logWrapper.Enable()
}

func (ab *AlgorandBuffer) DisableLogger() {
	ab.logWrapper.Disable()
}
