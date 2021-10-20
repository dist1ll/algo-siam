package core

import (
	"context"
	"errors"
	"reflect"
	"runtime"

	"github.com/algorand/go-algorand-sdk/types"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// AlgorandMock implements the AlgorandClient interface. All functions are simply
// returning the corresponding public field.
type AlgorandMock struct {
	AlwaysReturnError bool // When true, returns errors for every request
	Account           models.Account
	App               models.Application
	Params            types.SuggestedParams
	NodeStatus        models.NodeStatus
	RawTXNResponse    string
	PendingTXNInfo    models.PendingTransactionInfoResponse
	SignedTXN         types.SignedTxn
	CompileResponse   models.CompileResponse
	ErrorFunctions    map[string]bool
}

// wrapExecutionCondition wraps the execution of an AlgorandMock function and
// returns the expected value i with a nil error by default.
// AlgorandMock allows you to configure, which methods return errors or timeouts,
// configurable by SetError(...) or AlwaysReturnError. wrapExecutionCondition
// implements this behavior.
func (a *AlgorandMock) wrapExecutionCondition(i interface{}, def interface{}, f interface{}) (interface{}, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	err := errors.New("error at function " + funcName)
	if a.AlwaysReturnError {
		return def, err
	}
	if val, ok := a.ErrorFunctions[funcName]; ok {
		if val {
			return def, err
		}
	}

	return i, nil
}

func CreateAlgorandClientMock(URL string, token string) *AlgorandMock {
	err := make(map[string]bool)
	return &AlgorandMock{ErrorFunctions: err}
}

// SetError controls whether or not the specified functions return an error or not.
// If val is set to true, all provided methods belonging to this struct will return
// errors when called.
func (a *AlgorandMock) SetError(val bool, f ...interface{}) {
	for _, i := range f {
		funcName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
		a.ErrorFunctions[funcName] = val
	}
}

// ClearFunctionErrors resets the error function map to its default.
func (a *AlgorandMock) ClearFunctionErrors() {
	a.ErrorFunctions = make(map[string]bool)
}

func (a *AlgorandMock) AccountInformation(string, context.Context) (models.Account, error) {
	ret, err := a.wrapExecutionCondition(a.Account, models.Account{}, (*AlgorandMock).AccountInformation)
	return ret.(models.Account), err
}

func (a *AlgorandMock) GetApplicationByID(uint64, context.Context) (models.Application, error) {
	ret, err := a.wrapExecutionCondition(a.App, models.Application{}, (*AlgorandMock).GetApplicationByID)
	return ret.(models.Application), err
}

func (a *AlgorandMock) SuggestedParams(context.Context) (types.SuggestedParams, error) {
	ret, err := a.wrapExecutionCondition(a.Params, types.SuggestedParams{}, (*AlgorandMock).SuggestedParams)
	return ret.(types.SuggestedParams), err
}

func (a *AlgorandMock) HealthCheck(context.Context) error {
	_, err := a.wrapExecutionCondition(nil, nil, (*AlgorandMock).HealthCheck)
	return err
}

func (a *AlgorandMock) Status(context.Context) (models.NodeStatus, error) {
	ret, err := a.wrapExecutionCondition(a.NodeStatus, models.NodeStatus{}, (*AlgorandMock).Status)
	return ret.(models.NodeStatus), err
}

func (a *AlgorandMock) StatusAfterBlock(uint64, context.Context) (models.NodeStatus, error) {
	ret, err := a.wrapExecutionCondition(a.NodeStatus, models.NodeStatus{}, (*AlgorandMock).StatusAfterBlock)
	return ret.(models.NodeStatus), err
}

func (a *AlgorandMock) SendRawTransaction([]byte, context.Context) (string, error) {
	ret, err := a.wrapExecutionCondition(a.RawTXNResponse, "", (*AlgorandMock).SendRawTransaction)
	return ret.(string), err
}

func (a *AlgorandMock) PendingTransactionInformation(string, context.Context) (models.PendingTransactionInfoResponse, types.SignedTxn, error) {
	type txResponse struct {
		Info models.PendingTransactionInfoResponse
		TXN  types.SignedTxn
	}
	txr := txResponse{Info: a.PendingTXNInfo, TXN: a.SignedTXN}
	ret, err := a.wrapExecutionCondition(txr, txResponse{}, (*AlgorandMock).PendingTransactionInformation)
	txr = ret.(txResponse)
	return txr.Info, txr.TXN, err
}

func (a *AlgorandMock) TealCompile([]byte, context.Context) (models.CompileResponse, error) {
	ret, err := a.wrapExecutionCondition(a.CompileResponse, models.CompileResponse{}, (*AlgorandMock).TealCompile)
	return ret.(models.CompileResponse), err
}
