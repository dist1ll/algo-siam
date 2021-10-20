package core

import (
	"context"
	"errors"
	"reflect"
	"runtime"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// AlgorandClientMock implements the AlgorandClient interface. All functions are simply
// returning the corresponding public field.
type AlgorandClientSimpleMock struct {
	AlwaysReturnError bool // When true, returns errors for every request
	Account           models.Account
	App               models.Application
	ErrorFunctions    map[string]bool
}

func CreateAlgorandClientSimpleMock(URL string, token string) *AlgorandClientSimpleMock {
	err := make(map[string]bool)
	return &AlgorandClientSimpleMock{ErrorFunctions: err}
}

// SetError controls whether or not the specified functions return an error or not.
// If val is set to true, all provided methods belonging to this struct will return
// errors when called.
func (a *AlgorandClientSimpleMock) SetError(val bool, f...interface{}) {
	for i := range f {
		funcName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
		a.ErrorFunctions[funcName] = val
	}
}

// ClearFunctionErrors resets the error function map to its default.
func (a *AlgorandClientSimpleMock) ClearFunctionErrors() {
	a.ErrorFunctions = make(map[string]bool)
}

func (a *AlgorandClientSimpleMock) WrapExecutionCondition(i interface{}, err error, f interface{}) (interface{}, error) {
	if a.AlwaysReturnError {
		return nil, err
	}
	funcName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	if val, ok := a.ErrorFunctions[funcName]; ok {
		if val {
			return nil, err
		}
	}

	return i, nil
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
