package client

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/algorand/go-algorand-sdk/crypto"
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

// AddDummyApps adds applications with given IDs to the account with the default
// AEMA schema.
func (a *AlgorandMock) AddDummyApps(ids ...uint64) {
	if a.Account.CreatedApps == nil {
		return
	}

	loc, glob := GenerateSchemasModel()
	for _, val := range ids {
		params := models.ApplicationParams{GlobalStateSchema: glob, LocalStateSchema: loc}
		a.Account.CreatedApps = append(a.Account.CreatedApps, models.Application{Id: val, Params: params})
	}
}

// CreateDummyApps sets applications with given IDs to the account with the default
// AEMA schema. Note: existing applications are completely overridden.
func (a *AlgorandMock) CreateDummyApps(ids ...uint64) {
	a.Account.CreatedApps = make([]models.Application, len(ids))
	loc, glob := GenerateSchemasModel()
	for i, val := range ids {
		params := models.ApplicationParams{GlobalStateSchema: glob, LocalStateSchema: loc}
		a.Account.CreatedApps[i] = models.Application{Id: val, Params: params}
	}
}

// CreateDummyAppsWithSchema adds applications with given IDs and a given schema.
func (a *AlgorandMock) CreateDummyAppsWithSchema(s models.ApplicationStateSchema, ids ...uint64) {
	a.CreateDummyApps(ids...)
	for i, _ := range a.Account.CreatedApps {
		a.Account.CreatedApps[i].Params = models.ApplicationParams{GlobalStateSchema: s, LocalStateSchema: s}
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

func (a *AlgorandMock) ExecuteTransaction(crypto.Account, types.Transaction, context.Context) (models.PendingTransactionInfoResponse, error) {
	panic("AlgorandStub doesn't stub this method")
	return models.PendingTransactionInfoResponse{}, nil
}

func (a *AlgorandMock) DeleteApplication(acc crypto.Account, appId uint64) error {
	_, err := a.wrapExecutionCondition(nil, nil, (*AlgorandMock).DeleteApplication)
	if err != nil {
		return err
	}
	l := len(a.Account.CreatedApps)
	if l < 1 {
		return errors.New("no applications")
	}
	for idx, app := range a.Account.CreatedApps {
		if app.Id == appId {
			result := make([]models.Application, 0)
			result = append(result, a.Account.CreatedApps[:idx]...)
			a.Account.CreatedApps = append(result, a.Account.CreatedApps[idx+1:]...)
			return nil
		}
	}
	return errors.New("no app with given id found")
}

func (a *AlgorandMock) CreateApplication(account crypto.Account, approve string, clear string) (uint64, error) {
	l, g := GenerateSchemasModel()
	params := models.ApplicationParams{GlobalStateSchema: g, LocalStateSchema: l}
	app := models.Application{Id: 4512, Params: params}
	ret, err := a.wrapExecutionCondition(app, models.Application{}, (*AlgorandMock).CreateApplication)
	if err != nil {
		return 0, err
	}
	a.App = ret.(models.Application)
	a.Account.CreatedApps = []models.Application{a.App}
	return a.App.Id, nil
}

func (a *AlgorandMock) DeleteGlobals(acc crypto.Account, appId uint64, keys ...string) error {
	if a.App.Id != appId {
		return errors.New("incorrect appId provided")
	}
	state := a.App.Params.GlobalState
	for i, _ := range keys {
		keys[i] = base64.StdEncoding.EncodeToString([]byte(keys[i]))
	}
	for _, k := range keys {
		for j, kv := range state {
			if k == kv.Key {
				fmt.Println("deleting!!")
				result := make([]models.TealKeyValue, 0)
				result = append(result, state[:j]...)
				state = append(result, state[j+1:]...)
				break
			}
		}
	}
	a.App.Params.GlobalState = state
	a.Account.CreatedApps[0] = a.App
	return nil
}

func (a *AlgorandMock) StoreGlobals(acc crypto.Account, appId uint64, kv []models.TealKeyValue) error {
	if a.App.Id != appId {
		return errors.New("incorrect appId provided")
	}
	// Encode with base64 like reference implementation of Algorand sdk
	for i, _ := range kv {
		kv[i].Key = base64.StdEncoding.EncodeToString([]byte(kv[i].Key))
		kv[i].Value.Bytes = base64.StdEncoding.EncodeToString([]byte(kv[i].Value.Bytes))
	}

	// Attempt update
	state := a.App.Params.GlobalState
	for j, arg := range kv {
		noneFound := true
		for i, elem := range state {
			if elem.Key == arg.Key {
				state[i].Value.Bytes = kv[j].Value.Bytes
				noneFound = false
			}
		}
		// if no key exists, create new (as long as space is there)
		if noneFound && len(state) < GlobalBytes {
			state = append(state, arg)
		}
	}
	a.App.Params.GlobalState = state
	a.Account.CreatedApps[0] = a.App
	return nil
}
