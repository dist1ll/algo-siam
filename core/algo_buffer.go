package core

import (
	"encoding/base64"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
)

type AlgorandBuffer struct {
	Addr string
	Token string
	Client *algod.Client

}


func NewAlgorandBuffer() *AlgorandBuffer {
	buff := &AlgorandBuffer{}

	return buff
}

func (ab *AlgorandBuffer) Init(addr string, token string, base64key string) error {
	_, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		return err
	}

	return nil
}