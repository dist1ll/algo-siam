package core

import (
	"context"
	"errors"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// AlgorandClientMock implements the AlgorandClient interface. All functions are simply
// returning the corresponding public field.
type AlgorandClientSimpleMock struct {
	AlwaysReturnError bool // When true, returns errors for every request
	Account           models.Account
	App               models.Application
}

func (a *AlgorandClientSimpleMock) AccountInformation(string, context.Context) (models.Account, error) {
	if a.AlwaysReturnError {
		return models.Account{}, errors.New("generic account error")
	}
	return a.Account, nil
}

func (a *AlgorandClientSimpleMock) GetApplicationByID(uint64, context.Context) (models.Application, error) {
	if a.AlwaysReturnError {
		return models.Application{}, errors.New("generic application error")
	}
	return a.App, nil
}
