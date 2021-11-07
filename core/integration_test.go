// +build integration

package core

import (
	"context"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/types"
	"testing"
	"time"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/m2q/aema/core/client"
	"github.com/stretchr/testify/assert"
)

// Note: During integration tests, we need to make sure that there's no
// leak of goroutines. Because of this, every test will check for routine
// exits with WaitGroups returned by the SpawnManagingRoutine method of
// AlgorandBuffer

// Test if app removal works
func TestIntegration_RemoveAccount(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
}

func TestIntegration_ValidAccount(t *testing.T) {
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), client.AlgorandDefaultTimeout)
	info, err := buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), ctx)
	cancel()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(info.CreatedApps))
}

// Remove application, and see if Manage re-creates the application
func TestIntegration_AccountGetsRestored(t *testing.T) {
	buffer := createBufferAndRemoveApps(t)
	wg, cancel := buffer.SpawnManagingRoutine(nil)

	var info models.Account
	for !client.ValidAccount(info) {
		info, _ = buffer.Client.AccountInformation(buffer.AccountCrypt.Address.String(), context.Background())
		time.Sleep(time.Second)
	}

	cancel()
	if waitTimeout(wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

// Check if smart contract only allows deletion from an creator acc.
func TestSmartContract_DeleteWrongAcc(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	// Random Account deleting app should throw error
	randomAcc := crypto.GenerateAccount()
	err = buffer.Client.DeleteApplication(randomAcc, buffer.AppId)
	assert.NotNil(t, err)

	// Creator account deleting app should pass without problems
	err = buffer.Client.DeleteApplication(buffer.AccountCrypt, buffer.AppId)
	assert.Nil(t, err)
}

// Checks if client.GenerateApplicationCallTx creates a noop transaction that gets accepted
// by the smart contract
func TestSmartContract_GenerateTransaction(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	// Get Parameters
	params, err := buffer.Client.SuggestedParams(context.Background())
	assert.Nil(t, err)

	// NoOp Application call from original creator
	txn, err := client.GenerateApplicationCallTx(buffer.AppId, buffer.AccountCrypt, params, types.NoOpOC)
	assert.Nil(t, err)

	// Execute Transaction
	ctx, cancel := context.WithTimeout(context.Background(), client.AlgorandDefaultTimeout)
	_, err = buffer.Client.ExecuteTransaction(buffer.AccountCrypt, txn, ctx)
	cancel()
	assert.Nil(t, err)
}

// Test if the smart contract rejects ClearState, CloseOut, OptIn, and
// Update.
func TestSmartContract_RejectNonSupportedOps(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	// OnCompletion Teal ops that should be rejected
	denyOc := []types.OnCompletion {
		types.ClearStateOC,
		types.OptInOC,
		types.CloseOutOC,
		types.UpdateApplicationOC,
	}

	// Get Parameters
	params, err := buffer.Client.SuggestedParams(context.Background())
	assert.Nil(t, err)

	// Deny every transaction with the
	for _, oc := range denyOc {
		txn, err := client.GenerateApplicationCallTx(buffer.AppId, buffer.AccountCrypt, params, oc)
		assert.Nil(t, err)
		// Execute Transaction
		ctx, cancel := context.WithTimeout(context.Background(), client.AlgorandDefaultTimeout)
		_, err = buffer.Client.ExecuteTransaction(buffer.AccountCrypt, txn, ctx)
		cancel()
		assert.NotNil(t, err)
	}
}

// Push one data point
func TestSmartContract_PushData(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	// Fill with data
	data := map[string]string{
		"1000" : "Astralis",
	}
	wg, cancel := buffer.SpawnManagingRoutine(nil)
	err = putElementsAndWait(buffer, data, time.Second * 30)
	assert.Nil(t, err)

	// Make sure goroutine cancels in time
	cancel()
	if waitTimeout(wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

// Push multiple data points
func TestSmartContract_PushDataMultiple(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	// Fill with data
	data := map[string]string{
		"1000" : "Astralis",
		"1001" : "Vitality",
		"1002" : "Gambit",
	}
	wg, cancel := buffer.SpawnManagingRoutine(nil)
	err = putElementsAndWait(buffer, data, time.Second * 30)
	assert.Nil(t, err)

	// Make sure goroutine cancels in time
	cancel()
	if waitTimeout(wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

// Push data and subsequently delete several entries
func TestSmartContract_DeleteData(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	// Fill
	data := map[string]string {
		"1000" : "Astralis",
		"1001" : "Vitality",
		"1002" : "Gambit",
		"1003" : "OG",
		"1004" : "Na'Vi",
		"1005" : "Furia",
	}
	wg, cancel := buffer.SpawnManagingRoutine(nil)
	err = putElementsAndWait(buffer, data, time.Second * 30)
	assert.Nil(t, err)

	// Delete
	err = buffer.DeleteElements("1001", "1003")
	assert.Nil(t, err)

	assert.Nil(t, bufferLengthWithin(buffer, 4, time.Second * 10))

	// Make sure goroutine cancels in time
	cancel()
	if waitTimeout(wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}

func TestSmartContract_UpdateData(t *testing.T) {
	_ = createBufferAndRemoveApps(t)
	buffer, err := CreateAlgorandBufferFromEnv()
	assert.Nil(t, err)

	// Fill
	data := map[string]string {
		"1000" : "Astralis",
		"1001" : "Vitality",
		"1002" : "Gambit",
	}
	wg, cancel := buffer.SpawnManagingRoutine(nil)
	err = putElementsAndWait(buffer, data, time.Second * 30)
	assert.Nil(t, err)

	// Modify and insert data
	data["1000"] = "G2"
	err = buffer.PutElements(data)
	assert.Nil(t, err)

	// Check if value was updated (until timeout)
	assert.Nil(t, bufferEqualsWithin(buffer, "1000", "G2", time.Second * 2))

	// Make sure goroutine cancels in time
	cancel()
	if waitTimeout(wg, time.Second) {
		t.Fatalf("goroutine didn't finish in time")
	}
}