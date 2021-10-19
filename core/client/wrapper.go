package core

import (
	"context"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// AlgorandClientWrapper implements the AlgorandClient interface by wrapping the original
// algod.Client
type AlgorandClientWrapper struct {
	Client *algod.Client
}

func CreateAlgorandClientWrapper(URL string, token string) (*AlgorandClientWrapper, error) {
	c, err := algod.MakeClient(URL, token)
	return &AlgorandClientWrapper{Client: c}, err
}

func (a *AlgorandClientWrapper) HealthCheck(ctx context.Context) error {
	return a.Client.HealthCheck().Do(ctx)
}

func (a *AlgorandClientWrapper) Status(ctx context.Context) (models.NodeStatus, error) {
	return a.Client.Status().Do(ctx)
}

func (a *AlgorandClientWrapper) AccountInformation(s string, ctx context.Context) (models.Account, error) {
	return a.Client.AccountInformation(s).Do(ctx)
}

func (a *AlgorandClientWrapper) GetApplicationByID(id uint64, ctx context.Context) (models.Application, error) {
	return a.Client.GetApplicationByID(id).Do(ctx)
}
