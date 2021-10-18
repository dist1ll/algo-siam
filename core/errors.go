package core

import (
	"fmt"
	"github.com/algorand/go-algorand-sdk/crypto"
)

type NoApplication struct {
	Account crypto.Account
}

func (e *NoApplication) Error() string {
	return fmt.Sprintf("no application registered for given account {%s}", e.Account.Address)
}