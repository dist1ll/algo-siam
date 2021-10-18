package core

import (
	"context"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// AlgorandClientMock implements the AlgorandClient interface. All functions are simply
// returning the corresponding public field.
type AlgorandClientSimpleMock struct {
	Account models.Account
	App     models.Application
}

func (a *AlgorandClientSimpleMock) AccountInformation(string, context.Context) (models.Account, error) {
	return a.Account, nil
}

func (a *AlgorandClientSimpleMock) GetApplicationByID(uint64, context.Context) (models.Application, error) {
	return a.App, nil
}
