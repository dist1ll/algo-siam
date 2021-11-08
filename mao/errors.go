package mao

import (
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"

	"github.com/algorand/go-algorand-sdk/crypto"
)

// NoApplication is returned upon creation of an Algorand buffer for an account
// that owns no application.
type NoApplication struct {
	Account crypto.Account
}

func (e *NoApplication) Error() string {
	return fmt.Sprintf("no application registered for given account {%s}", e.Account.Address)
}

// TooManyApplications is returned upon creation of an Algorand buffer for an
// account that has more than 1 application registered.
type TooManyApplications struct {
	Account crypto.Account
	Apps    []models.Application
}

func (e *TooManyApplications) Error() string {
	return fmt.Sprintf("given account owns more than one application {%s}", e.Account.Address)
}
