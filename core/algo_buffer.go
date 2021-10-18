package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
)

const AlgorandDefaultTimeout time.Duration = time.Second * 5

type AlgorandBuffer struct {
	URL           string // URL is the address of the app creator.
	Token         string
	AppId         uint64
	Account       crypto.Account
	Client        *algod.Client
	timeoutLength time.Duration
}

// NewAlgorandBuffer creates a new instance of AlgorandBuffer. base64key is the
func NewAlgorandBuffer(URL string, token string, base64key string) (*AlgorandBuffer, error) {
	// Decode Base64 private key
	pk, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		return nil, err
	}

	account, err := crypto.AccountFromPrivateKey(pk)
	if err != nil {
		return nil, err
	}

	algodClient, err := algod.MakeClient(URL, token)
	if err != nil {
		return nil, err
	}

	buffer := &AlgorandBuffer{
		URL:           URL,
		Token:         token,
		Client:        algodClient,
		Account:       account,
		timeoutLength: AlgorandDefaultTimeout,
	}

	err = buffer.Health()
	if err != nil {
		return nil, fmt.Errorf("error checking health: %w", err)
	}

	err = buffer.VerifyToken()
	if err != nil {
		return nil, fmt.Errorf("error verifying token: %w", err)
	}

	app, err := buffer.GetApplication()
	if err != nil {
		return nil, fmt.Errorf("error querying app: %w", err)
	}
	buffer.AppId = app.Id

	return buffer, err
}

// NewAlgorandBufferFromEnv creates
func NewAlgorandBufferFromEnv() (*AlgorandBuffer, error) {
	url, token, base64key := GetAlgorandEnvironmentVars()
	return NewAlgorandBuffer(url, token, base64key)
}

// VerifyToken checks whether the URL and provided API token resolve to a correct
// Algorand node instance.
func (ab *AlgorandBuffer) VerifyToken() error {
	// status requires valid API token
	_, err := ab.Client.Status().Do(context.Background())
	return err
}

// Health returns nil if node is online and healthy
func (ab *AlgorandBuffer) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), ab.timeoutLength)
	err := ab.Client.HealthCheck().Do(ctx)
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
	info, err := ab.Client.AccountInformation(ab.Account.Address.String()).Do(context.Background())
	if err != nil {
		return models.Application{}, err
	}
	if len(info.CreatedApps) == 0 {
		return models.Application{}, &NoApplication{Account: ab.Account}
	}
	if len(info.CreatedApps) > 1 {
		return models.Application{}, &TooManyApplications{Account: ab.Account}
	}

	return info.CreatedApps[0], nil
}

func (ab *AlgorandBuffer) CreateApplication() error {
	return nil
}

// GetBuffer returns the stored global state of this buffers algorand application
func (ab *AlgorandBuffer) GetBuffer() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ab.timeoutLength)
	app, err := ab.Client.GetApplicationByID(ab.AppId).Do(ctx)
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
