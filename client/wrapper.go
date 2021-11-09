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
	params, err := a.Client.SuggestedParams().Do(ctx)
	if err == nil {
		params.FlatFee = true
		params.Fee = 1000
	}
	return params, err
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

func (a *AlgorandClientWrapper) ExecuteTransaction(acc crypto.Account, txn types.Transaction, ctx context.Context) (models.PendingTransactionInfoResponse, error) {
	_, signedTxn, err := crypto.SignTransaction(acc.PrivateKey, txn)
	if err != nil {
		return models.PendingTransactionInfoResponse{}, err
	}

	txID, err := a.SendRawTransaction(signedTxn, ctx)
	if err != nil {
		return models.PendingTransactionInfoResponse{}, err
	}

	_, err = WaitForConfirmation(txID, a, 5)
	if err != nil {
		return models.PendingTransactionInfoResponse{}, err
	}

	response, _, err := a.PendingTransactionInformation(txID, ctx)
	return response, err
}

func (a *AlgorandClientWrapper) DeleteApplication(acc crypto.Account, appId uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	params, err := a.SuggestedParams(ctx)
	cancel()
	if err != nil {
		return err
	}
	txn, _ := future.MakeApplicationDeleteTx(appId, nil, nil, nil, nil,
		params, acc.Address, nil, types.Digest{}, [32]byte{}, types.Address{})

	ctx, cancel = context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	_, err = a.ExecuteTransaction(acc, txn, ctx)
	cancel()
	return err
}

func (a *AlgorandClientWrapper) CreateApplication(acc crypto.Account, approve string, clear string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	params, err := a.SuggestedParams(ctx)
	cancel()
	if err != nil {
		return 0, err
	}
	localSchema, globalSchema := GenerateSchemas()
	appr := CompileProgram(a, []byte(approve))
	clr := CompileProgram(a, []byte(clear))

	txn, _ := future.MakeApplicationCreateTx(false, appr, clr, globalSchema, localSchema,
		nil, nil, nil, nil, params, acc.Address, nil,
		types.Digest{}, [32]byte{}, types.Address{})

	ctx, cancel = context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	result, err := a.ExecuteTransaction(acc, txn, ctx)
	cancel()
	if err != nil {
		return 0, err
	}
	return result.ApplicationIndex, nil
}

func (a *AlgorandClientWrapper) DeleteGlobals(acc crypto.Account, appId uint64, args ...string) error {
	// convert args from []string to [][]byte
	convArg := make([][]byte, len(args))
	for i, x := range args {
		convArg[i] = []byte(x)
	}
	return a.postArgumentsToApp(acc, appId, "delete", convArg)
}

func (a *AlgorandClientWrapper) StoreGlobals(acc crypto.Account, appId uint64, tkv []models.TealKeyValue) error {
	// convert TEAL kv pair to [][]byte arguments
	args := make([][]byte, len(tkv)*2)
	for i, kv := range tkv {
		args[i*2] = []byte(kv.Key)
		args[i*2+1] = []byte(kv.Value.Bytes)
	}
	return a.postArgumentsToApp(acc, appId, "put", args)
}

// postArgumentsToApp creates and publishes a No-Op transaction with given arguments
// to the application. A note is also added to the transaction. The note determines
// how the Arguments of the No-Op call get interpreted. You can distill note options
// from the approval.teal contract.
func (a *AlgorandClientWrapper) postArgumentsToApp(acc crypto.Account, appId uint64, note string, args [][]byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	params, err := a.SuggestedParams(ctx)
	cancel()
	if err != nil {
		return fmt.Errorf("error getting suggested tx params: %s", err)
	}
	txn, _ := future.MakeApplicationNoOpTx(appId, args,
		nil, nil, nil, params, acc.Address, []byte(note), types.Digest{}, [32]byte{}, types.Address{})

	ctx, cancel = context.WithTimeout(context.Background(), AlgorandDefaultTimeout)
	_, err = a.ExecuteTransaction(acc, txn, ctx)
	cancel()
	return err
}
