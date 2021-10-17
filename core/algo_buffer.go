package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
)

type AlgorandBuffer struct {
	Addr    string // Addr is the address of the app creator.
	Token   string
	Account crypto.Account
	Client  *algod.Client
}

// NewAlgorandBuffer creates a new instance of AlgorandBuffer.
func NewAlgorandBuffer(addr string, token string, base64key string) (*AlgorandBuffer, error) {
	// Decode Base64 private key
	pk, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		return nil, err
	}

	account, err := crypto.AccountFromPrivateKey(pk)
	if err != nil {
		return nil, err
	}

	algodClient, err := algod.MakeClient(addr, token)
	if err != nil {
		return nil, err
	}

	return &AlgorandBuffer{
		Addr:    addr,
		Token:   token,
		Client:  algodClient,
		Account: account,
	}, err
}

// NewAlgorandBufferFromEnv creates
func NewAlgorandBufferFromEnv() (*AlgorandBuffer, error) {
	addr, token, base64key := GetAlgorandEnvironmentVars()
	return NewAlgorandBuffer(addr, token, base64key)
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
	return ab.Client.HealthCheck().Do(context.Background())
}

// GetApplication returns the application that handles the algo buffer. Returns an error
// if the associated address Addr has zero or more than one application.
func (ab *AlgorandBuffer) GetApplication() (models.Application, error) {
	info, err := ab.Client.AccountInformation(ab.Addr).Do(context.Background())
	if err != nil {
		return models.Application{}, err
	}
	if len(info.CreatedApps) == 0 {
		return models.Application{}, errors.New(fmt.Sprintf("account <%s> has no applications", info.Address))
	}
	if len(info.CreatedApps) > 1 {
		return models.Application{}, errors.New(fmt.Sprintf("account <%s> has more than 1 application", info.Address))
	}

	return info.CreatedApps[0], nil
}

// GetBuffer returns the stored global state of this buffers algorand application
func (ab *AlgorandBuffer) GetBuffer() map[string]string {
	return nil
}