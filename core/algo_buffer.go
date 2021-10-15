package core

import (
	"encoding/base64"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
)

type AlgorandBuffer struct {
	Addr string
	Token string
	Account crypto.Account
	Client *algod.Client

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

	return &AlgorandBuffer{Client: algodClient, Account: account}, err
}

// NewAlgorandBufferFromEnv creates
func NewAlgorandBufferFromEnv() (*AlgorandBuffer, error) {
	addr, token, base64key := GetAlgorandEnvironmentVars()
	return NewAlgorandBuffer(addr, token, base64key)
}
