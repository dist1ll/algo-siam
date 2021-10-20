package core

import (
	"context"
	"errors"
	"reflect"
	"runtime"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

// AlgorandMock implements the AlgorandClient interface. All functions are simply
// returning the corresponding public field.
type AlgorandMock struct {
	AlwaysReturnError bool // When true, returns errors for every request
	Account           models.Account
	App               models.Application
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

func CreateAlgorandClientSimpleMock(URL string, token string) *AlgorandMock {
	err := make(map[string]bool)
	return &AlgorandMock{ErrorFunctions: err}
}

// SetError controls whether or not the specified functions return an error or not.
// If val is set to true, all provided methods belonging to this struct will return
// errors when called.
func (a *AlgorandMock) SetError(val bool, f...interface{}) {
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
	ret, err := a.wrapExecutionCondition(models.Account{}, models.Account{}, (*AlgorandMock).AccountInformation)
	return ret.(models.Account), err
}

func (a *AlgorandMock) GetApplicationByID(uint64, context.Context) (models.Application, error) {
	ret, err := a.wrapExecutionCondition(models.Application{}, models.Application{}, (*AlgorandMock).GetApplicationByID)
	return ret.(models.Application), err
}
