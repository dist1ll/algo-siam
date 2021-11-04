package client

import (
	"context"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/types"
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

func (a *AlgorandClientWrapper) SuggestedParams(ctx context.Context) (types.SuggestedParams, error) {
	return a.Client.SuggestedParams().Do(ctx)
}

func (a *AlgorandClientWrapper) HealthCheck(ctx context.Context) error {
	return a.Client.HealthCheck().Do(ctx)
}

func (a *AlgorandClientWrapper) Status(ctx context.Context) (models.NodeStatus, error) {
	return a.Client.Status().Do(ctx)
}

func (a *AlgorandClientWrapper) StatusAfterBlock(round uint64, ctx context.Context) (response models.NodeStatus, err error) {
	return a.Client.StatusAfterBlock(round).Do(ctx)
}

func (a *AlgorandClientWrapper) AccountInformation(s string, ctx context.Context) (models.Account, error) {
	return a.Client.AccountInformation(s).Do(ctx)
}

func (a *AlgorandClientWrapper) GetApplicationByID(id uint64, ctx context.Context) (models.Application, error) {
	return a.Client.GetApplicationByID(id).Do(ctx)
}

func (a *AlgorandClientWrapper) SendRawTransaction(txn []byte, ctx context.Context) (string, error) {
	return a.Client.SendRawTransaction(txn).Do(ctx)
}

func (a *AlgorandClientWrapper) PendingTransactionInformation(txid string, ctx context.Context) (models.PendingTransactionInfoResponse, types.SignedTxn, error) {
	return a.Client.PendingTransactionInformation(txid).Do(ctx)
}

func (a *AlgorandClientWrapper) TealCompile(b []byte, ctx context.Context) (response models.CompileResponse, err error) {
	return a.Client.TealCompile(b).Do(ctx)
}

func (a *AlgorandClientWrapper) DeleteApplication(acc crypto.Account, appId uint64) error {
	_, err := a.SuggestedParams(context.Background())
	if err != nil {
		return err
	}

	params, err := a.SuggestedParams(context.Background())
	if err != nil {
		return err
	}
	params.FlatFee = true
	params.Fee = 1000

	txn, _ := future.MakeApplicationDeleteTx(appId, nil, nil, nil, nil,
		params, acc.Address, nil, types.Digest{}, [32]byte{}, types.Address{})

	_, signedTxn, err := crypto.SignTransaction(acc.PrivateKey, txn)
	if err != nil {
		return err
	}

	txID, err := a.SendRawTransaction(signedTxn, context.Background())
	if err != nil {
		return err
	}

	_, err = WaitForConfirmation(txID, a, 5)
	if err != nil {
		return err
	}

	_, _, err = a.PendingTransactionInformation(txID, context.Background())
	return err
}

func (a *AlgorandClientWrapper) CreateApplication(account crypto.Account, approve string, clear string) (uint64, error) {
	_, err := a.SuggestedParams(context.Background())
	if err != nil {
		return 0, fmt.Errorf("error getting suggested tx params: %s", err)
	}

	localSchema, globalSchema := GenerateSchemas()

	ctx, cancel := context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	params, _ := a.SuggestedParams(ctx)
	params.FlatFee = true
	params.Fee = 1000
	cancel()

	appr := CompileProgram(a, []byte(approve))
	clr := CompileProgram(a, []byte(clear))

	txn, _ := future.MakeApplicationCreateTx(false, appr, clr, globalSchema, localSchema,
		nil, nil, nil, nil, params, account.Address, nil,
		types.Digest{}, [32]byte{}, types.Address{})

	// Sign the transaction
	txID, signedTxn, err := crypto.SignTransaction(account.PrivateKey, txn)
	if err != nil {
		return 0, nil
	}

	// Submit the transaction
	_, err = a.SendRawTransaction(signedTxn, context.Background())
	if err != nil {
		return 0, err
	}
	fmt.Println("create done")
	// Wait for confirmation
	_, err = WaitForConfirmation(txID, a, 5)
	if err != nil {
		return 0, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	confirmedTxn, _, err := a.PendingTransactionInformation(txID, ctx)
	cancel()
	if err != nil {
		return 0, err
	}
	return confirmedTxn.ApplicationIndex, nil
}

func (a *AlgorandClientWrapper) StoreGlobals(account crypto.Account, appId uint64, tkv []models.TealKeyValue) error {
	_, err := a.SuggestedParams(context.Background())
	if err != nil {
		return fmt.Errorf("error getting suggested tx params: %s", err)
	}
	fmt.Println("attempting to store globals")
	ctx, cancel := context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	params, _ := a.SuggestedParams(ctx)
	params.FlatFee = true
	params.Fee = 1000
	cancel()

	txn, _ := future.MakeApplicationNoOpTx(appId, nil,
		nil, nil, nil, params, account.Address, nil, types.Digest{}, [32]byte{}, types.Address{})

	// Sign the transaction
	txID, signedTxn, err := crypto.SignTransaction(account.PrivateKey, txn)
	if err != nil {
		return nil
	}

	// Submit the transaction
	_, err = a.SendRawTransaction(signedTxn, context.Background())
	if err != nil {
		return err
	}
	// Wait for confirmation
	_, err = WaitForConfirmation(txID, a, 5)
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	_, _, err = a.PendingTransactionInformation(txID, ctx)
	cancel()
	if err != nil {
		return err
	}
	return nil
}
