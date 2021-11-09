//go:build unit

package client

import (
	"context"
	"encoding/base64"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/crypto"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertEqualBase64(t *testing.T, actualEncoded string, expectedRaw string) {
	res, _ := base64.StdEncoding.DecodeString(actualEncoded)
	assert.Equal(t, expectedRaw, string(res))
}

// TestAlgorandMock_DummyApps tests if created dummy apps have the correct ID
func TestAlgorandMock_DummyApps(t *testing.T) {
	c := CreateAlgorandClientMock("", "")
	c.CreateDummyApps(2, 5, 8)
	assert.Equal(t, 3, len(c.Account.CreatedApps))
	assert.EqualValues(t, 2, c.Account.CreatedApps[0].Id)
	assert.EqualValues(t, 8, c.Account.CreatedApps[2].Id)

	for _, val := range c.Account.CreatedApps {
		assert.EqualValues(t, globalBytes, val.Params.GlobalStateSchema.NumByteSlice)
		assert.EqualValues(t, globalInts, val.Params.GlobalStateSchema.NumUint)
		assert.EqualValues(t, localBytes, val.Params.LocalStateSchema.NumByteSlice)
		assert.EqualValues(t, localInts, val.Params.LocalStateSchema.NumUint)
	}
}

func TestAlgorandMock_ErrorFunctions(t *testing.T) {
	client := CreateAlgorandClientMock("", "")

	client.SetError(true, (*AlgorandMock).AccountInformation, (*AlgorandMock).GetApplicationByID)

	if _, err := client.AccountInformation("", context.Background()); err == nil {
		t.Errorf("expected error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err == nil {
		t.Errorf("expected error")
	}

	client.SetError(false, (*AlgorandMock).AccountInformation)

	if _, err := client.AccountInformation("", context.Background()); err != nil {
		t.Errorf("expected no error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err == nil {
		t.Errorf("expected error")
	}

	client.ClearFunctionErrors()

	if _, err := client.AccountInformation("", context.Background()); err != nil {
		t.Errorf("expected no error")
	}
	if _, err := client.GetApplicationByID(5, context.Background()); err != nil {
		t.Errorf("expected no error")
	}
}

// Make sure the store function updates, and creates new only when not exceeding
// the limit defined by the application schema
func TestAlgorandMock_StoreGlobalSemantics(t *testing.T) {
	client := CreateAlgorandClientMock("", "")
	appId, err := client.CreateApplication(crypto.GenerateAccount(), "", "")
	assert.Nil(t, err)

	// Schema model defines application storage size
	_, global := GenerateSchemasModel()

	kv := make([]models.TealKeyValue, global.NumByteSlice)
	for i, _ := range kv {
		kv[i].Key = strconv.Itoa(i)
		kv[i].Value.Bytes = "dummy"
	}

	// We store MAX number the buffer can handle
	err = client.StoreGlobals(crypto.Account{}, appId, kv)

	// New values, same keys
	kv = make([]models.TealKeyValue, global.NumByteSlice)
	for i, _ := range kv {
		kv[i].Key = strconv.Itoa(i)
		kv[i].Value.Bytes = "dummy2"
	}
	err = client.StoreGlobals(crypto.Account{}, appId, kv)
	state, _ := client.GetApplicationByID(appId, context.Background())
	assert.Len(t, state.Params.GlobalState, int(global.NumByteSlice))
	for _, x := range state.Params.GlobalState {
		assertEqualBase64(t, x.Value.Bytes, "dummy2")
	}

	// New keys, old values
	for i, _ := range kv {
		kv[i].Key = "new" + strconv.Itoa(i)
		kv[i].Value.Bytes = "dummy"
	}
	err = client.StoreGlobals(crypto.Account{}, appId, kv)
	state, _ = client.GetApplicationByID(appId, context.Background())
	assert.Len(t, state.Params.GlobalState, int(global.NumByteSlice))
	// Values and keys should NOT change, because buffer is already maxed out
	for _, x := range state.Params.GlobalState {
		assertEqualBase64(t, x.Value.Bytes, "dummy2")
	}
}
